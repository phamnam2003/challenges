## Compaction in ScyllaDB

### What is Compaction?

- `Compaction` is the process of merging multiple small `SSTables` into larger ones in ScyllaDB. It `reduces the number` of data files, eliminates duplicates or deleted rows (`tombstones`). This mechanism is crucial for optimizing read performance and managing disk usage.
- The goal of compaction is to keep data well-organized and avoid reading from too many SSTables at once. It also reclaims disk space occupied by outdated or deleted data. As a result, the system maintains stable performance even as data volume grows.
- In ScyllaDB, compaction ensures fast and predictable reads regardless of write volume. Depending on the workload type (`OLTP`, `time-series`, `logging`), users select the appropriate strategy such as `STCS`, `LCS`, `TWCS`, or `ICS`. It is one of the core components enabling Scylla’s high performance.

### Type of Compaction

#### Size-Tiered Compaction Strategy (STCS)

- `STCS` is the original default compaction strategy in Cassandra and Scylla. It `merges` SSTables of similar sizes into a larger one. This approach is simple and efficient for write-heavy workloads.
- The goal of STCS is to `reduce` the number of small SSTables, improving `write efficiency`. However, because data remains spread across many SSTables, *reads can be slower* and *unpredictable*. It is suitable when *read load is light or latency is not critical*.
- In ScyllaDB, STCS is often used for bulk data loading or workloads dominated by writes. It is simple, requires little configuration, and its behavior is easy to predict. However, for OLTP systems where stable read latency is required, STCS is usually not the best choice.

#### Leveled Compaction Strategy (LCS)

- LCS organizes SSTables into multiple “levels” of fixed size, ensuring data is evenly distributed. Each level contains SSTables without overlapping data, providing fast and predictable reads. This strategy is commonly used in OLTP systems requiring low latency.
- The goal of LCS is to optimize read performance by minimizing the number of SSTables to scan. As a result, read latency is much more stable compared to STCS. However, it comes with higher storage overhead and significant write amplification.
- In ScyllaDB, LCS is suitable for read-heavy workloads that require predictable latency, such as OLTP, e-commerce, or financial applications. It ensures that range queries or point lookups remain fast. However, under heavy write workloads, LCS can put pressure on the system.

#### Time-Window Compaction Strategy (TWCS)

- TWCS groups SSTables by “time windows,” such as hourly or daily intervals. SSTables within the same window are compacted together. This is especially useful for time-series or log data.
- The goal of TWCS is to optimize for data with TTL or data typically accessed by recent time ranges. It enables efficient cleanup of old data without affecting new data. However, workloads requiring cross-window queries may experience higher latency.
- In ScyllaDB, TWCS is recommended for time-series, IoT, logging, metrics, or any system with time-based data and TTL. It reduces write amplification and keeps disk usage low. It is not the best choice for OLTP workloads.

#### Incremental Compaction Strategy (ICS)

- ICS is the default compaction strategy in ScyllaDB (since version 4.0), developed as an improvement over LCS. Instead of compacting an entire large level, ICS incrementally merges new SSTables with existing ones. This balances write efficiency with stable read performance.
- The goal of ICS is to maintain the stable read latency of LCS while reducing write amplification and disk overhead. It ensures sustainable performance for workloads with both heavy writes and reads. ICS typically performs well without much parameter tuning.
- In ScyllaDB, ICS fits most OLTP or mixed workloads. It is the default choice because it optimizes storage cost while keeping read/write performance stable. This makes it the safe option if you are unsure which compaction strategy to choose.

### Compaction Configuration

1. STCS (Size-Tiered Compaction Strategy)

- `min_threshold`: minimum number of SSTables to trigger compaction (default: 4).
- `max_threshold`: maximum number of SSTables to compact at once (default: 32).
- `bucket_low`: lower bound for size buckets (default: 0.5).
- `bucket_high`: upper bound for size buckets (default: 1.5).
- `max_sstable_size_in_mb`: maximum size of an SSTable to be considered for compaction.

2. LCS (Leveled Compaction Strategy)

- `sstable_size_in_mb`: target size for SSTables in each level (default: 160 MB).
- `fanout_size`: number of SSTables per level before compaction is triggered (default: 10), capacity of each level is `fanout_size`^level.
- `min_threshold`: minimum number of SSTables to trigger compaction (default: 4).
- `max_threshold`: maximum number of SSTables to compact at once (default: 32).
- `tombstones_compaction_interval`: interval in seconds to trigger compaction for SSTables with tombstones (default: 86400 seconds).
- `tombstone_threshold`: fraction of tombstones in an SSTable to trigger compaction (default: 0.2 - 20%).
