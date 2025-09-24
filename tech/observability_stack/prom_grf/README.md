# Prometheus

## What is Prometheus?

- [Prometheus](https://prometheus.io/) is an `open-source` *systems monitoring* and *alerting toolkit* originally built at [SoundCloud](https://soundcloud.com/). It is designed for reliability and scalability, making it a popular choice for monitoring modern cloud-native applications and infrastructure. Prometheus joined the [Cloud Native Computing Foundation](https://www.cncf.io/) in 2016 as the second hosted project, after [Kubernetes](https://kubernetes.io/).
- Prometheus `collects` and `stores` it's *metrics* as *time series data*, i.e. metrics information is stored with the timestamp at which it was recorded, alongside optional key-value pairs called labels.

## Features

- a `multi-dimensional` [data model](https://prometheus.io/docs/concepts/data_model/) with time series data identified by metric name and `key\/value pairs`
- `PromQL`, a [flexible query language](https://prometheus.io/docs/prometheus/latest/querying/basics/) to leverage this dimensionality
- no reliance on distributed storage; single server nodes are autonomous
- time series collection happens via a pull model over HTTP
- [pushing time series](https://prometheus.io/docs/instrumenting/pushing/) is supported via an intermediary gateway
- targets are discovered via service discovery or static configuration
- multiple modes of graphing and dashboarding support

## What are metrics?

- `Metrics` are *numerical measurements* in layperson terms. The term time series refers to the recording of changes over time. What users want to measure differs from application to application. For a web server, it could be request times; for a database, it could be the number of active connections or active queries, and so on.
- `Metrics` play an important role in *understanding why your application is working* in a *certain way*. Let's assume you are running a web application and discover that it is slow. To learn what is happening with your application, you will need some information. For example, when the number of requests is high, the application may become slow. If you have the request count metric, you can determine the cause and increase the number of servers to handle the load.

## Components

- The Prometheus ecosystem consists of multiple components, many of which are optional:
  - the main [Prometheus server](https://github.com/prometheus/prometheus) which `scrapes` and `stores` time series data
  - [client libraries](https://prometheus.io/docs/instrumenting/clientlibs/) for instrumenting application code
  - a [push gateway](https://github.com/prometheus/pushgateway) for supporting short-lived jobs
  - special-purpose [exporters](https://prometheus.io/docs/instrumenting/exporters/) for services like `HAProxy`, `StatsD`, `Graphite`, etc.
  - an [alertmanager](https://github.com/prometheus/alertmanager) to handle alerts
  - various support tools
- *Most Prometheus components* are *written in `Go`*, making them easy to build and deploy as static binaries.

## Architecture

- This diagram illustrates the architecture of Prometheus and some of its ecosystem components:
![architecture diagram](../../../images/architecture.svg)

- Prometheus scrapes metrics from *instrumented jobs*, either *directly* or via an *intermediary push gateway* for short-lived jobs. It stores all scraped samples locally and runs rules over this data to either aggregate and record new time series from existing data or generate alerts. [Grafana](https://grafana.com/) or other API consumers can be used to visualize the collected data.

## Installation

### Prometheus Server and Grafana

- Install via binary package or via `Docker`

### Node Exporter

- Once the Prometheus server is up and running, you can start monitoring your system metrics by installing the [Node Exporter](https://prometheus.io/download/).
- If you use Linux, you can install it via package manager with package name `prometheus-node-exporter`.

```bash
sudo apt-get install prometheus-node-exporter
sudo pacman -S prometheus-node-exporter
```

- Start `systemctl` service

```bash
sudo systemctl start prometheus-node-exporter
sudo systemctl enable prometheus-node-exporter (optional)
```

- Add data source prometheus in grafana:
  - In sidebar, click the gear icon to open the `Connections` menu -> `Add Datasource` -> Choose `Prometheus` -> Enter URL to connect Prometheus.
- Default Port of `Node Exporter` is `9100`
- Configuration Grafana Dashboard for Node Exporter: [Node Exporter Full](https://grafana.com/grafana/dashboards/1860-node-exporter-full/): use Dashboard ID `1860` for *Node Exporter Full Version* or `14513` for *Linux Node Exporter*

## PromQL

- `PromQL` is a *powerful query language* used in Prometheus to query and manipulate time series data. It allows users to select and aggregate time series data in real-time, making it a crucial tool for *monitoring* and *alerting* in Prometheus.
- Is a *specialized query language* in Prometheus for working with time-series data. It enables users to *filter*, *aggregate*, and *compute metrics* to create visual `dashboards` or define `alerting rules`. PromQL transforms raw exporter data into clear and actionable insights, making system monitoring more effective.

## Rules

- Recording Rules: allow users to *precompute* frequently needed or *computationally* expensive expressions and save their result as a new set of time series. Example: instead of query `rate(node_cpu_seconds_total{mode="user"}[5m])` every time, you can create a recording rule to save the result as `instance:cpu_usage:rate5m`. This new time series can then be queried like any other time series.

```yaml
groups:
  - name: recording_rules
    interval: 30s
    rules:
      - record: instance:cpu_usage:rate5m
        expr: rate(node_cpu_seconds_total{mode="user"}[5m])
```

- Alerting Rules: allow users to define alert conditions based on `PromQL` expressions. When the condition is met, an alert is generated and sent to the `Alertmanager` for further processing.

```yaml
groups:
  - name: alerting_rules
    rules:
      - alert: HighCPULoad
        expr: avg(rate(node_cpu_seconds_total{mode="user"}[5m])) by (instance) > 0.8
        for: 2m
        labels:
          severity: warning
        annotations:
          summary: "CPU usage is high on instance {{ $labels.instance }}"
          description: "CPU usage has been above 80% for more than 2 minutes."
```

### Import rules into Prometheus

- Rules can be defined in separate YAML files and imported into the Prometheus configuration file using the `rule_files` directive, then import into field `rule_files` in `prometheus.yml`

```yaml
rule_files:
  - "rules/*.yml"
```
