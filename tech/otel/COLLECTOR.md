# Collector

- Vendor-agnostic way to receive, process and export telemetry data.
- The OpenTelemetry Collector receives [traces](https://opentelemetry.io/docs/concepts/signals/traces/), [metrics](https://opentelemetry.io/docs/concepts/signals/metrics/), and [logs](https://opentelemetry.io/docs/concepts/signals/logs/), processes the telemetry, and exports it to a wide variety of observability backends using its components. For a conceptual overview of the Collector, see [Collector](https://opentelemetry.io/docs/collector).

## Introduction

- The `OpenTelemetry Collector` offers a *vendor-agnostic* implementation of how to *receive*, *process* and *export telemetry data*. It removes the need to run, operate, and maintain multiple `agents/collectors`. This works with improved scalability and supports open source observability data formats (e.g. `Jaeger`, `Prometheus`, `Fluent Bit`, etc.) sending to one or more open source or commercial backends.

## Objectives

- *Usability*: Reasonable default configuration, supports popular protocols, runs and collects out of the box.
- *Performance*: Highly stable and performant under varying loads and configurations.
- *Observability*: An exemplar of an observable service.
- *Extensibility*: Customizable without touching the core code.
- *Unification*: Single codebase, deployable as an agent or collector with support for traces, metrics, and logs.

## When to use a collector

- For most language specific instrumentation libraries you have exporters for popular backends and OTLP. You might wonder, under what circumstances does one use a collector to send data, as opposed to having each service send directly to the backend?
- For trying out and getting started with `OpenTelemetry`, sending your data directly to a backend is a great way to get value quickly. Also, in a *development* or *small-scale environment you can get decent results without a *collector*.
- However, in general we recommend using a collector alongside your service, since it allows your service to offload data quickly and the collector can take care of additional handling like *retries*, *batching*, *encryption* or *even sensitive data filtering*.
- It is also easier to [setup a collector](https://opentelemetry.io/docs/collector/quick-start) than you might think: the default `OTLP exporters` in each language assume a local collector endpoint, so if you launch a collector it will automatically start receiving telemetry.

## Collector Security

- Follow best practices to make sure your collectors are [hosted](https://opentelemetry.io/docs/security/hosting-best-practices/) and [configured](https://opentelemetry.io/docs/security/config-best-practices/) securely.

## Status

- The **Collector status** is: [mixed](https://opentelemetry.io/docs/specs/otel/document-status/#mixed), since core Collector components currently have mixed stability levels.
- `Collector components` differ in their maturity levels. Each component has its stability documented in its README.md. You can find a list of all available Collector components in the [registry](https://opentelemetry.io/ecosystem/registry/?language=collector).

# Installation

- You can deploy the `OpenTelemetry Collector` on a wide variety of operating systems and architectures.
- If you aren’t familiar with the deployment models, components, and repositories applicable to the `OpenTelemetry Collector`, first review the `Data Collection` and `Deployment Methods` page.

# Deployments

- Patterns you can apply to deploy the `OpenTelemetry collector`
- The `OpenTelemetry Collector` consists of a single binary which you can use in different ways, for different use cases. This section describes deployment patterns, their use cases along with pros and cons and best practices for collector configurations for cross-environment and multi-backend deployments. For deployment security considerations, see Collector hosting best practices.

## No Collector

## Agent

## Gateway
