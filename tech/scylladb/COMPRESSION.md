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

### Type Compression

#### LZ4 Compression

- Good compression ratio and very fast decompression speed, making it suitable for read-heavy workloads.
- It is the default compression algorithm in ScyllaDB and is a good choice for majority
- It only decompress blocks contains keys need to read, so it use less CPU resources when reading data.
- Conform with system have workloads latency sensitive, because CPU less affected and decompress high performance.

#### Snappy Compression

- Compression and decompression speed is very fast, but the compression ratio is lower than LZ4.
- It is suitable for workloads where speed is more critical than storage savings.
- When reading data, Snappy may use more CPU resources compared to LZ4, it only decompress blocks contains keys need to read.
- Conform with system have workloads fast read/write.

#### Zstd Compression

- Zstd (Zstandard) offers a very high compression ratio, often better than LZ4 and Snappy, especially at higher compression levels.
- It is **dictionary-based compression**, look like LZ4, but improve algorithms, suitable for workloads where storage savings are a priority, and some additional CPU usage is
- Balance between compress speed, decompression speed, and compression ratio.

### Compression Configuration

- `class`: Type compression algorithm. Options include:
  - `org.apache.cassandra.io.compress.LZ4Compressor` (default value - good for majority)
  - `org.apache.cassandra.io.compress.SnappyCompressor`
  - `org.apache.cassandra.io.compress.DeflateCompressor`
  - `org.apache.cassandra.io.compress.ZstdCompressor`
- `chunk_length_in_kb`: Size of each compressed block in KB. Default is `64` KB. Larger chunk sizes can improve compression ratio but may increase read latency for small queries.
- `crc_check_chance`: Probability of performing a CRC check on decompressed data. Default: 1.0
