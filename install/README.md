### Install operator by Helm

```bash
helm install --namespace clickhouse-system --name clickhouse-operator helm/clickhouse-operator
helm install --namespace clickhouse-system --name clickhouse-broker helm/clickhouse-broker
```
