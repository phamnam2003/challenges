# Security in ScyllaDB

## Authentication and Authorization

1. Authentication: ScyllaDB supports various authentication mechanisms, including password-based authentication and integration with external systems like LDAP and Kerberos. This ensures that only authorized users can access the database.

2. Authorization: ScyllaDB provides role-based access control (RBAC) to manage

3. Encryption: ScyllaDB supports encryption for data at rest and in transit. You can configure SSL/TLS for secure communication between clients and the database, as well as enable encryption for data stored on disk.

4. Auditing: ScyllaDB offers auditing capabilities to track and log database activities,

## Configuration Best Practices

1. Configure authentication directly on server:

- Configure strong passwords for all users and avoid using default credentials in file `/etc/scylla/scylla.yaml`:

```yaml
  authenticator: PasswordAuthenticator # required username/password to connect
  authorizer: CassandraAuthorizer # turn on controlled access (GRANT/REVOKE)
```

- Restart ScyllaDB after making changes to the configuration file. You have default user `cassandra` with password `cassandra`. Change the password immediately after the first login.
- Change password or create new user to connect to ScyllaDB:

```bash
cqlsh -u cassandra -p cassandra
```

```cql
ALTER USER cassandra WITH PASSWORD 'MyStrongPassword';
```

Create user with grant permission:

```cql
CREATE USER app_user WITH PASSWORD 'AppPass123' NOSUPERUSER;
GRANT ALL PERMISSIONS ON KEYSPACE my_keyspace TO app_user;
```

- You can repair user permissions if you lost them by `nodetool repair`:

```bash
nodetool repair system_auth
```

2. Configure via Docker

- You can set it via arg in command section in `docker-compose.yml` file:

```yaml
  scylla_node1:
    image: scylladb/scylla:2025.3 # Image and tag version
    container_name: scylla_node1 # container name when run local machine
    command: --smp 2 --memory 2G --overprovisioned 1 --authenticator PasswordAuthenticator --authorizer CassandraAuthorizer # authentication and authorization
    ports:
      - "9042:9042" # CQL
      - "7000:7000" # inter-node communication
      - "7001:7001" # TLS inter-node
      - "7199:7199" # JMX
    networks:
      scylla-cluster:
        ipv4_address:
```

3. Creating a Custom Superuser

- The default ScyllaDB superuser role is `cassandra` with password `cassandra`. Users with the `cassandra` role have full access to the database and can run any CQL command on the database resources.
To improve security, we recommend creating a custom superuser. You should:
  - Use the default `cassandra` superuser to log in.
  - Create *a custom superuser*.
  - Log in as the custom superuser.
  - *Remove* the `cassandra` role.

```cql
CREATE ROLE <custom_superuser name>  WITH SUPERUSER = true AND LOGIN = true and PASSWORD = '<custom_superuser_password>';
DROP ROLE cassandra;
```

4. Reset authenticator password.

- Stop all the nodes in the cluster.
- Remove system tables starting with role prefix from */var/lib/scylla/data/system* directory.
