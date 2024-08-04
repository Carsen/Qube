package bitcask

import (
	"fmt"
	"sync"

	"go.mills.io/bitcask/v2/internal"
	"go.mills.io/bitcask/v2/internal/codec"
	"go.mills.io/bitcask/v2/internal/config"
)

type batchOptions struct {
	maxKeySize   uint32
	maxValueSize uint64
}

func defaultBatchOptions(cfg *config.Config) *batchOptions {
	return &batchOptions{
		maxKeySize:   cfg.MaxKeySize,
		maxValueSize: cfg.MaxValueSize,
	}
}

// Batch ...
type Batch struct {
	db      *Bitcask
	mu      sync.RWMutex
	entries []internal.Entry
	opts    *batchOptions
}

// Clear clears the batch
func (b *Batch) Clear() {
	b.mu.Lock()
	b.entries = nil
	b.mu.Unlock()
}

// Entries returns all the entries in the batch
func (b *Batch) Entries() []internal.Entry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.entries
}

// Delete deletes a key from the batch
func (b *Batch) Delete(key Key) (internal.Entry, error) {
	entry := internal.NewEntry(key, Value(nil))

	b.mu.Lock()
	b.entries = append(b.entries, entry)
	b.mu.Unlock()

	return entry, nil
}

// Put writes a new key/value to the batch
func (b *Batch) Put(key Key, value Value) (internal.Entry, error) {
	if len(key) == 0 {
		return internal.Entry{}, ErrEmptyKey
	}
	if b.opts.maxKeySize > 0 && uint32(len(key)) > b.opts.maxKeySize {
		return internal.Entry{}, ErrKeyTooLarge
	}
	if b.opts.maxValueSize > 0 && uint64(len(value)) > b.opts.maxValueSize {
		return internal.Entry{}, ErrValueTooLarge
	}

	entry := internal.NewEntry(key, value)

	b.mu.Lock()
	b.entries = append(b.entries, entry)
	b.mu.Unlock()

	return entry, nil
}

// Batch returns a new batch that allows multiple keys to be deleted and written
// in a single operation.
func (b *Bitcask) Batch() *Batch {
	return &Batch{db: b, opts: defaultBatchOptions(b.config)}
}

// WriteBatch writes the batch to the database. The batch is not cleared when
// the write has completed, so call Clear() to reset the batch if needed.
func (b *Bitcask) WriteBatch(batch *Batch) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.current.Readonly() {
		return ErrDatabaseReadonly
	}

	b.metadata.IndexUpToDate = false

	for _, entry := range batch.Entries() {
		if err := b.maybeRotate(); err != nil {
			return fmt.Errorf("error rotating active datafile: %w", err)
		}

		offset, n, err := b.current.Write(entry)
		if err != nil {
			return err
		}

		if b.config.SyncWrites {
			if err := b.current.Sync(); err != nil {
				return err
			}
		}

		// in case of successful write, IndexUpToDate will be always be false
		b.metadata.IndexUpToDate = false

		if entry.Value != nil {
			if oldItem, found := b.trie.Root().Get(entry.Key); found {
				b.metadata.ReclaimableSpace += oldItem.Size
			}
			item := internal.Item{FileID: b.current.FileID(), Offset: offset, Size: n}
			b.trie, _, _ = b.trie.Insert(entry.Key, item)
		} else {
			if oldItem, found := b.trie.Root().Get(entry.Key); found {
				b.metadata.ReclaimableSpace += oldItem.Size + codec.MetaInfoSize + int64(len(entry.Key))
			}
			b.trie, _, _ = b.trie.Delete(entry.Key)
		}
	}

	return nil
}
