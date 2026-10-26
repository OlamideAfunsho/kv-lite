package storage

import (
	"fmt"
	"os"
	"sort"
)

// IndexEntry maps a key to its exact byte offset inside an SSTable file
type IndexEntry struct {
	Key string
	Offset int64
}

// This function flushes an in-memory map out to an immutable disk file,
// sorting all keys alphabetically to prepare for fast binary searches later
func WriteSSTable(path string, data map[string][]byte) ([]IndexEntry, error) {
	// Extract and sort keys alphabetically to maintain a sorted disk layout
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create the SSTable file on disk
	file, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("failed to create sstable file: %w", err)
	}
	defer file.Close()

	var index []IndexEntry
	var currentOffset int64 = 0

	// Iterate through sorted keys and write their binary frames sequentially
	for _, k := range keys {
		keyBytes := []byte(k)
		valBytes := data[k]

		// Reuse our binary frame encoder
		frame, err := EncodeRecord(keyBytes, valBytes)
		if err != nil {
			return nil, fmt.Errorf("failed to encode record for sstable: %w", err)
		}

		// Record the starting byte offset of this frame before writing it
		index = append(index, IndexEntry{
			Key: k,
			Offset: currentOffset,
		})

		// Write binary frame to disk
		n, err := file.Write(frame)
		if err != nil {
			return nil, fmt.Errorf("failed to write frame to sstable: %w", err)
		}

		// Advance our file offset tracker by the exact byte length of the written frame
		currentOffset += int64(n)
	}

	// Flush OS kernel buffer to disk storage
	err = file.Sync()
	if err != nil {
		return nil, fmt.Errorf("failed to sync sstable to disk: %w", err)
	}

	return index, nil
}


// This function seeks directly to a specific byte offset in an SSTable file
// and reads/decodes the record bypassing full-file scans
func ReadSSTableRecord(path string, targetOffset int64) (*Record, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open sstable for reading: %w", err)
	}
	defer file.Close()

	// Seek directly to the target byte offset in the file
	_, err = file.Seek(targetOffset, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to seek sstable offset: %w", err)
	}

	// Read the 16-byte fixed header
	headerBuf := make([]byte, HeaderSize)
	_, err = file.Read(headerBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read sstable frame header: %w", err)
	}

	// Unpack lengths to get payload size
	keyLen, valLen := DecodeHeaderLengths(headerBuf)
	payloadLen := keyLen + valLen

	// Read the exact payload bytes
	payloadBuf := make([]byte, payloadLen)
	_, err = file.Read(payloadBuf)
	if err != nil {
		return nil, fmt.Errorf("failed to read sstable frame payload: %w", err)
	}

	// Reconstruct full frame and decode
	fullFrame := append(headerBuf, payloadBuf...)
	record, err := DecodeRecord(fullFrame)
	if err != nil {
		return nil, fmt.Errorf("failed to decode sstable record: %w", err)
	}

	return record, nil
}