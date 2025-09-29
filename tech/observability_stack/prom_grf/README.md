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

# Alert Manager

## What is Alertmanager?

- `Alertmanager` handles alerts *sent by client applications* such as the Prometheus server. It takes care of `deduplicating`, `grouping`, and `routing` them to the correct receiver integration such as *email*, *PagerDuty*, or *OpsGenie*. It also takes care of *silencing and inhibition of alerts*.
- The main steps to setting up alerting and notifications are:
  - Setup and [configure](https://prometheus.io/docs/alerting/latest/configuration/) the Alertmanager.
  - [Configure Prometheus](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#alertmanager_config) to talk to the Alertmanager
  - Create [alerting rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/) in Prometheus

## Configuration

### Alerting Rules

- Alerting rules allow you to define alert conditions based on Prometheus expression language expressions and to send notifications about firing alerts to an external service. Whenever the alert expression results in one or more vector elements at a given point in time, the alert counts as active for these elements' label sets.
- Defining alerting rules:

```yaml
groups:
- name: example
  labels:
    team: myteam
  rules:
  - alert: HighRequestLatency
    expr: job:request_latency_seconds:mean5m{job="myjob"} > 0.5
    for: 10m
    keep_firing_for: 5m
    labels:
      severity: page
    annotations:
      summary: High request latency
```

- Templating:

```yaml
groups:
- name: example
  rules:

  # Alert for any instance that is unreachable for >5 minutes.
  - alert: InstanceDown
    expr: up == 0
    for: 5m
    labels:
      severity: page
    annotations:
      summary: "Instance {{ $labels.instance }} down"
      description: "{{ $labels.instance }} of job {{ $labels.job }} has been down for more than 5 minutes."

  # Alert for any instance that has a median request latency >1s.
  - alert: APIHighRequestLatency
    expr: api_http_request_latencies_second{quantile="0.5"} > 1
    for: 10m
    annotations:
      summary: "High request latency on {{ $labels.instance }}"
      description: "{{ $labels.instance }} has a median request latency above 1s (current value: {{ $value }}s)"
```

### Alert Configuration

- [Alertmanager](https://github.com/prometheus/alertmanager) is configured via command-line flags and a configuration file. While the command-line flags configure immutable system parameters, the configuration file defines inhibition rules, notification routing and notification receivers.
- [File Layout and global settings](https://prometheus.io/docs/alerting/latest/configuration/#file-layout-and-global-settings)

## Monitoring Services

- Node Exporter: provides hardware and OS metrics exposed by *\*NIX kernels*, `e.g. CPU`, `memory`, `disk`, `network`, etc.
- Nginx Exporter: exposes `Nginx` metrics in a format that Prometheus can scrape.
- Apache Exporter: exposes `Apache HTTP Server` metrics in a format that Prometheus can scrape.
- MongoDB Exporter: exposes `MongoDB` metrics in a format that Prometheus can scrape.
- Redis Exporter: exposes `Redis` metrics in a format that Prometheus can scrape.
- Postgres Exporter: exposes `PostgreSQL` metrics in a format that Prometheus can scrape.
- MySQL Exporter: exposes `MySQL` metrics in a format that Prometheus can scrape.

# Logging with Loki and Promtail

- Normally, logs are stored in a centralized logging system like `Elasticsearch`, `Splunk`, or `Graylog`. However, these systems can be complex to operate and expensive to scale. `Grafana Loki` is a new approach to log aggregation that is designed to be cost-effective and easy to operate.
- `Grafana Loki` and `Grafana Promtail` are two components of the Grafana Loki stack, which is a set of open-source tools for *log aggregation and analysis*.

## Loki

### Introduction

- `Grafana Loki` is a set of *open source components* that can be composed into a fully featured *logging stack*. A small index and *highly compressed chunks simplifies* the operation and *significantly lowers* the cost of Loki.
- Unlike other logging systems, Loki is built around the idea of only indexing `metadata` about your logs’ labels (just like Prometheus labels). Log data itself is then *compressed* and *stored in chunks* in object stores such as Amazon Simple Storage Service (S3) or Google Cloud Storage (GCS), or even locally on the filesystem.
- Loki is a *horizontally scalable*, *highly available*, *multi-tenant log aggregation* system inspired by Prometheus. It’s designed to be `very cost-effective` and `easy to operate`. It *doesn’t index* the contents of the logs, but *rather a set of labels* for each log stream.

### Install and Usage

- You can install `Loki` via binary package or via `Docker`. You can follow the instructions in the [Loki install manually](https://grafana.com/docs/loki/latest/setup/install/local/#install-manually).

```bash
# For Linux Arch
sudo pacman -S loki
# For Debian/Ubuntu
sudo apt-get install loki
# For MacOS
brew install loki
```

- If you use `Docker`, you can run Loki with the following `docker-compose.yml` or [Loki install by Docker](https://grafana.com/docs/loki/latest/setup/install/docker/#install-loki-with-docker-or-docker-compose)
- Alerting in Loki: The recommended alert rule type in Grafana Alerting. These alert rules can query a wide range of backend data sources, including multiple data sources in a single alert rule. They support expression-based transformations, advanced alert conditions, images in notifications, handling of error and no data states, and more. [Docs](https://grafana.com/docs/loki/latest/alert/#loki-alerting-and-recording-rules)

## Promtail

### Introduction

- `Promtail` is an `agent` which ships the contents of local logs to a Loki instance or Grafana Cloud. It is usually deployed to every machine that has applications needed to be monitored. `Promtail` is `responsible` for *gathering logs* and *sending them* to `Loki`. It can also `tail` log files and `push` them to `Loki`.
[!Note]
Promtail is now deprecated and will enter into Long-Term Support (LTS) beginning Feb. 13, 2025. This means that Promtail will no longer receive any new feature updates, but it will receive critical bug fixes and security fixes. Commercial support will end after the LTS phase, which we anticipate will extend for about 12 months until February 28, 2026. End-of-Life (EOL) phase for Promtail will begin once LTS ends. Promtail is expected to reach EOL on March 2, 2026, afterwards no future support or updates will be provided. All future feature development will occur in Grafana Alloy. If you are currently using Promtail, you should plan your migration to Alloy. The Alloy migration documentation includes a migration tool for converting your Promtail configuration to an [Alloy configuration](https://grafana.com/docs/alloy/latest/introduction/) with a single command.

### Install and Usage

- You can install `Promtail` via binary package or via `Docker`. You can follow the instructions in the [Promtail install manually](https://grafana.com/docs/loki/latest/clients/promtail/installation/#install-manually).

```bash
# For Linux Arch
sudo pacman -S loki
# For Debian/Ubuntu
sudo apt-get install loki
# For MacOS
brew install loki
```

- Install by `Docker` with the following `docker-compose.yml` or [Promtail install by Docker](https://grafana.com/docs/loki/latest/clients/promtail/installation/#install-with-docker-or-docker-compose)
- [Configuration](https://grafana.com/docs/loki/latest/send-data/promtail/configuration/) Promtail is configured via a YAML file. The configuration file is divided into several sections, each of which is described in detail below.
