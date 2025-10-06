# What is Observability?

- Observability lets you understand a system from the outside by letting you ask questions about that system without knowing its inner workings. Furthermore, it allows you to easily troubleshoot and handle novel problems, that is, “unknown unknowns”. It also helps you answer the question “Why is this happening?”
- To ask those questions about your system, your application *must be properly instrumented*. That is, the application code must emit [signals](https://opentelemetry.io/docs/concepts/signals/) such as [traces](https://opentelemetry.io/docs/concepts/signals/traces/), [metrics](https://opentelemetry.io/docs/concepts/signals/metrics/), and [logs](https://opentelemetry.io/docs/concepts/signals/logs/). An application is properly instrumented when developers don’t need to add more instrumentation to troubleshoot an issue, because they have all of the information they need.
- `OpenTelemetry` is the mechanism by which application code is instrumented to help make a system observable.

# Reliability and metrics

- **Telemetry** refers to data emitted from a system and its behavior. The data can come in the form of traces, metrics, and logs.
- Reliability answers the question: “Is the service doing what users expect it to be doing?” A system could be up 100% of the time, but if, when a user clicks “Add to Cart” to add a black pair of shoes to their shopping cart, the system doesn’t always add black shoes, then the system could be **unreliable**.

# Understanding distributed tracing

- Distributed tracing lets you observe requests as they propagate through complex, distributed systems. Distributed tracing improves the visibility of your application or system’s health and lets you debug behavior that is difficult to reproduce locally. It is essential for distributed systems, which commonly have nondeterministic problems or are too complicated to reproduce locally.
- To understand distributed tracing, you need to understand the role of each of its components: logs, spans, and traces.

# Logs

- A **log** is a timestamped message emitted by services or other components. Unlike traces, they aren’t necessarily associated with any particular user request or transaction. You can find logs almost everywhere in software. Logs have been heavily relied on in the past by both developers and operators to help them understand system behavior.
- Logs aren’t enough for tracking code execution, as they usually lack contextual information, such as where they were called from.

# Distributed traces

- A `distributed trace`, more commonly known as a **trace**, records the paths taken by requests (made by an application or end-user) as they propagate through multi-service architectures, like microservice and serverless applications.
- A **trace** is made of one or more spans. The first span represents the root span. Each root span represents a request from start to finish. The spans underneath the parent provide a more in-depth context of what occurs during a request (or what steps make up a request).
- Example `Spans`:

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

- Without tracing, finding the root cause of performance problems in a distributed system can be challenging. Tracing makes debugging and understanding distributed systems less daunting by breaking down what happens within a request as it flows through a distributed system.
