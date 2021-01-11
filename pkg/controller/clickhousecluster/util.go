package clickhousecluster

import (
	"encoding/xml"
	"fmt"
	"io"
	"math/rand"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"

	"strings"

	clickhousev1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
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
		//sts1.Spec.Template.Spec.Containers[i].Resources = sts2.Spec.Template.Spec.Containers[i].Resources
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

// Retry
func Retry(tries int, desc string, f func() error) error {
	var err error
	for try := 1; try <= tries; try++ {
		err = f()
		if err == nil {
			// All ok, no need to retry more
			if try > 1 {
				// Done, but after some retries, this is not 'clean'
				logrus.Infof("DONE attempt %d of %d: %s", try, tries, desc)
			}
			return nil
		}

		if try < tries {
			// Try failed, need to sleep and retry
			seconds := try * 5
			logrus.Infof("FAILED attempt %d of %d, sleep %d sec and retry: %s", try, tries, seconds, desc)
			select {
			case <-time.After(time.Duration(seconds) * time.Second):
			}
		} else if tries == 1 {
			// On single try do not put so much emotion. It just failed and user is not intended to retry
			logrus.Infof("FAILED single try. No retries will be made for %s", desc)
		} else {
			// On last try no need to wait more
			logrus.Infof("FAILED AND ABORT. All %d attempts: %s", tries, desc)
		}
	}

	return err
}

func decodeUsersXML(users string) map[string]string {
	result := make(map[string]string)
	if strings.TrimSpace(users) == "" {
		return result
	}
	decoder := xml.NewDecoder(strings.NewReader(users))
	var startRecord bool
	var lastKey, userKey string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return result
		}
		if err != nil {
			logrus.Errorf("parse Fail: %s", err.Error())
			return result
		}
		switch tp := token.(type) {
		case xml.StartElement:
			se := xml.StartElement(tp)
			if lastKey == "users" || startRecord {
				userKey = se.Name.Local
				result[userKey] = ""
				startRecord = false
			}
			lastKey = se.Name.Local
		case xml.EndElement:
			ee := xml.EndElement(tp)
			if ee.Name.Local == userKey {
				startRecord = true
			}
			if ee.Name.Local == "yandex" {
				return result
			}
		case xml.CharData:
			cd := xml.CharData(tp)
			data := strings.TrimSpace(string(cd))
			if len(data) != 0 && lastKey == "password" {
				result[userKey] = data
				return result
			}
		}
	}
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func validateResource(cc *clickhousev1.ClickHouseCluster) error {
	_, err := resource.ParseQuantity(cc.Spec.Resources.Limits.Memory)
	if err != nil {
		return err
	}
	_, err = resource.ParseQuantity(cc.Spec.Resources.Limits.CPU)
	if err != nil {
		return err
	}
	_, err = resource.ParseQuantity(cc.Spec.Resources.Requests.Memory)
	if err != nil {
		return err
	}
	_, err = resource.ParseQuantity(cc.Spec.Resources.Requests.CPU)
	if err != nil {
		return err
	}
	return nil
}
