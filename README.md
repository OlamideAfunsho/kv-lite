# kv-lite

A lightweight embedded key-value storage engine written in Go. Writes go to a write-ahead log on disk before they are applied in memory, so data survives a crash or restart.

## Features

- **Write-ahead log (WAL):** every write is appended to `data/wal.log` and synced to disk with `fsync` before it is acknowledged.
- **Crash recovery:** on startup, the engine replays the WAL to rebuild its in-memory state.
- **In-memory MemTable:** a thread-safe map (guarded by `sync.RWMutex`) that serves all reads.
- **Binary record format with checksums:** every record carries a CRC32 checksum, which is verified when the record is read back.
- **SSTable module:** writes key-value data to a sorted, immutable file and returns an offset index for direct, seek-based reads.
- **Graceful shutdown:** the CLI catches `SIGINT`/`SIGTERM` (Ctrl+C) and closes the database cleanly.

## Requirements

- Go 1.26.5 or later (as set in `go.mod`)

## Usage

Run commands from the project root:

```bash
# Store a value
go run cmd/cli/main.go set <key> <value>

# Retrieve a value
go run cmd/cli/main.go get <key>
```

Example:

```bash
$ go run cmd/cli/main.go set name olamide_prodigy
[Engine] Crash recovery complete. Restored 0 records into memory.
OK (written in 1.2ms)

$ go run cmd/cli/main.go get name
[Engine] Crash recovery complete. Restored 1 records into memory.
olamide_prodigy (read in 2µs)

$ go run cmd/cli/main.go get missing
[Engine] Crash recovery complete. Restored 1 records into memory.
(nil)
```

Write and read times will vary. Data is stored in the `./data` directory, which is created automatically if it does not exist.

## Project structure

```
kv-lite/
├── cmd/
│   └── cli/
│       └── main.go       # Command-line interface (set / get)
├── pkg/
│   └── storage/
│       ├── db.go         # DB engine: open, recover, Set, Get, Close
│       ├── wal.go        # Append-only write-ahead log
│       ├── memtable.go   # Thread-safe in-memory key-value map
│       ├── frame.go      # Binary record encoding/decoding
│       └── sstable.go    # Sorted string table writer and reader
├── go.mod
└── README.md
```

## How it works

### Write path

1. The key and value are encoded into a binary frame (`EncodeRecord`).
2. The frame is appended to the WAL and synced to disk.
3. The key-value pair is stored in the MemTable.

### Read path

`Get` looks up the key in the MemTable and returns a copy of the value, so callers cannot change the engine's internal state.

### Recovery

When the database opens, `wal.ReadAll()` reads every frame from the start of the log, validates its checksum, and loads the records into the MemTable. If the same key was written more than once, the latest write wins.

## Record format

Every record in the WAL and in SSTables uses the same frame layout (all integers are big-endian):

| Bytes   | Field        | Type     |
|---------|--------------|----------|
| 0–3     | CRC32        | `uint32` |
| 4–11    | Timestamp    | `int64` (Unix nanoseconds) |
| 12–13   | Key length   | `uint16` |
| 14–15   | Value length | `uint16` |
| 16…     | Key bytes    |          |
| …       | Value bytes  |          |

The checksum (CRC32 IEEE) covers everything from byte 4 to the end of the frame. A mismatch on decode is reported as data corruption.

## SSTables

`sstable.go` provides two functions:

- `WriteSSTable(path, data)`: sorts the keys alphabetically, writes each pair as a binary frame, syncs the file to disk, and returns an index of `{Key, Offset}` entries.
- `ReadSSTableRecord(path, offset)`: seeks to a byte offset, reads the frame header and payload, and decodes the record.

These functions are standalone. `DB` does not call them.

## Limits

- Keys and values can be at most 65,535 bytes each (lengths are stored as `uint16`).
- The CLI supports `set` and `get` only.
