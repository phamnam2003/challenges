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
