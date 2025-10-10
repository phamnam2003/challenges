# Exporters

- Send telemetry to the [`OpenTelemetry Collector`](https://opentelemetry.io/docs/collector/) to make sure it’s exported correctly. Using the `Collector` in *production environments* is a best practice. To visualize your telemetry, export it to a backend such as [Jaeger](https://jaegertracing.io/), [Zipkin](https://zipkin.io/), [Prometheus](https://prometheus.io/), or a vendor-specific backend.

## Available exporters

- The registry contains a list of [exporters for Go](https://opentelemetry.io/ecosystem/registry/?component=exporter&language=go).
- Among exporters, [OpenTelemetry Protocol (OTLP)](https://opentelemetry.io/docs/specs/otlp/) exporters are designed with the `OpenTelemetry` data model in mind, emitting `OTel` data without any loss of information. Furthermore, many tools that operate on telemetry data support OTLP (such as Prometheus, Jaeger, and most vendors), providing you with a high degree of flexibility when you need it. To learn more about OTLP, see [OTLP Specification](https://opentelemetry.io/docs/specs/otlp/).
- This page covers the main `OpenTelemetry Go exporters` and how to set them up

## Console

- The console exporter is useful for development and debugging tasks, and is the simplest to set up.

### Console traces

- The [`go.opentelemetry.io/otel/exporters/stdout/stdouttrace`](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/stdout/stdouttrace) package contains an implementation of the console trace exporter.

### Console metrics

- The [`go.opentelemetry.io/otel/exporters/stdout/stdoutmetric`](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/stdout/stdoutmetric) package contains an implementation of the console metrics exporter.

### Console logs (Experimental)

- The [`go.opentelemetry.io/otel/exporters/stdout/stdoutlog`](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/stdout/stdoutlog) package contains an implementation of the console log exporter.

## OTLP

- To send trace data to an OTLP endpoint (like the [collector](https://opentelemetry.io/docs/collector) or Jaeger >= v1.35.0) you’ll want to configure an OTLP exporter that sends to your endpoint.

### OTLP traces over HTTP

- [`go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp`] contains an implementation of the OTLP trace exporter using HTTP with binary protobuf payloads.

### OTLP traces over gRPC

- [`go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc`] contains an implementation of OTLP trace exporter using gRPC.

### Jeager

- To try out the `OTLP exporter`, since `v1.35.0` you can run `Jaeger` as an `OTLP` endpoint and for *trace visualization* in a Docker container:

```bash
docker run -d --name jaeger \
  -e COLLECTOR_OTLP_ENABLED=true \
  -p 16686:16686 \
  -p 4317:4317 \
  -p 4318:4318 \
  jaegertracing/all-in-one:latest
```

### OTLP metrics over HTTP

- [`go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp`](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp) contains an implementation of `OTLP metrics exporter` using HTTP with *binary protobuf payloads*.

### OTLP metrics over gRPC

- [`go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc`](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc) contains an implementation of `OTLP metrics exporter` using `gRPC`.

### OTLP logs over HTTP (Experimental)

- [`go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp`](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp) contains an implementation of `OTLP logs exporter` using HTTP with *binary protobuf payloads*.

### OTLP logs over gRPC (Experimental)

- [`go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc`] contains an implementation of `OTLP logs exporter` using `gRPC`.

### Prometheus (Experimental)

- A `Prometheus exporter` is used to report metrics via `Prometheus` *scrape HTTP endpoint*.
- [`go.opentelemetry.io/otel/exporters/prometheus`](https://pkg.go.dev/go.opentelemetry.io/otel/exporters/prometheus) contains an implementation of Prometheus metrics exporter.
- To learn more on how to use the `Prometheus exporter`, try the [prometheus example](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/examples/prometheus)
