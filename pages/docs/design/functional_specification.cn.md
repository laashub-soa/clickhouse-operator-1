---
title: "功能规格"
weight: 1
description: "clickhouse"
date: 2019-7-21T10:36:04+08:00
---

## 概述 Overview

提供基于 diamond-k8s 环境 clickhouse 部署方案，以满足工程实施及运维要求。

## 功能规格的细节

1. 提供基于 diamond-k8s 快捷部署功能；
2. 提供单节点或多节点的部署方式；
3. 支持客户端无感的扩容操作
4. 支持客户端以授权方式访问 ClickHouse 集群（K8s 内、外两种方式）
5. 提供 health check 及性能监控接口，以便与监控告警系统集成
6. 提供备份、恢复机制
7. 完备的文档：部署、维护

### 基于 Diamond-K8s 的快捷部署

- 提供离线安装包，包括：安装脚本及使用的镜像文件等
- 提供 kubectl、Helm 的安装方式
- 需要支持 Diamond OSS 集成
- 需要支持 Diamond LCM 管控

### 支持客户端无感的部署方式

- 在 K8s 内部提供 DNS 方式访问
- 在 K8s 外部提供 DNS 方式访问

### 支持授权方式访问

- 在 K8s 内部、外部均提供授权方式访问，以保证数据安全性
- 提供给客户配置 ClickHouse 账号密码的方式

### 监控

- 提供 healthcheck 接口
- 提供监控接口，如：使用 sidecar exportor 方式

### 文档

- 提供给实施人员部署实施操作手册
- 提供给运维同事运维操作手册
- Dev 编译、打包、CI 操作手册

## 非功能描述 Unfunctional Specifications & Details

1. 确定支持的 ClickHouse 版本
2. 需要基于 SenseGO 的需求给出性能测试报告
3. 建议使用 Amber 完成测试
4. 需要与 gitlab-ci 集成
5. 按照 OSB & K8s（CRD or AA）标准接口（方式）实现

### ClickHouse 版本

- 支持 v19.13.x.x

### 基于 SenseGO 场景的性能测试报告

- 按照使用场景准备数据及初始脚本，建议独立提供以便重复使用
- 验证场景：
  - 验证读取 1 亿数据的时间
  - 验证同时读写操作下的响应时间

### 建议使用 Amber 完成测试

- 需要模拟场景
  - 节点出错
  - 网络不连通
  - 网络丢包
- 固化测试脚本

### 实现规划：OSB、CRD or AA

- 需要以 OSB 标准集成，以便与 Service Catalog（kubeapp）集成
- 推荐使用 CRD 方式，备选方案为 AA（由于 AA 需要服务本身保证存储，增加运维复杂度）

## 开放问题 Open Issues

1. 需要收集客户对 clickhouse 集群规模的要求，我们希望提供有限的 plan
2. 客户对批量更新的需求：一次更新的数据量级、频次、响应速度
3. 是否存在 Legacy system 数据迁移，如：目前采用非 k8s 方式部署
4. 是否需要支持多 IDC 数据同步

## 其他 Others

强烈建议按照 functional spec 进行开发，并保持一定的灵活性和扩展性，并在出现非预期的情况下及时修正设计与实现，以保证按照 func spec 交付产品（因为功能描述是之前讨论确认的，除非有充足的理由打动大家，才可以修改）
