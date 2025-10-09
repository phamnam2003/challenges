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
