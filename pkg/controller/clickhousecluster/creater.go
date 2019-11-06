package clickhousecluster

import (
	corev1 "k8s.io/api/core/v1"
)

type Creater struct {
}

func NewCreater() *Creater {
	return &Creater{}
}

func (c *Creater) CreateConfigMap() *corev1.ConfigMap {
	return nil
}
