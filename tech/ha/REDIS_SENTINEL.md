# Redis Sentinel High Availability Setup

## Introduction

- [Redis Sentinel](https://redis.io/docs/latest/operate/oss_and_stack/management/sentinel/) provides high availability for Redis when not using [Redis Cluster](https://redis.io/docs/latest/operate/oss_and_stack/management/scaling/).
- Redis Sentinel also provides other collateral tasks such as monitoring, notifications and acts as a configuration provider for clients.
- This is the full list of Sentinel capabilities at a macroscopic level:
  - **Monitoring**
  - **Notification**
  - **Automatic failover**
  - **Configuration provider**

## Installation

- You can install `Redis Sentinel` using package managers like `apt`, `yum`, or `brew`, or by downloading the source code from the [official Redis website](https://redis.io/download).

```bash
sudo apt-get install redis-sentinel redis
```

## Configuration

- Running Sentinel requires a configuration file, typically named `sentinel.conf`. You can create this file manually or copy the default configuration file provided with the Redis installation.

```bash
redis-sentinel /path/to/sentinel.conf
```

- Otherwise you can use directly the redis-server executable starting it in Sentinel mode:

```bash
redis-server /path/to/sentinel.conf --sentinel
```

- Configure the Sentinel configuration file with the following parameters adjusted to your environment:

```bash
# /etc/redis/redis.conf
bind 0.0.0.0

masterauth <your_master_password> # password redis master connection

requirepass <your_sentinel_password> # password redis sentinel connection
```

- In slave nodes, configure the following parameters:

```bash
replica_of <master_ip> <master_port>
```

- Complete master slave with configuration above. But, when master down, slave will not be promoted to master without Sentinel.

## Configuration Sentinel

```bash
# /etc/redis/sentinel.conf
protected-mode no
port 26379

sentinel monitor mymaster <master_ip> <master_port> 2
sentinel auth-pass mymaster <your_master_password> # password redis master connection
sentinel down-after-milliseconds mymaster 5000
sentinel failover-timeout mymaster 10000
sentinel parallel-syncs mymaster 1
```

> [!Note]
> Adjust the `sentinel monitor` line to match your master Redis instance's IP address and port. The number `2` indicates that at least two Sentinels must agree that the master is down before a failover is initiated.
> Sentinel configuration file should be the same in all Sentinel nodes.
