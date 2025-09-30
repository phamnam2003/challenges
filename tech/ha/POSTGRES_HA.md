# PostgresSQL High Availability

## Introduction

- Use `Patroni` for PostgreSQL high availability. `Patroni` is a template for `PostgreSQL` HA with `ZooKeeper`, `etcd`, or `Consul`As mentioned ab as a distributed configuration store.

## Components

### ETCD

- `etcd` is one of the key components in high availability architecture, therefore, it’s important to understand it.
- `etcd` is a *distributed key-value consensus store* that helps applications store and manage cluster configuration data and *perform distributed coordination* of a `PostgreSQL` cluster.
- `etcd` runs as a cluster of nodes that communicate with each other to maintain a consistent state. The **primary node** in the cluster is called the “leader”, and the remaining nodes are the “followers”.
- `etcd lock` - `etcd` provides a *distributed locking mechanism*, which helps applications coordinate actions across multiple nodes and access to shared resources preventing conflicts. Locks ensure that only one process can hold a resource at a time, *avoiding race conditions* and *inconsistencies*. `Patroni` is an example of an application that uses `etcd` locks for primary election control in the `PostgreSQL` cluster.

### PATRONI

- `Patroni` is *an open-source tool* designed to manage and automate the *high availability (HA)* of `PostgreSQL` clusters. It ensures that your `PostgreSQL` database remains available even in the event of hardware failures, network issues or other disruptions. `Patroni` achieves this by using distributed consensus stores like `ETCD`, `Consul`, or `ZooKeeper` to manage cluster state and automate failover processes. We’ll use `etcd` in our architecture.
- Key benefits:
  - *Automated failover* and **promotion** of a new primary in case of a failure;
  - *Prevention* and *protection* from split-brain scenarios (where two nodes believe they are the primary and both accept transactions). Split-brain can lead to serious logical corruptions such as wrong, duplicate data or data loss, and to associated business loss and risk of litigation;
  - Simplifying the management of PostgreSQL clusters across multiple data centers;
  - Self-healing via automatic restarts of failed PostgreSQL instances or reinitialization of broken replicas.
  - Integration with tools like pgBackRest, HAProxy, and monitoring systems for a complete HA solution.

### HAProxy

- `HAProxy` (High Availability Proxy) is a *powerful, open-source* load balancer and proxy server used to improve the *performance and reliability* of web services by distributing network traffic across multiple servers. It is widely used to enhance the `scalability`, `availability`, and `reliability` of web applications by *balancing client requests* among backend servers.
- `HAProxy` architecture is optimized to move data as fast as possible with the least possible operations. It focuses on optimizing the CPU cache’s efficiency by sticking connections to the same CPU as long as possible.

### pgBackRest

- `pgBackRest` is *an advanced backup and restore tool* designed specifically for `PostgreSQL` databases. `pgBackRest` emphasizes simplicity, speed, and scalability. Its architecture is focused on minimizing the time and resources required for both backup and restoration processes.
- `pgBackRest` uses a custom protocol, which allows for more flexibility compared to traditional tools like **tar** and **rsync** and limits the types of connections that are required to perform a backup, thereby increasing security. `pgBackRest` is a simple, but feature-rich, reliable backup and restore system that can seamlessly scale up to the largest databases and workloads.

## Installation

- [Docs Percona](https://docs.percona.com/postgresql/17/solutions/ha-init-setup.html)
