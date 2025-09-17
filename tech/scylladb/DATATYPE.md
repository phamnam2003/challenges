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

## Collection Types

- CQL supports 3 kinds of collections: `Maps`, `Sets` and `Lists`. The types of those collections is defined by:

```yml
collection_type: MAP '<' `cql_type` ',' `cql_type` '>'
               : | SET '<' `cql_type` '>'
               : | LIST '<' `cql_type` '>'

map_example: MAP<text, int>
set_example: SET<text>
list_example: LIST<int>
```

## Noteworthy characteristics

- A collection can be ***frozen*** or ***non-frozen***.
- By default, a collection is `non-frozen`. To declare a frozen collection, use `FROZEN` keyword:

```yml
frozen_collection_type: FROZEN '<' MAP '<' `cql_type` ',' `cql_type` '>' '>'
                       : | FROZEN '<' SET '<' `cql_type` '>' '>'
                       : | FROZEN '<' LIST '<' `cql_type` '>' '>'
```

### Map

- A map is a (sorted) *set of key-value pairs*, where keys are **unique**, and the map is *sorted by its keys*. You can define a map column with:

```bash
CREATE TABLE users (
    id text PRIMARY KEY,
    name text,
    favs map<text, text> // A map of text keys, and text values
);
```

- A map column *can be assigned new contents* with either **INSERT** or **UPDATE**, as in the following examples. In both cases, the new contents replace the map’s old content, if any:

```bash
INSERT INTO users (id, name, favs)
           VALUES ('jsmith', 'John Smith', { 'fruit' : 'Apple', 'band' : 'Beatles' });

UPDATE users SET favs = { 'fruit' : 'Banana' } WHERE id = 'jsmith';
```

- Updating or inserting one or more elements:

```bash
UPDATE users SET favs['author'] = 'Ed Poe' WHERE id = 'jsmith';
UPDATE users SET favs = favs + { 'movie' : 'Cassablanca', 'band' : 'ZZ Top' } WHERE id = 'jsmith';
```

- Removing one or more element (if an element doesn’t exist, removing it is a no-op but no error is thrown):

```bash
DELETE favs['author'] FROM users WHERE id = 'jsmith';
UPDATE users SET favs = favs - { 'movie', 'band'} WHERE id = 'jsmith';
```

- Selecting one element:

```bash
SELECT favs['fruit'] FROM users WHERE id = 'jsmith';
```

- Lastly, *TTLs* are allowed for both **INSERT** and **UPDATE**, but in both cases, the TTL set only applies to the newly inserted/updated elements. In other words:

```bash
UPDATE users USING TTL 10 SET favs['color'] = 'green' WHERE id = 'jsmith';
```

- It will only apply the TTL to the `{ 'color' : 'green' }` record, the rest of the map `remaining unaffected`.
