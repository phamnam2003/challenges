# Custom Dashboard for Loki in Grafana

- Custom Dashboard for Loki in Grafana for filtering logs by labels or values from `JSON` logs.

## Variables

- `Variables` in dashboard Grafana configuration in Dashboard setting, in tab `Variables`. It make labels for filtering, it can be label or value logs.
- It have variable type, query options with data source, query type (label or values). Option `label values` is recommended for filter value of field `JSON` log
- Use `LOGQL` query for `Loki` data source, combines with `|=` or `|~` for filtering value of field `JSON` log.
