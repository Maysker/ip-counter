package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cespare/xxhash/v2"
	badger "github.com/dgraph-io/badger/v3"
)

const (
	blockSize      = 10 * 1024 * 1024 // 10 MB block size
	bufferSize     = 4 * 1024 * 1024  // 4 MB buffer for the scanner
	progressLog    = 1_000_000        // Log progress every 1 million lines
	maxInvalidLogs = 10               // Maximum invalid IPs to log
	batchSize      = 10_000           // Batch size for database writes
	dbPath         = "badger_db"      // Path to BadgerDB
	logFile        = "warnings.log"   // File for invalid IPs
	retryAttempts  = 3                // Retry attempts for file errors
)

var (
	invalidIPCount  int32 // Counter for invalid IPs
	invalidLogCount int32
	logWriter       *bufio.Writer
)

func main() {
	start := time.Now()
	defer func() {
		fmt.Printf("\nExecution time: %v\n", time.Since(start))
	}()

	// Clear screen for a cleaner view
	fmt.Print("\033[H\033[2J")
	fmt.Println("=== IP Address Processing Program ===")

	// Validate arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <file_path>")
		return
	}
	filePath := os.Args[1]
	fmt.Printf("\nProcessing file: %s\n", filePath)

	// Initialize log file
	logFile, err := os.Create(logFile)
	if err != nil {
		fmt.Printf("Error creating log file: %v\n", err)
		return
	}
	defer logFile.Close()
	logWriter = bufio.NewWriter(logFile)

	// Initialize BadgerDB
	opts := badger.DefaultOptions(dbPath).WithLogger(nil)
	db, err := badger.Open(opts)
	if err != nil {
		fmt.Printf("Error opening database: %v\n", err)
		return
	}
	defer db.Close()

	// Channels and synchronization
	linesChan := make(chan string, 10_000)
	batchChan := make(chan []uint64, 100)
	var wg sync.WaitGroup

	// Start monitoring memory usage
	go memoryMonitor()

	// Start database writer
	wg.Add(1)
	go databaseWriter(batchChan, db, &wg)

	// Start worker goroutines
	numWorkers := runtime.NumCPU()
	fmt.Printf("\nUsing %d workers...\n", numWorkers)
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(linesChan, batchChan, &wg)
	}

	// Read file
	totalLines := readFile(filePath, linesChan)

	// Wait for workers to finish
	close(linesChan)
	wg.Wait()
	close(batchChan)
	logWriter.Flush()

	// Print final results
	printResults(db, totalLines)
}

// Read file and send lines to channel
func readFile(filePath string, linesChan chan<- string) int {
	var totalLines int
	for attempt := 1; attempt <= retryAttempts; attempt++ {
		file, err := os.Open(filePath)
		if err != nil {
			fmt.Printf("Error opening file (attempt %d/%d): %v\n", attempt, retryAttempts, err)
			time.Sleep(2 * time.Second)
			continue
		}
		defer file.Close()

		reader := bufio.NewReaderSize(file, bufferSize)
		for {
			block := make([]byte, blockSize)
			n, err := reader.Read(block)
			if n > 0 {
				for _, line := range splitLines(block[:n]) {
					linesChan <- line
					totalLines++
					if totalLines%progressLog == 0 {
						fmt.Printf("Processed %d lines...\n", totalLines)
					}
				}
			}
			if err == io.EOF {
				return totalLines
			}
			if err != nil {
				fmt.Printf("Error reading file: %v\n", err)
				break
			}
		}
	}
	return totalLines
}

// Worker function
func worker(linesChan <-chan string, batchChan chan<- []uint64, wg *sync.WaitGroup) {
	defer wg.Done()
	var hashes []uint64
	for line := range linesChan {
		ip := net.ParseIP(line)
		if ip != nil {
			hash := xxhash.Sum64([]byte(ip.String()))
			hashes = append(hashes, hash)

			// Send batch to channel
			if len(hashes) >= batchSize {
				batchChan <- hashes
				hashes = nil
			}
		} else {
			atomic.AddInt32(&invalidIPCount, 1)
			if atomic.LoadInt32(&invalidLogCount) < maxInvalidLogs {
				logWriter.WriteString(fmt.Sprintf("Warning: invalid IP address: %s\n", line))
				atomic.AddInt32(&invalidLogCount, 1)
			}
		}
	}
	if len(hashes) > 0 {
		batchChan <- hashes
	}
}

// Database writer
func databaseWriter(batchChan <-chan []uint64, db *badger.DB, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range batchChan {
		err := db.Update(func(txn *badger.Txn) error {
			for _, hash := range batch {
				key := itob(hash)
				if err := txn.Set(key, []byte{}); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			fmt.Printf("Error writing to database: %v\n", err)
		}
	}
}

// Print final results
func printResults(db *badger.DB, totalLines int) {
	var uniqueCount int
	db.View(func(txn *badger.Txn) error {
		iter := txn.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			uniqueCount++
		}
		return nil
	})

	fmt.Printf("\nTotal lines processed: %d\n", totalLines)
	fmt.Printf("Number of unique valid IP addresses: %d\n", uniqueCount)
	fmt.Printf("Number of invalid IP addresses: %d\n", invalidIPCount)
}

// Split block into lines
func splitLines(block []byte) []string {
	var lines []string
	var currentLine []byte
	for _, b := range block {
		if b == '\n' {
			lines = append(lines, string(currentLine))
			currentLine = nil
		} else {
			currentLine = append(currentLine, b)
		}
	}
	if len(currentLine) > 0 {
		lines = append(lines, string(currentLine))
	}
	return lines
}

// Convert int to byte slice
func itob(v uint64) []byte {
	return []byte(fmt.Sprintf("%d", v))
}

// Monitor memory usage
func memoryMonitor() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		fmt.Printf("Memory usage: %v MB\n", m.Alloc/1024/1024)
	}
}
