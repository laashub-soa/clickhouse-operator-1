### Install operator by Helm

```bash
helm template --namespace clickhouse-system --name clickhouse-operator helm/clickhouse-operator
```


### Install operator by Shell

```bash
kubectl create -f shell/clickhouse-operator.yaml
```
If you want to check metrics, you can

```bash
kubectl create -f service-monitor.yaml
```

