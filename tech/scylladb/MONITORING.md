# Monitoring Scylla with Prometheus and Grafana

- Recommended setup
  - Prometheus
  - Grafana
  - [Scylla Monitoring Stack](https://github.com/scylladb/scylla-monitoring)

1. Clone SCylla Monitoring Stack repository

```bash
git clone https://github.com/scylladb/scylla-monitoring
cd scylla-monitoring
git checkout branch-4.11
```

2. Create file `prometheus/scylla_servers.yml`:

- Template for 3 nodes cluster, you can change the `targets` and `labels` as per your cluster configuration.

```yml
- targets:
  - localhost:9180
  labels:
    intance: "node1"

- targets:
  - localhost:9181
  labels:
    intance: "node2"

- targets:
  - localhost:9182
  labels:
    intance: "node3"
```

- Run Monitoring with command:

```bash
./start-all.sh -s prometheus/scylla_servers.yml -d prometheus_data # production
./start-all.sh -l -d prometheus_data # local host
```

Read more in the [Scylla Monitoring Stack documentation](https://monitoring.docs.scylladb.com/stable/install/monitoring-stack.html)
