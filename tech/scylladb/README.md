### Installation

1. Install ScyllaDB on your server or local machine. You can follow the official installation guide [here](https://docs.scylladb.com/manual/stable/).

- Recommendation: ScyllaDB works best on machines with a lot of RAM and CPU cores. A minimum of 8GB RAM and 4 CPU cores is recommended for development purposes.

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

### Query

- "Use prepared statements for queries that are executed multiple times in your application." - `ScyllaDB`. After use prepared statements, you can execute the query with the provided parameters, and release prepared statements.
