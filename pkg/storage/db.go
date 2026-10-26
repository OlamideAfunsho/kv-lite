package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

// DB represents the unified key-value storage engine instance
type DB struct {
	dirPath string
	wal *WAL
	memTable *MemTable
}

// This function initializes or opens the database directory, runs crash recovery from the WAL,
// and returns a ready to use DB instance
func OpenDB(dirPath string) (*DB, error) {
	// Ensure the database directory exits on disk
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create db directory: %w", err)
	}

	walPath := filepath.Join(dirPath, "wal.log")

	// Open the Write-Ahead Log
	wal, err := OpenWAL(walPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open wal: %w", err)
	}

	// Initialize MemTable
	memTable := NewMemTable()

	// Crash recovery pipeline: Read all records from WAL and load into MemTable
	records, err := wal.ReadAll()
	if err != nil {
		wal.Close()
		return nil, fmt.Errorf("failed to recover state from wal: %w", err)
	}

	memTable.LoadFromWAL(records)
	fmt.Printf("[Engine] Crash recovery complete. Restored %d records into memory.\n", len(records))

	return &DB{
		dirPath: dirPath,
		wal: wal,
		memTable: memTable,
	}, nil

}


// This function writes a key-value pair safely: first to the WAL on disk, and then to the MemTable in RAM
func (db *DB) Set(key, value []byte) error {
	// Encode binary frame
	frame, err := EncodeRecord(key, value)
	if err != nil {
		return fmt.Errorf("failed to encode record: %w", err)
	}

	// Write to WAL (note: durability first)
	if err := db.wal.Write(frame); err != nil {
		return fmt.Errorf("failed to write to wal: %w", err)
	}

	// Udpdate MemTable (speed)
	db.memTable.Set(key, value)

	return nil
}


// This function retries a value by key from the MemTable
func (db *DB) Get(key []byte) ([]byte, bool) {
	return db.memTable.Get(key)
}


// This function safely shuts down the database engine and releases the file handles
func (db *DB) Close() error {
	if err := db.wal.Close(); err != nil {
		return fmt.Errorf("failed to close wal: %w", err)
	}
	return nil
}