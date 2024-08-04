package bitcask

import (
	"bytes"
	"hash/crc32"
	"io"

	"github.com/abcum/lcp"
	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"go.mills.io/bitcask/v2/internal"
	"go.mills.io/bitcask/v2/internal/data"
)

// Txn is an transaction that represents a snapshot view of the current key space
// of the database.Transactions are isolated from each other, and key/value pairs
// written in a transaction are batched together. Transactions are not thread safe
// and should only be used in a single goroutine. Transactions writing the same
// key result in the last transaction wins strategy, and there are not versioning
// of keys or collision detection.
type Txn struct {
	db        *Bitcask
	current   data.Datafile
	previous  data.Datafile
	datafiles map[int]data.Datafile
	batch     *Batch
	trie      *iradix.Txn[internal.Item]
}

// Discard discards the transaction
func (t *Txn) Discard() {}

// Commit commits the transaction and writes the current batch to the database
func (t *Txn) Commit() error {
	err := t.db.WriteBatch(t.batch)
	if err != nil {
		return err
	}
	t.trie.Commit()
	return nil
}

// Has returns true if the key exists in the database, false otherwise.
func (t *Txn) Has(key Key) bool {
	_, found := t.trie.Root().Get(key)
	return found
}

// Get fetches value for a key
func (t *Txn) Get(key Key) (Value, error) {
	e, err := t.get(key)
	if err != nil {
		return nil, err
	}
	return e.Value, nil
}

// GetReader fetches value for a key and returns an io.ReadSeeker
func (t *Txn) GetReader(key Key) (io.ReadSeeker, error) {
	return t.getReader(key)
}

func (t *Txn) get(key []byte) (internal.Entry, error) {
	var df data.Datafile

	item, found := t.trie.Root().Get(key)

	if !found {
		return internal.Entry{}, ErrKeyNotFound
	}

	switch item.FileID {
	case t.current.FileID():
		df = t.current
	case t.previous.FileID():
		df = t.previous
	default:
		df = t.datafiles[item.FileID]
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

func (t *Txn) getReader(key []byte) (io.ReadSeeker, error) {
	var df data.Datafile

	item, found := t.trie.Root().Get(key)

	if !found {
		return nil, ErrKeyNotFound
	}

	switch item.FileID {
	case t.current.FileID():
		df = t.current
	case t.previous.FileID():
		df = t.previous
	default:
		df = t.datafiles[item.FileID]
	}

	return io.NewSectionReader(df.Reader(), item.Offset, item.Size), nil
}

// Delete deletes the named key.
func (t *Txn) Delete(key Key) error {
	entry, err := t.batch.Delete(key)
	if err != nil {
		return err
	}

	_, _, err = t.current.Write(entry)
	if err != nil {
		return err
	}

	_, _ = t.trie.Delete(key)

	return nil
}

// Put stores the key and value in the database.
func (t *Txn) Put(key Key, value Value) error {
	entry, err := t.batch.Put(key, value)
	if err != nil {
		return err
	}

	offset, n, err := t.current.Write(entry)
	if err != nil {
		return err
	}

	item := internal.Item{FileID: t.current.FileID(), Offset: offset, Size: n}

	_, _ = t.trie.Insert(key, item)

	return nil
}

// ForEach iterates over all keys in the database calling the function `f`
// for each key. If the function returns an error, no further keys are processed
// and the error is returned.
func (t *Txn) ForEach(f KeyFunc) (err error) {
	t.trie.Root().Walk(func(key []byte, item internal.Item) bool {
		if err = f(key); err != nil {
			return true
		}
		return false
	})

	return
}

// Iterator returns an iterator for iterating through keys in key order
func (t *Txn) Iterator(opts ...IteratorOption) *Iterator {
	it := &Iterator{db: t.db, opts: &iteratorOptions{}}
	for _, opt := range opts {
		opt(it)
	}
	if it.opts.reverse {
		it.itr = t.trie.Root().ReverseIterator()
	} else {
		it.itf = t.trie.Root().Iterator()
	}
	return it
}

// Range performs a range scan of keys matching a range of keys between the
// start key and end key and calling the function `f` with the keys found.
// If the function returns an error no further keys are processed and the first
// error returned.
func (t *Txn) Range(start Key, end Key, f KeyFunc) (err error) {
	if bytes.Compare(start, end) == 1 {
		return ErrInvalidRange
	}

	commonPrefix := lcp.LCP(start, end)
	if commonPrefix == nil {
		return ErrInvalidRange
	}

	t.trie.Root().WalkPrefix(commonPrefix, func(key []byte, item internal.Item) bool {
		if bytes.Compare(key, start) >= 0 && bytes.Compare(key, end) <= 0 {
			if err = f(key); err != nil {
				return true
			}
			return false
		} else if bytes.Compare(key, start) >= 0 && bytes.Compare(key, end) > 0 {
			return true
		}
		return false
	})
	return
}

// Scan performs a prefix scan of keys matching the given prefix and calling
// the function `f` with the keys found. If the function returns an error no
// further keys are processed and the first error is returned.
func (t *Txn) Scan(prefix Key, f KeyFunc) (err error) {
	t.trie.Root().WalkPrefix(prefix, func(key []byte, item internal.Item) bool {
		// Skip the root node
		if len(key) == 0 {
			return false
		}

		if err = f(key); err != nil {
			return true
		}
		return false
	})
	return
}

// Transaction returns a new transaction that is a snapshot of the key space of
// the database and isolated from other transactions. Key/Value pairs written in
// a transaction are batched together. Transactions are not thread safe and should
// only be used in a single goroutine.
func (b *Bitcask) Transaction() *Txn {
	b.mu.RLock()
	defer b.mu.RUnlock()

	current := data.NewInMemoryDatafile(-1, b.config.MaxKeySize, b.config.MaxValueSize)
	previous := b.current.ReopenReadonly()
	datafiles := b.datafiles

	return &Txn{
		db:        b,
		current:   current,
		previous:  previous,
		datafiles: datafiles,
		batch:     b.Batch(),
		trie:      b.trie.Txn(),
	}
}
