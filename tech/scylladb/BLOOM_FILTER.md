# Bloom Filter

## What is Bloom Filter

- A *Bloom filter* is a **probabilistic data structure** used to *quickly check* whether an element might exist in a set or definitely does not exist.
- Unlike traditional data structures, a Bloom filter does not store the actual elements; instead, it uses a bit array combined with multiple hash functions to mark the presence of elements.
- When an element is added, the hash functions determine specific positions in the bit array to set. When checking for an element, if all corresponding bits are set, the element might exist; if any bit is unset, the element definitely does not exist.
- *Bloom filters* are **highly memory-efficient** and allow for **fast lookups**, but they can *produce false positives*, meaning they may indicate an element exists when it actually does not.
- They are widely used in **databases**, **caching systems**, and **networking** to `reduce` unnecessary disk or network access. By filtering out non-existent elements early, Bloom filters help *improve* `performance` and `reduce` *I/O costs*.
- They are especially useful in systems like ScyllaDB and Cassandra to avoid reading SSTables that do not contain the requested key.
- Overall, *Bloom filters* offer a `fast`, `space-efficient`, and *probabilistic approach* to membership testing in large datasets.

## Configuration in ScyllaDB

- Configured at the table level using the `bloom_filter_fp_chance` option.
- This option sets the desired false positive probability for the Bloom filter.
- The value is a float between `0` and `1`, where lower values reduce false
- Values typically range from `0.01` - default value (1% false positive rate) to `0.1` (10% false positive rate).
- A lower false positive rate means the Bloom filter is more accurate but uses more memory.
