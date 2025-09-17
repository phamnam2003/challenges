# Data Type in ScyllaDB

- CQL is a typed language and supports a rich set of data types, including `native types` and `collection types`.

```yml
cql_type: `native_type` | `collection_type` | `user_defined_type` | `tuple_type` | `vector_type`
```

## Working With Bytes

- Bytes can be input with:
  - a hexadecimal literal.
  - one of the `typeAsBlob()` cql functions. There is such a function for each `native type`.

```sql
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

### Maps

- A map is a (sorted) *set of key-value pairs*, where keys are **unique**, and the map is *sorted by its keys*. You can define a map column with:

```sql
CREATE TABLE users (
    id text PRIMARY KEY,
    name text,
    favs map<text, text> // A map of text keys, and text values
);
```

- A map column *can be assigned new contents* with either **INSERT** or **UPDATE**, as in the following examples. In both cases, the new contents replace the map’s old content, if any:

```sql
INSERT INTO users (id, name, favs)
           VALUES ('jsmith', 'John Smith', { 'fruit' : 'Apple', 'band' : 'Beatles' });

UPDATE users SET favs = { 'fruit' : 'Banana' } WHERE id = 'jsmith';
```

- Updating or inserting one or more elements:

```sql
UPDATE users SET favs['author'] = 'Ed Poe' WHERE id = 'jsmith';
UPDATE users SET favs = favs + { 'movie' : 'Cassablanca', 'band' : 'ZZ Top' } WHERE id = 'jsmith';
```

- Removing one or more element (if an element doesn’t exist, removing it is a no-op but no error is thrown):

```sql
DELETE favs['author'] FROM users WHERE id = 'jsmith';
UPDATE users SET favs = favs - { 'movie', 'band'} WHERE id = 'jsmith';
```

- Selecting one element:

```sql
SELECT favs['fruit'] FROM users WHERE id = 'jsmith';
```

- Lastly, *TTLs* are allowed for both **INSERT** and **UPDATE**, but in both cases, the TTL set only applies to the newly inserted/updated elements. In other words:

```sql
UPDATE users USING TTL 10 SET favs['color'] = 'green' WHERE id = 'jsmith';
```

- It will only apply the TTL to the `{ 'color' : 'green' }` record, the rest of the map `remaining unaffected`.

### Sets

- A `set` is a (sorted) collection of *unique* values. You can define a set column with:

```sql
CREATE TABLE images (
    name text PRIMARY KEY,
    owner text,
    tags set<text> // A set of text values
);
```

- A set column *can be assigned new contents* with either **INSERT** or **UPDATE**, as in the following examples. In both cases, the new contents replace the set’s old content, if any:

```sql
INSERT INTO images (name, owner, tags)
            VALUES ('cat.jpg', 'jsmith', { 'pet', 'cute' });

UPDATE images SET tags = { 'kitten', 'cat', 'lol' } WHERE name = 'cat.jpg';
```

- Note that ScyllaDB does not *distinguish* an empty set from a missing value, thus assigning an empty set ({}) to a set is *the same as deleting it*.
- Adding one or multiple elements (as this is a set, inserting an already existing element is a no-op):

```sql
UPDATE images SET tags = tags + { 'gray', 'cuddly' } WHERE name = 'cat.jpg';
```

- Removing one or multiple elements (if an element doesn’t exist, removing it is a no-op but no error is thrown):

```sql
UPDATE images SET tags = tags - { 'cat' } WHERE name = 'cat.jpg';
```

- Selecting an element (if the element doesn’t exist, returns null):

```sql
SELECT tags['gray'] FROM images;
```

- Lastly, as for `maps`, *TTLs*, if used, only apply to the newly inserted values.

### Lists

[!Note]
As mentioned above and further discussed at the end of this section, lists have limitations and specific performance considerations that you should take into account before using them. In general, if you can use a set instead of a list, always prefer a set.

- A `list` is an ordered list of values (*not necessarily unique*). You can define a list column with:

```sql
CREATE TABLE plays (
    id text PRIMARY KEY,
    game text,
    players int,
    scores list<int> // A list of integers
);
```

- A list column *can be assigned new contents* with either **INSERT** or **UPDATE**, as in the following examples. In both cases, the new contents replace the list’s old content, if any:

```sql
INSERT INTO plays (id, game, players, scores)
           VALUES ('123-afde', 'quake', 3, [17, 4, 2]);

UPDATE plays SET scores = [3, 9, 4] WHERE id = '123-afde';
```

- Note that ScyllaDB does not distinguish an empty list from a missing value, thus assigning an empty list ([]) to a list is *the same as deleting it*.
- Appending and prepending values to a list:

```sql
UPDATE plays SET players = 5, scores = scores + [ 14, 21 ] WHERE id = '123-afde';
UPDATE plays SET players = 6, scores = [ 3 ] + scores WHERE id = '123-afde';
```

- Setting the value at a particular position in the list. This implies that the list has a pre-existing element for that position or an error will be thrown that the list is too small:

```sql
UPDATE plays SET scores[1] = 7 WHERE id = '123-afde';
```

- Deleting `all` the *occurrences of particular values* in the list (if a particular element doesn’t occur at all in the list, it is simply ignored, and no error is thrown):

```sql
UPDATE plays SET scores = scores - [ 12, 21 ] WHERE id = '123-afde';
```

- Selecting an element by its position in the list:

```sql
SELECT scores[1] FROM plays;
```

- Lastly, as for `maps`, *TTLs*, when used, only apply to the newly inserted values.

### [User-Defined Types (UDTs)](https://docs.scylladb.com/manual/stable/cql/types.html#user-defined-types)
