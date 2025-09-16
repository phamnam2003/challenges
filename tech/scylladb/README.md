### Installation

1. Install ScyllaDB on your server or local machine. You can follow the official installation guide [here](https://docs.scylladb.com/manual/stable/).

- Recommendation: ScyllaDB works best on machines with a lot of RAM and CPU cores. A minimum of 8GB RAM and 4 CPU cores is recommended for development purposes.
- Configure aio-max-nr: ScyllaDB uses asynchronous I/O for better performance. You may need to increase the limit of aio-max-nr on your system. You can do this by adding the following line to `/etc/sysctl.conf`:

```bash
echo "fs.aio-max-nr = 1048576" | sudo tee -a /etc/sysctl.conf
sysctl -p /etc/sysctl.conf
```

2. Install ScyllaDB via Docker (optional):

- Install docker, docker-compose

```bash
docker-compose -f tech/scylladb/docker-compose.yml up -d --build
```

- Options:
  - ScyllaDB version: Change the version in the `docker-compose.yml` file if you want to use a different version.
  - Ports: You can change the ports in the `docker-compose.yml` file if they are already in use on your machine.
  - Should you want to persist data, you can uncomment the volume section in the `docker-compose.yml` file.

```yml
  scylla_node1:
    image: scylladb/scylla:2025.3 # Image and tag version
    container_name: scylla_node1 # container name when run local machine
    command: --smp 2 --memory 2G --overprovisioned 1
    ports:
      - "9042:9042" # CQL
      - "7000:7000" # inter-node communication
      - "7001:7001" # TLS inter-node
      - "7199:7199" # JMX
    networks:
      scylla-cluster:
        ipv4_address: 172.28.1.2

  scylla_node2:
    image: scylladb/scylla:2025.3
    container_name: scylla_node2
    command: --smp 2 --memory 2G --overprovisioned 1 --seeds=scylla_node1
    networks:
      scylla-cluster:
        ipv4_address: 172.28.1.3
```

- Command options:
  - `--smp`: Number of CPU cores to use.
  - `--memory`: Amount of RAM to use.
  - `--overprovisioned`: Set to 1 if you are running ScyllaDB on a machine with less than 8GB RAM.
  - `seeds`: IP address of the seed node (only for nodes other than the first one).
  - `--developer-mode`: which relaxes checks for things like XFS and enables Scylla to run on unsupported configurations (which usually results in suboptimal performance). Default value is true. Should turn off in production.

### Keyspace and Table Creation

1. Keyspace Creation:
Command to create keyspace:

```bash
CREATE KEYSPACE IF NOT EXISTS <KEYSPACE_NAME> WITH replication = {'class': '<CLASS>', '<DATA_CENTER>': <REPLICATION_FACTOR_NUMS>};
```

- KEYSPACE_NAME: A keyspace is a namespace that defines data replication on nodes. It is the outermost container for data in ScyllaDB.
- CLASS: The replication strategy to use. Common options are 'SimpleStrategy' for single data center deployments and 'NetworkTopologyStrategy' for multi-data center deployments. I recommend using 'NetworkTopologyStrategy' for all of times, even you have only 1 data center.

2. Table Creation:
Command to create table:

```bash
CREATE TABLE catalog.mutant_data (
    <TABLE_COLUMNS> <DATA_TYPE>,
    PRIMARY KEY ((first_name, last_name))
) WITH bloom_filter_fp_chance = 0.01
    AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
    AND comment = ''
    AND compaction = {'class': 'IncrementalCompactionStrategy'}
    AND compression = {'sstable_compression': 'org.apache.cassandra.io.compress.LZ4Compressor'}
    AND crc_check_chance = 1
    AND default_time_to_live = 0
    AND gc_grace_seconds = 864000
    AND max_index_interval = 2048
    AND memtable_flush_period_in_ms = 0
    AND min_index_interval = 128
    AND speculative_retry = '99.0PERCENTILE'
    AND tombstone_gc = {'mode': 'repair', 'propagation_delay_in_seconds': '3600'};
```

- TABLE_OPTIONS: opts is really important for performance tuning. Some common options include:
  - `compaction`: Determines how data is compacted on disk. Options include 'SizeTieredCompactionStrategy' (default value - balance), 'LeveledCompactionStrategy' (sys strong read), and 'TimeWindowCompactionStrategy' (time series data).
  - `compression`: Specifies the compression algorithm to use for the table. Options include 'LZ4Compressor' (default value - good for majority), 'SnappyCompressor', and 'DeflateCompressor'.
  - `bloom_filter_fp_chance`: Sets the false positive chance for the Bloom filter. Lower values reduce false positives but increase memory usage.
  - `caching`: Configures caching options for the table. Options include 'ALL', 'KEYS_ONLY', and 'ROWS_ONLY'.
  - `comment`: A comment or description for the table.
  - `crc_check_chance`: Sets the probability of performing a CRC check on data read from disk. Default is 1 (always check).
  - `default_time_to_live`: Sets the default time-to-live (TTL) for data
  - `gc_grace_seconds`: Specifies the grace period for garbage collection of tombstones (deleted data). Default is 864000 seconds (10 days).
  - `max_index_interval` and `min_index_interval`: Control the size of the index. Options include values between 1 and 32768. Default values are 2048 and 128, respectively.
  - `memtable_flush_period_in_ms`: Sets the interval for flushing memtables to disk. Default is 0 (disabled).
  - `speculative_retry`: Configures speculative retry behavior for read operations. Options include 'NONE', 'ALWAYS', and percentile-based values like '99.0PERCENTILE'.
  - `tombstone_gc`: Configures tombstone garbage collection settings. Options include 'mode' (e.g., 'repair', 'disabled') and 'propagation_delay_in_seconds'.

### Query

- "Use prepared statements for queries that are executed multiple times in your application." - `ScyllaDB`. After use prepared statements, you can execute the query with the provided parameters, and release prepared statements.

### Schema

- Use appropriate data types for your columns. ScyllaDB supports a variety of data types, including text, int, float, boolean, and more. Choose the data type that best fits your data to optimize storage and performance.
- When use `gocqlx`, you can make table objects with struct tags.

```go
 // create table metadata
 mutantMetadata := table.Metadata{
  Name:    "mutant_data",
  Columns: []string{"first_name", "last_name", "address", "picture_location"},
  PartKey: []string{"first_name", "last_name"},
  SortKey: []string{},
 }
 mutantTable := table.New(mutantMetadata)
```

- You can automatically create tables with `schemagen`, install tool:

```bash
go install github.com/scylladb/gocqlx/v2/cmd/schemagen@latest
```

- Usage:

```bash
schemagen [flags]

Flags:
  -cluster string
     a comma-separated list of host:port tuples (default "127.0.0.1")
  -keyspace string
     keyspace to inspect (required)
  -output string
     the name of the folder to output to (default "models")
  -pkgname string
     the name you wish to assign to your generated package (default "models") 
```

- Running the following command for `catalog` keyspace:

```bash
schemagen -cluster="127.0.0.1:9042" -keyspace="catalog" -output="./tech/scylladb/models" -pkgname="models"
```

- Generates Go structs for each table in the specified keyspace and saves them in the specified output directory.

```go
// Code generated by "gocqlx/cmd/schemagen"; DO NOT EDIT.

package models

import (
 "github.com/scylladb/gocqlx/v2/table"
)

// Table models.
var (
 MutantData = table.New(table.Metadata{
  Name: "mutant_data",
  Columns: []string{
   "address",
   "first_name",
   "last_name",
   "picture_location",
  },
  PartKey: []string{
   "first_name",
   "last_name",
  },
  SortKey: []string{},
 })
)

type MutantDataStruct struct {
 Address         string
 FirstName       string
 LastName        string
 PictureLocation string
}
```

### Backup and Restore

1. Backup:

- Run cmd below to backup Schema.

```bash
cqlsh -e "DESCRIBE KEYSPACE my_keyspace" > my_keyspace_schema.cql
```

- With backup data, you can use `nodetool snapshot` to create a snapshot of your data.

```bash
nodetool snapshot my_keyspace -t backup_2025_09_15
```

2. Restore:

- Run cmd to restore Schema.

```bash
cqlsh -f my_keyspace_schema.cql
```

- Restore data from snapshot: Copy the snapshot files from the backup location to the appropriate data directory on your ScyllaDB nodes. After copying, run `nodetool refresh` to make ScyllaDB recognize the new data.

```bash
cp /var/lib/scylla/data/my_keyspace/my_table-<table_uuid>/snapshots/backup_2025_09_15/* \
   /var/lib/scylla/data/my_keyspace/my_table-<table_uuid>/

nodetool refresh my_keyspace my_table
```
