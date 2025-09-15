## Compression in ScyllaDB

### What is Compression?

- When a ***Memtable*** flushed to disk, make SSTable file. The data in SSTable is stored in a compressed format to save disk space and improve I/O performance. Compression reduces the amount of data that needs to be read from disk during queries, which can lead to faster read times and lower latency.
- ScyllaDB supports several compression algorithms, including `LZ4`, `Snappy`, `Zstd`, and `Deflate`. Each algorithm has its own trade-offs in terms of compression ratio, speed, and CPU usage. The choice of compression algorithm can impact the performance of both read and write operations.
- Data in **SSTable** save called blocks. Each block is compressed separately, allowing for efficient access to specific parts of the data without needing to decompress the entire file. This block-level compression helps optimize read performance, especially for queries that access a small subset of the data.
- Advantages of using compression in ScyllaDB include:
  - Reduced disk space usage, allowing for more data to be stored on the same hardware.
  - Improved read performance due to reduced I/O operations.
  - Potentially lower network bandwidth usage when replicating data between nodes.
  - Non reduces **read amplification** cause SSTable overlap.
