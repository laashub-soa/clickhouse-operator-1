---
title: "安装 Operator"
weight: 2
description: "clickhouse"
date: 2019-7-018T16:36:04+08:00
---

## 准备阶段

Clickhouse 服务的部署依赖于 helm，所以需要：

- [Helm](https://helm.sh/) version v2.12.2+.
- Kubernetes v1.13.3+ cluster.
- [Service-Catalog](https://github.com/kubernetes-incubator/service-catalog) [installed](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md) (Diamond 中已集成)

## 部署

为了简化部署流程，我们采用了 helm 作为部署工具。你可以 从 [这里](http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/tags/)下载最新的 chart 包。里面包含如下内容：

```bash
.
|____service-broker
| |____Chart.yaml
| |____templates
| | |____ssl-certs.yaml
| | |____NOTES.txt
| | |____broker-service-account.yaml
| | |____broker-service.yaml
| | |____broker-configmap.yaml
| | |_____helpers.tpl
| | |____broker-deployment.yaml
| | |____broker.yaml
| |____values.yaml
|____clickhouse-operator
| |____Chart.yaml
| |____.helmignore
| |____templates
| | |____service-monitor.yaml
| | |____deployment.yaml
| | |____zookeeper.yaml
| | |____NOTES.txt
| | |_____functions.tpl
| | |____v1_clickhouse_crd.yaml
| | |____service_account.yaml
| | |____service.yaml
| | |____clusterrole.yaml
| | |____rolebinding.yaml
| |____values.yaml
| |____index.yaml
```

**部署 clickhouse operator**:

```bash
$ helm install --namespace clickhouse-system --name clickhouse-operator ./clickhouse-operator

NAME:   clickhouse-operator
LAST DEPLOYED: Thu Nov 15 18:52:42 2019
NAMESPACE: clickhouse-system
STATUS: DEPLOYED

RESOURCES:
==> v1/Deployment
NAME                READY  UP-TO-DATE  AVAILABLE  AGE
clickhouse-operator  0/1    1           0          27s

==> v1/Pod(related)
NAME                                 READY  STATUS             RESTARTS  AGE
clickhouse-operator-7975b48959-tr4cv  0/1    ContainerCreating  0         27s

==> v1/RoleBinding
NAME                AGE
clickhouse-operator  27s

==> v1/ServiceAccount
NAME                SECRETS  AGE
clickhouse-operator  1        27s

==> v1beta1/Role
NAME                AGE
clickhouse-operator  27s


NOTES:
Check its status by running:
kubectl --namespace clickhouse-system get pods -l "release=clickhouse-operator"

```

**部署 clickhouse broker**:

```bash
$ helm install --namespace clickhouse-system --name clickhouse-broker ./service-broker

NAME:   clickhouse-broker
LAST DEPLOYED: Thu Nov 15 18:55:52 2019
NAMESPACE: clickhouse-system
STATUS: DEPLOYED

RESOURCES:
==> v1/ConfigMap
NAME                             DATA  AGE
clickhouse-service-broker-config  1     27s

==> v1/Pod(related)
NAME                                      READY  STATUS             RESTARTS  AGE
clickhouse-service-broker-fcf65c77b-dq5rr  0/1    ContainerCreating  0         27s

==> v1/Secret
NAME                                            TYPE    DATA  AGE
clickhouse-broker-clickhouse-service-broker-cert  Opaque  2     27s

==> v1/Service
NAME                                       TYPE       CLUSTER-IP   EXTERNAL-IP  PORT(S)  AGE
clickhouse-broker-clickhouse-service-broker  ClusterIP  10.99.119.1  <none>       80/TCP   27s

==> v1/ServiceAccount
NAME                                               SECRETS  AGE
clickhouse-broker-clickhouse-service-broker-client   1        27s
clickhouse-broker-clickhouse-service-broker-service  1        27s

==> v1beta1/ClusterRole
NAME                                              AGE
access-clickhouse-broker-clickhouse-service-broker  27s
clickhouse-broker-clickhouse-service-broker         27s

==> v1beta1/ClusterRoleBinding
NAME                                              AGE
clickhouse-broker-clickhouse-service-broker         27s
clickhouse-broker-clickhouse-service-broker-client  27s

==> v1beta1/ClusterServiceBroker
NAME                      AGE
clickhouse-service-broker  0s

==> v1beta1/Deployment
NAME                      READY  UP-TO-DATE  AVAILABLE  AGE
clickhouse-service-broker  0/1    1           0          27s


NOTES:
Now that you have installed clickhouse-service-broker broker, you can create a resource.
```

部署完成之后，你会看到一些 被 chart 创建的 pods:

```bash
$ kubectl get pod -n clickhouse-system
NAME                                       READY   STATUS    RESTARTS   AGE
clickhouse-operator-7975b48959-tr4cv        1/1     Running   0          34m
clickhouse-service-broker-fcf65c77b-dq5rr   1/1     Running   0          30m
```

同样的，你也会看到一些 新创建的 CRD 资源：

```bash
$ kubectl get crd
NAME                                             CREATED AT
...
clickhouseclusters.sensetime.com                  2019-08-15T10:52:15Z
...
```

```bash
$ svcat get broker
            NAME             NAMESPACE                                           URL                                           STATUS
+--------------------------+-----------+-------------------------------------------------------------------------------------+--------+
  clickhouse-service-broker               http://clickhouse-broker-clickhouse-service-broker.clickhouse-system.svc.cluster.local   Ready
```

**部署 Prometheus (可选)**:

{{% notice note %}}
由于 Diamond 已经集成 Prometheus 系统，因此如果在 diamond 系统中，可以忽略此步骤
{{% /notice %}}

```shell script
 helm install --namespace monitoring --name prometheus stable/prometheus-operator
```

**安装 grafana dashborad**:

下载 {{% button href="http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/raw/master/prometheus/prometheus-grafana-clickhouse-dashboard.json" icon="fas fa-download" %}}clickhouse-dashboard.json{{% /button %}}并导入到 grafana。

<br>
<br>
至此，Clickhouse service 已经部署完成，下面将介绍 创建 Clickhouse 实例的方法。
