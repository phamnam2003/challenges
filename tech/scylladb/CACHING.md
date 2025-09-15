# Caching in ScyllaDB

## What is cache in ScyllaDB?

- Check data exists in memory before going to disk. If the data is found in memory, it can be returned quickly without the need for a disk read. This can significantly improve read performance, especially for frequently accessed data.
- ScyllaDB uses two types of caches: the key cache and the row cache.
- Type cache of ScyllaDB: `key` and `row` cache.

## Configuration cache

- `keys`: values include:
  - `ALL`: Cache all keys (default value).
  - `NONE`: Do not cache any keys.
  - `ROWS_ONLY`: Cache only the rows, not the keys.
- `rows_per_partition`: number of rows to cache per partition. Values include:
  - `ALL`: Cache all rows in a partition.
  - `NONE`: Do not cache any rows.
  - A specific number: Cache up to that many rows per partition (e.g., `10` to cache the first 10 rows of each partition).
- `enabled`: Enable or disable the row cache. Values include:
  - `true`: Enable the row cache.
  - `false`: Disable the row cache (default value).
