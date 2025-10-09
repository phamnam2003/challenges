# Introduction

- A [resource](https://opentelemetry.io/docs/specs/otel/resource/sdk/) represents the entity producing telemetry as *resource attributes*. For example, a process producing telemetry that is running in a container on Kubernetes has a process name, a pod name, a namespace, and possibly a deployment name. All four of these attributes can be included in the resource.
- In your observability backend, you can use *resource information* to better *investigate interesting behavior*. For example, if your `trace` or `metrics` data indicate latency in your system, you can narrow it down to a specific container, pod, or Kubernetes deployment.
- A `resource` is added to the `TracerProvider` or `MetricProvider` when they are created during initialization. This association cannot be changed later. After a `resource` is added, all `spans` and `metrics` produced from a `Tracer` or `Meter` from the provider will have the resource associated with them.

# Semantic Attributes with SDK-provided Default Value

- There are attributes provided by the `OpenTelemetry SDK`. One of them is the `service.name`, which *represents the logical name* of the service. By default, `SDKs` will assign the value *unknown_service* for this value, so it is recommended to set it explicitly, either in code or via setting the environment variable `OTEL_SERVICE_NAME`.
- Additionally, the `SDK` will also provides the following resource attributes to identify itself: `telemetry.sdk.name`, `telemetry.sdk.language` and `telemetry.sdk.version`.

# Resource Detectors

- Most language-specific `SDKs` provide a set of resource detectors that can be used to automatically detect resource information from the environment. Common resource detectors include:
  - [Operating System](https://opentelemetry.io/docs/specs/semconv/resource/os/)
  - [Host](https://opentelemetry.io/docs/specs/semconv/resource/host/)
  - [Process and Process Runtime](https://opentelemetry.io/docs/specs/semconv/resource/process/)
  - [Container](https://opentelemetry.io/docs/specs/semconv/resource/container/)
  - [Kubernetes](https://opentelemetry.io/docs/specs/semconv/resource/k8s/)
  - [Cloud-Provider-Specific Attributes](https://opentelemetry.io/docs/specs/semconv/resource/#cloud-provider-specific-attributes)
  - [and more](https://opentelemetry.io/docs/specs/semconv/resource/)

# Custom resources

- You can also provide your own resource attributes. You can either provide them in code or via populating the environment variable `OTEL_RESOURCE_ATTRIBUTES`. If applicable, use the semantic conventions for your resource attributes. For example, you can provide the name of your deployment environment using `deployment.environment.name`:

```bash
env OTEL_RESOURCE_ATTRIBUTES=deployment.environment.name=production yourApp
```
