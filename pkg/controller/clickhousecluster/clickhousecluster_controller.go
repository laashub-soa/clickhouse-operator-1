package clickhousecluster

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"path/filepath"
	"strings"
	"time"

	clickhousev1 "github.com/mackwong/clickhouse-operator/pkg/apis/clickhouse/v1"
	"github.com/sirupsen/logrus"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_clickhousecluster")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new ClickHouseCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	commonConfigs := make(map[string]string)
	if err := filepath.Walk("./config", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".xml") || strings.HasSuffix(path, ".json") {
			content, err := ioutil.ReadFile(path)
			if err != nil {
				logrus.Error(err)
				return err
			}
			commonConfigs[info.Name()] = string(content)
		}
		return nil
	}); err != nil {
		logrus.Fatal(err)
	}
	return &ReconcileClickHouseCluster{client: mgr.GetClient(), scheme: mgr.GetScheme(), commonConfigs: commonConfigs}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clickhousecluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ClickHouseCluster
	err = c.Watch(&source.Kind{Type: &clickhousev1.ClickHouseCluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner ClickHouseCluster
	//err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
	//	IsController: true,
	//	OwnerType:    &clickhousev1.ClickHouseCluster{},
	//})
	//if err != nil {
	//	return err
	//}

	return nil
}

// blank assignment to verify that ReconcileClickHouseCluster implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileClickHouseCluster{}

// ReconcileClickHouseCluster reconciles a ClickHouseCluster object
type ReconcileClickHouseCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme

	commonConfigs map[string]string
}

func (r *ReconcileClickHouseCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling ClickHouseCluster")

	requeue5 := reconcile.Result{RequeueAfter: 5 * time.Second}
	//requeue := reconcile.Result{Requeue: true}
	forget := reconcile.Result{}

	// Fetch the ClickHouseCluster instance
	instance := &clickhousev1.ClickHouseCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			logrus.Info("Delete ClickHouseCluster")
			return forget, nil
		}
		return forget, err
	}

	if instance.Status.Status == "" {
		changed := instance.SetDefaults()
		if changed {
			logrus.WithFields(logrus.Fields{
				"cluster":   instance.Name,
				"namespace": instance.Namespace}).Info("Update ClickHouseCluster")
			err = r.client.Update(context.TODO(), instance)
			return forget, err
		}
	}

	status := instance.Status.DeepCopy()
	defer r.updateClickHouseStatus(instance, status)

	if r.CheckNonAllowedChanges(instance) {
		return requeue5, nil
	}

	if err = r.reconcile(instance); err != nil {
		return requeue5, err
	}

	return forget, nil
}

func (r *ReconcileClickHouseCluster) updateClickHouseStatus(
	instance *clickhousev1.ClickHouseCluster,
	status *clickhousev1.ClickHouseClusterStatus) {
	return
}

func (r *ReconcileClickHouseCluster) reconcile(instance *clickhousev1.ClickHouseCluster) error {
	var generator = NewGenerator(r, instance)

	//Service for Clickhouse
	service := generator.GenerateService()
	if err := r.ReconcileService(service); err != nil {
		logrus.WithFields(logrus.Fields{"namespace": service.Namespace, "name": service.Name, "error": err}).Error("create service error")
		return err
	}

	commonConfigMap := generator.GenerateCommonConfigMap()
	if err := r.ReconcileConfigMap(commonConfigMap); err != nil {
		logrus.WithFields(logrus.Fields{"namespace": commonConfigMap.Namespace, "name": commonConfigMap.Name, "error": err}).Error("create command configmap error")
		return err
	}

	//userConfigMap := generator.generateUserConfigMap()
	//if err := r.ReconcileConfigMap(userConfigMap); err != nil {
	//	logrus.WithFields(logrus.Fields{"namespace": userConfigMap.Namespace, "name": userConfigMap.Name, "error": err}).Error("create user configmap error")
	//	return err
	//}

	statefulSets := generator.GenerateStatefulSets()
	for _, s := range statefulSets {
		if err := r.ReconcileStatefulSet(s); err != nil {
			logrus.WithFields(logrus.Fields{"namespace": s.Namespace, "name": s.Name, "error": err}).Error("create statefulSets error")
			return err
		}
	}
	return nil
}

// CheckNonAllowedChanges - checks if there are some changes on CRD that are not allowed on statefulset
// If a non Allowed Changed is Find we won't Update associated kubernetes objects, but we will put back the old value
// and Patch the CRD with correct values
func (r *ReconcileClickHouseCluster) CheckNonAllowedChanges(instance *clickhousev1.ClickHouseCluster) bool {
	var oldCRD clickhousev1.ClickHouseCluster
	if instance.Annotations[clickhousev1.AnnotationLastApplied] == "" {
		return false
	}

	//We retrieved our last-applied-configuration stored in the CRD
	err := json.Unmarshal([]byte(instance.Annotations[clickhousev1.AnnotationLastApplied]), &oldCRD)
	if err != nil {
		logrus.WithFields(logrus.Fields{"cluster": instance.Name}).Error("Can't get Old version of CRD")
		return false
	}
	//DataCapacity change is forbidden
	if instance.Spec.DataCapacity != oldCRD.Spec.DataCapacity {
		logrus.WithFields(logrus.Fields{"cluster": instance.Name}).
			Warningf("The Operator has refused the change on DataCapacity from [%s] to NewValue[%s]",
				oldCRD.Spec.DataCapacity, instance.Spec.DataCapacity)
		instance.Spec.DataCapacity = oldCRD.Spec.DataCapacity
		return true
	}
	//DataStorage
	if instance.Spec.DataStorageClass != oldCRD.Spec.DataStorageClass {
		logrus.WithFields(logrus.Fields{"cluster": instance.Name}).
			Warningf("The Operator has refused the change on DataStorageClass from [%s] to NewValue[%s]",
				oldCRD.Spec.DataStorageClass, instance.Spec.DataStorageClass)
		instance.Spec.DataStorageClass = oldCRD.Spec.DataStorageClass
		return true
	}
	return false
}

func (r *ReconcileClickHouseCluster) ReconcileService(service *corev1.Service) error {
	var curService corev1.Service
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: service.Namespace, Name: service.Name}, &curService)
	// Object with such name does not exist or error happened
	if err != nil {
		if apierrors.IsNotFound(err) {
			logrus.WithFields(logrus.Fields{
				"service":   service.Name,
				"namespace": service.Namespace}).Info("Create Service")
			// Object with such name not found - create it
			return r.client.Create(context.TODO(), service)
		}
		return err
	}

	logrus.WithFields(logrus.Fields{
		"service":   service.Name,
		"namespace": service.Namespace}).Info("Update Service")
	// spec.resourceVersion is required in order to update Service
	service.ResourceVersion = curService.ResourceVersion
	// spec.clusterIP field is immutable, need to use already assigned value
	// From https://kubernetes.io/docs/concepts/services-networking/service/#defining-a-service
	// Kubernetes assigns this Service an IP address (sometimes called the “cluster IP”), which is used by the Service proxies
	// See also https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies
	// You can specify your own cluster IP address as part of a Service creation request. To do this, set the .spec.clusterIP
	service.Spec.ClusterIP = curService.Spec.ClusterIP
	return r.client.Update(context.TODO(), service)
}

func (r *ReconcileClickHouseCluster) ReconcileConfigMap(configMap *corev1.ConfigMap) error {
	var curConfigMap corev1.ConfigMap
	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: configMap.Namespace, Name: configMap.Name}, &curConfigMap)
	// Object with such name does not exist or error happened
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object with such name not found - create it
			logrus.WithFields(logrus.Fields{
				"configmap": configMap.Name,
				"namespace": configMap.Namespace}).Info("Create ConfigMap")
			return r.client.Create(context.TODO(), configMap)
		}
		return err
	}

	logrus.WithFields(logrus.Fields{
		"configmap": configMap.Name,
		"namespace": configMap.Namespace}).Info("Update ConfigMap")
	return r.client.Update(context.TODO(), configMap)
}

func (r *ReconcileClickHouseCluster) ReconcileStatefulSet(statefulSet *appsv1.StatefulSet) error {
	// Check whether object with such name already exists in k8s
	var curStatefulSet appsv1.StatefulSet

	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: statefulSet.Namespace, Name: statefulSet.Name}, &curStatefulSet)

	// Object with such name does not exist or error happened
	if err != nil {
		if apierrors.IsNotFound(err) {
			// Object with such name not found - create it
			logrus.WithFields(logrus.Fields{
				"statefulset": statefulSet.Name,
				"namespace":   statefulSet.Namespace}).Info("Create StatefulSet")
			return r.client.Create(context.TODO(), statefulSet)
		}
		return err
	}

	logrus.WithFields(logrus.Fields{
		"statefulSet": statefulSet.Name,
		"namespace":   statefulSet.Namespace}).Info("Update StatefulSet")
	return r.client.Update(context.TODO(), statefulSet)
}

////Todo: use object replace statefulset/service/configmap
//func (r *ReconcileClickHouseCluster) ReconcileObject(obj runtime.Object) error {
//	// Check whether object with such name already exists in k8s
//	var curObject runtime.Object
//
//	err := r.client.Get(context.TODO(), types.NamespacedName{Namespace: statefulSet.Namespace, Name: statefulSet.Name}, &curObject)
//
//	// Object with such name does not exist or error happened
//	if err != nil {
//		if apierrors.IsNotFound(err) {
//			// Object with such name not found - create it
//			logrus.WithFields(logrus.Fields{
//				"statefulset":   statefulSet.Name,
//				"namespace": statefulSet.Namespace}).Info("Create StatefulSet")
//			return r.client.Create(context.TODO(), statefulSet)
//		}
//		return err
//	}
//
//	logrus.WithFields(logrus.Fields{
//		"configmap":   statefulSet.Name,
//		"namespace": statefulSet.Namespace}).Info("Update ConfigMap")
//	return r.client.Update(context.TODO(), statefulSet)
//}
