// Package bitcask is a high performance embedded key value store that uses an on-disk LSM and WAL data structures
// and in-memory radix tree of key/value pairs as per the bitcask paper and seen in the Riak database.
package bitcask

import (
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"sync"

	"github.com/gofrs/flock"
	iradix "github.com/hashicorp/go-immutable-radix/v2"

	"go.mills.io/bitcask/v2/internal"
	"go.mills.io/bitcask/v2/internal/config"
	"go.mills.io/bitcask/v2/internal/data"
	"go.mills.io/bitcask/v2/internal/index"
	"go.mills.io/bitcask/v2/internal/metadata"
)

const lockfile = "lock"

// Key is a slice of bytes that represents a key in a key/value pair
type Key []byte

// Value is a slice of bytes that represents a value in key/value pair
type Value []byte

// Item is a single key/value pair
type Item struct {
	key   Key
	value Value
}

// Key returns the key
func (i *Item) Key() Key { return i.key }

// Value returns the value
func (i *Item) Value() Value { return i.value }

// KeySize returns the size of the key
func (i *Item) KeySize() int { return len(i.key) }

// ValueSize returns the value of the value
func (i *Item) ValueSize() int { return len(i.value) }

// String implements the fmt.Stringer interface and returns a representation of a key
func (i *Item) String() string { return fmt.Sprintf("key=%q", i.key) }

// KeyFunc is a function that takes a key and performs some operation on it possibly returning an error
type KeyFunc func(Key) error

// Bitcask ...
type Bitcask struct {
	mu        sync.RWMutex
	flock     *flock.Flock
	config    *config.Config
	options   []Option
	path      string
	current   data.Datafile
	datafiles map[int]data.Datafile
	trie      *iradix.Tree[internal.Item]
	indexer   index.Indexer[internal.Item]
	metadata  *metadata.MetaData
	isMerging bool
}

// Close closes the database and removes the lock. It is important to call
// Close() as this is the only way to cleanup the lock held by the open
// database.
func (b *Bitcask) Close() error {
	// Acquire an exclusive write lock as we're closing the database now.
	b.mu.Lock()
	defer func() {
		// First unlock our file lock on the database.
		// Note we're ignoring any errors here as we're closing the database
		// anyway and we're inside a defer, and there isn't much we can do.
		b.flock.Unlock()

		// And finally release our exclusive in-process lock.
		b.mu.Unlock()
	}()

	return b.close()
}

func (b *Bitcask) close() error {
	if !b.current.Readonly() {
		if err := b.saveIndexes(); err != nil {
			return err
		}

		b.metadata.IndexUpToDate = true
		if err := b.saveMetadata(); err != nil {
			return err
		}
	}

	for _, df := range b.datafiles {
		if err := df.Close(); err != nil {
			return err
		}
	}

	return b.current.Close()
}

// Sync flushes all buffers to disk ensuring all data is written
func (b *Bitcask) Sync() error {
	b.mu.RLock()
	if b.current.Readonly() {
		b.mu.RUnlock()
		return nil
	}
	b.mu.RUnlock()

	if err := b.saveMetadata(); err != nil {
		return err
	}

	return b.current.Sync()
}

// Readonly returns true if the database is currently opened in readonly mode, false otherwise
func (b *Bitcask) Readonly() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.current.Readonly()
}

// Get fetches value for a key
func (b *Bitcask) Get(key Key) (Value, error) {
	return b.Transaction().Get(key)
}

// GetReader fetches value for a key and returns an io.ReadSeeker
func (b *Bitcask) GetReader(key Key) (io.ReadSeeker, error) {
	return b.Transaction().GetReader(key)
}

// Has returns true if the key exists in the database, false otherwise.
func (b *Bitcask) Has(key Key) bool {
	return b.Transaction().Has(key)
}

// Put stores the key and value in the database.
func (b *Bitcask) Put(key Key, value Value) error {
	b.mu.RLock()
	if b.current.Readonly() {
		b.mu.RUnlock()
		return ErrDatabaseReadonly
	}
	b.mu.RUnlock()

	tx := b.Transaction()
	defer tx.Discard()

	if err := tx.Put(key, value); err != nil {
		return err
	}

	return tx.Commit()
}

// Delete deletes the named key.
func (b *Bitcask) Delete(key Key) error {
	tx := b.Transaction()
	defer tx.Discard()

	if err := tx.Delete(key); err != nil {
		return err
	}

	return tx.Commit()
}

// Scan performs a prefix scan of keys matching the given prefix and calling
// the function `f` with the keys found. If the function returns an error
// no further keys are processed and the first error is returned.
func (b *Bitcask) Scan(prefix Key, f KeyFunc) (err error) {
	return b.Transaction().Scan(prefix, f)
}

// Range performs a range scan of keys matching a range of keys between the
// start key and end key and calling the function `f` with the keys found.
// If the function returns an error no further keys are processed and the
// first error returned.
func (b *Bitcask) Range(start, end Key, f KeyFunc) (err error) {
	return b.Transaction().Range(start, end, f)
}

// Len returns the total number of keys in the database
func (b *Bitcask) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.trie.Len()
}

// ForEach iterates over all keys in the database calling the function `f` for
// each key. If the function returns an error, no further keys are processed
// and the error is returned.
func (b *Bitcask) ForEach(f KeyFunc) (err error) {
	return b.Transaction().ForEach(f)
}

func (b *Bitcask) read(key []byte) (internal.Entry, error) {
	var df data.Datafile

	b.mu.RLock()
	item, found := b.trie.Root().Get(key)
	b.mu.RUnlock()

	if !found {
		return internal.Entry{}, ErrKeyNotFound
	}

	if item.FileID == b.current.FileID() {
		df = b.current
	} else {
		df = b.datafiles[item.FileID]
	}

	e, err := df.ReadAt(item.Offset, item.Size)
	if err != nil {
		return internal.Entry{}, err
	}

	checksum := crc32.ChecksumIEEE(e.Value)
	if checksum != e.Checksum {
		return internal.Entry{}, ErrChecksumFailed
	}

	return e, nil
}

func (b *Bitcask) maybeRotate() error {
	size := b.current.Size()
	if size < int64(b.config.MaxDatafileSize) {
		return nil
	}

	err := b.current.CloseWriter()
	if err != nil {
		return err
	}

	id := b.current.FileID()
	b.datafiles[id] = b.current
	current, err := data.NewOnDiskDatafile(
		b.path, id+1, false,
		b.config.MaxKeySize,
		b.config.MaxValueSize,
		b.config.FileMode,
	)
	if err != nil {
		return err
	}
	b.current = current

	return nil
}

// closeCurrentFile closes current datafile and makes it read only.
func (b *Bitcask) closeCurrentFile() error {
	if err := b.current.Close(); err != nil {
		return err
	}

	id := b.current.FileID()
	df, err := data.NewOnDiskDatafile(
		b.path, id, true,
		b.config.MaxKeySize,
		b.config.MaxValueSize,
		b.config.FileMode,
	)
	if err != nil {
		return err
	}

	b.datafiles[id] = df
	return nil
}

// openNewWriteableFile opens new datafile for writing data
func (b *Bitcask) openNewWriteableFile() error {
	id := b.current.FileID() + 1
	current, err := data.NewOnDiskDatafile(
		b.path, id, false,
		b.config.MaxKeySize,
		b.config.MaxValueSize,
		b.config.FileMode,
	)
	if err != nil {
		return err
	}
	b.current = current
	return nil
}

// reopen reloads a bitcask object with index and datafiles
// caller of this method should take care of locking
func (b *Bitcask) reopen(readonly bool) error {
	datafiles, lastID, err := loadDatafiles(
		b.path,
		b.config.MaxKeySize,
		b.config.MaxValueSize,
		b.config.FileMode,
	)
	if err != nil {
		return err
	}
	t, err := loadIndexes(b, datafiles, lastID)
	if err != nil {
		return err
	}

	current, err := data.NewOnDiskDatafile(
		b.path, lastID, readonly,
		b.config.MaxKeySize,
		b.config.MaxValueSize,
		b.config.FileMode,
	)
	if err != nil {
		return err
	}

	b.trie = t
	b.current = current
	b.datafiles = datafiles

	return nil
}

// Merge merges all datafiles in the database. Old keys are squashed
// and deleted keys removes. Duplicate key/value pairs are also removed.
// Call this function periodically to reclaim disk space.
func (b *Bitcask) Merge() error {
	b.mu.Lock()

	if b.current.Readonly() {
		b.mu.Unlock()
		return ErrDatabaseReadonly
	}

	if b.isMerging {
		b.mu.Unlock()
		return ErrMergeInProgress
	}

	b.isMerging = true
	b.mu.Unlock()
	defer func() {
		b.isMerging = false
	}()
	b.mu.Lock()
	err := b.closeCurrentFile()
	if err != nil {
		b.mu.RUnlock()
		return err
	}
	filesToMerge := make([]int, 0, len(b.datafiles))
	for k := range b.datafiles {
		filesToMerge = append(filesToMerge, k)
	}
	err = b.openNewWriteableFile()
	if err != nil {
		b.mu.RUnlock()
		return err
	}
	b.mu.Unlock()
	sort.Ints(filesToMerge)

	// Temporary merged database path
	temp, err := os.MkdirTemp(b.path, "merge")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temp)

	// Create a merged database
	mdb, err := Open(temp, withConfig(b.config))
	if err != nil {
		return err
	}

	// Rewrite all key/value pairs into merged database
	// Doing this automatically strips deleted keys and
	// old key/value pairs
	b.trie.Root().Walk(func(key []byte, item internal.Item) bool {
		// if key was updated after start of merge operation, nothing to do
		if item.FileID > filesToMerge[len(filesToMerge)-1] {
			return false
		}
		e, err := b.read(key)
		if err != nil {
			return true
		}

		if err := mdb.Put(key, e.Value); err != nil {
			return true
		}

		return false
	})

	if err := mdb.Close(); err != nil {
		return err
	}

	// no reads and writes till we reopen
	b.mu.Lock()
	defer b.mu.Unlock()
	if err = b.close(); err != nil {
		return err
	}

	// Remove data files
	files, err := os.ReadDir(b.path)
	if err != nil {
		return err
	}
	for _, file := range files {
		if file.IsDir() || file.Name() == lockfile {
			continue
		}
		ids, err := internal.ParseIds([]string{file.Name()})
		if err != nil {
			return err
		}
		// if datafile was created after start of merge, skip
		if len(ids) > 0 && ids[0] > filesToMerge[len(filesToMerge)-1] {
			continue
		}
		err = os.RemoveAll(path.Join(b.path, file.Name()))
		if err != nil {
			return err
		}
	}

	// Rename all merged data files
	files, err = os.ReadDir(mdb.Path())
	if err != nil {
		return err
	}
	for _, file := range files {
		// see #225
		if file.Name() == lockfile {
			continue
		}
		err := os.Rename(
			path.Join([]string{mdb.Path(), file.Name()}...),
			path.Join([]string{b.path, file.Name()}...),
		)
		if err != nil {
			return err
		}
	}
	b.metadata.ReclaimableSpace = 0

	// And finally reopen the database
	return b.reopen(false)
}

// Open opens the database at the given path with optional options.
// Options can be provided with the `WithXXX` functions that provide
// configuration options as functions.
func Open(path string, options ...Option) (*Bitcask, error) {
	var (
		cfg  *config.Config
		err  error
		meta *metadata.MetaData
	)

	configPath := filepath.Join(path, "config.json")
	if internal.Exists(configPath) {
		cfg, err = config.Load(configPath)
		if err != nil {
			return nil, &ErrBadConfig{err}
		}
	} else {
		cfg = newDefaultConfig()
	}

	for _, opt := range options {
		if err := opt(cfg); err != nil {
			return nil, err
		}
	}

	if err := os.MkdirAll(path, cfg.DirMode); err != nil {
		return nil, err
	}

	meta, err = loadMetadata(path)
	if err != nil {
		return nil, &ErrBadMetadata{err}
	}

	db := &Bitcask{
		flock:    flock.New(filepath.Join(path, lockfile)),
		config:   cfg,
		options:  options,
		path:     path,
		trie:     iradix.New[internal.Item](),
		indexer:  index.NewIndexer(),
		metadata: meta,
	}

	ok, err := db.flock.TryLock()
	if err != nil {
		return nil, err
	}

	if !ok {
		if !cfg.OpenReadonly {
			return nil, ErrDatabaseLocked
		}

		if err := db.reopen(true); err != nil {
			return nil, err
		}
		return db, nil
	}

	if err := cfg.Save(configPath); err != nil {
		return nil, err
	}

	if cfg.AutoRecovery {
		if err := data.CheckAndRecover(path, cfg); err != nil {
			return nil, fmt.Errorf("recovering database: %s", err)
		}
	}
	if err := db.reopen(false); err != nil {
		return nil, err
	}

	return db, nil
}

// Path returns the database path
func (b *Bitcask) Path() string { return b.path }

// Backup copies db directory to given path
// it creates path if it does not exist
func (b *Bitcask) Backup(path string) error {
	if !internal.Exists(path) {
		if err := os.MkdirAll(path, b.config.DirMode); err != nil {
			return err
		}
	}
	return internal.Copy(b.path, path, []string{lockfile})
}

// saveIndex saves index currently in memory to disk
func (b *Bitcask) saveIndexes() error {
	return b.indexer.Save(b.trie, filepath.Join(b.path, "index"))
}

// saveMetadata saves metadata into disk
func (b *Bitcask) saveMetadata() error {
	return b.metadata.Save(filepath.Join(b.path, "meta.json"), b.config.FileMode)
}

func loadDatafiles(path string, maxKeySize uint32, maxValueSize uint64, fileModeBeforeUmask os.FileMode) (datafiles map[int]data.Datafile, lastID int, err error) {
	fns, err := internal.GetDatafiles(path)
	if err != nil {
		return nil, 0, err
	}

	ids, err := internal.ParseIds(fns)
	if err != nil {
		return nil, 0, err
	}

	datafiles = make(map[int]data.Datafile, len(ids))
	for _, id := range ids {
		datafiles[id], err = data.NewOnDiskDatafile(
			path, id, true,
			maxKeySize,
			maxValueSize,
			fileModeBeforeUmask,
		)
		if err != nil {
			return
		}

	}
	if len(ids) > 0 {
		lastID = ids[len(ids)-1]
	}
	return
}

func getSortedDatafiles(datafiles map[int]data.Datafile) []data.Datafile {
	out := make([]data.Datafile, len(datafiles))
	idx := 0
	for _, df := range datafiles {
		out[idx] = df
		idx++
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].FileID() < out[j].FileID()
	})
	return out
}

// loadIndexes loads index from disk to memory. If index is not available or partially available (last bitcask process crashed)
// then it iterates over last datafile and construct index
func loadIndexes(b *Bitcask, dataFiles map[int]data.Datafile, _ int) (*iradix.Tree[internal.Item], error) {
	t, err := b.indexer.Load(filepath.Join(b.path, "index"), b.config.MaxKeySize)
	if err != nil {
		return loadIndexFromDatafiles(dataFiles)
	}
	if !b.metadata.IndexUpToDate {
		return loadIndexFromDatafiles(dataFiles)
	}
	return t, err
}

func loadIndexFromDatafiles(dataFiles map[int]data.Datafile) (t *iradix.Tree[internal.Item], err error) {
	t = iradix.New[internal.Item]()

	sortedDatafiles := getSortedDatafiles(dataFiles)
	for _, df := range sortedDatafiles {
		t, err = loadIndexFromDatafile(t, df)
		if err != nil {
			return t, err
		}
	}

	return
}

func loadIndexFromDatafile(t *iradix.Tree[internal.Item], df data.Datafile) (*iradix.Tree[internal.Item], error) {
	var offset int64
	for {
		e, n, err := df.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return t, err
		}

		// Tombstone value  (deleted key)
		if len(e.Value) == 0 {
			t, _, _ = t.Delete(e.Key)
			offset += n
			continue
		}
		item := internal.Item{FileID: df.FileID(), Offset: offset, Size: n}
		t, _, _ = t.Insert(e.Key, item)
		offset += n
	}
	return t, nil
}

func loadMetadata(path string) (*metadata.MetaData, error) {
	if !internal.Exists(filepath.Join(path, "meta.json")) {
		meta := new(metadata.MetaData)
		return meta, nil
	}
	return metadata.Load(filepath.Join(path, "meta.json"))
}
