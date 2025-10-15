- `OpenTelemetry SDK` components are highly configurable. This specification outlines the mechanisms by which `OpenTelemetry` components can be configured. It does not attempt to specify the details of what can be configured.

# Configuration Interfaces

## Programmatic

- The `SDK` **MUST** provide a programmatic interface for all configuration. This interface **SHOULD** be written in the language of the `SDK` itself. All other configuration mechanisms **SHOULD** be built on top of this interface.
- An example of this **programmatic interface** is accepting a well-defined struct on an *SDK builder class*. From that, one could build a CLI that accepts a file (`YAML`, `JSON`, `TOML`, …) and then transforms into that well-defined struct consumable by the *programmatic interface* (see [declarative configuration](https://opentelemetry.io/docs/specs/otel/configuration/#declarative-configuration)).

## Environment variables

- Environment variable configuration defines a set of language agnostic *environment variables* for common configuration goals.
- See [OpenTelemetry Environment Variable Specification](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/).

## Declarative configuration

- Declarative configuration provides a mechanism for configuring `OpenTelemetry` which is more expressive and full-featured than the [environment variable](https://opentelemetry.io/docs/specs/otel/configuration/#environment-variables) based scheme, and language agnostic in a way not possible with [programmatic configuration](https://opentelemetry.io/docs/specs/otel/configuration/#programmatic). Notably, declarative configuration defines tooling allowing users to load `OpenTelemetry` components according to a file-based representation of a standardized configuration data model.
- Declarative configuration consists of the following main components:
  - [Data model](https://opentelemetry.io/docs/specs/otel/configuration/data-model/) defines data structures which allow users to specify an intended configuration of `OpenTelemetry SDK` components and `instrumentation`. The data model includes a file-based representation.
  - [Instrumentation configuration API](https://opentelemetry.io/docs/specs/otel/configuration/api/) allows instrumentation libraries to consume configuration by reading relevant configuration options during initialization.
  - [Configuration SDK](https://opentelemetry.io/docs/specs/otel/configuration/sdk/) defines `SDK` capabilities around file configuration, including an In-Memory configuration model, support for referencing custom extension plugin interfaces in configuration files, and operations to parse configuration files and interpret the configuration data model.

## Other Mechanisms

- Additional configuration mechanisms **SHOULD** be provided in whatever `language/format/style` is idiomatic for the language of the `SDK`. The `SDK` can include as many configuration mechanisms as appropriate.

# Instrumentation Configuration API

## Overview

- The instrumentation configuration API is part of the [declarative configuration interface](https://opentelemetry.io/docs/specs/otel/configuration/#declarative-configuration).
- The API allows [instrumentation libraries](https://opentelemetry.io/docs/specs/otel/glossary/#instrumentation-library) to consume configuration by reading relevant configuration during initialization. For example, an instrumentation library for an HTTP client can read the set of HTTP request and response headers to capture.
- It consists of the following main components:
  - [`ConfigProvider`](https://opentelemetry.io/docs/specs/otel/configuration/api/#configprovider) is the entry point of the API.
  - [`ConfigProperties`](https://opentelemetry.io/docs/specs/otel/configuration/api/#configproperties) is a programmatic representation of a configuration mapping node.

## Config Provider

- `ConfigProvider` provides access to configuration properties relevant to instrumentation.
- Instrumentation libraries access `ConfigProvider` during initialization. `ConfigProvider` may be passed as an argument to the instrumentation library, or the instrumentation library may access it from a central place. Thus, the **API SHOULD** provide a way to access a global default `ConfigProvider`, and set/register it.

### Config Provider operations

- The `ConfigProvider` **MUST** provide the following functions:
  - [Get instrumentation config](https://opentelemetry.io/docs/specs/otel/configuration/api/#get-instrumentation-config)
- TODO: decide if additional operations are needed to improve API ergonomics

#### Get instrumentation config

- Obtain configuration relevant to instrumentation libraries.
- **Returns**: `ConfigProperties` representing the `.instrumentation` configuration mapping node.
- If the `.instrumentation` node is not set, get instrumentation config **MUST** return `nil`, `null`, `undefined` or another language-specific idiomatic pattern denoting empty.

## Config Properties

- `ConfigProperties` is a *programmatic representation* of a configuration mapping node (i.e. a `YAML` mapping node).
- `ConfigProperties` **MUST** provide accessors for reading all properties from the mapping node it represents, including:
  - scalars (string, boolean, double precision floating point, 64-bit integer)
  - mappings, which SHOULD be represented as `ConfigProperties`
  - sequences of scalars
  - sequences of mappings, which SHOULD be represented as `ConfigProperties`
  - the set of property keys present
- `ConfigProperties` **SHOULD** provide access to properties in a type safe manner, based on what is idiomatic in the language.
- `ConfigProperties` **SHOULD** allow a caller to determine if a property is present with a *null value*, *versus not set*.

# Configuration SDK

## Overview

- The `SDK` is an implementation of [`Instrumenation Config API`](https://opentelemetry.io/docs/specs/otel/configuration/api/) and other user facing declarative configuration capabilities. It consists of the following main components:
  - [In-Memory configuration model](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#in-memory-configuration-model) is an in-memory representation of the [configuration model](https://opentelemetry.io/docs/specs/otel/configuration/data-model/).
  - [ConfigProvider](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#configprovider) defines the `SDK` implementation of the `ConfigProvider API`.
  - [SDK extension components](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#sdk-extension-components) defines how users and libraries extend file configuration with custom SDK extension plugin interfaces (*exporters*, *processors*, etc).
  - [SDK operations](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#sdk-operations) defines user APIs to parse configuration files and produce SDK components from their contents.

## In-Memory configuration model

- `SDK`s **SHOULD** provide an in-memory representation of the [configuration model](https://opentelemetry.io/docs/specs/otel/configuration/data-model/). Whereas [`ConfigProperties`](https://opentelemetry.io/docs/specs/otel/configuration/api/#configproperties) is a schemaless representation of any mapping node, the in-memory configuration model **SHOULD** reflect the schema of the configuration model.
- `SDK`s are encouraged to provide this in-memory representation in a manner that is idiomatic for their language. If an `SDK` needs to expose a class or interface, the name Configuration is **RECOMMENDED**.

## Config Provider

- The SDK implementation of [`ConfigProvider`](https://opentelemetry.io/docs/specs/otel/configuration/api/#configprovider) **MUST** be created using a [`ConfigProperties`](https://opentelemetry.io/docs/specs/otel/configuration/api/#configproperties) representing the [`.instrumentation`](https://github.com/open-telemetry/opentelemetry-configuration/blob/670901762dd5cce1eecee423b8660e69f71ef4be/examples/kitchen-sink.yaml#L438-L439) mapping node of the [configuration model](https://opentelemetry.io/docs/specs/otel/configuration/data-model/).

## SDK extension components

- The `SDK` supports a variety of extension [plugin interfaces](https://opentelemetry.io/docs/specs/otel/glossary/#sdk-plugins), allowing users and libraries to customize behaviors including the sampling, processing, and exporting of data. In general, the configuration data model defines specific types for built-in implementations of these plugin interfaces. For example, the BatchSpanProcessor type refers to the built-in Batching span processor. The schema SHOULD also support the ability to specify custom implementations of plugin interfaces defined by libraries or users.
- For example, a custom [span exporter](https://opentelemetry.io/docs/specs/otel/trace/sdk/#span-exporter) might be configured as follows:

```yaml
tracer_provider:
  processors:
    - batch:
        exporter:
          my-exporter:
            config-parameter: value
```

- Here we specify that the `tracer provider` has a `batch span processor` paired with a custom span exporter named my-exporter, which is configured with `config-parameter: value`. For this configuration to succeed, a `ComponentProvider` must be registered with type: `SpanExporter`, and `name: my-exporter`. When parse is called, the implementation will encounter my- exporter and translate the corresponding configuration to an equivalent `ConfigProperties` representation (i.e. `properties: {config-parameter: value}`). When create is called, the implementation will encounter my-exporter and invoke create plugin on the registered `ComponentProviderwith` the `ConfigProperties` determined during `parse`.
- Given the inherent differences across languages, the details of extension component mechanisms are likely to vary to a greater degree than is the case with other `API`s defined by `OpenTelemetry`. This is to be expected and is acceptable so long as the implementation results in the defined behaviors.

### Component Provider

- A `ComponentProvider` is responsible for interpreting configuration and returning an implementation of a particular type of `SDK extension plugin interface`.
- `ComponentProvider`s are registered with an SDK implementation of configuration via register. This **MAY** be done automatically or require manual intervention by the user based on what is possible and idiomatic in the language ecosystem. For example in Java, `ComponentProviders` might be registered automatically using the `service provider interface (SPI)` mechanism.
- See [create](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#create), which details `ComponentProvider` usage in configuration model interpretation.

#### Supported SDK extension plugins

- The [configuration data model](https://opentelemetry.io/docs/specs/otel/configuration/data-model/) **SHOULD** support configuration of all SDK extension plugin interfaces. `SDK`s **SHOULD** support [registration](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#register-componentprovider) of custom implementations of SDK extension plugin interfaces via the `ComponentProvider mechanism`.

#### ComponentsProvider operations

- The `ComponentsProvider` **MUST** provide the following functions:
  - [Create Plugin](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#create-plugin)
Interpret configuration to create a instance of a SDK extension plugin interface.
- **Parameters**: `properties` - The [`ConfigProperties`](https://opentelemetry.io/docs/specs/otel/configuration/api/#configproperties) representing the configuration specified for the component in the [configuration model](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#in-memory-configuration-model).
- **Returns**: A configured SDK extension plugin interface implementation.
- The plugin interface **MAY** have properties which are optional or required, and have specific requirements around type or format. The set of properties a `ComponentProvider` accepts, along with their requirement level and expected type, comprise a configuration schema. A `ComponentProvider` **SHOULD** document its configuration schema and include examples.
- When Create Plugin is invoked, the `ComponentProvider` interprets properties and attempts to extract data according to its configuration schema. If this fails (e.g. a required property is not present, a type is mismatches, etc.), Create Plugin SHOULD return an error.

## SDK Operations

- SDK implementations of configuration **MUST** provide the following operations.
- Note: Because these operations are stateless pure functions, they are not defined as part of any type, class, interface, etc. SDKs may organize them in whatever manner is idiomatic for the language.

### Parse

- **Parameters**:
  - `file`: The [configuration file](https://opentelemetry.io/docs/specs/otel/configuration/data-model/#file-based-configuration-model) to parse. This **MAY** be a file path, or language specific file data structure, or a stream of a file’s content.
  - `file_format`: The file format of the file (e.g. yaml). Implementations **MAY** accept a `file_format` parameter, or infer it from the file extension, or include file format specific overloads of parse, e.g. `parseYaml(file)`. If parse accepts `file_format`, the *API SHOULD* be structured so a user is obligated to provide it.
- **Returns**: [configuration model](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#in-memory-configuration-model)
- Parse **MUST** differentiate between properties that are missing and properties that are present but null. For example, consider the following snippet, noting `.meter_provider.views[0].stream.drop` is present but null:

```yaml
meter_provider:
  views:
    - selector:
        name: some.metric.name
      stream:
        aggregation:
          drop:
```

- As a result, the view stream should be configured with the `drop aggregation`. Note that some aggregations have additional arguments, but `drop` does not. The user **MUST** not be required to specify an empty object (i.e. `drop: {}`) in these cases.
- When encountering a reference to a [SDK extension component](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#sdk-extension-components) which is not built in to the `SDK`, Parse **MUST** resolve corresponding configuration to a generic [`ConfigProperties`](https://opentelemetry.io/docs/specs/otel/configuration/api/#configproperties) representation as described in [Create Plugin](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#create-plugin).

### Create

- Interpret configuration model and return SDK components.
- **Parameters**: `configuration` - An [in-memory configuration model](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#in-memory-configuration-model).
- **Returns**: Top level SDK components:
  - [`TracerProvider`](https://opentelemetry.io/docs/specs/otel/trace/sdk/#tracer-provider)
  - [`MeterProvider`](https://opentelemetry.io/docs/specs/otel/metrics/sdk/#meterprovider)
  - [`LoggerProvider`](https://opentelemetry.io/docs/specs/otel/logs/sdk/#loggerprovider)
  - [`Propagators`](https://opentelemetry.io/docs/specs/otel/context/api-propagators/#composite-propagator)
  - [`ConfigProvider`](https://opentelemetry.io/docs/specs/otel/configuration/sdk/#configprovider)
