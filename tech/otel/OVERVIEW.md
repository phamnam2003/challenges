# `OpenTelemetry` Client Architecture

![Client Architecture](../../images/otel_client_architecture.png)

- At the highest architectural level, `OpenTelemetry` clients are organized into [**signals**](https://opentelemetry.io/docs/specs/otel/glossary/#signals). Each signal provides a *specialized form* of observability. For example, `tracing`, `metrics`, and `baggage` are three separate signals. Signals share a common subsystem – **context propagation** – but they function independently from each other.
- Each signal provides a mechanism for software to describe itself. A codebase, such as web framework or a database client, takes a dependency on various signals in order to describe itself. `OpenTelemetry instrumentation` code can then be mixed into the other code within that codebase. This makes `OpenTelemetry` a [cross-cutting concern](https://en.wikipedia.org/wiki/Cross-cutting_concern) - a piece of software which is mixed into many other pieces of software in order to provide value. Cross-cutting concerns, by their very nature, violate a core design principle – separation of concerns. As a result, `OpenTelemetry` client design requires extra care and *attention to avoid* creating issues for the codebases which depend upon these cross-cutting APIs.
- `OpenTelemetry` clients are designed to separate the portion of each signal which must be imported as cross-cutting concerns from the portions which can be managed independently. `OpenTelemetry` clients are also designed to be an extensible framework. To accomplish these goals, each signal consists of four types of packages: `API`, `SDK`, `Semantic Conventions`, and `Contrib`.

# API

- API packages consist of the cross-cutting public interfaces used for instrumentation. Any portion of an `OpenTelemetry` client which is imported into third-party libraries and application code is considered part of the API.

# SDK

- The `SDK` is the implementation of the API provided by the `OpenTelemetry` project. Within an application, the `SDK` is installed and managed by the [application owner](https://opentelemetry.io/docs/specs/otel/glossary/#application-owner). Note that the `SDK` includes additional public interfaces which are not considered part of the `API` package, as they are not *cross-cutting concerns*. These public interfaces are defined as [*constructors*](https://opentelemetry.io/docs/specs/otel/glossary/#constructors) and [*plugin interfaces*](https://opentelemetry.io/docs/specs/otel/glossary/#sdk-plugins). Application owners use the `SDK constructors`; [plugin authors](https://opentelemetry.io/docs/specs/otel/glossary/#plugin-author) use the `SDK plugin interfaces`. [`Instrumentation authors`](https://opentelemetry.io/docs/specs/otel/glossary/#instrumentation-author) **MUST NOT** directly reference any `SDK package` of any kind, only the `**API**`.

# Semantic Conventions

- The **Semantic Conventions** define the keys and values which describe commonly observed concepts, protocols, and operations used by applications.
- Semantic Conventions are now located in their own repository: [https://github.com/open-telemetry/semantic-conventions](https://github.com/open-telemetry/semantic-conventions)
- Both the collector and the client libraries **SHOULD** autogenerate `semantic convention` keys and enum values into constants (or language idiomatic equivalent). Generated values shouldn’t be distributed in stable packages until semantic conventions are stable. The `YAML` files *MUST* be used as the source of truth for generation. Each language implementation **SHOULD** provide language-specific support to the [code generator](https://github.com/open-telemetry/semantic-conventions/tree/main/model/go).
- Additionally, attributes required by the specification will be listed [here](https://opentelemetry.io/docs/specs/otel/semantic-conventions/)

# Contrib Packages

- The `OpenTelemetry` project maintains integrations with popular **OSS projects** which have been identified as important for observing modern web services. Example API integrations include instrumentation for web frameworks, database clients, and message queues. Example SDK integrations include plugins for exporting telemetry to popular analysis tools and telemetry storage systems.
- Some plugins, such as `OTLP Exporters` and `TraceContext Propagators`, are required by the `OpenTelemetry specification`. These required plugins are included as part of the SDK.
- *Plugins and instrumentation packages* which are optional and separate from the SDK are referred to as **Contrib** packages. **`API Contrib`** refers to packages which depend solely upon the API; **SDK Contrib** refers to packages which also depend upon the SDK.
- The term `Contrib specifically` refers to the collection of plugins and instrumentation maintained by the `OpenTelemetry project`; it does not refer to *third-party plugins* hosted elsewhere.

# Versioning and Stability

- `OpenTelemetry` values *stability* and *backwards compatibility*. Please see the [versioning and stability guide for details](https://opentelemetry.io/docs/specs/otel/versioning-and-stability/).
