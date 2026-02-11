# Kafka Admin Client Requests

This package is a **focused demo** of Kafka admin operations with `franz-go`.

It shows the same admin intent through **two API levels**:
- `kadm` (*high-level admin API*)
- `kmsg` (*low-level Kafka protocol API*)

## Why this package exists

When building admin tools, teams usually face this question:
- choose **fast development** with simpler APIs
- or choose **full protocol control** for advanced needs

This package helps you understand that tradeoff quickly.

## What this demo does

In `main.go`, you select a mode with `-mode`:
- `kadm`: list topics, check if a topic exists, then create it.
- `kmsg`: build and send raw `CreateTopicsRequest` and `MetadataRequest`.

So you can compare *developer experience* and *control level* side by side.

## Quick Start

```bash
go run . -brokers localhost:9092 -topic admin-client-req-topic -mode kadm
```

```bash
go run . -brokers localhost:9092 -topic admin-client-req-topic -mode kmsg
```

**Flags**
- `-brokers`: comma-separated seed brokers (for example: `localhost:9092`)
- `-topic`: topic name to create/query
- `-mode`: `kadm` or `kmsg`

## `kadm` and `kmsg` in Practice

### `kadm`: high-level admin client

`kadm` is a **convenience admin layer** on top of `kgo`.

It is designed for common operations such as:
- topic management (*list/create/delete topics, partitions*)
- config/admin operations (*describe or alter settings*)
- operational tasks (*consumer groups, offsets, ACL workflows*)

It solves this problem:
- **"I need reliable admin tooling quickly without writing protocol boilerplate."**

Use `kadm` when you want **clean code**, **faster delivery**, and **safer defaults** for day-to-day admin tasks.

### `kmsg`: low-level protocol requests

`kmsg` exposes Kafka APIs as **request/response structs** that map closely to the wire protocol.

It is designed for cases where you need:
- precise control over request fields
- direct handling of protocol responses and error codes
- access to less common or highly specific API behavior

It solves this problem:
- **"I need exact protocol-level control that high-level helpers may abstract away."**

Use `kmsg` when you need **maximum flexibility**, **deep debugging**, or **advanced/custom admin workflows**.

## Rule of Thumb

- Start with `kadm` for most internal admin tools.
- Move to `kmsg` when requirements become protocol-specific.
- In complex systems, it is normal to use **both**: `kadm` for standard flows, `kmsg` for edge cases.
