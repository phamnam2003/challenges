# Overview

- `Baggage` is a set of application-defined properties contextually associated with a distributed request or workflow execution (see also the W3C Baggage Specification). Baggage can be used, among other things, to annotate telemetry, adding contextual information to metrics, traces, and logs.
- In OpenTelemetry Baggage is represented as a set of name/value pairs describing user-defined properties. Each name in Baggage MUST be associated with exactly one value. This is more restrictive than the W3C Baggage Specification, § 3.2.1.1 which allows duplicate entries for a given name.
- Baggage names are any valid, non-empty UTF-8 strings. Language API SHOULD NOT restrict which strings are used as baggage names. However, the specific Propagators that are used to transmit baggage entries across component boundaries may impose their own restrictions on baggage names. For example, the W3C Baggage specification restricts the baggage keys to strings that satisfy the token definition from RFC7230, Section 3.2.6. For maximum compatibility, alpha-numeric names are strongly recommended to be used as baggage names.
- Baggage values are any valid UTF-8 strings. Language API MUST accept any valid UTF-8 string as baggage value in Set and return the same value from Get.
- Language API MUST treat both baggage names and values as case sensitive. See also W3C Baggage Rationale.

```go
baggage.Set('a', 'B% 💼');
baggage.Set('A', 'c');
baggage.Get('a'); // returns "B% 💼"
baggage.Get('A'); // returns "c"
```

- The Baggage API consists of:
  - the Baggage as a logical container
  - functions to interact with the Baggage in a Context
- The functions described here are one way to approach interacting with the `Baggage` via having struct/object that represents the entire Baggage content. Depending on language idioms, a language API MAY implement these functions by interacting with the baggage via the `Context` directly.
- The Baggage API MUST be fully functional in the absence of an installed SDK. This is required in order to enable transparent cross-process Baggage propagation. If a Baggage propagator is installed into the API, it will work with or without an installed SDK.
- The `Baggage` container MUST be immutable, so that the containing `Context` also remains immutable.

# Operations

## Get Value

- To access the value for a name/value pair set by a prior event, the Baggage API MUST provide a function that takes the name as input, and returns a value associated with the given name, or null if the given name is not present.
REQUIRED parameters:
`Name` the name to return the value for.

## Get All Values

## Set Value

## Remove Value

# Context Interaction

## Clear Baggage in the Context

# Propagation

# Conflict Resolution
