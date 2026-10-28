package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kv-lite/pkg/storage"
)

const dbDir = "./data"

func main() {
	// Setup graceful shutdown (Intercept Unix SIGINT / CTRL+C signals)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Initialize our unified db engine
	db, err := storage.OpenDB(dbDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}

	// Background goroutine to catch shutdown gracefully
	go func() {
		<-sigChan
		fmt.Println("\n[CLI] Received shutdown signal. Closing database gracefully...")
		db.Close()
		os.Exit(0)
	}()

	// Parse CLI arguments (kv-lite set key value)
	args := os.Args[1:]
	if len(args) == 0 {
		printUsage()
		return
	}

	command := args[0]

	switch command {
	case "set":
		if len(args) < 3 {
			fmt.Println("Usage: kv-lite set <key> <value>")
			os.Exit(1)
		}
		key := []byte(args[1])
		val := []byte(args[2])

		start := time.Now()
		err := db.Set(key, val)
		if err != nil {
			fmt.Printf("Failed to set key: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("OK (written in %v)\n", time.Since(start))

	case "get":
		if len(args) < 2 {
			fmt.Println("Usage: kv-lite get <key>")
			os.Exit(1)
		}
		key := []byte(args[1])

		start := time.Now()
		val, exists := db.Get(key)
		if !exists {
			fmt.Println("(nil)")
			return
		}
		fmt.Printf("%s (read in %v)\n", string(val), time.Since(start))

	default:
		fmt.Printf("Unknown command: '%s'\n", command)
		printUsage()
	}

	// Ensure DB closes cleanly on normal exit
	db.Close()

}

func printUsage() {
	fmt.Println("kv-lite - A high performace embedded key-value storage engine")
	fmt.Println("Usage")
	fmt.Println(" go run cmd/cli/main.go set <key> <value>")
	fmt.Println(" go run cmd/cli/main.go get <key>")
}