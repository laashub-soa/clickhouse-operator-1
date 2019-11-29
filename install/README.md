### Install operator by Helm

```bash
helm install --namespace clickhouse-system --name clickhouse-operator helm/clickhouse-operator
helm install --namespace clickhouse-system --name clickhouse-broker helm/clickhouse-broker
```


### Install operator by Shell

```bash
kubectl create -f shell/clickhouse-operator.yaml
```
If you want to check metrics, you can

```bash
kubectl create -f service-monitor.yaml
```

