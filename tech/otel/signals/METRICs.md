# Overview Open Telemetry Metrics

## Design Goals

- Given there are many well-established metrics solutions that exist today, it is important to understand the goals of `OpenTelemetry’s metrics` effort:
  - **Being able to connect metrics to other signals**. For example, `metrics` and `traces` can be correlated via [exemplars](https://opentelemetry.io/docs/specs/otel/metrics/data-model/#exemplars), and metrics attributes can be enriched via [Baggage](https://opentelemetry.io/docs/specs/otel/baggage/api/) and [Context](https://opentelemetry.io/docs/specs/otel/context/). Additionally, [Resource](https://opentelemetry.io/docs/specs/otel/resource/sdk/) can be applied to `logs/metrics/traces` in a consistent way.
  - **Providing a path for OpenCensus customers to migrate to OpenTelemetry**
  - **Working with existing metrics instrumentation protocols and standards**. Here is the minimum set of goals:
    - Providing full support for [Prometheus](https://prometheus.io/) - users should be able to use `OpenTelemetry` clients and [Collector](https://opentelemetry.io/docs/specs/otel/overview/#collector) to collect and export metrics, with the ability to achieve the same functionality as the native Prometheus clients.
    - Providing the ability to collect StatsD metrics using the OpenTelemetry Collector.

# Concepts

## API

- The **`OpenTelemetry Metrics API`** (“the API” hereafter) serves two purposes:
  - Capturing raw measurements efficiently and simultaneously.
  - Decoupling the instrumentation from the `SDK`, allowing the `SDK` to be *specified/included* in the application.

## SDK

- The **`OpenTelemetry Metrics SDK`** (“the SDK” hereafter) implements the API, providing functionality and extensibility such as *configuration*, *aggregation*, *processors* and *exporters*.
- `OpenTelemetry` requires a [separation of the API from the SDK](https://opentelemetry.io/docs/specs/otel/library-guidelines/#requirements), so that different `SDK`s can be configured at run time. Please refer to the overall [`OpenTelemetry SDK`](https://opentelemetry.io/docs/specs/otel/overview/#sdk) concept for more information.

## Programming Model

```md
+------------------+
| MeterProvider    |                 +-----------------+             +--------------+
|   Meter A        | Measurements... |                 | Metrics...  |              |
|     Instrument X +-----------------> In-memory state +-------------> MetricReader |
|     Instrument Y |                 |                 |             |              |
|   Meter B        |                 +-----------------+             +--------------+
|     Instrument Z |
|     ...          |                 +-----------------+             +--------------+
|     ...          | Measurements... |                 | Metrics...  |              |
|     ...          +-----------------> In-memory state +-------------> MetricReader |
|     ...          |                 |                 |             |              |
|     ...          |                 +-----------------+             +--------------+
+------------------+
```
