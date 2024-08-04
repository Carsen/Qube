package bitcask

import (
	"errors"

	iradix "github.com/hashicorp/go-immutable-radix/v2"
	"go.mills.io/bitcask/v2/internal"
)

var (
	// ErrIteratorClosed ...
	ErrIteratorClosed = errors.New("error: iterator is closed")

	// ErrStopIteration ...
	ErrStopIteration = errors.New("error: iterator has no more items")
)

type iteratorOptions struct {
	reverse bool
}

// IteratorOption ...
type IteratorOption func(it *Iterator)

// Reverse ...
func Reverse() IteratorOption {
	return func(it *Iterator) {
		it.opts.reverse = true
	}
}

// Iterator ...
type Iterator struct {
	db   *Bitcask
	itf  *iradix.Iterator[internal.Item]
	itr  *iradix.ReverseIterator[internal.Item]
	opts *iteratorOptions
}

func (it *Iterator) Close() error {
	if it.itf == nil && it.itr == nil {
		return ErrIteratorClosed
	}
	it.itf = nil
	it.itr = nil
	return nil
}

func (it *Iterator) Next() (*Item, error) {
	var (
		key  []byte
		more bool
	)

	if it.opts.reverse {
		key, _, more = it.itr.Previous()
	} else {
		key, _, more = it.itf.Next()
	}

	if !more {
		defer it.Close()
		return nil, ErrStopIteration
	}
	value, err := it.db.Get(key)
	if err != nil {
		defer it.Close()
		return nil, err
	}
	return &Item{key, value}, nil
}

func (it *Iterator) SeekPrefix(prefix Key) (*Item, error) {
	if it.opts.reverse {
		it.itr.SeekPrefix(prefix)
	} else {
		it.itf.SeekPrefix(prefix)
	}
	return it.Next()
}

// Iterator returns an iterator for iterating through keys in key order
func (b *Bitcask) Iterator(opts ...IteratorOption) *Iterator {
	it := &Iterator{db: b, opts: &iteratorOptions{}}
	for _, opt := range opts {
		opt(it)
	}
	if it.opts.reverse {
		it.itr = b.trie.Root().ReverseIterator()
	} else {
		it.itf = b.trie.Root().Iterator()
	}
	return it
}
