# bitcask

[![Go Reference](https://pkg.go.dev/badge/go.mills.io/bitcask/v2.svg)](https://pkg.go.dev/go.mills.io/bitcask/v2)
[![Build Status](https://git.mills.io/prologic/bitcask/actions/workflows/build.yml/badge.svg)](https://git.mills.io/prologic/bitcask/actions?workflow=build.yml)
[![Test Status](https://git.mills.io/prologic/bitcask/actions/workflows/test.yml/badge.svg)](https://git.mills.io/prologic/bitcask/actions?workflow=test.yml)

A high performance Key/Value store written in [Go](https://golang.org) with a predictable read/write performance and high throughput. Uses a [Bitcask](https://en.wikipedia.org/wiki/Bitcask) on-disk layout (LSM+WAL) similar to [Riak](https://riak.com/)

For a more feature-complete Redis-compatible server, distributed key/value store have a look at [Bitraft](https://git.mills.io/prologic/bitraft) which uses this library as its backend. Use [Bitcask](https://go.mills.io/bitcask/v2) as a starting point or if you want to embed in your application, use [Bitraft](https://git.mills.io/prologic/bitraft) if you need a complete server/client solution with high availability with a Redis-compatible API.

Table of Contents:

<!--toc-->
- [bitcask](#bitcask)
    * [Features](#features)
    * [Migrating from Bitcask v1](#migrating-from-bitcask-v1)
    * [Is Bitcask right for my project?](#is-bitcask-right-for-my-project)
    * [Transactions](#transactions)
    * [Development](#development)
    * [Install](#install)
    * [Usage (library)](#usage-library)
        * [Configuration Options](#configuration-options)
    * [Usage (tool)](#usage-tool)
    * [Usage (server)](#usage-server)
    * [Docker](#docker)
    * [Performance](#performance)
    * [Contributors](#contributors)
    * [Related Projects](#related-projects)
    * [License](#license)

<!-- tocstop -->

## Features

* Embedded (`import "go.mills.io/bitcask/v2"`)
* Builtin CLI (`bitcask`)
* Builtin Redis-compatible server (`bitcaskd`)
* Predictable read/write performance
* High throughput (See: [Performance](README.md#Performance) )
* Full Transactions support (ACID)
* Low latency

## Migrating from Bitcask v1

If you are migrating from Bitcask v1 ([git.mills.io/prologic/bitcask)[bitcask-v1]),
to Bitcask v2 ([go.mills.io/bitcask/v2][bitcask-v2]), please update your code as follows:

- Update your import path to [go.mills.io/bitcask/v2][bitcask-v2]:

```diff
import (
- "git.mills.io/prologic/bitcask"
+ "go.mills.io/bitcask/v2"
)
```

- `WithSync(true)` was renamed to `WithSyncWrites(true)`

- Iterators that take a `bitcask.KeyFunc` as input now use `bitcask.Key` as the type for keys rather than `[] byte]`.

```diff
- db.Scan(prefix, func (key []byte) error {
+ db.Scan(prefix, func (key bitcask.Key) error {
  //
})
```

- `Fold()` was renamed to `ForEach()` (_see other changes above_)

## Is Bitcask right for my project?

> [!NOTE]
> Please read this carefully to identify whether using Bitcask is
> suitable for your needs.

`bitcask` is a **great fit** for:

- Storing hundreds of thousands to millions of key/value pairs based on
  default configuration. With the default configuration (_configurable_)
  of 64 bytes per key and 64kB values, 1M keys would consume roughly ~600-700MB
  of memory ~65-70GB of disk storage. These are all configurable when you
  create a new database with `bitcask.Open(...)` with functional-style options
  you can pass with `WithXXX()`.

- As the backing store to a distributed key/value store. See for example the
  [bitraft](https://git.mills.io/prologic/bitraft) as an example of this.

- For high performance, low latency read/write workloads where you cannot fit
  a typical hash-map into memory, but require the highest level of performance
  and predicate read latency. Bitcask ensures only 1 read/write IOPS are ever
  required for reading and writing key/value pairs.

- As a general purpose embedded key/value store where you would have used
  [BoltDB](https://github.com/boltdb/bolt),
  [LevelDB](https://github.com/syndtr/goleveldb),
  [BuntDB](https://github.com/tidwall/buntdb)
  or similar...

`bitcask` is not suited for:

- Storing billions of records
  The reason for this is the key-space is held in memory using radix tree.
  This means the more keys you have in your key space, the more memory is
  consumed. Consider using a disk-backed B-Tree like [BoltDB](https://github.com/boltdb/bolt)
  or [LevelDB](https://github.com/syndtr/goleveldb) if you intend to store a
  large quantity of key/value pairs.

> [!NOTE]
> However that storing large amounts of data in terms of value(s) is
> totally fine. In other words, thousands to millions of keys with large values
> will work just fine.

## Transactions

Bitcask supports transactions with ACID semantics. A call to `Txn()` returns a
new transaction that is a snapshot of the current trie of keys. Keys written to
a transaction are committed as a single batch operation, providing **A**tomicity.

As writes are performed in the transaction, we maintain an internal cache of
new entries written within the transaction. Thus, any follow up reads on the
same key by this transaction would see this write.
But, other transactions won’t, providing **I**solation and **C**onsistency.

Finally **D**urability in Bitcask is guaranteed with by a write-ahead-log of the
current datafile and further guaranteed by enabling synchronous writes with the
`WithSyncWrites(true)` option.

> [!WARNING]
> A transaction is not thread safe and should only be used by a single goroutine.

## Development

```sh
git clone https://git.mills.io/prologic/bitcask.git
make
```

## Install

```sh
go get go.mills.io/bitcask/v2
```

## Usage (library)

Install the package into your project:

```sh
go get go.mills.io/bitcask/v2
```

```go
package main

import (
	"log"
	"go.mills.io/bitcask/v2"
)

func main() {
    db, _ := bitcask.Open("/tmp/db")
    defer db.Close()
    db.Put([]byte("Hello"), []byte("World"))
    val, _ := db.Get([]byte("Hello"))
    log.Printf(string(val))
}
```

See the [GoDoc](https://godoc.org/go.mills.io/bitcask/v2) for further
documentation and other examples.

See also [examples](./examples)

### Configuration Options

The default options (_if none are specified_) default to a Bitcask instance with:

- Maximum Keys of 64 bytes
- Maximum Values of 65 KB
- Maximum Datafiles of 1 MB (_before rotating_)
- Synchronous Writes: off
- Auto Recovery: on

The defaults are designed for high performance in mind with recovery on startup and support limits of ~16M keys
and ~1GB of persitent storage with the default file descriptor limits on most Linux systems.

Any of these options can be changed with any of the `WithXXX(...)` options.

> [!NOTE]
> If you require better reliability over performance,
> please enable synchronous writes with the `WithSyncWrites(true)`.

Bitcask is an embedded key/value store designed for handling write-intensive workloads. However, frequent write operations leading to a large number of new key-value pairs over time can result in issues like "Too many open files" (#193) errors due to the creation of numerous data files. These problems can be mitigated by periodically compacting the data through issuing a `.Merge()` operation, increasing the maximum value size with the `MaxDatafileSize()` option, and increasing the process file descriptor limit. Example: With a `MaxDatafileSize(1<<30)` (1GB) and a file descriptor limit of 1M (million) files, you are able to store up to 1PB (Petabytes) of (compacted) data before you hit "Too many open files", assuming a single machine can even handle this.

> You should consider your read/write workloads carefully and ensure you set
> appropriate file descriptor limits with `ulimit -n` that suit your needs.

## Usage (tool)

```sh
bitcask -p /tmp/db set Hello World
bitcask -p /tmp/db get Hello
World
```

## Usage (server)

There is also a builtin very  simple Redis-compatible server called `bitcaskd`:

```sh
./bitcaskd ./tmp
INFO[0000] starting bitcaskd v0.0.7@146f777              bind=":6379" path=./tmp
```

Example session:

```sh
telnet localhost 6379
Trying ::1...
Connected to localhost.
Escape character is '^]'.
SET foo bar
+OK
GET foo
$3
bar
DEL foo
:1
GET foo
$-1
PING
+PONG
QUIT
+OK
Connection closed by foreign host.
```

## Docker

You can also use the [Bitcask Docker Image](https://cloud.docker.com/u/prologic/repository/docker/prologic/bitcask):

```sh
docker pull prologic/bitcask
docker run -d -p 6379:6379 prologic/bitcask
```

## Performance

```sh
make bench
...

BenchmarkGet/128B-10             1029229              1191 ns/op         107.46 MB/s        4864 B/op         14 allocs/op
BenchmarkGet/256B-10              916785              1190 ns/op         215.16 MB/s        4992 B/op         14 allocs/op
BenchmarkGet/512B-10              831607              1261 ns/op         406.19 MB/s        5280 B/op         14 allocs/op
BenchmarkGet/1K-10                796448              1384 ns/op         740.06 MB/s        5856 B/op         14 allocs/op
BenchmarkGet/2K-10                612469              1724 ns/op        1187.83 MB/s        7008 B/op         14 allocs/op
BenchmarkGet/4K-10                515680              2314 ns/op        1770.19 MB/s        9568 B/op         14 allocs/op
BenchmarkGet/8K-10                375813              3231 ns/op        2535.32 MB/s       14176 B/op         14 allocs/op
BenchmarkGet/16K-10               236959              5115 ns/op        3203.28 MB/s       23136 B/op         14 allocs/op
BenchmarkGet/32K-10               129828              9449 ns/op        3467.77 MB/s       45664 B/op         14 allocs/op

BenchmarkPut/128BNoSync-10        249405              5116 ns/op          25.02 MB/s        6737 B/op         46 allocs/op
BenchmarkPut/256BNoSync-10        155542              6896 ns/op          37.12 MB/s        6867 B/op         46 allocs/op
BenchmarkPut/1KNoSync-10           72939             19902 ns/op          51.45 MB/s        7740 B/op         46 allocs/op
BenchmarkPut/2KNoSync-10           37819             33780 ns/op          60.63 MB/s        8904 B/op         46 allocs/op
BenchmarkPut/4KNoSync-10           18554             70200 ns/op          58.35 MB/s       18914 B/op         47 allocs/op
BenchmarkPut/8KNoSync-10            8276            167674 ns/op          48.86 MB/s       20249 B/op         47 allocs/op
BenchmarkPut/16KNoSync-10           3660            333656 ns/op          49.10 MB/s       29291 B/op         47 allocs/op
BenchmarkPut/32KNoSync-10           2275            561683 ns/op          58.34 MB/s       52000 B/op         48 allocs/op

BenchmarkPut/128BSync-10             258           5149745 ns/op           0.02 MB/s        6736 B/op         46 allocs/op
BenchmarkPut/256BSync-10             211           5138904 ns/op           0.05 MB/s        6864 B/op         46 allocs/op
BenchmarkPut/1KSync-10               207           5356101 ns/op           0.19 MB/s        7728 B/op         46 allocs/op
BenchmarkPut/2KSync-10               247           5212069 ns/op           0.39 MB/s        8932 B/op         46 allocs/op
BenchmarkPut/4KSync-10               207           5043624 ns/op           0.81 MB/s       18924 B/op         47 allocs/op
BenchmarkPut/8KSync-10               208           5411918 ns/op           1.51 MB/s       20204 B/op         47 allocs/op
BenchmarkPut/16KSync-10              234           5367222 ns/op           3.05 MB/s       29261 B/op         47 allocs/op
BenchmarkPut/32KSync-10              198           5594519 ns/op           5.86 MB/s       51996 B/op         48 allocs/op

BenchmarkScan-10                 1112818              1066 ns/op            4986 B/op         22 allocs/op
```

For 128B values:

* ~1,000,000 reads/sec
* ~250,000 writes/sec
* ~1,100,000 scans/sec

The full benchmark above shows linear performance as you increase key/value sizes.

As far as benchmarks go, this is all contrived and generally not typical of any real workloads. These benchmarks were run on a 2022 Mac Studio M1 Max with 32GB of RAM. Your results may differ.

## Contributors

Thank you to all those that have contributed to this project, battle-tested it,
used it in their own projects or products, fixed bugs, improved performance
and even fix tiny typos in documentation! Thank you and keep contributing!

You can find an [AUTHORS](/AUTHORS) file where we keep a list of contributors
to the project. If you contribute a PR please consider adding your name there.

## Related Projects

- [bitraft](https://git.mills.io/prologic/bitraft) -- A Distributed Key/Value store (_using Raft_) with a Redis compatible protocol.
- [bitcaskfs](https://go.mills.io/bitcask/v2fs) -- A FUSE file system for mounting a Bitcask database.
- [bitcask-bench](https://go.mills.io/bitcask/v2-bench) -- A benchmarking tool comparing Bitcask and several other Go key/value libraries.

## License

bitcask is licensed under the term of the [MIT License](https://go.mills.io/bitcask/v2/blob/master/LICENSE)

----

[bitcask-v1]: https://pkg.go.dev/git.mills.io/prologic/bitcask
[bitcask-v2]: https://pkg.go.dev/go.mills.io/bitcask/v2
