# What is Open Telemetry (OTel)?

- `Open Telemetry (OTel)` is an `open-source` [*observability framework*](https://opentelemetry.io/docs/concepts/observability-primer/#what-is-observability) for *cloud-native software*, providing a set of APIs, libraries, agents, and instrumentation to collect and export telemetry data such as `traces`, `metrics`, `logs`. `Open Telemetry` is a project under the `Cloud Native Computing Foundation (CNCF)` and is a merger of two previous projects: `OpenTracing` and `OpenCensus`.
- `Otel` designed to facilitate the:
  - [Generation](https://opentelemetry.io/docs/concepts/instrumentation)
  - Export
  - [Collection](https://opentelemetry.io/docs/concepts/components/#collector)
of [telemetry data](https://opentelemetry.io/docs/concepts/signals/) such as [traces](https://opentelemetry.io/docs/concepts/signals/traces/), [metrics](https://opentelemetry.io/docs/concepts/signals/metrics/), and [logs](https://opentelemetry.io/docs/concepts/signals/logs/)
- A major goal of `OpenTelemetry` is to enable easy *instrumentation* of your applications and systems, *regardless* of the programming language, infrastructure, and runtime environments used.

# Why Open Telemetry?

- With the rise of cloud computing, microservices architectures, and increasingly complex business requirements, the need for software and infrastructure [observability](https://opentelemetry.io/docs/concepts/observability-primer/#what-is-observability) is greater than ever.
- `OpenTelemetry` satisfies the need for observability while following two key principles:
  - You own the data that you generate. There’s no vendor lock-in.
  - You only have to learn a single set of APIs and conventions.

# Components of Open Telemetry

- `OpenTelemetry` is currently made up of several main components:
  - [Specification](https://opentelemetry.io/docs/concepts/components/#specification)
  - [Collector](https://opentelemetry.io/docs/concepts/components/#collector)
  - [Language SDKs](https://opentelemetry.io/docs/concepts/components/#language-specific-api--sdk-implementations)
    - [Instrumentation Libraries](https://opentelemetry.io/docs/concepts/components/#instrumentation-libraries)
    - [Exporter](https://opentelemetry.io/docs/concepts/components/#exporters)
    - [Zero-Code Instrumentation](https://opentelemetry.io/docs/concepts/components/#zero-code-instrumentation)
    - [Resource Detectors](https://opentelemetry.io/docs/concepts/components/#resource-detectors)
    - [Copy Service Propagation](https://opentelemetry.io/docs/concepts/components/#cross-service-propagators)
    - [Sampler](https://opentelemetry.io/docs/concepts/components/#samplers)
  - [K8s operator](https://opentelemetry.io/docs/concepts/components/#kubernetes-operator)
  - [Function as a Service assets](https://opentelemetry.io/docs/concepts/components/#function-as-a-service-assets)

## Specification

- **API**: Defines data types and operations for generating and *correlating tracing*, *metrics*, and *logging data*.
- **SDK**: Defines requirements for a language-specific implementation of the API. *Configuration*, *data processing*, and *exporting* concepts are also defined here
- **DATA**: Defines the `OpenTelemetry Protocol (OTLP)` and `vendor-agnostic` *semantic conventions* that a telemetry backend can provide support for.
- For more information, see the [specifications](https://opentelemetry.io/docs/specs/).

## Collector

- The `OpenTelemetry Collector` is a *vendor-agnostic proxy* that can *receive*, *process*, and *export* telemetry data. It supports receiving telemetry data in multiple formats (for example, `OTLP`, `Jaeger`, `Prometheus`, as well as many commercial/proprietary tools) and sending data to one or more backends. It also supports processing and filtering telemetry data before it gets exported.
- For more information, see [Collector](https://opentelemetry.io/docs/collector/).

## Language-specific API & SDK implementations

- `OpenTelemetry` also has language *SDKs* that let you use the `OpenTelemetry` API to generate telemetry data with *your language of choice* and export that data to a preferred backend. These *SDKs* also let you *incorporate instrumentation libraries* for common libraries and frameworks that you can use to connect to manual instrumentation in your application.
- For more information, see [Instrumenting](<https://opentelemetry.io/docs/concepts/instrumentation/Instrumentation> libraries).

### Instrumentation libraries

- `OpenTelemetry` supports a broad number of components that generate relevant telemetry data from popular libraries and frameworks for supported languages. For example, inbound and outbound HTTP requests from an HTTP library generate data about those requests.
- And *aspirational goal* of `OpenTelemetry` is that all popular libraries are built to be observable by default, so that separate dependencies are not required.
- For more information, see [Instrumenting libraries](https://opentelemetry.io/docs/concepts/instrumentation/libraries/).

### Exporters

- Send telemetry to the `OpenTelemetry Collector` to make sure it’s exported correctly. Using the Collector in production environments is a best practice. To visualize your telemetry, export it to a backend such as [`Jaeger`](https://jaegertracing.io/), [`Zipkin`](https://zipkin.io/), [`Prometheus`](https://prometheus.io/), or a vendor-specific backend.
- Among exporters, [OpenTelemetry Protocol (OTLP)](https://opentelemetry.io/docs/specs/otlp/) exporters are designed with the `OpenTelemetry` data model in mind, *emitting OTel data* without any loss of information. Furthermore, many tools that operate on telemetry data support OTLP (such as `Prometheus`, `Jaeger`, and most vendors), providing you with a high degree of flexibility when you need it. To learn more about OTLP, see [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/)

### Zero-code instrumentation

- If applicable, a language specific implementation of `OpenTelemetry` provides a way to instrument your application without touching your source code. While the underlying mechanism depends on the language, zero-code instrumentation adds the  `OpenTelemetry` *API and SDK capabilities* to your application. Additionally, it might add a set of instrumentation libraries and exporter dependencies.
- For more information, see [Zero-code instrumentation](https://opentelemetry.io/docs/concepts/instrumentation/zero-code/)

### Resource detectors

- A `resource` represents the entity producing telemetry as resource attributes. For example, a process that produces telemetry that is running in a container on k8s has a Pod name, a namespace, and possibly a deployment name. You can include all these attributes in the resource.
- The language specific implementations of `OpenTelemetry` provide resource detection from the `OTEL_RESOURCE_ATTRIBUTES` environment variable and for many common entities, like process runtime, service, host, or operating system.
- For more information, see [Resources](https://opentelemetry.io/docs/concepts/resources/)

### Cross-service propagators

- Propagation is the mechanism that *moves data* between *services* and *processes*. Although not limited to tracing, propagation allows traces to build causal information about a system across services that are arbitrarily distributed across process and network boundaries.
- For the vast majority of the use cases, context propagation happens through *instrumentation libraries*. If needed, you can use propagators yourself to *serialize* and *deserialize* cross-cutting concerns such as the context of a `span` and [`baggage`](https://opentelemetry.io/docs/concepts/signals/baggage/).

### Samplers

- `Sampling` is process that restricts the amount of traces that are generated by a system. Each language-specific implementation of `OpenTelemetry` offers several [head samplers](https://opentelemetry.io/docs/concepts/sampling/#head-sampling).
- For more information, see [Sampling](https://opentelemetry.io/docs/concepts/sampling).

## Kubernetes operator

- The `OpenTelemetry Operator` is an implementation of a `Kubernetes Operator`. The operator manages the `OpenTelemetry Collector` and auto-instrumentation of the workloads using `OpenTelemetry`.
- For more information, see [K8s Operator](https://opentelemetry.io/docs/platforms/kubernetes/operator/).

## Function as a Service assets

- `OpenTelemetry` supports various methods of monitoring *Function-as-a-Service* provided by different cloud vendors. The `OpenTelemetry` community currently provides pre-built Lambda layers able to auto-instrument your application as well as the option of a standalone Collector Lambda layer that can be used when instrumenting applications manually or automatically.
- For more information, see [Functions as a Service](https://opentelemetry.io/docs/platforms/faas/).
