---
title: "Install Operator"
weight: 2
description: "clickhouse"
date: 2019-8-022T9:30:00+08:00
---

## Prerequisite

Clickhouse Service required Helm to deploy. So you should have:

- [Helm](https://helm.sh/) version v2.12.2+.
- Kubernetes v1.13.3+ cluster.
- [Service-Catalog](https://github.com/kubernetes-incubator/service-catalog) [installed](https://github.com/kubernetes-incubator/service-catalog/blob/master/docs/install.md) (integrated in Diamond)

## Deployment

To simplify the deployment, we use `Helm` as deploying tool. You could download the latest helm chart from [here](http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/tags/), which contains YAML files below:

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
| |____readme.md
| |____templates
| | |____deployment.yaml
| | |____clusterrole.yaml
| | |____NOTES.txt
| | |_____functions.tpl
| | |____service_account.yaml
| | |____service_monitor.yaml
| | |____service.yaml
| | |____db_v1alpha1_clickhousecluster_crd.yaml
| | |____psp.yaml
| |____values.yaml
| |____index.yaml
```

**Deploy Clickhouse Operator**:

```bash
$ helm install --namespace clickhouse-system --name clickhouse-operator ./clickhouse-operator

NAME:   clickhouse-operator
LAST DEPLOYED: Thu Aug 15 18:52:42 2019
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

Visit https://github.com/Orange-OpenSource/clickhouse-k8s-operator for instructions on hot to create & configure Clickhouse clusters using the operator.
```

**Deploy Clickhouse Broker**:

```bash
$ helm install --namespace clickhouse-system --name clickhouse-broker ./service-broker

NAME:   clickhouse-broker
LAST DEPLOYED: Thu Aug 15 18:55:52 2019
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

When the deployment finished, you could see some pods created by the chart:

```bash
$ kubectl get pod -n clickhouse-system
NAME                                       READY   STATUS    RESTARTS   AGE
clickhouse-operator-7975b48959-tr4cv        1/1     Running   0          34m
clickhouse-service-broker-fcf65c77b-dq5rr   1/1     Running   0          30m
```

Also, you could see some new created CustomResourceDefinitions (CRDs).

```bash
$ kubectl get crd
NAME                                             CREATED AT
...
clickhouseclusters.db.orange.com                  2019-08-15T10:52:15Z
...
```

```bash
$ svcat get broker
            NAME             NAMESPACE                                           URL                                           STATUS
+--------------------------+-----------+-------------------------------------------------------------------------------------+--------+
  clickhouse-service-broker               http://clickhouse-broker-clickhouse-service-broker.clickhouse-system.svc.cluster.local   Ready
```

**Deploy Prometheus (Optional)**:

{{% notice note %}}
Because Prometheus is integrated into Diamond platform, you could ignore this step when using Diamond.
{{% /notice %}}

```shell script
 helm install --namespace monitoring --name prometheus stable/prometheus-operator
```

**Install Grafana dashborad**:

Please download {{% button href="http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/raw/master/prometheus/prometheus-grafana-clickhouse-dashboard.json" icon="fas fa-download" %}}clickhouse-dashboard.json{{% /button %}} and import it into Grafana.

<br>
<br>
Clickhouse Service is installed completely so far. We will introduce you about how to create a Clickhouse intance.
