# Secondary Index

## GSI (Global Secondary Index)

- **Global Secondary indexes** (named “*Secondary indexes*” for the rest of this doc) are a *mechanism* in ScyllaDB which *allows efficient searches* on *non-partition keys* by creating an index
- Secondary indexes provide the following advantages:
  - *Secondary Indexes* are (mostly) transparent to your application. Queries have access to all the columns in the table, and you can add or remove indexes on the fly without changing the application.
  - We can use the value of the indexed column to find the corresponding index table row in the cluster so that reads are scalable.
  - Updates can be more efficient with secondary indexes than materialized views because only changes to the primary key and indexed column cause an update in the index view.

### How Secondary Index Queries Work

- ScyllaDB breaks indexed queries into two parts:
  - A query on the index table to retrieve partition keys for the indexed table, and
  - A query to the indexed table using the retrieved partition keys.
- Example, given the following schema:

```sql
CREATE TABLE buildings  (name text, city text, height int, PRIMARY KEY (name));

INSERT INTO buildings(name,city,height) VALUES ('Burj Khalifa','Dubai',828);
INSERT INTO buildings(name,city,height) VALUES ('Shanghai Tower','Shanghai',632);
INSERT INTO buildings(name,city,height) VALUES ('Abraj Al-Bait Clock Tower','Mecca',601);
INSERT INTO buildings(name,city,height) VALUES ('Ping An Finance Centre','Shenzhen',599);
INSERT INTO buildings(name,city,height) VALUES ('Lotte World Tower','Seoul',554);
INSERT INTO buildings(name,city,height) VALUES ('One World Trade Center','New York City',541);
INSERT INTO buildings(name,city,height) VALUES ('Guangzhou CTF Finance Centre','Guangzhou',530);
INSERT INTO buildings(name,city,height) VALUES ('Tianjin CTF Finance Centre','Tianjin',530);
INSERT INTO buildings(name,city,height) VALUES ('China Zun','Beijing',528);
INSERT INTO buildings(name,city,height) VALUES ('Taipei 101','Taipei',508);
```

```sql
SELECT * FROM buildings WHERE city = 'Shenzhen'; -- will error because city is not partition key
```

- *Secondary indexes* are designed to *allow efficient querying* of *non-partition* key columns. We can create an index on city by with the following CQL statements:

```sql
CREATE INDEX buildings_by_city ON buildings (city);

-- Query using the secondary index
SELECT * FROM buildings WHERE city = 'Shenzhen';

-- Result:
name                   | city     | height
-----------------------+----------+--------
Ping An Finance Centre | Shenzhen |    599
(1 rows)
```

- Note that you can use the `DESCRIBE` command to see the whole *schema* for the buildings table, including *created indexes* and *views*:

```sql
cqlsh:mykeyspace> DESC buildings;

CREATE TABLE mykeyspace.buildings (
             name text PRIMARY KEY,
             city text,
             height int
) WITH bloom_filter_fp_chance = 0.01
AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
...;

CREATE INDEX buildings_by_city ON mykeyspace.buildings (city);

CREATE MATERIALIZED VIEW mykeyspace.buildings_by_city_index AS
SELECT city, idx_token, name
FROM mykeyspace.buildings
WHERE city IS NOT NULL
PRIMARY KEY (city, idx_token, name)
WITH CLUSTERING ORDER BY (idx_token ASC, name ASC)
AND bloom_filter_fp_chance = 0.01
AND caching = {'keys': 'ALL', 'rows_per_partition': 'ALL'}
...
```
