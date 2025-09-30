# MongoDB Cluster High Availability

- Configuring a MongoDB cluster for high availability (HA) involves setting up a replica set, which is a group of MongoDB instances that maintain the same data set, providing redundancy and failover capabilities. Here’s a step-by-step guide to setting up a MongoDB replica set for high availability:
  - Bind MongoDB to All Network Interfaces: Edit the MongoDB configuration file (usually located at `/etc/mongod.conf` or `/etc/mongodb.conf`) to bind MongoDB to all network interfaces. Change the `bindIp` setting to `0.0.0.0`
  - In `/etc/mongod.conf` in all the files configuration of the nodes in the cluster, add or modify the following lines:

```yaml
net:
  bindIp: 0.0.0.0
replication:
  replSetName: "rs0"
```

- Cmd `mongosh` to access the `MongoDB` shell to configuration HA. Run cmd `rs.initiate()` to initiate the replica set. This command should be run on the primary node.

```javascript
rs.initiate(
  {
    _id: <replSetName>,
    members: [
      { _id: 0, host: "<primary_node_ip>:27017" },
      { _id: 1, host: "<secondary_node_1_ip>:27017" },
      { _id: 2, host: "<secondary_node_2_ip>:27017" }
    ]
  }
)
```

- To testing the replica set status, run the following command in the `MongoDB` shell:

```javascript
rs.status()
```

- Check Node Primary: You can check which node is primary by running:

```javascript
rs.isMaster()
```

# Install KeepAlived for Automatic Failover

- `Keepalived` is a routing software written in C. It provides *simple and robust* facilities for load balancing and *high-availability* to Linux system and Linux based infrastructures. It is used to provide high availability by implementing the *Virtual Router Redundancy Protocol* (`VRRP`).
- Install `Keepalived` on all nodes in the MongoDB cluster. Use the following command to install it:

```bash
sudo apt-get install keepalived
```

- Create or edit the `Keepalived` configuration file, usually located at `/etc/keepalived/keepalived.conf`. Below is an example configuration for a three-node MongoDB cluster:

```conf
vrrp_instance VI_1 {
    state MASTER # State can be MASTER or BACKUP
    interface eth0  # Replace with your network interface, `ens33` for example
    virtual_router_id 51 # Unique ID for the VRRP instance
    priority 101  # Higher priority for the primary node
    advert_int 1 # Advertisement interval in seconds
    authentication {
        auth_type PASS
        auth_pass your_password  # Replace with a secure password
    }
    virtual_ipaddress {
      <VIRTUAL_IP_ADDRESS> # Replace with your virtual IP address
    }
```

- But, `Keepalived` and `MongoDB` will be working wrong. It can be one node hold `virtual_ipaddress` but not `MongoDB Primary`. To fix this, you connect to `Master` node and open `MongoDB Shell`.

```javascript
mongosh
cfd = rs.conf()
cfg.members[0].priority = 101  // Primary node priority
cfg.members[1].priority = 100  // Secondary node priority
cfg.members[2].priority = 100  // Secondary node priority

// reconfigure the replica set with the new configuration
rs.reconfig(cfg)
```

- Connect to `virtual_ipaddress` to access the `MongoDB` cluster. You can use the following command to connect:

```bash
mongosh --host <VIRTUAL_IP_ADDRESS>:27017
```
