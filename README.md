# ClickHouse Operator

The ClickHouse Operator creates, configures and manages ClickHouse clusters running on Kubernetes.

[![build status](http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/badges/master/build.svg)](http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/commits/master)
[![coverage report](http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/badges/master/coverage.svg)](http://gitlab.bj.sensetime.com/diamond/service-providers/clickhouse/commits/master)

**Warning!**
**ClickHouse Operator is in beta. You can use it at your own risk. There may be backwards incompatible API changes up until the first major release.**

## Features

The ClickHouse Operator for Kubernetes currently provides the following:

- Creates ClickHouse cluster based on Custom Resource provided
- Storage customization (VolumeClaim templates)
- Pod template customization (Volume and Container templates)
- ClickHouse configuration customization (including Zookeeper integration)
- ClickHouse cluster scaling including automatic schema propagation
- ClickHouse cluster version upgrades
- Exporting ClickHouse metrics to Prometheus

## Requirements

- Kubernetes 1.12.6+

## Documentation
