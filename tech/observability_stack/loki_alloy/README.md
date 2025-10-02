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
