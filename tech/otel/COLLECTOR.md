# Collector

- Vendor-agnostic way to receive, process and export telemetry data.
- The OpenTelemetry Collector receives [traces](https://opentelemetry.io/docs/concepts/signals/traces/), [metrics](https://opentelemetry.io/docs/concepts/signals/metrics/), and [logs](https://opentelemetry.io/docs/concepts/signals/logs/), processes the telemetry, and exports it to a wide variety of observability backends using its components. For a conceptual overview of the Collector, see [Collector](https://opentelemetry.io/docs/collector).

## Introduction

- The `OpenTelemetry Collector` offers a *vendor-agnostic* implementation of how to *receive*, *process* and *export telemetry data*. It removes the need to run, operate, and maintain multiple `agents/collectors`. This works with improved scalability and supports open source observability data formats (e.g. `Jaeger`, `Prometheus`, `Fluent Bit`, etc.) sending to one or more open source or commercial backends.

## Objectives

- *Usability*: Reasonable default configuration, supports popular protocols, runs and collects out of the box.
- *Performance*: Highly stable and performant under varying loads and configurations.
- *Observability*: An exemplar of an observable service.
- *Extensibility*: Customizable without touching the core code.
- *Unification*: Single codebase, deployable as an agent or collector with support for traces, metrics, and logs.

## When to use a collector

- For most language specific instrumentation libraries you have exporters for popular backends and OTLP. You might wonder, under what circumstances does one use a collector to send data, as opposed to having each service send directly to the backend?
- For trying out and getting started with `OpenTelemetry`, sending your data directly to a backend is a great way to get value quickly. Also, in a *development* or *small-scale environment you can get decent results without a *collector*.
- However, in general we recommend using a collector alongside your service, since it allows your service to offload data quickly and the collector can take care of additional handling like *retries*, *batching*, *encryption* or *even sensitive data filtering*.
- It is also easier to [setup a collector](https://opentelemetry.io/docs/collector/quick-start) than you might think: the default `OTLP exporters` in each language assume a local collector endpoint, so if you launch a collector it will automatically start receiving telemetry.

## Collector Security

- Follow best practices to make sure your collectors are [hosted](https://opentelemetry.io/docs/security/hosting-best-practices/) and [configured](https://opentelemetry.io/docs/security/config-best-practices/) securely.

## Status

- The **Collector status** is: [mixed](https://opentelemetry.io/docs/specs/otel/document-status/#mixed), since core Collector components currently have mixed stability levels.
- `Collector components` differ in their maturity levels. Each component has its stability documented in its README.md. You can find a list of all available Collector components in the [registry](https://opentelemetry.io/ecosystem/registry/?language=collector).

# Installation

- You can deploy the `OpenTelemetry Collector` on a wide variety of operating systems and architectures.
- If you aren’t familiar with the deployment models, components, and repositories applicable to the `OpenTelemetry Collector`, first review the `Data Collection` and `Deployment Methods` page.

# Deployments

- Patterns you can apply to deploy the `OpenTelemetry collector`
- The `OpenTelemetry Collector` consists of a single binary which you can use in different ways, for different use cases. This section describes deployment patterns, their use cases along with pros and cons and best practices for collector configurations for cross-environment and multi-backend deployments. For deployment security considerations, see Collector hosting best practices.

## No Collector

- The simplest pattern is not to use a collector at all. This pattern consists of applications instrumented with an OpenTelemetry SDK that export telemetry signals (traces, metrics, logs) directly into a backend: SDK → Backend.
- Tradeoff:
  - Pros:
    - Simple to use (especially in a dev/test environment)
    - No additional moving parts to operate (in production environments)
  - Cons:
    - Requires code changes if collection, processing, or ingestion changes
    - Strong coupling between the application code and the backend
    - There are limited number of exporters per language implementation

## Agent

- The agent collector deployment pattern consists of applications — [instrumented](https://opentelemetry.io/docs/languages/) with an `OpenTelemetry SDK` using `OpenTelemetry protocol (OTLP)` — or other collectors (using the `OTLP exporter`) that send telemetry signals to a collector instance running with the application or on the same host as the application (such as a *sidecar* or a *daemonset*).
- Each client-side SDK or downstream collector is configured with a collector location: SDK → Agent Collector → Backend.
- In the app, the SDK is configured to send OTLP data to a collector.
- The collector is configured to send telemetry data to one or more backends
- In the context of the app, you would set the `OTEL_METRICS_EXPORTER` to otlp (which is the default value) and configure the `OTLP exporter` with the address of your `collector`, for example (in `Bash` or `zsh` shell):

```bash
export OTEL_EXPORTER_OTLP_ENDPOINT=http://collector.example.com:4318
```

The collector serving at `collector.example.com:4318` would then be configured like so:

```yaml
# Traces
receivers:
  otlp: # the OTLP receiver the app is sending traces to
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:

exporters:
  otlp/jaeger: # Jaeger supports OTLP directly
    endpoint: https://jaeger.example.com:4317

service:
  pipelines:
    traces/dev:
      receivers: [otlp]
      processors: [batch]
      exporters: [otlp/jaeger]


# Metrics
receivers:
  otlp: # the OTLP receiver the app is sending metrics to
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:

exporters:
  prometheusremotewrite: # the PRW exporter, to ingest metrics to backend
    endpoint: https://prw.example.com/v1/api/remote_write

service:
  pipelines:
    metrics/prod:
      receivers: [otlp]
      processors: [batch]
      exporters: [prometheusremotewrite]


# Logs
receivers:
  otlp: # the OTLP receiver the app is sending logs to
    protocols:
      http:
        endpoint: 0.0.0.0:4318

processors:
  batch:

exporters:
  file: # the File Exporter, to ingest logs to local file
    path: ./app42_example.log
    rotation:

service:
  pipelines:
    logs/dev:
      receivers: [otlp]
      processors: [batch]
      exporters: [file]
```

- Tradeoff:
  - Pros:
    - Simple to get started
    - Clear 1:1 mapping between application and collector
  - Cons:
    - Scalability challenges (human and load-wise)
    - Inflexible

## Gateway

### Introduction

- The gateway collector deployment pattern consists of applications (or other collectors) sending telemetry signals to a single `OTLP endpoint` provided by one or more collector instances running as a standalone service (for example, a deployment in `Kubernetes`), typically per cluster, per data center or per region.
- In the general case you can use an out-of-the-box load balancer to distribute the load amongst the collectors: SDK + OTLP -> Load Balancer → Gateway Collector(s) → Backend.
- For use cases where the processing of the telemetry data processing has to happen in a specific collector, you would use a two-tiered setup with a collector that has a pipeline configured with the [Trace ID/Service-name aware load-balancing exporter](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/exporter/loadbalancingexporter) in the first tier and the collectors handling the scale out in the second tier. For example, you will need to use the load-balancing exporter when using the [Tail Sampling processor](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/tailsamplingprocessor) so that all spans for a given trace reach the same collector instance where the tail sampling policy is applied.

1. In the app, the SDK is configured to send OTLP data to a central location.
2. A collector configured using the load-balancing exporter that distributes signals to a group of collectors.
3. The collectors are configured to send telemetry data to one or more backends.

### Example

```nginx
server {
    listen 4317 http2;
    server_name _;

    location / {
            grpc_pass      grpc://collector4317;
            grpc_next_upstream     error timeout invalid_header http_500;
            grpc_connect_timeout   2;
            grpc_set_header        Host            $host;
            grpc_set_header        X-Real-IP       $remote_addr;
            grpc_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}

server {
    listen 4318;
    server_name _;

    location / {
            proxy_pass      http://collector4318;
            proxy_redirect  off;
            proxy_next_upstream     error timeout invalid_header http_500;
            proxy_connect_timeout   2;
            proxy_set_header        Host            $host;
            proxy_set_header        X-Real-IP       $remote_addr;
            proxy_set_header        X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}

upstream collector4317 {
    server collector1:4317;
    server collector2:4317;
    server collector3:4317;
}

upstream collector4318 {
    server collector1:4318;
    server collector2:4318;
    server collector3:4318;
}
```

### load-balancing exporter

- For a concrete example of the centralized collector deployment pattern we first need to have a closer look at the load-balancing exporter. It has two main configuration fields:
  - The `resolver`, which determines where to find the downstream collectors (or: backends). If you use the static sub-key here, you will have to manually enumerate the collector URLs. The other supported `resolver` is the DNS resolver which will periodically check for updates and resolve IP addresses. For this resolver type, the `hostname` sub-key specifies the hostname to query in order to obtain the list of IP addresses.
  - With the `routing_key` field you tell the load-balancing exporter to route spans to specific downstream collectors. If you set this field to `traceID` (default) then the Load-balancing exporter exports spans based on their `traceID`. Otherwise, if you use `service` as the value for `routing_key`, it exports spans based on their service name which is useful when using connectors like the [Span Metrics connector](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/connector/spanmetricsconnector), so all spans of a service will be send to the same downstream collector for metric collection, guaranteeing accurate aggregations.
- The first-tier collector servicing the `OTLP endpoint` would be configured as shown below:

```yaml
# Static
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  loadbalancing:
    protocol:
      otlp:
        tls:
          insecure: true
    resolver:
      static:
        hostnames:
          - collector-1.example.com:4317
          - collector-2.example.com:5317
          - collector-3.example.com

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [loadbalancing]


# DNS
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  loadbalancing:
    protocol:
      otlp:
        tls:
          insecure: true
    resolver:
      dns:
        hostname: collectors.example.com

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [loadbalancing]


# DNS with Service
receivers:
  otlp:
    protocols:
      grpc:
        endpoint: 0.0.0.0:4317

exporters:
  loadbalancing:
    routing_key: service
    protocol:
      otlp:
        tls:
          insecure: true
    resolver:
      dns:
        hostname: collectors.example.com
        port: 5317

service:
  pipelines:
    traces:
      receivers: [otlp]
      exporters: [loadbalancing]
```

- The load-balancing exporter emits metrics including `otelcol_loadbalancer_num_backends` and `otelcol_loadbalancer_backend_latency` that you can use for health and performance monitoring of the OTLP endpoint collector.

### Combined deployment of Collectors as agents and gateways

- Often a deployment of multiple `OpenTelemetry collectors` involves running both Collector as gateways and as [agents](https://opentelemetry.io/docs/collector/deployment/agent/).
- The following diagram shows an architecture for such a combined deployment:
  - We use the Collectors running in the agent deployment pattern (running on each host, similar to Kubernetes daemonsets) to collect telemetry from services running on the host and host telemetry, such as host metrics and scrap logs.
  - We use Collectors running in the gateway deployment pattern to process data, such as filtering, sampling, and exporting to backends etc.
- This combined deployment pattern is necessary, when you use components in your Collector that either need to be unique per host or that consume information that is only available on the same host as the application is running:
  - Receivers like the [`hostmetricsreceiver`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/hostmetricsreceiver) or [`filelogreceiver`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver) need to be unique per host instance. Running multiple instances of these receivers will result in duplicated data.
  - Processors like the [`resourcedetectionprocessor`](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/processor/resourcedetectionprocessor) are used to add information about the host, the collector and the application are running on. Running them within a Collector on a remote machine will result in incorrect data.

### Tradeoff

- Pros:
  - Separation of concerns such as centrally managed credentials
  - Centralized policy management (for example, filtering certain logs or sampling)
- Cons:
  - It’s one more thing to maintain and that can fail (complexity)
  - Added latency in case of cascaded collectors
  - Higher overall resource usage (costs)

### Multiple collectors and the single-writer principle

- All metric data streams within OTLP must have a [single writer](https://opentelemetry.io/docs/specs/otel/metrics/data-model/#single-writer). When deploying multiple collectors in a gateway configuration, it’s important to ensure that all metric data streams have a single writer and a globally unique identity.

### Potential problems

- Concurrent access from multiple applications that modify or report on the same data can lead to data loss or degraded data quality. For example, you might see inconsistent data from multiple sources on the same resource, where the different sources can overwrite each other because the resource is not uniquely identified.
- There are patterns in the data that may provide some insight into whether this is happening or not. For example, upon visual inspection, a series with unexplained gaps or jumps in the same series may be a clue that multiple collectors are sending the same samples. You might also see errors in your backend. For example, with a Prometheus backend:
- `Error on ingesting out-of-order samples`
- This error could indicate that identical targets exist in two jobs, and the order of the timestamps is incorrect. For example:
  - Metric `M1` received at `T1` with a timestamp 13:56:04 with value `100`
  - Metric `M1` received at `T2` with a timestamp 13:56:24 with value `120`
  - Metric `M1` received at `T3` with a timestamp 13:56:04 with value `110`
  - Metric `M1` received at time 13:56:24 with value `120`
  - Metric `M1` received at time 13:56:04 with value `110`
