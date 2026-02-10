# Kafka Producer: Configuration Principles

## Core Concepts

**Kafka Producer**: Client that publishes `events` to topics. Serializes objects, routes to partitions, batches for efficiency.

**Key Characteristics**:
- **Asynchronous**: `send()` buffers in memory, sends in batches (non-blocking)
- **Batching**: Multiple events per request (reduces overhead, improves compression)
- **Ordering**: Strict within partition only (same key → same partition)
- **Idempotent**: Deduplicates retries automatically (enable by default)

**Delivery Guarantees**:
- **At-most-once** (`acks=0`): Fast, can lose messages
- **At-least-once** (`acks=1`): No loss, may duplicate
- **Exactly-once** (`acks=all` + idempotence): No loss, no duplicates

**Architecture**: `send()` → Serialize → Partition → Buffer → Batch → Send → Acknowledge → Callback

## Configuration Principles

**The Tradeoff Triangle**: Throughput ↔ Latency ↔ Durability

| Priority | Configuration Approach |
|----------|----------------------|
| **High Throughput** | Large batches (`batch.size=1MB`), wait for batches (`linger.ms=100`), compression |
| **Low Latency** | Small batches (`batch.size=16KB`), send immediately (`linger.ms=0`), no compression |
| **High Durability** | `acks=all`, `enable.idempotence=true`, retries enabled |

**Configuration Strategy**:
1. Start with **safe defaults** (correctness first)
2. **Measure** actual performance
3. **Tune one parameter** at a time
4. **Measure again**

**Producer Instance**: Thread-safe and expensive to create. ✅ One per application, shared across threads. ❌ Never create per message/request.

## Essential Configurations

### Required

| Config | Value | Purpose |
|--------|-------|---------|
| `bootstrap.servers` | `broker1:9092,broker2:9092` | Broker addresses (list 2-3) |
| `key.serializer` | `StringSerializer` | Key serialization |
| `value.serializer` | `StringSerializer` | Value serialization |
| `client.id` | `order-service-prod-1` | Instance identifier |

### Durability

| Config | Default | Recommended | Purpose |
|--------|---------|-------------|---------|
| `acks` | `1` | `all` | Wait for all ISR replicas (no data loss) |
| `enable.idempotence` | `false` | `true` | **Always enable** - prevents duplicates |
| `retries` | `2147483647` | `2147483647` | Retry until timeout |
| `delivery.timeout.ms` | `120000` (2m) | `120000` | Total delivery timeout |

**Acks Comparison**:

| Setting | Behavior | Durability | Use Case |
|---------|----------|------------|----------|
| `acks=0` | No acknowledgment | Can lose data | Metrics, non-critical logs |
| `acks=1` | Leader only | Loss if leader fails | General apps (default) |
| `acks=all` | All ISR replicas | No data loss | Financial, critical events |

### Performance

| Config | Low Latency | Balanced | High Throughput |
|--------|-------------|----------|-----------------|
| `batch.size` | `16384` (16 KB) | `32768` (32 KB) | `1048576` (1 MB) |
| `linger.ms` | `0` | `10` | `100` |
| `compression.type` | `none` | `snappy` | `zstd` |
| `buffer.memory` | `67108864` (64 MB) | `67108864` | `134217728` (128 MB) |

**How it works**:
- `batch.size`: Max bytes per batch
- `linger.ms`: Max wait time before sending incomplete batch
- `compression.type`: `snappy` (balanced), `lz4` (fast), `zstd` (best ratio), `gzip` (CPU heavy)
- `buffer.memory`: Total memory for buffering

**`max.in.flight.requests.per.connection`**
```
max.in.flight.requests.per.connection=5  # With idempotence (default)
```
With idempotence enabled, use **5** for throughput while maintaining ordering.

### Transactions (Exactly-Once)

| Config | Value | Purpose |
|--------|-------|---------|
| `transactional.id` | `payment-service-1` | **Unique per instance** - enables exactly-once |

Use for: Read-process-write pipelines, atomic multi-partition writes.

**`security.protocol`**
```
security.protocol=SASL_SSL  # Authentication + Encryption (recommended)
```

## Configuration Patterns

| Pattern | Use Case | Key Settings |
|---------|----------|-------------|
| **High Throughput** | Logs, metrics, clickstreams | `batch.size=1MB`, `linger.ms=100`, `compression=lz4`, `acks=1` |
| **Real-Time** | User actions, notifications | `batch.size=16KB`, `linger.ms=0`, `acks=1`, no compression |
| **Critical** | Payments, financial data | `acks=all`, `idempotence=true`, transactions enabled |
| **Balanced** | General services | `acks=all`, `batch.size=32KB`, `linger.ms=10`, `compression=snappy` |

### High-Throughput Ingestion
```
acks=1
batch.size=1048576
linger.ms=100
compression.type=lz4
buffer.memory=134217728
```

### Real-Time Events (Low Latency)
```
acks=1
batch.size=16384
linger.ms=0
compression.type=snappy
enable.idempotence=true
```

### Critical Transactions (No Loss)
```
acks=all
enable.idempotence=true
transactional.id=service-instance-id
delivery.timeout.ms=300000
batch.size=32768
linger.ms=10
```

### Balanced (Recommended Default)
```
acks=all
enable.idempotence=true
batch.size=32768
linger.ms=10
compression.type=snappy
buffer.memory=67108864
delivery.timeout.ms=120000
```

## Best Practices

**Lifecycle Management**:
- ✅ Create **once**, reuse across threads (thread-safe)
- ❌ Never create per message/request
- Close on shutdown to flush buffers

**Error Handling**:
- **Retriable** (auto-retry): `NetworkException`, `TimeoutException`
- **Non-retriable** (handle in app): `RecordTooLargeException`, `AuthorizationException`
- Always implement callbacks for failures

**Partitioning**:
- Use **keys** when: Related events need ordering (userId, orderId)
- Use **null keys** when: Independent events, want even distribution

**Message Size**:
- Target: < 100 KB
- Max: < 1 MB
- Large data: Store elsewhere, send reference

**Schema Management**:
- Use **Schema Registry** with Avro/Protobuf
- Ensure backward/forward compatibility

## Troubleshooting Guide

| Problem | Solution |
|---------|----------|
| **Low throughput** | Increase `batch.size`, add `linger.ms > 0`, enable compression, verify producer reuse |
| **High latency** | Reduce `linger.ms`, reduce `batch.size`, use `acks=1`, lighter compression |
| **Buffer exhaustion** | Increase `buffer.memory`, implement backpressure, scale brokers |
| **Message loss** | Use `acks=all`, enable idempotence, implement error callbacks, call `close()` on shutdown |
| **Duplicates** | Enable `enable.idempotence=true`, use transactions, make consumers idempotent |

## Monitoring Metrics

**Critical Metrics**:
- `record-error-rate` - Should be near 0
- `buffer-available-bytes` - Buffer exhaustion indicator
- `request-latency-avg` - End-to-end latency
- `record-retry-rate` - High = broker issues

**Alerts**: Error rate > 1%, buffer < 10%, latency > SLA

## Quick Reference

### Recommended Configuration
```properties
# Essential
bootstrap.servers=broker1:9092,broker2:9092,broker3:9092
client.id=my-service-prod-1
key.serializer=org.apache.kafka.common.serialization.StringSerializer
value.serializer=org.apache.kafka.common.serialization.StringSerializer

# Durability
acks=all
enable.idempotence=true
retries=2147483647
delivery.timeout.ms=120000

# Performance
batch.size=32768
linger.ms=10
compression.type=snappy
buffer.memory=67108864
max.in.flight.requests.per.connection=5

# Security (production)
security.protocol=SASL_SSL
sasl.mechanism=SCRAM-SHA-256
```

### Production Flow
```
1. Create producer once (expensive, thread-safe)
2. Send messages:
   - Async: producer.send(record, callback)
   - Sync: producer.send(record).get()
3. On shutdown:
   - producer.flush()  # Wait for in-flight
   - producer.close()  # Clean shutdown
```

## Key Takeaways

1. **Asynchronous by design** - Buffers and batches for efficiency
2. **Always enable idempotence** - Prevents duplicates with no cost
3. **Use `acks=all`** for durability - Unless latency critical
4. **One producer per application** - Thread-safe, expensive to create
5. **Batch size + linger.ms** control throughput/latency tradeoff
6. **Compression** improves throughput (start with `snappy`)
7. **Keys** enable ordering within partition (same key → same partition)
8. **Implement callbacks** - Don't ignore send failures
9. **Monitor buffer** - Prevent memory exhaustion
10. **Close on shutdown** - Flush pending messages
