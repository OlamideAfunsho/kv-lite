package storage

import (
	"fmt"
	"io"
	"os"
	"sync"
)


// WAL represents an Append-Only Write-Ahead Log file handle
type WAL struct {
	mu sync.Mutex
	file *os.File
}

// OpenWAL creates or opens a Write-Ahead Log file at the specified path
func OpenWAL(path string) (*WAL, error) {
	// Open file with flags:
	// - O_CREATE: Create file if it does not exist
	// - O_APPEND: All writes automatically append to the end of the file
	// - O_RDWR:   Open for both reading and writing
	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open wal file: %w", err)
	}

	return &WAL{
		file: file,	
	}, nil
}

// Write appends an encoded binary record frame to the WAL on disk in a thread-safe manner
func (w *WAL) Write(frame []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Append bytes to file
	_, err := w.file.Write(frame)
	if err != nil {
		return fmt.Errorf("failed to write frame to wal: %w", err)
	}

	// Flush file OS buffers directly to physical disk storage
	err = w.file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync wal to disk: %w", err)
	}

	return nil

}


// ReadAll iterates sequentailly through the WAL file from the start,
// decoding all valid binary records to recover memory state on startup
func (w *WAL) ReadAll() ([]*Record, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Move file pointer to the very beginning of the WAL file
	_, err := w.file.Seek(0, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek start of wal file: %w", err)
	}

	var records []*Record

	for {
		// Read the 16-byte fixed header first
		headerBuf := make([]byte, HeaderSize)
		_, err :=io.ReadFull(w.file, headerBuf)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break // Reached end of file cleanly
		} 
		if err != nil {
			return nil, fmt.Errorf("failed to read frame header from wal: %w", err)
		}

		// Unpack key and value lengths from header to know how many more bytes to read
		keyLen, valLen := DecodeHeaderLengths(headerBuf)
		payloadLen := keyLen + valLen

		// Read the exact payload (Key + Value)
		payloadBuf  := make([]byte, payloadLen)
		_, err = io.ReadFull(w.file, payloadBuf)
		if err != nil {
			return nil, fmt.Errorf("failed to read frame payload from wal: %w", err)
		}

		// Combine header + payload into a complete frame buffer and decode
		fullFrame := append(headerBuf, payloadBuf...)
		record, err := DecodeRecord(fullFrame)
		if err != nil {
			return nil, fmt.Errorf("wal record corruption detected: %w", err)
		}

		records = append(records, record)

	}

	return records, nil

}


// Close safely closes the WAL file handle
func (w *WAL) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.file.Close()
}