# Kafka Producer: Principles and Configuration

## Producer Theory

### What is a Kafka Producer?

A **Kafka Producer** is a client application that publishes `events` to Kafka `topics`. It transforms application state changes into durable, distributed messages.

**Core Responsibilities**:
- **Serialize** domain objects to byte arrays
- **Route** events to correct partition
- **Batch** messages for efficiency
- **Ensure delivery** with configured guarantees

### Six Core Principles

#### 1. Asynchronous by Design

- `send()` **doesn't immediately** transmit to broker
- Events **buffered in memory**, sent in batches
- **Non-blocking** operation for high throughput
- **Callbacks** notify success/failure

**Why**: Synchronous = *hundreds/second*. Asynchronous = *millions/second*.

#### 2. Batching for Efficiency

- **Multiple events** sent together in one request
- Reduces **network overhead** and **broker load**
- Improves **compression** ratio
- **Trade-off**: Adds latency for higher throughput

#### 3. Ordering Within Partitions

- **Strict ordering** within a partition only
- Events with **same key** → **same partition**
- **No ordering** across different partitions

**Use keys when**: Related events need ordering (user actions, order lifecycle).

#### 4. Delivery Guarantees

**At-most-once** (`acks=0`): Fastest, can lose messages
**At-least-once** (`acks=1`): No loss, may duplicate (on retry)
**Exactly-once** (`acks=all` + idempotence): No loss, no duplicates

#### 5. Idempotence and Transactions

**Idempotence**: Producer deduplicates retries automatically
**Transactions**: Atomic writes across multiple partitions/topics

**Enable by default** for data integrity.

#### 6. Backpressure and Flow Control

- **Buffer limit** prevents unbounded memory growth
- When full, `send()` **blocks** or throws exception
- Forces application to **slow down** if needed

### Producer Architecture

**Internal Components**:
- **Serializer**: Object → byte array
- **Partitioner**: Selects target partition
- **Record Accumulator**: In-memory buffer, batches events
- **Sender Thread**: Background I/O, sends batches to brokers
- **Metadata Manager**: Caches cluster topology

**Request Flow**:
`send()` → Serialize → Partition → Buffer → Batch → Send → Acknowledge → Callback

## Configuration Principles

### The Performance Triangle

```
        Throughput
           /\
          /  \
         /    \
   Latency -- Durability
```

**Choose two**—you can't optimize all three:
- **High throughput + Low latency** = Lower durability
- **High throughput + High durability** = Higher latency
- **Low latency + High durability** = Lower throughput

### Configuration Strategy

1. **Start with safe defaults** (prioritize correctness)
2. **Measure** actual performance
3. **Identify bottleneck**
4. **Tune one parameter** at a time
5. **Measure again**

**Don't guess—measure first, then optimize.**

### Producer Instance Management

**Critical Rule**: Producers are **thread-safe** and **expensive to create**.

✅ **Correct**: One producer per application, shared across threads
❌ **Wrong**: Creating producer per message/request (resource exhaustion)

## Essential Configurations

### Required Settings

**`bootstrap.servers`**
```
bootstrap.servers=broker1:9092,broker2:9092,broker3:9092
```
List **2-3 brokers** for redundancy. Producer discovers rest automatically.

**`key.serializer` and `value.serializer`**
```
key.serializer=org.apache.kafka.common.serialization.StringSerializer
value.serializer=org.apache.kafka.common.serialization.StringSerializer
```
Common: `StringSerializer`, `AvroSerializer`, `ByteArraySerializer`

**`client.id`**
```
client.id=order-service-prod-1
```
Pattern: `{service}-{environment}-{instance}`

### Durability Configuration

**`acks` - Acknowledgment Level**

| Setting | Behavior | Latency | Durability | Use Case |
|---------|----------|---------|------------|----------|
| `acks=0` | No acknowledgment | Lowest | Can lose data | Metrics, non-critical logs |
| `acks=1` | Leader only | Medium | Loss if leader fails | General applications |
| `acks=all` | All ISR replicas | Highest | No data loss | Financial, critical events |

**Recommendation**: Use `acks=all` unless latency is unacceptable.

**`enable.idempotence`**
```
enable.idempotence=true
```
**Always enable**—prevents duplicates from retries with no performance cost.

**`retries` and `delivery.timeout.ms`**
```
retries=2147483647          # Retry until timeout
delivery.timeout.ms=120000  # 2 minutes total
```
Let producer **handle transient failures** automatically.

### Performance Configuration

**`batch.size` - Batch Size**
```
batch.size=16384    # 16 KB (low latency)
batch.size=32768    # 32 KB (balanced)
batch.size=1048576  # 1 MB (high throughput)
```

**`linger.ms` - Batching Delay**
```
linger.ms=0    # Send immediately
linger.ms=10   # Wait up to 10ms (balanced)
linger.ms=100  # Wait up to 100ms (high throughput)
```

**`compression.type`**
```
compression.type=snappy  # Good default
compression.type=lz4     # Faster
compression.type=zstd    # Best compression
compression.type=gzip    # High ratio, CPU heavy
```
**Recommendation**: Start with `snappy`.

**`buffer.memory`**
```
buffer.memory=67108864  # 64 MB (typical)
```
Increase if hitting buffer limits. Monitor `buffer-available-bytes` metric.

**`max.in.flight.requests.per.connection`**
```
max.in.flight.requests.per.connection=5  # Default with idempotence
```
With idempotence enabled, use **5** for throughput while maintaining ordering.

### Transaction Configuration

**`transactional.id`**
```
transactional.id=payment-service-instance-1
```
Enables **exactly-once** across partitions. Must be **unique per instance**.

**Use for**: Read-process-write pipelines, atomic multi-partition writes.

### Security Configuration

**`security.protocol`**
```
security.protocol=SASL_SSL  # Authentication + Encryption (recommended)
```

**`sasl.mechanism`**
```
sasl.mechanism=SCRAM-SHA-256  # Modern, secure
```

## Configuration Patterns

### Pattern 1: High-Throughput Ingestion

**Use case**: Logs, metrics, clickstreams (millions/second)

```
acks=1
batch.size=1048576          # 1 MB
linger.ms=100
compression.type=lz4
buffer.memory=134217728     # 128 MB
```

### Pattern 2: Real-Time Events

**Use case**: User actions, notifications (low latency critical)

```
acks=1
batch.size=16384            # 16 KB
linger.ms=0
compression.type=snappy
enable.idempotence=true
```

### Pattern 3: Critical Transactions

**Use case**: Payments, financial data (no data loss)

```
acks=all
enable.idempotence=true
transactional.id=service-instance-id
delivery.timeout.ms=300000  # 5 min
batch.size=32768
linger.ms=10
```

### Pattern 4: Balanced Production (Recommended)

**Use case**: General backend services

```
acks=all
enable.idempotence=true
batch.size=32768            # 32 KB
linger.ms=10
compression.type=snappy
buffer.memory=67108864      # 64 MB
delivery.timeout.ms=120000  # 2 min
```

## Backend Developer Best Practices

### 1. Lifecycle Management

**✅ Do**: Create **once**, reuse across threads, close on shutdown
**❌ Don't**: Create per message or per request

### 2. Error Handling

**Retriable Errors** (auto-retry):
- `NetworkException`, `LeaderNotAvailableException`, `TimeoutException`

**Non-Retriable Errors** (handle in app):
- `RecordTooLargeException`, `AuthorizationException`, `InvalidTopicException`

**Always implement callbacks** to handle failures.

### 3. Monitoring

**Critical Metrics**:
- `record-error-rate`: Should be near 0
- `buffer-available-bytes`: Buffer exhaustion indicator
- `request-latency-avg`: End-to-end latency
- `record-retry-rate`: High = broker issues

**Alert on**: Error rate > 1%, buffer < 10%, latency > SLA

### 4. Partitioning Strategy

**Use keys when**:
- Events are **related** and need **ordering**
- Implementing **stateful processing**
- Using **log compaction**

**Use null keys when**:
- Events are **independent**
- Want **even distribution**

**Examples**: 
- User events → `userId` key
- Order lifecycle → `orderId` key
- Metrics/logs → no key

### 5. Message Size

**Keep messages small**: < 100 KB (target), < 1 MB (max)
**Large data**: Store in object storage, send reference in Kafka

### 6. Schema Management

**Use Schema Registry** with Avro or Protobuf:
- **Type safety**
- **Schema evolution**
- **Backward/forward compatibility**

## Troubleshooting

### Low Throughput
**Solutions**: Increase `batch.size`, add `linger.ms > 0`, enable compression, reuse producer instance

### High Latency
**Solutions**: Reduce `linger.ms`, reduce `batch.size`, use `acks=1`, lighter compression

### Buffer Exhaustion
**Solutions**: Increase `buffer.memory`, increase `max.block.ms`, scale brokers, implement backpressure

### Message Loss
**Solutions**: Use `acks=all`, enable idempotence, implement error handling, call `Close()` on shutdown

### Duplicates
**Solutions**: Enable `enable.idempotence=true`, use transactions, make downstream idempotent

## Quick Reference

### Safe Production Defaults
```
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

## Key Takeaways

1. **Enable idempotence by default** for data integrity
2. **Use `acks=all`** unless proven bottleneck
3. **Reuse producer instances**—thread-safe, expensive to create
4. **Start with safe defaults**, optimize based on measurements
5. **Monitor everything**: Metrics reveal actual behavior
6. **Choose configs** based on your requirements: throughput vs latency vs durability
7. **Test under load** to validate behavior

The producer is powerful and flexible—understand the principles and trade-offs to configure it correctly for your use case.
