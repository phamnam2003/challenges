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

## Config Provider

## SDK extension components

## SDK Operations

## Examples
