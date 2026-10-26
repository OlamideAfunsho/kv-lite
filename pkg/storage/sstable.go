package storage

import (
	"fmt"
	"os"
	"sort"
)

// This function flushes an in-memory map out to an immutable disk file,
// sorting all keys alphabetically to prepare for fast binary searches later
func WriteSSTable(path string, data map[string][]byte) error {
	// Extract and sort keys alphabetically to maintain a sorted disk layout
	var keys []string
	for k := range data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create the SSTable file on disk
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create sstable file: %w", err)
	}
	defer file.Close()

	// Iterate through sorted keys and write their binary frames sequentially
	for _, k := range keys {
		keyBytes := []byte(k)
		valBytes := data[k]

		// Reuse our binary frame encoder
		frame, err := EncodeRecord(keyBytes, valBytes)
		if err != nil {
			return fmt.Errorf("failed to encode record for sstable: %w", err)
		}

		_, err = file.Write(frame)
		if err != nil {
			return fmt.Errorf("failed to write frame to sstable: %w", err)
		}
	}

	// Flush OS kernel buffer to disk storage
	err = file.Sync()
	if err != nil {
		return fmt.Errorf("failed to sync sstable to disk: %w", err)
	}

	return nil
}