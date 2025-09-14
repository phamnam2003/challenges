## Compaction in ScyllaDB

1. **What is Compaction?**

- `Compaction` is the process of merging multiple small `SSTables` into larger ones in ScyllaDB. It `reduces the number` of data files, eliminates duplicates or deleted rows (`tombstones`). This mechanism is crucial for optimizing read performance and managing disk usage.
- The goal of compaction is to keep data well-organized and avoid reading from too many SSTables at once. It also reclaims disk space occupied by outdated or deleted data. As a result, the system maintains stable performance even as data volume grows.
- In ScyllaDB, compaction ensures fast and predictable reads regardless of write volume. Depending on the workload type (`OLTP`, `time-series`, `logging`), users select the appropriate strategy such as `STCS`, `LCS`, `TWCS`, or `ICS`. It is one of the core components enabling Scylla’s high performance.

