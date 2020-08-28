package clickhousecluster

import (
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"math/rand"
	"reflect"
	"time"

	clickhousev1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"strings"
)

func doParse(v reflect.Value, indent int, father string) string {
	var out string
	space := strings.Repeat("   ", indent)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Map {
		for _, e := range v.MapKeys() {
			tag := e.String()
			out += fmt.Sprintf("\n%s<%s>%s\n%s</%s>", space, tag, doParse(v.MapIndex(e), indent+1, tag), space, e)
		}
	} else if v.Kind() == reflect.Struct {
		for i := 0; i < v.NumField(); i++ {
			tag := v.Type().Field(i).Tag.Get("xml")
			if tag != "" {
				kind := v.Field(i).Kind()
				if kind == reflect.Slice {
					out += fmt.Sprintf("%s", doParse(v.Field(i), indent, tag))
					continue
				}
				if kind == reflect.String || kind == reflect.Bool || kind == reflect.Int || kind == reflect.Int32 {
					out += fmt.Sprintf("\n%s<%s>%s</%s>", space, tag, doParse(v.Field(i), indent+1, tag), tag)
					continue
				}
				out += fmt.Sprintf("%s<%s>%s\n%s</%s>", space, tag, doParse(v.Field(i), indent+1, tag), space, tag)
			}
		}
	} else if v.Kind() == reflect.Slice {
		for i := 0; i < v.Len(); i++ {
			out += fmt.Sprintf("\n%s<%s>%s\n%s</%s>", space, father, doParse(v.Index(i), indent+1, ""), space, father)
		}
	} else {
		out = fmt.Sprintf("%v", v)
	}
	return out
}

func ParseXML(s interface{}) string {
	v := reflect.ValueOf(s)
	return fmt.Sprintf("<yandex>\n%s\n</yandex>", doParse(v, 1, ""))
}

func isStatefulSetReady(statefulSet *appsv1.StatefulSet) bool {
	return statefulSet.Status.Replicas == *statefulSet.Spec.Replicas &&
		statefulSet.Status.ReadyReplicas == *statefulSet.Spec.Replicas
}

// sts1 = stored statefulset and sts2 = new generated statefulset
func statefulSetsAreEqual(sts1, sts2 *appsv1.StatefulSet) bool {

	sts1.Spec.Template.Spec.SchedulerName = sts2.Spec.Template.Spec.SchedulerName
	sts1.Spec.Template.Spec.DNSPolicy = sts2.Spec.Template.Spec.DNSPolicy // ClusterFirst
	sts1.Spec.Template.Spec.TerminationGracePeriodSeconds = sts2.Spec.Template.Spec.TerminationGracePeriodSeconds
	sts1.Spec.Template.Spec.RestartPolicy = sts2.Spec.Template.Spec.RestartPolicy
	sts1.Spec.Template.Spec.SecurityContext = sts2.Spec.Template.Spec.SecurityContext

	for i := 0; i < len(sts1.Spec.Template.Spec.Containers); i++ {
		sts1.Spec.Template.Spec.Containers[i].LivenessProbe = sts2.Spec.Template.Spec.Containers[i].LivenessProbe
		sts1.Spec.Template.Spec.Containers[i].ReadinessProbe = sts2.Spec.Template.Spec.Containers[i].ReadinessProbe
		sts1.Spec.Template.Spec.Containers[i].ImagePullPolicy = sts2.Spec.Template.Spec.Containers[i].ImagePullPolicy

		sts1.Spec.Template.Spec.Containers[i].TerminationMessagePath = sts2.Spec.Template.Spec.Containers[i].TerminationMessagePath
		sts1.Spec.Template.Spec.Containers[i].TerminationMessagePolicy = sts2.Spec.Template.Spec.Containers[i].TerminationMessagePolicy

		sts1.Spec.Template.Spec.Containers[i].SecurityContext = sts2.Spec.Template.Spec.Containers[i].SecurityContext
		sts1.Spec.Template.Spec.Containers[i].Resources = sts2.Spec.Template.Spec.Containers[i].Resources
	}

	for i := 0; i < len(sts1.Spec.Template.Spec.InitContainers); i++ {
		sts1.Spec.Template.Spec.InitContainers[i].LivenessProbe = sts2.Spec.Template.Spec.InitContainers[i].LivenessProbe
		sts1.Spec.Template.Spec.InitContainers[i].ReadinessProbe = sts2.Spec.Template.Spec.InitContainers[i].ReadinessProbe
		sts1.Spec.Template.Spec.InitContainers[i].ImagePullPolicy = sts2.Spec.Template.Spec.InitContainers[i].ImagePullPolicy

		sts1.Spec.Template.Spec.InitContainers[i].TerminationMessagePath = sts2.Spec.Template.Spec.InitContainers[i].TerminationMessagePath
		sts1.Spec.Template.Spec.InitContainers[i].TerminationMessagePolicy = sts2.Spec.Template.Spec.InitContainers[i].TerminationMessagePolicy
	}

	//some defaultMode changes make falsepositif, so we bypass this, we already have check on configmap changes
	sts1.Spec.VolumeClaimTemplates = sts2.Spec.VolumeClaimTemplates
	sts1.Spec.PodManagementPolicy = sts2.Spec.PodManagementPolicy
	sts1.Spec.RevisionHistoryLimit = sts2.Spec.RevisionHistoryLimit
	sts1.Spec.UpdateStrategy = sts2.Spec.UpdateStrategy

	if !apiequality.Semantic.DeepEqual(sts1.Spec, sts2.Spec) {
		logrus.WithFields(logrus.Fields{"statefulset": sts1.Name,
			"namespace": sts1.Namespace}).Info("Template is different: " + pretty.Compare(sts1.Spec, sts2.Spec))
		return false
	}

	return true
}

func generateResourceList(cpuMem clickhousev1.CPUAndMem) v1.ResourceList {
	cpu, memory := cpuMem.CPU, cpuMem.Memory
	resources := v1.ResourceList{}
	if cpu != "" {
		resources[v1.ResourceCPU], _ = resource.ParseQuantity(cpu)
	}
	if memory != "" {
		resources[v1.ResourceMemory], _ = resource.ParseQuantity(memory)
	}
	return resources
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}
