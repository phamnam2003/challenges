# MariaDB with High Availability (HA)

- `MariaDB` is a popular *open-source relational database management system* that is a *fork* of `MySQL`. Setting up `MariaDB` for `high availability`` (HA) typically involves configuring a master-slave replication setup or using`Galera Cluster`for multi-master replication. Below are the steps to set up`MariaDB` with HA using Galera Cluster.

## MariaDB Galera Cluster Setup

- MariaDB Galera Cluster is a Linux-exclusive, multi-primary cluster designed for MariaDB, offering features such as active-active topology, read/write capabilities on any node, automatic membership and node joining, true parallel replication at the row level, and direct client connections, with an emphasis on the native MariaDB experience.
- Install `MariaDB` and `Galera` on all nodes in the cluster. Use the following commands to install them:

```bash
sudo apt-get install mariadb-server mariadb-client galera-4 rsync
```

- Configure the `MariaDB` configuration file (`/etc/mysql/conf.d/galera.cnf` or `/etc/mysql/mariadb.conf.d/galera.cnf`) on all nodes.

```ini
[mysql]
binlog_format=ROW
default_storage_engine=InnoDB # supported engines: InnoDB, MyISAM, Aria. InnoDB is recommended because support transaction, row-lock.
innodb_autoinc_lock_mode=2
bind-address=0.0.0.0

wsrep_on=ON # Enable Galera replication
wsrep_provider=/usr/lib/galera/libgalera_smm.so # Path to Galera library
wsrep_cluster_name="my_galera_cluster" # Name of the cluster
wsrep_cluster_address="gcomm://node1_ip,node2_ip,node3_ip" # List of node IPs

wsrep_stt_method=rsync # State transfer method

wsrep_node_address="this_node_ip" # IP address of this node
wsrep_node_name="this_node_name" # Name of this node
```

- [Docs Installation Galera Cluster](https://mariadb.com/docs/galera-cluster/galera-cluster-quickstart-guides/mariadb-galera-cluster-guide)
- Install `MaxScale` for load balancing and high availability. Use the following commands to install it:

```bash
# Debian/Ubuntu
sudo apt update
sudo apt install -y curl
curl -LsS https://r.mariadb.com/downloads/mariadb_repo_setup | sudo bash
sudo apt install -y maxscale

# RHEL/Rocky Linux/Alma Linux
curl -LsS https://r.mariadb.com/downloads/mariadb_repo_setup | sudo bash
sudo dnf install -y maxscale
```

- Steps to configure [MaxScale](https://mariadb.com/docs/maxscale/maxscale-quickstart-guides/mariadb-maxscale-installation-guide)
- `MaxScale's` configuration is primarily done in its main configuration file in `/etc/maxscale.cnf`.

```ini
# Define servers in the cluster
[server1]
type=server
# IP address or hostname of your first MariaDB server
address=192.168.1.101

[server2]
type=server
# IP address or hostname of your second MariaDB server
address=192.168.1.102
# Set the port if MariaDB is listening on a non-default port
port=3307

# Define a Monitor to check the status of the MariaDB servers
[MariaDB-Cluster]
type=monitor
# The  MariaDB asynchronous replication monitoring module
module=mariadbmon
# List of servers to monitor
servers=server1,server2
# The user used for monitoring
user=maxscale_monitor
password=monitor_password
# Check every 5 seconds
monitor_interval=5s
```

- ***Important***: Create the `maxscale_monitor` user on each `MariaDB` server with the necessary privileges for monitoring.

```sql
CREATE USER 'maxscale_monitor'@'%' IDENTIFIED BY 'monitor_password';
GRANT BINLOG ADMIN, BINLOG MONITOR, CONNECTION ADMIN, READ_ONLY ADMIN, REPLICATION SLAVE ADMIN, SLAVE MONITOR, RELOAD, PROCESS, SUPER, EVENT, SET USER, SHOW DATABASES ON *.* TO `maxscale_monitor`@`%`;
GRANT SELECT ON mysql.global_priv TO 'maxscale_monitor'@'%';
```

- Define a Service (e.g., Read-Write Split), continue configuration in `/etc/maxscale.cnf`.

```ini
[Read-Write-Service]
type=service
# The readwritesplit router module load balances reads and routes writes to the primary node
router=readwritesplit
# Servers available for this service
cluster=MariaDB-Cluster
# The user account used to fetch the user information from MariaDB
user=maxscale_user
password=maxscale_password
```
