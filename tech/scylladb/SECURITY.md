# Security in ScyllaDB

## Authentication and Authorization

1. Authentication: ScyllaDB supports various authentication mechanisms, including password-based authentication and integration with external systems like LDAP and Kerberos. This ensures that only authorized users can access the database.

2. Authorization: ScyllaDB provides role-based access control (RBAC) to manage

3. Encryption: ScyllaDB supports encryption for data at rest and in transit. You can configure SSL/TLS for secure communication between clients and the database, as well as enable encryption for data stored on disk.

4. Auditing: ScyllaDB offers auditing capabilities to track and log database activities.

5. SSL/TLS: ScyllaDB supports SSL/TLS encryption for client-server communication, ensuring that data transmitted over the network is secure, permissions for users and roles. You can define roles with specific privileges and assign them to users, allowing

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
- Start ScyllaDB nodes.

5. Grant authorization CQL References

- [docs.scylladb.com](https://docs.scylladb.com/manual/stable/operating-scylla/security/authorization.html)
- Database Role, Alter Role, Drop Role, List Roles, Grant, Revoke, List Permissions, List Role Permissions

6. Certificate-based Authentication

- Enable CQL transport TLS using client certificate verification in each node by configuring the `client_encryption_options` option in the **/etc/scylla/scylla.yaml** file:

```yaml
client_encryption_options:
   enabled: True
   certificate: <server cert>
   keyfile: <server key>
   truststore: <shared trust>
   require_client_auth: True
```

- `enabled`: Set to `True` to enable TLS encryption for client connections.
- `certificate`: Path to the server's SSL certificate file.
- `keyfile`: Path to the server's private key file.
- `truststore`: Path to the truststore file containing trusted CA certificates.
- `require_client_auth`: Set to `True` to enforce client certificate authentication.

- Configure the certificate based authenticator in each node.

```yaml
  authenticator: CertificateAuthenticator
```

- ***Encryption: Data in Transit Client to Node***:
  - Run `nodetool drain` on each node to flush memtables to SSTables and stop accepting writes.
  - Stop ScyllaDB on each node.
  - Update the `client_encryption_options` section in the /etc/scylla/scylla.yaml file with the appropriate paths to the certificate, keyfile, and truststore.

```yaml
client_encryption_options:
    enabled: true
    certificate: /etc/scylla/db.crt
    keyfile: /etc/scylla/db.key
    truststore: <path to a PEM-encoded trust store> (optional)
    certficate_revocation_list: <path to a PEM-encoded CRL file> (optional)
    require_client_auth: ...
    priority_string: SECURE128:-VERS-TLS1.0:-VERS-TLS1.1
```

- ***Encryption: Data in Transit Node to Node***:
  - Communication between all or some nodes can be encrypted. The controlling parameter is `server_encryption_options`.
  - Configure the internode_encryption, under `/etc/scylla/scylla.yaml`.

```yaml
server_encryption_options:
    internode_encryption: <none|rack|dc|all|transitional>
    certificate: <path to a PEM-encoded certificate file>
    keyfile: <path to a PEM-encoded key for certificate>
    truststore: <path to a PEM-encoded trust store> (optional)
    certficate_revocation_list: <path to a PEM-encoded CRL file> (optional)
```

- Options `sever_encryption_options`:
  - `internode_encryption` can be one of the following:
    - `none` (default) - No traffic is encrypted.
    - `all` - Encrypts all traffic
    - `dc` - Encrypts the traffic between the data centers.
    - `rack` - Encrypts the traffic between the racks.
    - `transitional` - Encrypts all outgoing traffic, but allows non-encrypted incoming. Used for upgrading cluster(s) without downtime.
  - `certificate` - A PEM format certificate, either self-signed, or provided by a certificate authority (CA).
  - `keyfile` - The corresponding PEM format key for the certificate.
  - `truststore` - Optional path to a PEM format certificate store of trusted CAs. If not provided, ScyllaDB will attempt to use the system trust store to authenticate certificates.
  - `certficate_revocation_list` - The path to a PEM-encoded certificate revocation list (CRL) - a list of issued certificates that have been revoked before their expiration date.
  - `require_client_auth` - Set to True to require client side authorization. False by default.
  - `priority_string` - Specifies session’s handshake algorithms and options to use. By default there are none
