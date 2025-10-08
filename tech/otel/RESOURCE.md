# OpenTelemetry Resource Theory

## Core Concept

A Resource defines the source of telemetry data - the immutable entity (service, host, container) that produces observability signals.

## Fundamental Principles

### Resource Identity

- **Immutable attributes** that don't change during process lifetime
- **Unique identification** of telemetry source
- **Hierarchical composition** from multiple detectors

### Key Attributes

- **service.name**: Logical service name (required)
- **service.instance.id**: Unique instance identifier (required)  
- **deployment.environment**: prod/staging/dev context
- **Infrastructure**: host, OS, container, cloud metadata

## Architecture

### Detection Layers

- Service → Process → Container → Host → Cloud

### Detection Methods

- **Automatic**: Environment variables, platform APIs, file inspection
- **Manual**: Explicit configuration by developers
- **Hybrid**: Auto-detection with manual overrides

## 🔧 Key Features

### Composition

- Multiple resource detectors merge attributes
- Conflict resolution: manual > automatic
- Layered context accumulation

### Context Propagation

- Resource attributes attached to all telemetry
- Enables source identification across distributed systems
- Supports aggregation and correlation

## 🎯 Purpose

- **Identify** where telemetry originated
- **Contextualize** data with environmental metadata  
- **Correlate** signals across services and infrastructure
- **Attribute** costs and ownership

## 📊 Output

All telemetry data (traces, metrics, logs) carries resource context, enabling unified analysis across your observability stack.
