<p align="center">
  <img src="https://raw.githubusercontent.com/Maysker/ip-counter/refs/heads/master/assets/logo.png" alt="Logo" width="200">
</p>

## Overview

This application efficiently processes a large file of IP addresses, identifying unique valid IPs and logging invalid entries. Designed to handle extremely large datasets, the application demonstrates professional-grade performance optimization and scalability.

This project was developed as part of an assignment from a potential employer, but due to visa issues, the process was not continued. However, the task was completed, as it was much more interesting than typical tests or banal questions.

## Key Features

- **Multithreaded processing**: Utilizes all available CPU cores for maximum efficiency.
- **BadgerDB integration**: High-performance key-value database for storing unique IP hashes.
- **Error handling and logging**: Ensures robustness and reliability.
- **Memory usage tracking**: Provides insights during runtime.
- **Progress reporting**: Displays the current processing status in real-time.

## Tools and Technologies Used

- **Programming Language**: Go (Golang)
- **Database**: BadgerDB
- **Hashing**: `cespare/xxhash` for fast and efficient hashing.
- **Memory Management**: Built-in `sync` and `sync/atomic` packages for concurrency control.

## Installation and Setup

### Clone the repository:
```bash
git clone https://github.com/Maysker/ip-counter.git
cd ip-counter
```

## Install dependencies:

- go mod tidy
    
- Run the application:
    
- go run main.go <file_path>

## How It Works

File Reading:
- Reads the IP file in chunks to optimize memory usage.
- Splits chunks into individual IP lines.

IP Validation:

- Validates each IP using net.ParseIP.
- Logs invalid IPs in warnings.log.

Unique IP Tracking:

- Hashes valid IPs using xxhash.
- Stores unique hashes in BadgerDB.

Progress Reporting:

- Displays the number of processed lines and memory usage.

Error Handling:

- Retries file reading on errors.
- Logs critical issues for debugging.

## Example Output
```bash
=== IP Address Processing Program ===
Processing file: ip_addresses
Using 24 workers...
Progress: 1,000,000 lines processed...
Progress: 2,000,000 lines processed...
Memory usage: 150 MB
...
Number of unique valid IP addresses: 1,234,567
Number of invalid IP addresses: 12
Execution time: 15m30s
```

## Key Optimizations Implemented:

- Switched to BadgerDB for better performance with large datasets.
- Implemented batch writes to minimize database overhead.
- Added real-time memory usage tracking.

## Future Improvements

- Add configuration options for batch size, logging verbosity, and worker count.
- Implement a web-based UI for easier monitoring.
- Explore further optimizations with low-level I/O operations.

## Acknowledgments

Special thanks to the creators of:

- Golang
- BadgerDB
- cespare/xxhash
## 
For questions or feedback, please contact [Maysker](https://github.com/Maysker).
