# 🚀 Challenges Repository

> A collection of ***algorithm solutions***, ***design patterns***, ***concurrency examples***, and ***technology explorations*** in **Go**.

---

## 📦 Packages Overview

### 🧠 **LeetCode** | Algorithm Solutions

***18+ interview-ready algorithm solutions*** with clean, optimized Go implementations. Perfect for ***problem-solving mastery***.

**Key Features:**
- ✨ Complete solutions with comprehensive test cases
- ✨ Time & space complexity analysis
- ✨ Multiple approaches per problem
- ✨ Real interview patterns & tips
- ✨ Guided learning path (Easy → Medium → Hard)

**Problem Categories:**
- 🟢 **Easy (13)** — Arrays, Hash Maps, Strings, Basics
- 🟠 **Medium (5)** — DP, Stacks, Heaps, Design

**All Problems & Algorithms:**

| # | Title | Difficulty | Algorithm / Data Structure |
|:---:|:---|:---:|:---|
| 1 | Two Sum | Easy | Hash Map |
| 21 | Merge Sorted Lists | Easy | Linked List, Merge |
| 53 | Maximum Subarray | Medium | Dynamic Programming (Kadane's) |
| 146 | LRU Cache | Medium | Hash Map + Doubly Linked List |
| 167 | Two Sum II | Easy | Two Pointers |
| 198 | House Robber | Medium | Dynamic Programming |
| 206 | Reverse Linked List | Easy | Linked List, Recursion |
| 242 | Valid Anagram | Easy | Hash Map, Frequency Counting |
| 347 | Top K Frequent Elements | Medium | Heap, Min-Heap |
| 509 | Fibonacci Number | Easy | Recursion, Memoization |
| 1431 | Kids With Candies | Easy | Array, Iteration |
| 1436 | Destination City | Easy | Hash Set, Graph |
| 1441 | Build Array With Stack | Medium | Stack, String Manipulation |
| 1480 | Running Sum | Easy | Prefix Sum, Array |
| 1486 | XOR Operation | Easy | Bitwise Operations |
| 1491 | Average Salary | Easy | Array, Math |
| 1496 | Path Crossing | Easy | Hash Set, Coordinate Tracking |
| 1512 | Number of Good Pairs | Easy | Hash Map, Combinatorics |

**Algorithms & Techniques Used:**
- 🔹 **Hash Tables** (Hash Map, Hash Set) — O(1) lookup, frequency counting
- 🔹 **Two Pointers** — Converging/diverging approaches for sorted arrays
- 🔹 **Dynamic Programming** — Kadane's, Memoization, State optimization
- 🔹 **Linked Lists** — Node traversal, reversal, merge operations
- 🔹 **Heaps** — Min-Heap, Max-Heap for top-K problems
- 🔹 **Bit Operations** — XOR, AND, OR for efficient computation
- 🔹 **Stack** — LIFO structure for nested operations
- 🔹 **Recursion** — Base cases, induction, memoization
- 🔹 **Prefix Sum** — Running totals for range operations

**[→ LeetCode Guide](./leetcode/README.md)** — Detailed explanations, complexity analysis, and interview patterns

---

### 🎨 **Design Patterns** | OOP & Architecture

***20+ Gang of Four Design Patterns*** with practical, real-world Go implementations. Learn to build ***flexible, maintainable systems***.

**Behavioral Patterns** (9) — Object interaction & responsibility distribution:
| Pattern | Use Case |
|:---|:---|
| **Chain of Responsibility** | Sequential handler pipeline |
| **Command** | Encapsulate request as object |
| **Iterator** | Safe collection traversal |
| **Mediator** | Centralize object communication |
| **Memento** | Capture & restore state |
| **Observer** | Notify multiple observers |
| **State** | Alter behavior based on state |
| **Strategy** | Encapsulate algorithms |
| **Template Method** | Define algorithm skeleton |
| **Visitor** | Add operations to objects |

**Creational Patterns** (4) — Object creation mechanisms:
| Pattern | Use Case |
|:---|:---|
| **Abstract Factory** | Create families of objects |
| **Builder** | Construct complex objects |
| **Factory Method** | Create objects via methods |
| **Singleton** | Single instance guarantee |

**Structural Patterns** (7) — Object composition & relationships:
| Pattern | Use Case |
|:---|:---|
| **Adapter** | Make incompatible objects work |
| **Bridge** | Decouple abstraction from implementation |
| **Composite** | Treat individual & group uniformly |
| **Decorator** | Add behavior dynamically |
| **Facade** | Simplify complex subsystems |
| **Flyweight** | Share object instances efficiently |
| **Proxy** | Control access to objects |

**Key Concepts:**
- 🎯 SOLID principles (Open/Closed, Single Responsibility, etc.)
- 🎯 Decoupling & loose coupling
- 🎯 Flexible state & behavior management
- 🎯 Composition over inheritance

**[→ Design Patterns Complete Guide](./pattern/README.md)** — Pattern explanations, real-world examples, when to use, and source code walkthrough

---

### ⚡ **Concurrency** | Go Goroutines & Channels

Master ***concurrent programming*** in Go with ***practical, production-ready patterns***. Build responsive, scalable systems.

**Patterns Covered:**

| Pattern | Use Case | Key Feature |
|:---|:---|:---|
| **PubSub** | Event distribution | Decouple producers & consumers |
| **Select** | Channel multiplexing | Handle multiple channels simultaneously |

**Core Concepts:**
- 🚀 Goroutines — Lightweight concurrent execution
- 📡 Channels — Type-safe communication
- ⏱️ Timeouts — Prevent indefinite waiting
- 🛑 Cancellation — Graceful shutdown
- 🔀 Multiplexing — Handle multiple events

**Real-World Applications:**
- Event-driven systems
- Rate limiting & worker pools
- Request multiplexing
- Non-blocking operations

**[→ Go Concurrency Guide](./concurrency/README.md)** — Pattern deep-dives, pitfalls, best practices, and concurrency mental models

---

### 🔧 **Tech** | Technology Deep Dives

***Comprehensive explorations*** of distributed systems, databases, observability, and infrastructure.

**Technologies Covered:**

| Tech | Use Case | Example |
|:---|:---|:---|
| **Kafka** | Message streaming & event distribution | Offset commit strategies, admin clients |
| **ScyllaDB** | High-performance distributed database | CQL queries, prepared statements, RCM |
| **OpenTelemetry** | Observability framework | Traces, metrics, logs collection |
| **Kubernetes** | Container orchestration | Ingress, pod management |
| **ELK/Loki** | Log aggregation & analysis | Log processing, visualization |
| **Consistent Hashing** | Distributed system partitioning | Load balancing, cache distribution |
| **Dependency Injection** | Loose coupling & testability | DI patterns, containers |

**[→ Kafka](./tech/kafka/)** — Offset committing, admin clients, message handling
**[→ ScyllaDB](./tech/scylladb/)** — Performance tuning, prepared statements, schema design
**[→ OpenTelemetry](./tech/otel/)** — Traces, metrics, logs, collector architecture
**[→ Kubernetes](./tech/k8s/)** — Ingress, pod lifecycle, networking
**[→ Logging](./tech/logging/)** — ELK stack, Loki, log aggregation strategies
**[→ Hashing](./tech/hashing/)** — Consistent hashing algorithms
**[→ DI](./tech/di/)** — Dependency injection patterns

---

## 🛠 Tech Stack

### **Language & Runtime**
- **Go** (1.18+) — Statically typed, compiled, concurrent language

### **Go Standard Library** 📚
| Package | Purpose |
|:---|:---|
| `sync` | Mutexes, WaitGroups, Once, RWMutex for goroutine synchronization |
| `atomic` | Atomic operations for thread-safe counters |
| `context` | Context propagation, cancellation, timeouts |
| `testing` | Unit testing, benchmarks, subtests |
| `io` | I/O abstractions, Reader/Writer interfaces |
| `fmt` | String formatting and printing |
| `strings` | String manipulation and operations |
| `math` | Mathematical functions |
| `time` | Time operations, timers, tickers |
| `sort` | Sorting algorithms |

### **Third-Party Libraries**

**Message Streaming & Distributed Systems:**
- [`franz-go`](https://github.com/twmb/franz-go) — High-performance Kafka client
- [`kadm`](https://pkg.go.dev/github.com/twmb/franz-go/pkg/kadm) — Kafka admin operations
- [`kmsg`](https://pkg.go.dev/github.com/twmb/franz-go/pkg/kmsg) — Kafka protocol messages

**Databases:**
- [`gocqlx`](https://github.com/scylladb/gocqlx) — ScyllaDB/Cassandra Go driver

**Observability:**
- [`go.opentelemetry.io`](https://opentelemetry.io/docs/instrumentation/go/) — OpenTelemetry SDK
- Traces, Metrics, Logs exporters

**Infrastructure:**
- Kubernetes API
- Docker
- Elasticsearch, Kibana (logging)
- Prometheus, Grafana (monitoring)

---

## 📖 Detailed Documentation

Each package includes ***comprehensive guides*** with explanations, examples, and learning paths:

### **Package Guides**
- 🧠 [LeetCode Solutions Guide](./leetcode/README.md) — 18+ problems, complexity analysis, interview patterns
- 🎨 [Design Patterns Master Guide](./pattern/README.md) — 20+ patterns (behavioral, creational, structural), real-world examples
- ⚡ [Go Concurrency Patterns](./concurrency/README.md) — PubSub, Select, best practices, common pitfalls
- 🔧 [Kafka Offset Commit Strategies](./tech/kafka/group_committing/README.md) — 5 commit styles explained

### **Quick Reference**
- ⏱️ [Complexity Cheat Sheet](https://www.bigocheatsheet.com/)
- 🎯 [Go Best Practices](https://golang.org/doc/effective_go)
- 📚 [Design Patterns (Refactoring Guru)](https://refactoring.guru/design-patterns)

---

## 🎯 Key Concepts

### Why This Repository?

This repository demonstrates:

| Concept | Purpose |
|:---|:---|
| **Problem-Solving** | Algorithmic thinking & optimization |
| **Design Patterns** | OOP principles & code reusability |
| **Concurrency** | Safe parallel processing in Go |
| **Distributed Systems** | Message streaming, databases, observability |

---