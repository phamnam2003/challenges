# What is Traces

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
- More information about `Traces` can be found in the [OpenTelemetry Traces Concept](https://opentelemetry.io/docs/concepts/signals/traces/).
