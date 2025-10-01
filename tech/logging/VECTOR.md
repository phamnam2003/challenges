# Log Agent

- `Log Agent`: A lightweight software that collects and forwards log data from various sources to a centralized logging system.
- It is typically installed on servers, containers, or applications to gather logs in real-time.

## Vector Log Agent

- Install [Vector](https://vector.dev/docs/setup/installation/), enabled and start service.
- `Vector` have 3 components: `sources`, `transform` and `sinks`. Configuration file located at `/etc/vector/vector.yaml`.

```yaml
                                    __   __  __
#                                      / / / /
#                                      V / / /
#                                      _/  /
#
#                                    V E C T O R
#                                   Configuration
#
# ------------------------------------------------------------------------------
# Website: https://vector.dev
# Docs: https://vector.dev/docs
# Chat: https://chat.vector.dev
# ------------------------------------------------------------------------------

# Change this to use a non-default directory for Vector data storage:
# data_dir: "/var/lib/vector"

# Random Syslog-formatted logs
sources:
  dummy_logs:
    type: "demo_logs"
    format: "syslog"
    interval: 1

# Parse Syslog logs
# See the Vector Remap Language reference for more info: https://vrl.dev
transforms:
  parse_logs:
    type: "remap"
    inputs: ["dummy_logs"]
    source: |
      . = parse_syslog!(string!(.message))

# Print parsed logs to stdout
sinks:
  print:
    type: "console"
    inputs: ["parse_logs"]
    encoding:
      codec: "json"
      json:
        pretty: true

# Vector's GraphQL API (disabled by default)
# Uncomment to try it out with the `vector top` command or
# in your browser at http://localhost:8686
# api:
#   enabled: true
#   address: "127.0.0.1:8686"
```

- [Vector Configuration Reference](https://vector.dev/docs/reference/configuration/)
- `sources` in `/etc/vector/vector.yaml`, you can setting to collect logs from different sources:
  - `file`: Collect logs from files, e.g., `/var/log/*.log`.
  - `syslog`: Collect logs from syslog servers.
  - `journald`: Collect logs from systemd journal.
  - `docker_logs`: Collect logs from Docker containers.
  - `kubernetes_logs`: Collect logs from Kubernetes pods.
  ...
  - [More information about sources](https://vector.dev/docs/reference/configuration/sources/)
- `transforms` in `/etc/vector/vector.yaml`, you can setting to process and transform logs:
  - `remap`: Use Vector Remap Language (VRL) to manipulate log data.
  - `filter`: filter events based on a set of conditions.
  - `log_to_metric`: Convert log data into metrics.
  - `metric_to_log`: Convert metrics back into log data.
  ...
  - [More information about transforms](https://vector.dev/docs/reference/configuration/transforms/)
- `sinks` in `/etc/vector/vector.yaml`, you can setting to send logs to different destinations:
  - `elasticsearch`: Send logs to Elasticsearch for storage and search.
  - `kafka`: Send logs to Apache Kafka for further processing.
  - `http`: Send logs to an HTTP endpoint.
  - `file`: Write logs to files.
  - `console`: Print logs to the console (useful for debugging).
  ...
  - [More information about sinks](https://vector.dev/docs/reference/configuration/sinks/)

- `Vector` is highly configurable and can be tailored to fit various logging needs. You can combine multiple sources, transforms, and sinks to create a comprehensive logging pipeline that suits your infrastructure. `Vector` can be combine with multiple `Log Agents` to collect logs from different servers and send them to a centralized logging system like `Elasticsearch`, `Loki`, `Splunk`, etc. That make `vector` become a powerful and flexible solution for log management in modern IT environments.

# Configuration Vector to collect logs from file and send to Elasticsearch

- Configure `Vector` to collect logs from a specific file and send them to `Elasticsearch` for storage and analysis. Below is an example configuration that demonstrates how to achieve this.
- [`/etc/vector/vector.yml`](./vector.yml), and [`explain`](./TUTORIAL_CONFIGURE_VECTOR.md) the configuration file.
