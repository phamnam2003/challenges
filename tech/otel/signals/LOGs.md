# Open Telemetry Solution

- Distributed tracing introduced the notion of `trace context propagation`.
![Open Telemetry Solution](../../../images/otel_solution.png)
- We emit `logs`, `traces` and `metrics` in a way that is compliant with `OpenTelemetry` data models, send the data through `OpenTelemetry Collector`, where it can be enriched and processed in a uniform manner

## Logs API

- The `Logs API` is provided for logging library authors to build [log appenders](https://opentelemetry.io/docs/specs/otel/logs/supplementary-guidelines/#how-to-create-a-log4j-log-appender), which use this API to bridge between existing logging libraries and the `OpenTelemetry` log data model.
- The `Logs API` can also be directly called by instrumentation libraries as well as instrumented libraries or applications.
- Consist of these main concepts:
  - [`LoggerProvider`](https://opentelemetry.io/docs/specs/otel/logs/api/#loggerprovider) is the entry point of the API. It provides access to `Logger`s.
  - [`Logger`](https://opentelemetry.io/docs/specs/otel/logs/api/#logger) is responsible for emitting logs as [`LogRecords`](https://opentelemetry.io/docs/specs/otel/logs/data-model/#log-and-event-record-definition).

## Data Model

### Log and Event Record Definition

| Field Name            | Description                                   |
| ----------------------| --------------------------------------------- |
| **Timestamp**         | Time when the event occurred.                 |
| **ObservedTimestamp** | Time when the event was observed.             |
| **TraceId**           | Request trace id.                             |
| **SpanId** | Request span id. |
| **TraceFlags** | W3C trace flag. |
| **SeverityText** | The severity text (also known as log level). |
| **SeverityNumber** | Numerical value of the severity. |
| **Body** | The body of the log record. |
| **Resource** | Describes the source of the log. |
| **InstrumentationScope** | Describes the scope that emitted the log. |
| **Attributes** | Additional information about the event. |
| **EventName** | Name that identifies the class / type of event. |

## SDK
