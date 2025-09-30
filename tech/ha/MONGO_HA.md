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
