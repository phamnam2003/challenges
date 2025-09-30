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
