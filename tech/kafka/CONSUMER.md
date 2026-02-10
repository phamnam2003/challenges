# Kafka Consumer: Configuration Principles

## Core Concepts

**Kafka Consumer**: Client that **pulls** events from topics, processes them, and tracks position via **offsets**.

**Consumer Group**: Logical group where each partition assigned to **one consumer**. Enables parallel processing and fault tolerance.

**Offset**: Consumer's position in partition. Stored in `__consumer_offsets` topic.

**Delivery Semantics**:
- **At-most-once**: Commit before processing (risk: loss)
- **At-least-once**: Commit after processing (risk: duplicates) - *most common*
- **Exactly-once**: Transactional consumers + producers

**Rebalancing**: Partition reassignment when consumers join/leave. Use `CooperativeStickyAssignor` for minimal disruption.

**Key Rule**: `max_parallelism = partition_count` (can't have more consumers than partitions in a group)

## Configuration Principles

**The Tradeoff Triangle**: Throughput ↔ Latency ↔ Reliability

| Priority | Configuration Approach |
|----------|----------------------|
| **High Throughput** | Large batches, less frequent commits |
| **Low Latency** | Small batches, frequent polls |
| **High Reliability** | Manual commits, careful error handling |

**Configuration Strategy**:
1. Start with **safe defaults** (manual commits, small batches)
2. **Measure** actual performance (lag, throughput)
3. **Tune incrementally** (one parameter at a time)
4. **Monitor consumer lag** (most critical metric)

## Essential Configurations

### Required

| Config | Value | Purpose |
|--------|-------|---------|
| `bootstrap.servers` | `broker1:9092,broker2:9092` | Broker addresses (list 2-3) |
| `group.id` | `order-processing-service` | **Critical**: Consumer group identifier |
| `key.deserializer` | `StringDeserializer` | Key deserialization |
| `value.deserializer` | `StringDeserializer` | Value deserialization |
| `client.id` | `order-processor-1` | Instance identifier |

### Offset Management

| Config | Recommended | Purpose |
|--------|-------------|---------|
| `enable.auto.commit` | `false` | **Manual commits** for control |
| `auto.offset.reset` | `earliest` | Start from beginning if no offset |

**Auto vs Manual Commits**:
- **Manual** (`false`): Commit after successful processing - **at-least-once** semantics
- **Auto** (`true`): Commit periodically - risk of loss or duplicates

### Fetch Configuration

| Config | Default | Low Latency | High Throughput |
|--------|---------|-------------|-----------------|
| `fetch.min.bytes` | `1` | `1` | `50000` |
| `fetch.max.wait.ms` | `500` | `100` | `1000` |
| `max.partition.fetch.bytes` | `1048576` (1 MB) | `524288` | `2097152` |
| `max.poll.records` | `500` | `100` | `1000` |

**How it works**:
- `fetch.min.bytes`: Broker waits to accumulate this much data
- `fetch.max.wait.ms`: Max wait time if `fetch.min.bytes` not reached
- `max.poll.records`: Max events per `poll()` call

### Session Management

| Config | Default | Recommended | Purpose |
|--------|---------|-------------|---------|
| `session.timeout.ms` | `10000` (10s) | `45000` (45s) | Heartbeat timeout before rebalance |
| `heartbeat.interval.ms` | `3000` (3s) | `3000` (3s) | Heartbeat frequency |
| `max.poll.interval.ms` | `300000` (5m) | `300000` (5m) | Max time between polls |

**Critical Rule**: `heartbeat.interval.ms` < `session.timeout.ms` < `max.poll.interval.ms`

**`max.poll.interval.ms`**: Set based on **max processing time** for batch. If exceeded, consumer kicked from group.

### Rebalancing

| Config | Recommended Value |
|--------|------------------|
| `partition.assignment.strategy` | `CooperativeStickyAssignor` |

**Strategies**:
- `RangeAssignor`: Contiguous ranges (default, unbalanced)
- `CooperativeStickyAssignor`: **Recommended** - incremental rebalancing, minimal disruption

### Isolation (Transactions)

| Config | Default | When to Use |
|--------|---------|-------------|
| `isolation.level` | `read_uncommitted` | Default |
| | `read_committed` | With transactional producers for exactly-once |

## Configuration Patterns

| Pattern | Use Case | Key Settings |
|---------|----------|-------------|
| **Real-Time** | Notifications, dashboards | `max.poll.records=100`, `fetch.min.bytes=1`, `fetch.max.wait.ms=100` |
| **Batch Processing** | Analytics, ETL | `max.poll.records=1000`, `fetch.min.bytes=50000`, `max.poll.interval.ms=600000` |
| **Mission-Critical** | Payments, finance | `max.poll.records=50`, `isolation.level=read_committed`, manual commits |
| **Balanced** | General services | `max.poll.records=500`, `session.timeout.ms=45000`, cooperative rebalancing |

### Real-Time Processing (Low Latency)
```
group.id=realtime-processor
enable.auto.commit=false
max.poll.records=100
fetch.min.bytes=1
fetch.max.wait.ms=100
session.timeout.ms=10000
max.poll.interval.ms=60000
partition.assignment.strategy=CooperativeStickyAssignor
```

### Batch Processing (High Throughput)
```
group.id=batch-analytics
enable.auto.commit=false
max.poll.records=1000
fetch.min.bytes=50000
fetch.max.wait.ms=1000
session.timeout.ms=45000
max.poll.interval.ms=600000
partition.assignment.strategy=CooperativeStickyAssignor
```

### Mission-Critical (High Reliability)
```
group.id=payment-processor
enable.auto.commit=false
max.poll.records=50
isolation.level=read_committed
session.timeout.ms=45000
max.poll.interval.ms=300000
partition.assignment.strategy=CooperativeStickyAssignor
```

### Balanced (Recommended Default)
```
group.id=order-service
enable.auto.commit=false
max.poll.records=500
fetch.min.bytes=1
fetch.max.wait.ms=500
session.timeout.ms=45000
max.poll.interval.ms=300000
partition.assignment.strategy=CooperativeStickyAssignor
```

## Best Practices

**Offset Commits**:
- ✅ Commit **after** processing (at-least-once)
- ❌ Avoid committing **before** processing (risk: loss)
- Make processing **idempotent** (handle duplicates)

**Error Handling**:
- Transient errors → retry with backoff
- Permanent errors → skip and log, or dead letter queue

**Rebalancing**:
- Use `CooperativeStickyAssignor` for minimal disruption
- Set `session.timeout.ms=45000` for stability
- Ensure processing stays within `max.poll.interval.ms`

**Scaling**:
- Max consumers = partition count (more = idle consumers)
- Monitor **consumer lag** (most critical metric)

**Thread Safety**:
- ❌ Consumers are **NOT thread-safe**
- ✅ One consumer per thread

**Lifecycle**:
- On shutdown: finish processing → commit offsets → `close()` consumer

## Troubleshooting Guide

| Problem | Solution |
|---------|----------|
| **High consumer lag** | Add consumers (up to partition count), optimize processing, increase `max.poll.records` |
| **Frequent rebalancing** | Increase `max.poll.interval.ms` / `session.timeout.ms`, reduce `max.poll.records`, use cooperative rebalancing |
| **Consumer stuck** | Check thread alive, add processing timeout, verify `poll()` frequency |
| **Duplicate processing** | Make processing idempotent, commit smaller batches, use exactly-once semantics |
| **Message loss** | Use manual commits, commit **after** processing, implement dead letter queue |

## Quick Reference

### Recommended Configuration
```properties
# Essential
bootstrap.servers=broker1:9092,broker2:9092,broker3:9092
group.id=my-consumer-group
client.id=my-service-1
key.deserializer=org.apache.kafka.common.serialization.StringDeserializer
value.deserializer=org.apache.kafka.common.serialization.StringDeserializer

# Offset Management
enable.auto.commit=false
auto.offset.reset=earliest

# Fetch
fetch.min.bytes=1
fetch.max.wait.ms=500
max.partition.fetch.bytes=1048576
max.poll.records=500

# Session
session.timeout.ms=45000
heartbeat.interval.ms=3000
max.poll.interval.ms=300000

# Rebalancing
partition.assignment.strategy=org.apache.kafka.clients.consumer.CooperativeStickyAssignor

# Security (production)
security.protocol=SASL_SSL
sasl.mechanism=SCRAM-SHA-256
```

### Consumption Flow
```
1. consumer.subscribe(topics)
2. Loop:
   - messages = consumer.poll(100ms)
   - Process messages
   - consumer.commit()
3. On shutdown:
   - Commit offsets
   - consumer.close()
```

### Key Monitoring Metrics
- `records-lag-max` - Consumer lag (critical)
- `records-consumed-rate` - Throughput
- `fetch-latency-avg` - Fetch performance
- `commit-latency-avg` - Commit performance

### Common CLI Commands
```bash
# Check consumer group status
kafka-consumer-groups --describe --group my-group

# Reset offsets to beginning
kafka-consumer-groups --reset-offsets --to-earliest --group my-group --topic orders --execute

# Reset to timestamp
kafka-consumer-groups --reset-offsets --to-datetime 2024-01-01T00:00:00.000 --group my-group --topic orders --execute
```

## Key Takeaways

1. **Pull-based consumption** - Consumer controls rate and backpressure
2. **Consumer groups** - Enable parallel processing and fault tolerance
3. **Manual commits** - Commit **after** processing for at-least-once semantics
4. **Rebalancing** - Use `CooperativeStickyAssignor` for minimal disruption
5. **Consumer lag** - Most critical metric to monitor
6. **Idempotent processing** - Essential for at-least-once delivery
7. **Thread safety** - One consumer per thread (not thread-safe)
8. **Proper shutdown** - Commit offsets and close consumer cleanly
9. **Scaling limit** - Max consumers = partition count
10. **Configuration tradeoffs** - Balance throughput, latency, reliability based on use case
