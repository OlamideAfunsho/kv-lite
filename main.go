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

	// walPath := "wal.log"

	// // Cleanup any existing wal.log file from previous test runs
	// os.Remove(walPath)

	// // Open new WAL instance
	// wal, err := storage.OpenWAL(walPath)
	// if err != nil {
	// 	panic(err)
	// }
	// memTable := storage.NewMemTable()

	// fmt.Println("--- Writing & Setting Data ---")

	// // Write key 1
	// k1, v1 := []byte("theme"), []byte("dark")
	// frame1, _ := storage.EncodeRecord(k1, v1)
	// wal.Write(frame1)
	// memTable.Set(k1, v1)
	// fmt.Println("Set theme = dark")

	// // Write key 2
	// k2, v2 := []byte("version"), []byte("v1.0.0")
	// frame2, _ := storage.EncodeRecord(k2, v2)
	// wal.Write(frame2)
	// memTable.Set(k2, v2)
	// fmt.Println("Set version = v1.0.0")

	// // Close WAL to simulate shutdown
	// wal.Close()
	// fmt.Println("\nSimulated system shutdown...\n")

	// // Reboot: Recover WAL and load into a fresh MemTable
	// fmt.Println("--- Rebooting & Recovering State ---")
	// recoveredWAL, err := storage.OpenWAL(walPath)
	// if err != nil {
	// 	panic(err)
	// }
	// defer recoveredWAL.Close()

	// records, err := recoveredWAL.ReadAll()
	// if err != nil {
	// 	panic(err)
	// }

	// freshMemTable := storage.NewMemTable()
	// freshMemTable.LoadFromWAL(records)

	// // Query fresh MemTable
	// if val, ok := freshMemTable.Get([]byte("theme")); ok {
	// 	fmt.Printf("Get 'theme': %s\n", string(val))
	// }
	// if val, ok := freshMemTable.Get([]byte("version")); ok {
	// 	fmt.Printf("Get 'version': %s\n", string(val))
	// }



	// fmt.Println("--- Writing Records to WAL ---")

	// // Write entry 1
	// frame1, _ := storage.EncodeRecord([]byte("user_1"), []byte("Olamide"))
	// if err := wal.Write(frame1); err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Appended key: user_1")

	// // Write entry 2
	// frame2, _ := storage.EncodeRecord([]byte("user_2"), []byte("Prodigy"))
	// if err := wal.Write(frame2); err != nil {
	// 	panic(err)
	// }
	// fmt.Println("Appended key: user_2")

	// // Close the file to simulate system shutdown
	// wal.Close()
	// fmt.Println("Simulated system shutdown (WAL file closed). \n")


	// // Re-open WAL to test crash recovery pipeline
	// fmt.Println("--- Recovering State on Database Reboot ---")
	// recoveredWAL, err := storage.OpenWAL(walPath)
	// if err != nil {
	// 	panic(err)
	// }
	// defer recoveredWAL.Close()

	// records, err := recoveredWAL.ReadAll()
	// if err != nil {
	// 	panic(err)
	// }

	// fmt.Printf("Successfully recovered %d records from WAL file on disk:\n", len(records))
	// for i, rec := range records {
	// 	fmt.Printf(" Record %d -> Key: %s, Value: %s, Timestamp: %d\n", i+1, string(rec.Key), string(rec.Value), rec.Timestamp)
	// }

	sstablePath := "sstable_1.db"
	os.Remove(sstablePath)

	// Initialize a MemTable and insert unsorted keys
	memTable := storage.NewMemTable()
	memTable.Set([]byte("Olamide"), []byte("name"))
	memTable.Set([]byte("Male"), []byte("gender"))
	memTable.Set([]byte("21"), []byte("age"))

	fmt.Println("--- Flushing MemTable to SSTable ---")

	// Access internal map directly for flushing (or add an export method)
	// Write an helper, or pass data directly
	data := map[string][]byte{
		"Olamide": []byte("name"),
		"Male": []byte("gender"),
		"21": []byte("age"),
	}

	err := storage.WriteSSTable(sstablePath, data)
	if err != nil {
		panic(err)
	}

	// Verify file exists and has size
	fileInfo, err := os.Stat(sstablePath)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Successfully flushed SSTable to disk: %s\n", fileInfo.Name())
	fmt.Printf("SSTable file size: %d bytes\n", fileInfo.Size())

}