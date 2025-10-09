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
