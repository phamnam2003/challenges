- `OpenTelemetry SDK` components are highly configurable. This specification outlines the mechanisms by which `OpenTelemetry` components can be configured. It does not attempt to specify the details of what can be configured.

# Configuration Interfaces

## Programmatic

- The `SDK` **MUST** provide a programmatic interface for all configuration. This interface **SHOULD** be written in the language of the `SDK` itself. All other configuration mechanisms **SHOULD** be built on top of this interface.
- An example of this **programmatic interface** is accepting a well-defined struct on an *SDK builder class*. From that, one could build a CLI that accepts a file (`YAML`, `JSON`, `TOML`, …) and then transforms into that well-defined struct consumable by the *programmatic interface* (see [declarative configuration](https://opentelemetry.io/docs/specs/otel/configuration/#declarative-configuration)).
