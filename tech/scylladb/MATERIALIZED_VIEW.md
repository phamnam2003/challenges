# Materialized View

## [CREATE MATERIALIZED VIEW](https://docs.scylladb.com/manual/stable/cql/mv.html#create-materialized-view)

-You can create a materialized view on a table using a CREATE MATERIALIZED VIEW statement: You can create a materialized view on a table using a `CREATE MATERIALIZED VIEW` statement:

```bash
create_materialized_view_statement: CREATE MATERIALIZED VIEW [ IF NOT EXISTS ] `view_name` AS
                                  :     `select_statement`
                                  :     PRIMARY KEY '(' `primary_key` ')'
                                  :     WITH `table_options`
```

- The `CREATE MATERIALIZED VIEW` statement creates a new materialized view. *Each view is a set of rows* that *corresponds* to the **rows** that are present in the ***underlying***, or base table, as specified in the SELECT statement. A materialized view cannot be directly updated, but updates to the base table will cause *corresponding updates* in the `view`.
