# `OpenTelemetry` Client Architecture

![Client Architecture](../../images/otel_client_architecture.png)

- At the highest architectural level, `OpenTelemetry` clients are organized into [**signals**](https://opentelemetry.io/docs/specs/otel/glossary/#signals). Each signal provides a *specialized form* of observability. For example, `tracing`, `metrics`, and `baggage` are three separate signals. Signals share a common subsystem – **context propagation** – but they function independently from each other.
- Each signal provides a mechanism for software to describe itself. A codebase, such as web framework or a database client, takes a dependency on various signals in order to describe itself. `OpenTelemetry instrumentation` code can then be mixed into the other code within that codebase. This makes `OpenTelemetry` a [cross-cutting concern](https://en.wikipedia.org/wiki/Cross-cutting_concern) - a piece of software which is mixed into many other pieces of software in order to provide value. Cross-cutting concerns, by their very nature, violate a core design principle – separation of concerns. As a result, `OpenTelemetry` client design requires extra care and *attention to avoid* creating issues for the codebases which depend upon these cross-cutting APIs.
- `OpenTelemetry` clients are designed to separate the portion of each signal which must be imported as cross-cutting concerns from the portions which can be managed independently. `OpenTelemetry` clients are also designed to be an extensible framework. To accomplish these goals, each signal consists of four types of packages: `API`, `SDK`, `Semantic Conventions`, and `Contrib`.

# API

- API packages consist of the cross-cutting public interfaces used for instrumentation. Any portion of an `OpenTelemetry` client which is imported into third-party libraries and application code is considered part of the API.
