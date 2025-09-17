# Data Type in ScyllaDB

- CQL is a typed language and supports a rich set of data types, including `native types` and `collection types`.

```yml
cql_type: `native_type` | `collection_type` | `user_defined_type` | `tuple_type` | `vector_type`
```

## Working With Bytes

- Bytes can be input with:
  - a hexadecimal literal.
  - one of the `typeAsBlob()` cql functions. There is such a function for each `native type`.

```bash
INSERT INTO blobstore (id, data) VALUES (4375645, 0xabf7971528cfae76e00000008bacdf);
INSERT INTO blobstore (id, data) VALUES (4375645, intAsBlob(33));
```

## Native Types

- The native types supported by CQL are:

```yml
native_type: ASCII
           : | BIGINT
           : | BLOB
           : | BOOLEAN
           : | COUNTER
           : | DATE
           : | DECIMAL
           : | DOUBLE
           : | DURATION
           : | FLOAT
           : | INET
           : | INT
           : | SMALLINT
           : | TEXT
           : | TIME
           : | TIMESTAMP
           : | TIMEUUID
           : | TINYINT
           : | UUID
           : | VARCHAR
           : | VARINT
```
