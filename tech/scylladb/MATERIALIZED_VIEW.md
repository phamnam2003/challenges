# Materialized View

## [CREATE MATERIALIZED VIEW](https://docs.scylladb.com/manual/stable/cql/mv.html#create-materialized-view)

-You can create a materialized view on a table using a CREATE MATERIALIZED VIEW statement: You can create a materialized view on a table using a `CREATE MATERIALIZED VIEW` statement:

```sql
create_materialized_view_statement: CREATE MATERIALIZED VIEW [ IF NOT EXISTS ] `view_name` AS
                                  :     `select_statement`
                                  :     PRIMARY KEY '(' `primary_key` ')'
                                  :     WITH `table_options`
```

- Example:

```sql
CREATE MATERIALIZED VIEW monkeySpecies_by_population AS
    SELECT * FROM monkeySpecies
    WHERE population IS NOT NULL AND species IS NOT NULL
    PRIMARY KEY (population, species)
    WITH comment='Allow query by population instead of species';
```

- The `CREATE MATERIALIZED VIEW` statement creates a new materialized view. *Each view is a set of rows* that *corresponds* to the **rows** that are present in the ***underlying***, or base table, as specified in the SELECT statement. A materialized view cannot be directly updated, but updates to the base table will cause *corresponding updates* in the `view`.

## [Materialized View Select Statement](https://docs.scylladb.com/manual/stable/cql/mv.html#mv-select-statement)

- The *select statement* of a `materialized view` creation defines which of the base table is included in the view. That statement is limited in a number of ways:
  - The selection is limited to those that only select columns of the base table. In other words, you *can’t use any function* (aggregate or not), *casting*, *term*, etc. `Aliases` are also not supported. You can, however, use *as a shortcut to *selecting all columns*. Further, ***static columns*** *cannot* be included in a `materialized view` (which means SELECT* *isn’t allowed* if the *base table* has *static columns*).
  - The ***WHERE*** clause has the following restrictions:
    - It *cannot include* any `bind_marker`.
    - The columns that *are not part* of the view table **primary key** can’t be restricted.
    - As the columns that *are part of* the view **primary key** *cannot be null*, they *must always be at least restricted* by a **IS NOT NULL** restriction (or any other restriction, but they must have one).
    - They can also be restricted by relational operations (=, >, <).
  - The **SELECT** statement cannot include any of the following:
    - `Ordering` by clause
    - `Limit` clause
    - `ALLOW FILTERING` clause

## [Materialized View Primary Key](https://docs.scylladb.com/manual/stable/cql/mv.html#mv-primary-key)

- A view must have a primary key, and that primary key must conform to the following restrictions:
  - It must `contain` *all* the **primary key** columns of the base table. This ensures that every row in the view corresponds to *exactly* one row of the base table.
  - It can only contain a single column that is not a primary key column in the base table.

```sql
CREATE TABLE t (
    k int,
    c1 int,
    c2 int,
    v1 int,
    v2 int,
    PRIMARY KEY (k, c1, c2)
);

-- ALLOW CREATE MATERIALIZED VIEW statements:
CREATE MATERIALIZED VIEW mv1 AS
    SELECT * FROM t WHERE k IS NOT NULL AND c1 IS NOT NULL AND c2 IS NOT NULL
    PRIMARY KEY (c1, k, c2);

CREATE MATERIALIZED VIEW mv1 AS
    SELECT * FROM t WHERE v1 IS NOT NULL AND k IS NOT NULL AND c1 IS NOT NULL AND c2 IS NOT NULL
    PRIMARY KEY (v1, k, c1, c2);
-- END ALLOW CREATE MATERIALIZED VIEW statements

-- NOT ALLOW CREATE MATERIALIZED VIEW statements:
-- error: cannot include both v1 and v2 in the primary key as both are not in the base table primary key
CREATE MATERIALIZED VIEW mv1 AS
    SELECT * FROM t WHERE k IS NOT NULL AND c1 IS NOT NULL AND c2 IS NOT NULL AND v1 IS NOT NULL
    PRIMARY KEY (v1, v2, k, c1, c2)

-- error: must include k in the primary as it's a base table primary key column
CREATE MATERIALIZED VIEW mv1 AS
    SELECT * FROM t WHERE c1 IS NOT NULL AND c2 IS NOT NULL
    PRIMARY KEY (c1, c2)
-- END NOT ALLOW CREATE MATERIALIZED VIEW statements
```

- Note that, although each materialized view is a separate table, a user *cannot modify* a view directly:

```bash
cqlsh:mykeyspace> DELETE FROM building_by_city WHERE city='Taipei';

InvalidRequest: code=2200 [Invalid query] message="Cannot directly modify a materialized view"
```

## Compaction Strategies with Materialized Views

- Materialized views, just like regular tables, use one of the available *`compaction strategies`*. When a materialized view is created, it does not inherit its base table compaction strategy settings, because the data model of a view does not necessarily have the same characteristics as the one from its base table. Instead, the default compaction strategy (*SizeTieredCompactionStrategy*) is used.

```sql
CREATE MATERIALIZED VIEW ks.mv AS SELECT a,b FROM ks.t WHERE
  a IS NOT NULL
  AND b IS NOT NULL
  PRIMARY KEY (a,b)
  WITH COMPACTION = {'class': 'LeveledCompactionStrategy'};
```

- You can also change the compaction strategy of an already existing materialized view, using an `ALTER MATERIALIZED VIEW` statement.
