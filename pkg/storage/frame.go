package storage

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"time"
)


// HeaderSize defines the fixed size of our binary record header in bytes
// CRC32 (4) + Timestamp (8) + Key Size (2) + Value Size (2) = 16 bytes
const HeaderSize = 16

// Record represents a single key-value entry in memory
type Record struct {
	Checksum uint32
	Timestamp int64
	Key []byte
	Value []byte
}


// This function transforms a key and value into contiguous byte slice according to our wire format
func EncodeRecord(key, value []byte) ([]byte, error) {
	if len(key) > 65535 || len(value) > 65535 {
		return nil, errors.New("key or value size exceeds maximum allowed length of 65535 bytes")
	}

	keyLen := uint16(len(key))
	valLen := uint16(len(value))
	totalSize := HeaderSize + int(keyLen) + int(valLen)

	// Pre-allocate contiguous memory buffer for the full frame
	buf := make([]byte, totalSize)

	// Set Timestamp (bytes 4 to 11)
	timestamp := time.Now().UnixNano()
	binary.BigEndian.PutUint64(buf[4:12], uint64(timestamp))
	
	// Set Key Length (bytes 12 to 13)
	binary.BigEndian.PutUint16(buf[12:14], keyLen)

	// Set Value Length (bytes 14 to 15)
	binary.BigEndian.PutUint16(buf[14:16], valLen)

	// Copy Key and Value payload bytes into the buffer
	copy(buf[HeaderSize:HeaderSize+int(keyLen)], key)
	copy(buf[HeaderSize+int(keyLen):], value)

	// Calculate CRC32 Checksum over all payload data (Timestamp + KeyLen + ValLen + Key + Value)
	// and write it to the first 4 Bytes (Bytes 0 to 3)
	checksum := crc32.ChecksumIEEE(buf[4:])
	binary.BigEndian.PutUint32(buf[0:4], checksum)

	return buf, nil


} 


// This function parses a binary byte slice into a Record struct
// It verifies the 16-byte fixed header, unpacks the key/value payloads, 
// and validates the CRC32 checksum for data corruption
func DecodeRecord(buf []byte) (*Record, error) {
	if len(buf) < HeaderSize {
		return nil, fmt.Errorf("buffer too small: expected at least %d bytes, got %d", HeaderSize, len(buf))
	}

	// Read stored CRC32 Checksum (bytes 0 to 3)
	storedChecksum := binary.BigEndian.Uint32(buf[0:4])

	// Validate CRC32 Checksum over payload (bytes 4 onward)
	calculatedChecksum := crc32.ChecksumIEEE(buf[4:])
	if storedChecksum != calculatedChecksum {
		return nil, fmt.Errorf("data corruption detected: checksum mismatch (stored: %d, calculated: %d)", storedChecksum, calculatedChecksum)
	}

	// Unpack Timestamp (bytes 4 to 11)
	timestamp := int64(binary.BigEndian.Uint64(buf[4:12]))

	// Unpack Key length (bytes 12 to 13) and Value length (bytes 14 to 15)
	keyLen := int(binary.BigEndian.Uint16(buf[12:14]))
	valLen := int(binary.BigEndian.Uint16(buf[14:16]))

	// Ensure the buffer is large enough for the full header + payload
	expectedTotalSize := HeaderSize + keyLen + valLen
	if len(buf) < expectedTotalSize {
		return nil, fmt.Errorf("incomplete buffer: expected %d bytes, got %d", expectedTotalSize, len(buf))
	}

	// Slice key and value bytes
	key := make([]byte, keyLen)
	val := make([]byte, valLen)
	copy(key, buf[HeaderSize:HeaderSize+keyLen])
	copy(val, buf[HeaderSize+keyLen:expectedTotalSize])

	return &Record{
		Checksum: storedChecksum,
		Timestamp: timestamp,
		Key: key,
		Value: val,
	}, nil
}

// DecodeHeaderLengths extracts key length and value length from a 16-byte fixed header slice.
func DecodeHeaderLengths(headerBuf []byte) (int, int) {
	keyLen := int(binary.BigEndian.Uint16(headerBuf[12:14]))
	valLen := int(binary.BigEndian.Uint16(headerBuf[14:16]))
	return keyLen, valLen
}