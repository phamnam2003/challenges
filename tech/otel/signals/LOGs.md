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

### Logger Provider

- A `LoggerProvider` **MUST** provide a way to allow a [Resource](https://opentelemetry.io/docs/specs/otel/resource/sdk/) to be specified. If a Resource is specified, it SHOULD be associated with all the `LogRecords` produced by any Logger from the `LoggerProvider`.

### Additional LogRecord interfaces

- In this document we refer to `ReadableLogRecord` and `ReadWriteLogRecord`, defined as follows.

#### ReadableLogRecord

- A function receiving this as an argument **MUST** be able to access all the information added to the [`LogRecord`](https://opentelemetry.io/docs/specs/otel/logs/data-model/#log-and-event-record-definition). It **MUST** also be able to access the *Instrumentation Scope* and *Resource information* (implicitly) associated with the `LogRecord`.
- The trace context fields **MUST** be populated from the resolved `Context` (either the explicitly *passed* `Context` or the *current* `Context`) when [emitted](https://opentelemetry.io/docs/specs/otel/logs/api/#emit-a-logrecord).

#### ReadWriteLogRecord

- `ReadWriteLogRecord` is a superset of `ReadableLogRecord`.
- A function receiving this as an argument **MUST** additionally be able to modify the following information added to the `LogRecord`:
  - [Timestamp](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-timestamp)
  - [ObservedTimestamp](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-observedtimestamp)
  - [SeverityText](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-severitytext)
  - [SeverityNumber](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-severitynumber)
  - [Body](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-body)
  - [Attributes](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-attributes) (addition, modification, removal)
  - [TraceId](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-traceid)
  - [SpanId](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-spanid)
  - [TraceFlags](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-traceflags)
  - [EventName](https://opentelemetry.io/docs/specs/otel/logs/data-model/#field-eventname)
- The `SDK` **MAY** provide an operation that makes a deep clone of a `ReadWriteLogRecord`. The operation can be used by *asynchronous processors* (e.g. `Batching processor`) to avoid *race conditions* on the log record that is not required to be **concurrent-safe**.
