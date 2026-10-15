package main

import (
	"fmt"
	"time"
	"kv-lite/pkg/storage"
)

func main() {
	key := []byte("user_001")
	val := []byte("Olamide")

	// Encode record
	encoded, err := storage.EncodeRecord(key, val)
	if err != nil {
		panic(err)
	}

	// fmt.Printf("Key length: %d bytes\n", len(key))
	// fmt.Printf("Value length: %d bytes\n", len(val))
	// fmt.Printf("Total binary frame length: %d bytes\n\n", len(encoded))
	// fmt.Printf("Raw Byte Stream: %v\n", encoded)

	fmt.Print("Encoded record into %d bytes\n", len(encoded))


	// Decode record
	decoded, err := storage.DecodeRecord(encoded)
	if err != nil {
		panic(err)
	}

	fmt.Println("--- Successfully Decoded Record ---")
	fmt.Printf("Key:       %s\n", string(decoded.Key))
	fmt.Printf("Value:       %s\n", string(decoded.Value))
	fmt.Printf("Timestamp: %s\n", time.Unix(0, decoded.Timestamp).Format(time.RFC3339))
	fmt.Printf("Checksum:  %d\n\n", decoded.Checksum)

	// Simulate data corruption
	fmt.Println("--- Simulating Bit Corruption ---")
	encoded[20] ^= 0xFF // Flip bits at byte position 20 in the payload

	_, err = storage.DecodeRecord(encoded)
	if err != nil {
		fmt.Printf("Caught expected error: %v\n", err)
	}

}