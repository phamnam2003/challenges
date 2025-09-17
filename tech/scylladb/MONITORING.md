# Monitoring Scylla with Prometheus and Grafana

- Recommended setup
  - Prometheus
  - Grafana
  - [Scylla Monitoring Stack](https://github.com/scylladb/scylla-monitoring)

## Using Scylla Monitoring Stack

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

## Using Docker Compose

- Read more in the [Scylla Monitoring Stack documentation](https://monitoring.docs.scylladb.com/stable/install/docker-compose.html)

1. Setting Prometheus

- You can use ./prometheus-config.sh to generate the file, for example:

```bash
./prometheus-config.sh --compose
```

2. Setting Grafana Provisioning

- Grafana Data-Source file

```bash
./grafana-datasource.sh --compose
```

- Grafana Dashboard Load file

```bash
./generate-dashboards.sh -t -v 2020.1
```

3. Docker Compose file

```yml
services:
  alertmanager:
    container_name: aalert
    image: prom/alertmanager:v0.26.0
    ports:
    - 9093:9093
    volumes:
    - ./prometheus/rule_config.yml:/etc/alertmanager/config.yml
  grafana:
    container_name: agraf
    environment:
    - GF_PANELS_DISABLE_SANITIZE_HTML=true
    - GF_PATHS_PROVISIONING=/var/lib/grafana/provisioning
    - GF_PLUGINS_ALLOW_LOADING_UNSIGNED_PLUGINS=scylladb-scylla-datasource
    # This is where you set Grafana security
    - GF_AUTH_BASIC_ENABLED=false
    - GF_AUTH_ANONYMOUS_ENABLED=true
    - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
    - GF_SECURITY_ADMIN_PASSWORD=admin
    # To set your home dashboard uncomment the following line, set VERSION to be your current version
    #- GF_DASHBOARDS_DEFAULT_HOME_DASHBOARD_PATH=/var/lib/grafana/dashboards/ver_VERSION/scylla-overview.VERSION.json
    image: grafana/grafana:10.4.1
    ports:
    - 3000:3000
    user: 1000:1000
    volumes:
    - ./grafana/build:/var/lib/grafana/dashboards
    - ./grafana/plugins:/var/lib/grafana/plugins
    - ./grafana/provisioning:/var/lib/grafana/provisioning
    # Uncomment the following line for grafana persistency
    # - path/to/grafana/dir:/var/lib/grafana
  loki:
    command:
    - --config.file=/mnt/config/loki-config.yaml
    container_name: loki
    image: grafana/loki:2.9.5
    ports:
    - 3100:3100
    volumes:
    - ./loki/rules:/etc/loki/rules
    - ./loki/conf:/mnt/config
  promotheus:
    command:
    - --config.file=/etc/prometheus/prometheus.yml
    container_name: aprom
    image: prom/prometheus:v2.51.1
    ports:
    - 9090:9090
    volumes:
    - ./prometheus/build/prometheus.yml:/etc/prometheus/prometheus.yml
    - ./prometheus/prom_rules/:/etc/prometheus/prom_rules/
    # instead of the following three targets, you can place three files under one directory and mount that directory
    # If you do, uncomment the following line and delete the three lines afterwards
    #- /path/to/targets:/etc/scylla.d/prometheus/targets/
    - ./prometheus/scylla_servers.yml:/etc/scylla.d/prometheus/targets/scylla_servers.yml
    - ./prometheus/scylla_manager_servers.yml:/etc/scylla.d/prometheus/targets/scylla_manager_servers.yml
    - ./prometheus/scylla_servers.yml:/etc/scylla.d/prometheus/targets/node_exporter_servers.yml
    
    # Uncomment the following line for prometheus persistency 
    # - path/to/data/dir:/prometheus/data
  promtail:
    command:
    - --config.file=/etc/promtail/config.yml
    container_name: promtail
    image: grafana/promtail:2.7.3
    ports:
    - 1514:1514
    - 9080:9080
    volumes:
    - ./loki/promtail/promtail_config.compose.yml:/etc/promtail/config.yml
version: '3'
```

4. Run Docker Compose

```bash
docker-compose up -d
```
