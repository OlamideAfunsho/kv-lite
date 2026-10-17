package main

import (
	"fmt"
	"os"
	"kv-lite/pkg/storage"
)

func main() {
	// key := []byte("user_001")
	// val := []byte("Olamide")

	// // Encode record
	// encoded, err := storage.EncodeRecord(key, val)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("Key length: %d bytes\n", len(key))
	// fmt.Printf("Value length: %d bytes\n", len(val))
	// fmt.Printf("Total binary frame length: %d bytes\n\n", len(encoded))
	// fmt.Printf("Raw Byte Stream: %v\n", encoded)

	// fmt.Print("Encoded record into %d bytes\n", len(encoded))


	// // Decode record
	// decoded, err := storage.DecodeRecord(encoded)
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Println("--- Successfully Decoded Record ---")
	// fmt.Printf("Key:       %s\n", string(decoded.Key))
	// fmt.Printf("Value:       %s\n", string(decoded.Value))
	// fmt.Printf("Timestamp: %s\n", time.Unix(0, decoded.Timestamp).Format(time.RFC3339))
	// fmt.Printf("Checksum:  %d\n\n", decoded.Checksum)

	// // Simulate data corruption
	// fmt.Println("--- Simulating Bit Corruption ---")
	// encoded[20] ^= 0xFF // Flip bits at byte position 20 in the payload

	// _, err = storage.DecodeRecord(encoded)
	// if err != nil {
	// 	fmt.Printf("Caught expected error: %v\n", err)
	// }

	walPath := "wal.log"

	// Cleanup any existing wal.log file from previous test runs
	os.Remove(walPath)

	// Open new WAL instance
	wal, err := storage.OpenWAL(walPath)
	if err != nil {
		panic(err)
	}

	fmt.Println("--- Writing Records to WAL ---")

	// Write entry 1
	frame1, _ := storage.EncodeRecord([]byte("user_1"), []byte("Olamide"))
	if err := wal.Write(frame1); err != nil {
		panic(err)
	}
	fmt.Println("Appended key: user_1")

	// Write entry 2
	frame2, _ := storage.EncodeRecord([]byte("user_2"), []byte("Prodigy"))
	if err := wal.Write(frame2); err != nil {
		panic(err)
	}
	fmt.Println("Appended key: user_2")

	// Close the file to simulate system shutdown
	wal.Close()
	fmt.Println("Simulated system shutdown (WAL file closed). \n")


	// Re-open WAL to test crash recovery pipeline
	fmt.Println("--- Recovering State on Database Reboot ---")
	recoveredWAL, err := storage.OpenWAL(walPath)
	if err != nil {
		panic(err)
	}
	defer recoveredWAL.Close()

	records, err := recoveredWAL.ReadAll()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully recovered %d records from WAL file on disk:\n", len(records))
	for i, rec := range records {
		fmt.Printf(" Record %d -> Key: %s, Value: %s, Timestamp: %d\n", i+1, string(rec.Key), string(rec.Value), rec.Timestamp)
	}

}