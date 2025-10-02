# LOKI AND ALLOY

- `Loki` is a log aggregation system designed to store and query logs from various sources. It is optimized for efficiency and scalability, making it suitable for large-scale environments.
- `Alloy` is a managed service that provides a simplified way to deploy and manage Loki instances in the cloud. It offers features like automatic scaling, backups, and monitoring.

## Quick start

- Run shell to get template configuration run `Loki` and `Alloy` in `Docker` container.

```bash
wget https://raw.githubusercontent.com/grafana/loki/main/examples/getting-started/loki-config.yaml -O loki-config.yaml
wget https://raw.githubusercontent.com/grafana/loki/main/examples/getting-started/alloy-local-config.yaml -O alloy-local-config.yaml
wget https://raw.githubusercontent.com/grafana/loki/main/examples/getting-started/docker-compose.yaml -O docker-compose.yaml
```

```bash
# Run docker-compose to up template service
docker-compose up -d
```

## View your logs in Grafana

- [Docs](https://grafana.com/docs/loki/latest/get-started/quick-start/quick-start/#view-your-logs-in-grafana)

## Loki data source in Grafana

- In this example, the Loki data source is already configured in Grafana. This can be seen within the `docker-compose.yaml` file:

```yaml
grafana:
  image: grafana/grafana:latest
  environment:
    - GF_PATHS_PROVISIONING=/etc/grafana/provisioning
    - GF_AUTH_ANONYMOUS_ENABLED=true
    - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
  depends_on:
    - gateway
  entrypoint:
    - sh
    - -euc
    - |
      mkdir -p /etc/grafana/provisioning/datasources
      cat <<EOF > /etc/grafana/provisioning/datasources/ds.yaml
      apiVersion: 1
      datasources:
        - name: Loki
          type: loki
          access: proxy
          url: http://gateway:3100
          jsonData:
            httpHeaderName1: "X-Scope-OrgID"
          secureJsonData:
            httpHeaderValue1: "tenant1"
      EOF
      /run.sh
```

## LOGQL

- `LOGQL` is a query language used in Loki to filter and analyze log data. It allows users to extract specific information from logs, perform aggregations, and create visualizations.
- It makes it easy to search and analyze large volumes of log data efficiently, it can be format logs, extract fields, and perform statistical operations.

```logql
{service_name="my_service"} |= "error" | logfmt | duration > 5s
{service_name="my_service"} |~ "timeout|failed" | json | status="500"
{service_name="my_service"} | json | status=500 | line_format "{{.timestamp}} - {{.message}}"
```

## Alloy

- `Alloy` is a managed service that provides a simplified way to deploy and manage Loki instances in the cloud. It offers features like automatic *scaling*, *backups*, and *monitoring*.
- `Alloy` is a *flexible*, *high performance*, *vendor-neutral distribution* of the `OpenTelemetry Collector`. It’s *fully compatible* with the most popular *open source observability standards* such as `OpenTelemetry` and `Prometheus`.
- `Alloy` focuses on ease-of-use and the `ability` to adapt to the needs of power users.

### Migrate to Alloy

- [Migrate from Prometheus to Alloy](https://grafana.com/docs/alloy/latest/set-up/migrate/from-prometheus/#migrate-from-prometheus-to-grafana-alloy)
  - Migrate your existing Prometheus setup to Grafana Alloy with minimal downtime and effort.
  - Components used in this topic:
    - [prometheus.scrape](https://grafana.com/docs/alloy/latest/reference/components/prometheus/prometheus.scrape/)
    - [prometheus.remote_write](https://grafana.com/docs/alloy/latest/reference/components/prometheus/prometheus.remote_write/)
  - [Convert a Prometheus configuration](https://grafana.com/docs/alloy/latest/set-up/migrate/from-prometheus/#migrate-from-prometheus-to-grafana-alloy)

  ```bash
  alloy convert --source-format=prometheus --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
  ```

Example:

```yaml
global:
  scrape_timeout:    45s

scrape_configs:
  - job_name: "prometheus"
    static_configs:
      - targets: ["localhost:12345"]

remote_write:
  - name: "grafana-cloud"
    url: "https://prometheus-us-central1.grafana.net/api/prom/push"
    basic_auth:
      username: <USERNAME>
      password: <PASSWORD>
```

```bash
alloy convert --source-format=prometheus --output=<OUTPUT_CONFIG_PATH> <INPUT_CONFIG_PATH>
```

```alloy
prometheus.scrape "prometheus" {
  targets = [{
    __address__ = "localhost:12345",
  }]
  forward_to     = [prometheus.remote_write.default.receiver]
  job_name       = "prometheus"
  scrape_timeout = "45s"
}

prometheus.remote_write "default" {
  endpoint {
    name = "grafana-cloud"
    url  = "https://prometheus-us-central1.grafana.net/api/prom/push"

    basic_auth {
      username = "USERNAME"
      password = "PASSWORD"
    }

    queue_config {
      capacity             = 2500
      max_shards           = 200
      max_samples_per_send = 500
    }

    metadata_config {
      max_samples_per_send = 500
    }
  }
}
```
