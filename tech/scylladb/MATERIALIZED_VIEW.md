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
