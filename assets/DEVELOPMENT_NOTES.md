# Development Notes

## Architectural Decisions
- **BadgerDB** was chosen for its high performance and efficient handling of key-value data. It replaced BoltDB after profiling revealed bottlenecks during write operations.
- **Bloom Filters** were considered for faster deduplication but excluded due to the potential for false positives, as the task required 100% accuracy in identifying unique IP addresses.

## Optimization and Profiling
To identify performance bottlenecks, we used Go's built-in `pprof` tool. The profiling revealed:
- A high number of `cgocalls` during database writes.
- Suboptimal memory usage when processing large files.

### Fixes Implemented:
1. Switched to **batch writes** to reduce the overhead of frequent database transactions.
2. Added memory usage logging to monitor and optimize performance during runtime.

## Tools and Libraries
- **Go (v1.23.4)**: Primary programming language.
- **BadgerDB**: Key-value storage engine.
- **xxhash**: High-speed hashing for deduplication.

## Performance
- Successfully processed a 10 million line file in ~5 minutes using 24 workers.
- Memory usage stabilized around 1.5 GB during peak processing.

## Future Improvements
- Explore distributed processing for even larger datasets.
- Add support for real-time monitoring and visualization of processing progress.
