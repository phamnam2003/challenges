# Introduction

- Learn about the categories of telemetry supported by `OpenTelemetry`
- The purpose of `OpenTelemetry` is to collect, process, and export [signals](https://opentelemetry.io/docs/specs/otel/glossary/#signals). Signals are system outputs that describe the underlying activity of the operating system and applications running on a platform. A signal can be something you want to measure at a specific point in time, like *temperature* or *memory usage*, or *an event* that goes through the components of your *distributed system* that you’d like to `trace`. You can group different signals together to observe the inner workings of the same piece of technology under different angles.
- `OpenTelemetry` currently supports:
  - [Traces](https://opentelemetry.io/docs/concepts/signals/traces/)
  - [Metrics](https://opentelemetry.io/docs/concepts/signals/metrics)
  - [Logs](https://opentelemetry.io/docs/concepts/signals/logs)
  - [Baggage](https://opentelemetry.io/docs/concepts/signals/baggage/)
- Also under development or at the [proposal](https://github.com/open-telemetry/opentelemetry-specification/tree/main/oteps/#readme) stage:
  - [Events](https://opentelemetry.io/docs/specs/otel/logs/data-model/#events), a specific type of [log](https://opentelemetry.io/docs/concepts/signals/logs/)
  - [Profiles](https://github.com/open-telemetry/opentelemetry-specification/blob/main/oteps/profiles/0212-profiling-vision.md) are being worked on by the Profiling Working Group.

## Traces

### What is Traces

- The path of a request through your application.
- **Traces** give us the big picture of what happens when a request is made to an application. Whether your application is a monolith with a single database or a sophisticated mesh of services, traces are essential to understanding the full “path” a request takes in your application.
- `hello` spans:

```json
{
  "name": "hello",
  "context": {
    "trace_id": "5b8aa5a2d2c872e8321cf37308d69df2",
    "span_id": "051581bf3cb55c13"
  },
  "parent_id": null,
  "start_time": "2022-04-29T18:52:58.114201Z",
  "end_time": "2022-04-29T18:52:58.114687Z",
  "attributes": {
    "http.route": "some_route1"
  },
  "events": [
    {
      "name": "Guten Tag!",
      "timestamp": "2022-04-29T18:52:58.114561Z",
      "attributes": {
        "event_attributes": 1
      }
    }
  ]
}
```

- This is the root span, denoting the beginning and end of the entire operation. Note that it has a `trace_id` field indicating the `trace`, but has no *parent_id*. That’s how you know it’s the root span.
- `hello-greetings` span:

```json
{
  "name": "hello-greetings",
  "context": {
    "trace_id": "5b8aa5a2d2c872e8321cf37308d69df2",
    "span_id": "5fb397be34d26b51"
  },
  "parent_id": "051581bf3cb55c13",
  "start_time": "2022-04-29T18:52:58.114304Z",
  "end_time": "2022-04-29T22:52:58.114561Z",
  "attributes": {
    "http.route": "some_route2"
  },
  "events": [
    {
      "name": "hey there!",
      "timestamp": "2022-04-29T18:52:58.114561Z",
      "attributes": {
        "event_attributes": 1
      }
    },
    {
      "name": "bye now!",
      "timestamp": "2022-04-29T18:52:58.114585Z",
      "attributes": {
        "event_attributes": 1
      }
    }
  ]
}
```

- This span encapsulates specific tasks, like saying greetings, and its parent is the `hello` span. Note that it shares the *same trace_id* as the `root span`, indicating it’s a part of the *same trace*. Additionally, it has a `parent_id` that *matches* the `span_id` of the `hello` span.
- `hello-salutations` span:

```json
{
  "name": "hello-salutations",
  "context": {
    "trace_id": "5b8aa5a2d2c872e8321cf37308d69df2",
    "span_id": "93564f51e1abe1c2"
  },
  "parent_id": "051581bf3cb55c13",
  "start_time": "2022-04-29T18:52:58.114492Z",
  "end_time": "2022-04-29T18:52:58.114631Z",
  "attributes": {
    "http.route": "some_route3"
  },
  "events": [
    {
      "name": "hey there!",
      "timestamp": "2022-04-29T18:52:58.114561Z",
      "attributes": {
        "event_attributes": 1
      }
    }
  ]
}
```

- This span represents the *third operation* in this trace and, like the previous one, it’s *a child* of the `hello` span. That also makes it a sibling of the `hello-greetings` span.
- These three blocks of `JSON` all share the same trace_id, and the parent_id field represents a hierarchy. That makes it a `Trace`!
- Another thing you’ll note is that each `Span` looks like a *structured log*. That’s because it kind of is! One way to think of `Traces` is that they’re a collection of structured logs with *context*, *correlation*, *hierarchy*, and *more baked in*. However, these “structured logs” can come from different *processes*, *services*, *VMs*, *data centers*, and so on. This is what allows tracing to represent an end-to-end view of any system.

### Trace Providers

- A Tracer Provider (sometimes called `TracerProvider`) is a *factory for Tracers*. In most applications, a Tracer Provider is initialized once and `its lifecycle` matches the *application’s lifecycle*. `Tracer Provider` initialization also includes `Resource` and `Exporter` initialization. It is typically the first step in tracing with `OpenTelemetry`. In some language `SDKs`, a global `Tracer Provider` is already initialized for you.

### Tracer

- `Trace Exporters` send *traces* to a *consumer*. This *consumer* can be standard output for debugging and development-time, the `OpenTelemetry Collector`, or any open source or vendor backend of your choice.

### Context Propagation

- `Context Propagation` is the core concept that enables `Distributed Tracing`. With `Context Propagation`, `Spans` can be *correlated* with each other and *assembled into a trace*, regardless of where Spans are generated. To learn more about this topic, see the concept page on [Context Propagation](https://opentelemetry.io/docs/concepts/context-propagation).

### Spans

- A **span** represents a *unit of work or operation*. `Spans` are the building *blocks of Traces*. In `OpenTelemetry`, they include the following information:
  - Name
  - Parent span ID (empty for root spans)
  - Start and End Timestamps
  - [Span Context](https://opentelemetry.io/docs/concepts/signals/traces/#span-context)
  - [Attributes](https://opentelemetry.io/docs/concepts/signals/traces/#attributes)
  - [Span Events](https://opentelemetry.io/docs/concepts/signals/traces/#span-events)
  - [Span Links](https://opentelemetry.io/docs/concepts/signals/traces/#span-links)
  - [Span Status](https://opentelemetry.io/docs/concepts/signals/traces/#span-status)
- Sample Spans:

```json
{
  "name": "/v1/sys/health",
  "context": {
    "trace_id": "7bba9f33312b3dbb8b2c2c62bb7abe2d",
    "span_id": "086e83747d0e381e"
  },
  "parent_id": "",
  "start_time": "2021-10-22 16:04:01.209458162 +0000 UTC",
  "end_time": "2021-10-22 16:04:01.209514132 +0000 UTC",
  "status_code": "STATUS_CODE_OK",
  "status_message": "",
  "attributes": {
    "net.transport": "IP.TCP",
    "net.peer.ip": "172.17.0.1",
    "net.peer.port": "51820",
    "net.host.ip": "10.177.2.152",
    "net.host.port": "26040",
    "http.method": "GET",
    "http.target": "/v1/sys/health",
    "http.server_name": "mortar-gateway",
    "http.route": "/v1/sys/health",
    "http.user_agent": "Consul Health Check",
    "http.scheme": "http",
    "http.host": "10.177.2.152:26040",
    "http.flavor": "1.1"
  },
  "events": [
    {
      "name": "",
      "message": "OK",
      "timestamp": "2021-10-22 16:04:01.209512872 +0000 UTC"
    }
  ]
}
```

### Span Context

- Span context is an *immutable object* on every span that contains the following:
  - The Trace ID representing the trace that the span is a part of
  - The span’s Span ID
  - Trace Flags, a binary encoding containing information about the trace
  - Trace State, a list of key-value pairs that can carry vendor-specific trace information
- Span context is the part of a span that is serialized and propagated alongside [Distributed Context](https://opentelemetry.io/docs/concepts/signals/traces/#context-propagation) and [Baggage](https://opentelemetry.io/docs/concepts/signals/baggage).
- Because Span Context contains the Trace ID, it is used when creating [Span Links](https://opentelemetry.io/docs/concepts/signals/traces/#span-links).

### Attributes

- Attributes are key-value pairs that contain metadata that you can use to annotate a Span to carry information about the operation it is tracking.
- You can *add attributes* to *spans* during or after span creation. Prefer adding attributes at span creation to make the attributes available to SDK sampling. If you have to add a value after span creation, update the span with the value.
- Attributes have the following rules that each language SDK implements:
  - Keys must be non-null string values
  - Values must be a non-null string, boolean, floating point value, integer, or an array of these values
- Additionally, there are `Semantic Attributes`, which are known naming conventions for metadata that is typically present in common operations. It’s helpful to use semantic attribute naming wherever possible so that common kinds of metadata are standardized across systems.

### Span Events

- A `Span Event` can be thought of as a structured log message (or annotation) on a `Span`, typically used to denote a meaningful, singular point in time during the Span’s duration.
- For example, consider two scenarios in a web browser:
  - Tracking a page load
  - Denoting when a page becomes interactive
- A Span is best used to the first scenario because it’s an operation with a start and an end.
- A Span Event is best used to track the second scenario because it represents a meaningful, singular point in time.

### When to use span events versus span attributes

- Since span events also contain attributes, the question of when to use events instead of attributes might not always have an obvious answer. To inform your decision, consider whether a specific timestamp is meaningful.
- If the timestamp in which the operation completes is meaningful or relevant, attach the data to a span event. If the timestamp isn’t meaningful, attach the data as span attributes.

### Span Links

- Links exist so that you can associate one span with one or more spans, implying a causal relationship. For example, let’s say we have a distributed system where some operations are tracked by a trace.
- In response to some of these operations, an additional operation is queued to be executed, but its execution is asynchronous. We can track this subsequent operation with a trace as well.
- We would like to associate the trace for the subsequent operations with the first trace, but we cannot predict when the subsequent operations will start. We need to associate these two traces, so we will use a span link.
- You can link the last span from the first trace to the first span in the second trace. Now, they are causally associated with one another.
- Links are optional but serve as a good way to associate trace spans with one another.
- For more information see [Span Links](https://opentelemetry.io/docs/specs/otel/trace/api/#link)

### Span Status

- Each span has a status. The three possible values are:
  - `Unset`: The default value. It means the span has no status.
  - `Error`: The operation represented by the span has failed.
  - `Ok`: The operation represented by the span has completed successfully.

### Span Kinds

- When a span is created, it is one of `Client`, `Server`, `Internal`, `Producer`, or `Consumer`. This span kind provides a hint to the tracing backend as to *how the trace* should be assembled. According to the `OpenTelemetry specification`, the parent of a server span is often a *remote client span*, and the child of a client span is usually a server span. Similarly, the parent of a *consumer span* is always a *producer* and the child of a *producer span is always a consumer*. If not provided, the span kind is assumed to be internal.

For more information regarding [SpanKind](https://opentelemetry.io/docs/specs/otel/trace/api/#spankind)

#### Client

- A `client span` represents a *synchronous* outgoing remote call such as an outgoing HTTP request or database call. Note that in this context, *“synchronous”* does not refer to `async/await`, but to the fact that it is not queued for later processing.

#### Server

- A server span represents a synchronous incoming remote call such as an incoming HTTP request or remote procedure call.

#### Internal

- Internal spans represent operations which do not cross a process boundary. Things like instrumenting a function call or an Express middleware may use internal spans.

#### Producer

- Producer spans represent the creation of a job which may be asynchronously processed later. It may be a remote job such as one inserted into a job queue or a local job handled by an event listener.

#### Consumer

- Consumer spans represent the processing of a job created by a producer and may start long after the producer span has already ended.

## Metrics

- A measurement captured at runtime.
- A `metric` is a *measurement* of a service *captured at runtime*. The *moment of capturing* a measurement is known as a *metric event*, which consists not only of the measurement itself, but also the time at which it was captured and *associated metadata*.
- Application and request metrics are important indicators of *availability* and *performance*. Custom metrics can provide insights into how availability indicators impact user experience or the business. Collected data can be used to alert of an outage or trigger scheduling decisions to scale up a deployment automatically upon high demand.

### Meter Provider

- A *Meter Provider* (sometimes called `MeterProvider`) is a factory for `Meters`. In most applications, a `Meter Provider` is initialized once and its lifecycle matches the application’s lifecycle. Meter Provider initialization also includes `Resource` and `Exporter` initialization. It is typically the first step in metering with `OpenTelemetry`. In some language `SDKs`, a global *Meter Provider* is already initialized for you.

### Meter

- A `Meter` creates [metric instruments](https://opentelemetry.io/docs/concepts/signals/metrics/#metric-instruments), capturing measurements about a service at runtime. Meters are created from `Meter Providers`.

### Metric Exporter

- `Metric Exporters` send metric data to a consumer. This consumer can be standard output for debugging during development, the `OpenTelemetry Collector`, or any open source or vendor backend of your choice.

### Metric Instruments

- In `OpenTelemetry` measurements are captured by **metric instruments**. A metric instrument is defined by:
  - Name
  - Kind
  - Unit (optional)
  - Description (optional)
- The name, unit, and description are chosen by the developer or defined via semantic conventions for common ones like request and process metrics.
- The instrument kind is one of the following:
  - **Counter**: A value that accumulates over time – you can think of this like an odometer on a car; it only ever goes up.
  - **Asynchronous Counter**: Same as the Counter, but is collected once for each export. Could be used if you don’t have access to the continuous increments, but only to the aggregated value.
  - **UpDownCounter**: A value that accumulates over time, but can also go down again. An example could be a queue length, it will increase and decrease with the number of work items in the queue.
  - **Asynchronous UpDownCounter**: Same as the `UpDownCounter`, but is collected once for each export. Could be used if you don’t have access to the continuous changes, but only to the aggregated value (e.g., current queue size).
  - **Gauge**: Measures a current value at the time it is read. An example would be the fuel gauge in a vehicle. Gauges are synchronous.
  - **Asynchronous Gauge**: Same as the `Gauge`, but is collected once for each export. Could be used if you don’t have access to the continuous changes, but only to the aggregated value.
  - **Histogram**: A client-side aggregation of values, such as request latencies. A histogram is a good choice if you are interested in value statistics. For example: How many requests take fewer than 1s?

### Aggregation

- In addition to the *metric instruments*, the concept of **aggregations** is an important one to understand. An aggregation is a technique whereby a large number of measurements are combined into either exact or estimated statistics about metric events that took place during a time window. The `OTLP protocol` transports such aggregated metrics. The `OpenTelemetry API` provides a default *aggregation* for each instrument which can be overridden using the `Views`. The `OpenTelemetry` project aims to provide default aggregations that are supported by visualizers and telemetry backends.
- Unlike [request tracing](https://opentelemetry.io/docs/concepts/signals/traces/), which is intended to capture request lifecycles and provide context to the individual pieces of a request, metrics are intended to provide statistical information in aggregate. Some examples of use cases for metrics include:
  - Reporting the total number of bytes read by a service, per protocol type.
  - Reporting the total number of bytes read and the bytes per request.
  - Reporting the duration of a system call.
  - Reporting request sizes in order to determine a trend.
  - Reporting CPU or memory usage of a process.
  - Reporting average balance values from an account.
  - Reporting current active requests being handled.

### Views

- A view provides SDK users with the flexibility to customize the metrics output by the SDK. You can customize which metric instruments are to be processed or ignored. You can also customize aggregation and what attributes you want to report on metrics.

## Logs

- A recording of an event.
- A **log** is a `timestamped` text record, either structured (recommended) or unstructured, with optional metadata. Of all telemetry signals, `logs` have the *biggest legacy*. Most programming languages have built-in logging capabilities or well-known, widely used logging libraries.

### OpenTelemetry logs

- `OpenTelemetry` does not define a bespoke *API or SDK* to create logs. Instead, `OpenTelemetry logs` are the existing logs you already have from a logging framework or infrastructure component. `OpenTelemetry SDKs` and autoinstrumentation utilize several components to automatically correlate logs with traces.
- `OpenTelemetry’s` support for *logs* is designed to be fully compatible with what you already have, providing capabilities to wrap those logs with additional context and a common toolkit to parse and manipulate logs into a common format across many different sources.

### OpenTelemetry logs in the OpenTelemetry Collector

- The [OpenTelemetry Collector](https://opentelemetry.io/docs/collector/) provides several tools to work with logs:
  - Several receivers which parse logs from specific, known sources of log data.
  - The `filelogreceiver`, which reads logs from any file and provides features to parse them from different formats or use a regular expression.
  - Processors like the `transformprocessor` which lets you parse nested data, flatten nested structures, *add/remove/update* values, and more.
  - `Exporters` that let you emit log data in a *non-OpenTelemetry format*.
- The first step in adopting `OpenTelemetry` frequently involves deploying a Collector as a general-purposes logging agent.

### OpenTelemetry logs for applications

- In applications, `OpenTelemetry logs` are created with any logging library or built-in logging capabilities. When you add `autoinstrumentation` or activate an *SDK*, `OpenTelemetry` will automatically correlate your existing logs with any *active trace and span*, *wrapping the log body with their IDs*. In other words, `OpenTelemetry` automatically correlates your `logs` and `traces`.

### Structured, unstructured, and semistructured logs

- `OpenTelemetry` does not technically distinguish between structured and unstructured logs. You can use any log you have with `OpenTelemetry`. However, not all log formats are equally useful! Structured logs, in particular, are recommended for production observability because they are easy to parse and analyze at scale. The following section explains the differences between *structured, unstructured, and semistructured logs*.

#### Structured logs

- A structured log is a log whose textual format follows a consistent, machine-readable format. For applications, one of the most common formats is JSON:

```json
{
  "timestamp": "2024-08-04T12:34:56.789Z",
  "level": "INFO",
  "service": "user-authentication",
  "environment": "production",
  "message": "User login successful",
  "context": {
    "userId": "12345",
    "username": "johndoe",
    "ipAddress": "192.168.1.1",
    "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36"
  },
  "transactionId": "abcd-efgh-ijkl-mnop",
  "duration": 200,
  "request": {
    "method": "POST",
    "url": "/api/v1/login",
    "headers": {
      "Content-Type": "application/json",
      "Accept": "application/json"
    },
    "body": {
      "username": "johndoe",
      "password": "******"
    }
  },
  "response": {
    "statusCode": 200,
    "body": {
      "success": true,
      "token": "jwt-token-here"
    }
  }
}
```

- To make the most use of this log, parse both the JSON and the ELF-related pieces into a shared format to make analysis on an observability backend easier. The `filelogreceiver` in the `OpenTelemetry Collector` contains standardized ways to parse logs like this.
- Structured logs are the preferred way to use logs. Because structured logs are emitted in a consistent format, they are straightforward to parse, which makes them easier to *preprocess* in an `OpenTelemetry Collector`, correlate with other data, and ultimate analyze in an Observability backend.
