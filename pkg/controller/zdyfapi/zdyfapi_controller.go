package zdyfapi

import (
	"context"
	_ "encoding/asn1"
	"encoding/json"
	_ "go/ast"
	_ "k8s.io/api/extensions/v1beta1"
	_ "k8s.io/api/storage/v1"
	v1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	_ "runtime/pprof"
	_ "time"
	zdyfv1alpha1 "zdyfop/pkg/apis/zdyf/v1alpha1"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	_ "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_zdyfapi")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Zdyfapi Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileZdyfapi{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("zdyfapi-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Zdyfapi
	err = c.Watch(&source.Kind{Type: &zdyfv1alpha1.Zdyfapi{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource deployment and requeue the owner Zdyfapi
	//the deployment's owner is zdyfapi
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &zdyfv1alpha1.Zdyfapi{},
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileZdyfapi implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileZdyfapi{}

// ReconcileZdyfapi reconciles a Zdyfapi object
type ReconcileZdyfapi struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Zdyfapi object and makes changes based on the state read
// and what is in the Zdyfapi.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates two deployment for different nginx's version as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileZdyfapi) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Zdyfapi")



	// Fetch the Zdyfapi instance for hdfs
	//next := 0
	instance := &zdyfv1alpha1.Zdyfapi{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("instance resource not found.Ignoring since object nust be deleted.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get instance.")
		return reconcile.Result{}, err
	}
	//如果不为空，就会阻止删除
	if instance.DeletionTimestamp != nil {
		return reconcile.Result{}, nil
	}

	reqLogger.Info("instance is", *instance, *instance.Spec.NameReplicas)

	//HDFS sc
	foundsc := v1.StorageClass{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &foundsc)
	if err != nil && errors.IsNotFound(err) {
		namesc := r.NewSc(instance)
		reqLogger.Info("Creating a new sc ", "StorageClass.Namespace", namesc.Namespace, "StorageClass.Name", namesc.Name)
		err = r.client.Create(context.TODO(), namesc)
		if err != nil {
			reqLogger.Error(err, "Failed to create new sc", "StorageClass.Namespace", namesc.Namespace, "StorageClass.Name", namesc.Name)
			return reconcile.Result{}, err
		}
		reqLogger.Info("Creating a new sc successfully")

		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		//return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get pv .")
		return reconcile.Result{}, err
	}

	reqLogger.Info("next is name pv!!!!")

	//namenode pv
	//create pv pvc before the deployment and service
	foundpv := &corev1.PersistentVolume{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundpv)
	if err != nil && errors.IsNotFound(err) {
		namepv := r.NameNewPv(instance)
		reqLogger.Info("Creating a new Pv.", "PersistentVolume.Namespace", namepv.Namespace, "PersistentVolume.Name", namepv.Name)
		err = r.client.Create(context.TODO(), namepv) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new pv .", "PersistentVolume.Namespace", namepv.Namespace, "PersistentVolume.Name", namepv.Name)
			return reconcile.Result{}, err
		}
		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		// return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get name pv .")
		return reconcile.Result{}, err
	}

	//namenode pvc
	foundpvc := &corev1.PersistentVolumeClaim{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundpvc)
	if err != nil && errors.IsNotFound(err) {
		namepvc := r.NameNewPvc(instance)
		reqLogger.Info("Creating a new pvc.", "PersistentVolumeClaim.Namespace", namepvc.Namespace, "PersistentVolumeClaim.Name", namepvc.Name)
		err = r.client.Create(context.TODO(), namepvc) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Info("the err is ", err)
			reqLogger.Error(err, "Failed to create new pvc.", "PersistentVolumeClaim.Namespace", namepvc.Namespace, "PersistentVolumeClaim.Name", namepvc.Name)
			return reconcile.Result{}, err
		}

		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}

	} else if err != nil {
		reqLogger.Error(err, "Failed to get pvc .")
		return reconcile.Result{}, err
	}
	///
	//datanode pv
	dfoundpv := &corev1.PersistentVolume{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, dfoundpv)
	if err != nil && errors.IsNotFound(err) {
		datapv := r.DataNewPv(instance)
		reqLogger.Info("Creating a new data Pv.", "PersistentVolume.Namespace", datapv.Namespace, "PersistentVolume.Name", datapv.Name)
		err = r.client.Create(context.TODO(), datapv) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new data pv .", "PersistentVolume.Namespace", datapv.Namespace, "PersistentVolume.Name", datapv.Name)
			return reconcile.Result{}, err
		}
		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		// return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get data pv .")
		return reconcile.Result{}, err
	}
	///////
	//datanode the second pv
	dfoundpvse := &corev1.PersistentVolume{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, dfoundpvse)
	if err != nil && errors.IsNotFound(err) {
		datapvs := r.DataNewPvse(instance)
		reqLogger.Info("Creating next new data Pv two.", "PersistentVolume.Namespace", datapvs.Namespace, "PersistentVolume.Name", datapvs.Name)
		err = r.client.Create(context.TODO(), datapvs) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new data pv .", "PersistentVolume.Namespace", datapvs.Namespace, "PersistentVolume.Name", datapvs.Name)
			return reconcile.Result{}, err
		}
		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		// return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get data pv .")
		return reconcile.Result{}, err
	}
	///////
	//datanode the 3 pv
	dfoundpvth := &corev1.PersistentVolume{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, dfoundpvth)
	if err != nil && errors.IsNotFound(err) {
		datapvs := r.DataNewPvth(instance)
		reqLogger.Info("Creating next new data Pv two.", "PersistentVolume.Namespace", datapvs.Namespace, "PersistentVolume.Name", datapvs.Name)
		err = r.client.Create(context.TODO(), datapvs) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new data pv .", "PersistentVolume.Namespace", datapvs.Namespace, "PersistentVolume.Name", datapvs.Name)
			return reconcile.Result{}, err
		}
		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		// return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get data pv .")
		return reconcile.Result{}, err
	}
	//////
	//datanode the 4 pv
	dfoundpvfo := &corev1.PersistentVolume{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, dfoundpvfo)
	if err != nil && errors.IsNotFound(err) {
		datapvs := r.DataNewPvfo(instance)
		reqLogger.Info("Creating next new data Pv two.", "PersistentVolume.Namespace", datapvs.Namespace, "PersistentVolume.Name", datapvs.Name)
		err = r.client.Create(context.TODO(), datapvs) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new data pv .", "PersistentVolume.Namespace", datapvs.Namespace, "PersistentVolume.Name", datapvs.Name)
			return reconcile.Result{}, err
		}
		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		// return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get data pv .")
		return reconcile.Result{}, err
	}
	///////
	//datanode the 5 pv
	dfoundpvfi := &corev1.PersistentVolume{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, dfoundpvfi)
	if err != nil && errors.IsNotFound(err) {
		datapvs := r.DataNewPvfi(instance)
		reqLogger.Info("Creating next new data Pv five.", "PersistentVolume.Namespace", datapvs.Namespace, "PersistentVolume.Name", datapvs.Name)
		err = r.client.Create(context.TODO(), datapvs) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new data pv .", "PersistentVolume.Namespace", datapvs.Namespace, "PersistentVolume.Name", datapvs.Name)
			return reconcile.Result{}, err
		}
		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		// return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get data pv .")
		return reconcile.Result{}, err
	}
	////

	//Check if the deployment and service  already exists, if not create a new one
	founddep := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, founddep) //request.NamespacedName,
	if err != nil && errors.IsNotFound(err) {
		//define a new deployment namenodedep  (instance)
		namenodedep := r.NewDeploy(instance)
		reqLogger.Info("Creating a new Deployment namenodedep.", "Deployment.Namespace", namenodedep.Namespace, "Deployment.Name", namenodedep.Name)
		err = r.client.Create(context.TODO(), namenodedep) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment namenodedep.", "Deployment.Namespace", namenodedep.Namespace, "Deployment.Name", namenodedep.Name)
			return reconcile.Result{}, err
		}
		//deployment namenodedep created successfully - return and requeue
		reqLogger.Info("After create namenodedep successfully, create the next nginx version")

		//define a new service namenodeser
		namenodeser := r.NameNewService(instance)
		reqLogger.Info("Creating a new service namenodeser.", "Service.Namespace", namenodedep.Namespace, "Service.Name", namenodedep.Name)
		err = r.client.Create(context.TODO(), namenodeser) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service namenodeser.", "Service.Namespace", namenodedep.Namespace, "Service.Name", namenodedep.Name)
			return reconcile.Result{}, err
		}
		//service namenodeser created successfully - return and requeue
		reqLogger.Info("After create namenodeser successfully", "Test", instance.Spec.NameEnvsName2)

		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		//return reconcile.Result{Requeue: true}, nil

	} else if err != nil {
		reqLogger.Error(err, "Failed to get namenode Deployment and service .")
		return reconcile.Result{}, err
	}
	reqLogger.Info("namenode is ok!!!!!!")
	/////////

	//datanode 先service 后statefulset
	foundser := corev1.Service{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Namespace: instance.Namespace, Name: instance.Name}, &foundser)
	if err != nil && errors.IsNotFound(err) {
		//define a new service datanodeser
		datanodeser := r.DataNewService(instance)
		reqLogger.Info("Creating a new service datanodeser.", "Service.Namespace", datanodeser.Namespace, "Service.Name", datanodeser.Name)
		err = r.client.Create(context.TODO(), datanodeser) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service datanodeser.", "Service.Namespace", datanodeser.Namespace, "Service.Name", datanodeser.Name)
			return reconcile.Result{}, err
		}
		//service namenodeser created successfully - return and requeue
		reqLogger.Info("After create datanodeser successfully")

		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		//return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Statefulset datanode.")
		return reconcile.Result{}, err
	}

	//datanode statefulset
	foundset := &appsv1.StatefulSet{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, foundset)
	if err != nil && errors.IsNotFound(err) {

		datanodeset := r.NewState(instance)
		reqLogger.Info("Creating a new StatefulSet datanodeset.", "StatefulSet.Namespace", datanodeset.Namespace, "StatefulSet.Name", datanodeset.Name)
		err = r.client.Create(context.TODO(), datanodeset) //saves the object obj in the Kubernetes cluster
		if err != nil {
			reqLogger.Error(err, "Failed to create new StatefulSet datanodeset.", "StatefulSet.Namespace", datanodeset.Namespace, "StatefulSet.Name", datanodeset.Name)
			return reconcile.Result{}, err
		}
		reqLogger.Info("After create datanodeset successfull")
		// 关联 Annotations - spec
		data, _ := json.Marshal(instance.Spec)
		if instance.Annotations != nil {
			instance.Annotations["spec"] = string(data)
		} else {
			instance.Annotations = map[string]string{"spec": string(data)}
		}

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, nil
		}
		//return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Statefulset datanode.")
		return reconcile.Result{}, err
	}

	//////////////////////////datanode////////////////////////////////

	oldSpec := &zdyfv1alpha1.Zdyfapi{}
	if err := json.Unmarshal([]byte(instance.Annotations["spec"]), oldSpec); err != nil {
		return reconcile.Result{}, err
	}
	// update deployment and service 's status
	if !reflect.DeepEqual(instance.Spec, oldSpec) {
		// 更新关联资源
		//namenode
		namenewDeploy := r.NewDeploy(instance)
		nameoldDeploy := &appsv1.Deployment{}
		nameoldDeploy.Spec = namenewDeploy.Spec
		if err := r.client.Get(context.TODO(), request.NamespacedName, nameoldDeploy); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.client.Update(context.TODO(), nameoldDeploy); err != nil {
			return reconcile.Result{}, err
		}

		//update the service for namenode
		namenewSer := r.NameNewService(instance)
		nameoldSer := &corev1.Service{}
		nameoldSer.Spec = namenewSer.Spec
		if err := r.client.Get(context.TODO(), request.NamespacedName, nameoldSer); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.client.Update(context.TODO(), nameoldSer); err != nil {
			return reconcile.Result{}, err
		}
		// update the pv for namenode
		namenewpv := r.NameNewPv(instance)
		nameoldpv := &corev1.PersistentVolume{}
		nameoldpv.Spec = namenewpv.Spec
		if err := r.client.Get(context.TODO(), request.NamespacedName, nameoldpv); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.client.Update(context.TODO(), nameoldpv); err != nil {
			return reconcile.Result{}, err
		}

		namenewpvc := r.NameNewPvc(instance)
		nameoldpvc := &corev1.PersistentVolumeClaim{}
		nameoldpvc.Spec = namenewpvc.Spec
		if err := r.client.Get(context.TODO(), request.NamespacedName, nameoldpvc); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.client.Update(context.TODO(), nameoldpvc); err != nil {
			return reconcile.Result{}, err
		}
		//datanode
		newState := r.NewState(instance)
		oldState := &appsv1.StatefulSet{}
		oldState.Spec = newState.Spec
		if err := r.client.Get(context.TODO(), request.NamespacedName, oldState); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.client.Update(context.TODO(), oldState); err != nil {
			return reconcile.Result{}, err
		}

		datanewSer := r.DataNewService(instance)
		dataoldSer := &corev1.Service{}
		dataoldSer.Spec = datanewSer.Spec
		if err := r.client.Get(context.TODO(), request.NamespacedName, dataoldSer); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.client.Update(context.TODO(), dataoldSer); err != nil {
			return reconcile.Result{}, err
		}

		datanewPv := r.DataNewPv(instance)
		dataoldPv := &corev1.PersistentVolume{}
		dataoldPv.Spec = datanewPv.Spec
		if err := r.client.Get(context.TODO(), request.NamespacedName, dataoldPv); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.client.Update(context.TODO(), dataoldPv); err != nil {
			return reconcile.Result{}, err
		}

		datanewPvs := r.DataNewPv(instance)
		dataoldPvs := &corev1.PersistentVolume{}
		dataoldPvs.Spec = datanewPvs.Spec
		if err := r.client.Get(context.TODO(), request.NamespacedName, dataoldPvs); err != nil {
			return reconcile.Result{}, err
		}
		if err := r.client.Update(context.TODO(), dataoldPvs); err != nil {
			return reconcile.Result{}, err
		}

		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil

}

func labelsFornamenode(name string) map[string]string {
	return map[string]string{"app": name, "hasuraService": "custom"}
}

//newdeploy for namenode
func (r *ReconcileZdyfapi) NewDeploy(m *zdyfv1alpha1.Zdyfapi) *appsv1.Deployment {
	ls := labelsFornamenode(m.Name)
	time := int64(0)
	dep := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "namenode-dep",
			Namespace: m.Namespace,
			////CreationTimestamp: x,
			Labels: ls,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: m.Spec.NameReplicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: ls,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					//CreationTimestamp: nil,
					Labels: ls,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:    "namenode-dep-pod",
						Image:   m.Spec.NameImage,
						Command: []string{"/run.sh"},
						//Resources:       m.Spec.Resources,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{{
							Name:  m.Spec.NameEnvsName,
							Value: m.Spec.NameEnvsValue,
						},
							{
								Name:  m.Spec.NameEnvsName2,
								Value: m.Spec.NameEnvsValue2,
							},
							/*		{
										Name:  m.Spec.NameEnvsName3,
										Value: m.Spec.NameEnvsValue3,
									},
										{
										Name:  m.Spec.NameEnvsName4,
										Value: m.Spec.NameEnvsValue4,
									},*/
						},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      m.Spec.NamePVName, //namenode-v
							MountPath: m.Spec.NameVolumeMountsPath,
						}},
						Ports: []corev1.ContainerPort{
							{
								Name:          m.Spec.NameContainerPortName1, //nn-rpc
								ContainerPort: m.Spec.NameContainerPort1,
							},
							{
								Name:          m.Spec.NameContainerPortName2, //nn-web
								ContainerPort: m.Spec.NameContainerPort2,
							},
						},
					}},
					Volumes: []corev1.Volume{{
						Name: m.Spec.NamePVName, //namenode-v
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: m.Spec.NamePVCName, //namenode-pvc
							},
						},
					}},
					SecurityContext:               &corev1.PodSecurityContext{},
					TerminationGracePeriodSeconds: &time,
				},
			},
		},
	}
	return dep
}

//newservice for namenode
func (r *ReconcileZdyfapi) NameNewService(m *zdyfv1alpha1.Zdyfapi) *corev1.Service {
	ls := labelsFornamenode(m.Name)
	ser := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "namenode-ser",
			Namespace: m.Namespace,
			//CreationTimestamp: nil,
			Labels: ls,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Port:     m.Spec.NameContainerPort1,
					Protocol: corev1.Protocol(corev1.ProtocolTCP),
					Name:     m.Spec.NameContainerPortName1,
				},
				{
					Port:     m.Spec.NameContainerPort2,
					Protocol: corev1.Protocol(corev1.ProtocolTCP),
					Name:     m.Spec.NameContainerPortName2,
				},
			},
			Selector: map[string]string{
				"app": "namenode",
			},
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{},
		},
	}
	return ser
}

//Namenode pv just one.
func (r *ReconcileZdyfapi) NameNewPv(m *zdyfv1alpha1.Zdyfapi) *corev1.PersistentVolume {

	npv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Spec.NamePVName,
			Namespace: m.Namespace,
			Labels:    map[string]string{"app": "nn-pv-1"},
			//Annotations: map[string]string{"type": "namenode"},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse(m.Spec.NamePvStorage),
			},
			//VolumeMode:  &corev1.PersistentVolumeFilesystem,
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName:              m.Spec.NameSCName, //local volume 只提供了卷的延迟绑定
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				Local: &corev1.LocalVolumeSource{
					Path: "/var/lib/docker/volumes/namenode/_data",
				},
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{ //pv node 关联没设置好
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"docker-desktop"}, //设置节点
						}},
					},
					},
				},
			},
		},
	}
	return npv
}

//Namenode's pvc
func (r *ReconcileZdyfapi) NameNewPvc(m *zdyfv1alpha1.Zdyfapi) *corev1.PersistentVolumeClaim {
	scname := m.Spec.NameSCName
	npvc := &corev1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolumeClaim",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Spec.NamePVCName,
			Namespace: m.Namespace,
			//Annotations: map[string]string{"type": "namenode"},
			Labels: map[string]string{"app": "nn-pv-1"},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"type": "namenode"},
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			VolumeName:       m.Spec.NamePVName,
			StorageClassName: &scname, //设置 ！！！！
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					"storage": resource.MustParse(m.Spec.NamePVCStorage),
				},
			},
		},
	}
	return npvc
}

//newstate for datanode
func (r *ReconcileZdyfapi) NewState(m *zdyfv1alpha1.Zdyfapi) *appsv1.StatefulSet {
	foundsc := m.Spec.NameSCName //r.NewSc(m *zdyfv1alpha1.Zdyfapi)
	time := int64(0)

	stateset := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "datanode-set",
			Namespace: m.Namespace,
			//CreationTimestamp: nil,
			Labels: map[string]string{"app": "datanode", "hasuraService": "custom"},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName: m.Spec.ServiceName,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "datanode"},
			},
			Replicas: m.Spec.DataReplicas,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					//CreationTimestamp: nil,
					Labels: map[string]string{"app": "datanode"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{ //newContainers(m),
						Image:   m.Spec.DataImage,
						Name:    "datanode-set-pod",
						Command: []string{"/run.sh"},
						//Resources:       m.Spec.Resources,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Env: []corev1.EnvVar{{
							Name:  m.Spec.DataEnvsName,
							Value: m.Spec.DataEnvsValue,
						}},
						VolumeMounts: []corev1.VolumeMount{{
							Name:      m.Spec.DataPVCName1, //与PVC保持一致
							MountPath: m.Spec.DataVolumeMountsPath,
						}},
					}},
					SecurityContext:               &corev1.PodSecurityContext{},
					TerminationGracePeriodSeconds: &time,

					Volumes: []corev1.Volume{{
						Name: m.Spec.DataPVCName1,
						VolumeSource: corev1.VolumeSource{
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
								ClaimName: m.Spec.DataPVCName1,
							},
						},
					}},
				},
			},
			//datanode 的；两个pvc
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:        m.Spec.DataPVCName1,
						Annotations: map[string]string{"volume.beta.kubernetes.io/storage-class": m.Spec.NameSCName},
						Labels:      map[string]string{"pv-pvc": "dn-pv-1"},
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							corev1.ReadWriteOnce,
						},
						//VolumeName:      m.Spec.DataPVName1,
						StorageClassName: &foundsc, //&defaultStorageClass,/////////////////
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								"storage": resource.MustParse(m.Spec.DataPVCStorage1), //  2-》3
							},
						}, /*
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"pv-pvc": "dn-pv-1"},
							},*/
					},
				},
				//pvc2
				/*
					{

						ObjectMeta: metav1.ObjectMeta{
							Name:        m.Spec.DataPVCName2,
							Annotations: map[string]string{"volume.beta.kubernetes.io/storage-class": m.Spec.NameSCName},
							Labels:      map[string]string{"pv-pvc": "dn-pv-2"},
						},
						Spec: corev1.PersistentVolumeClaimSpec{
							AccessModes: []corev1.PersistentVolumeAccessMode{
								corev1.ReadWriteOnce,
							},
							VolumeName:       m.Spec.DataPVName2,
							StorageClassName: &foundsc, //&defaultStorageClass,/////////////////
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									"storage": resource.MustParse(m.Spec.DataPVCStorage2),
								},
							},
							Selector: &metav1.LabelSelector{
								MatchLabels: map[string]string{"pv-pvc": "dn-pv-2"},
							},
						},
					}, */
			},
		},
		Status: appsv1.StatefulSetStatus{},
	}
	return stateset
}

//newservice for datanode
func (r *ReconcileZdyfapi) DataNewService(m *zdyfv1alpha1.Zdyfapi) *corev1.Service {
	//ls := labelsFornamenode(m.Name)
	ser := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Spec.ServiceName,
			Namespace: m.Namespace,
			//CreationTimestamp: nil,
			Labels: map[string]string{"app": "datanode", "hasuraService": "custom"},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Port:     m.Spec.DataSport,
				Protocol: corev1.ProtocolTCP,
			}},
			Selector: map[string]string{"app": "datanode"},
			//Type:     corev1.ServiceType(corev1.ServiceTypeClusterIP, corev1.ServiceTypeNodePort),
			//ClusterIP: nil,
		},
		Status: corev1.ServiceStatus{
			LoadBalancer: corev1.LoadBalancerStatus{},
		},
	}
	return ser
}

//storageclass for local volumes

func (r *ReconcileZdyfapi) NewSc(m *zdyfv1alpha1.Zdyfapi) *v1.StorageClass {
	rp := corev1.PersistentVolumeReclaimRetain
	mode := v1.VolumeBindingWaitForFirstConsumer
	tr := true
	sc := &v1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StorageClass",
			APIVersion: "storage.k8s.io/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Spec.NameSCName,
			Namespace:   m.Namespace,
			Annotations: map[string]string{"storage-class": m.Name}, //sc 与 PVC
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Provisioner: "kubernetes.io/no-provisioner", //no-provisioner
		Parameters:  map[string]string{
			//"type": m.Spec.DataSC,
		},
		AllowVolumeExpansion: &tr, //expansion the volume
		VolumeBindingMode:    &mode,
		ReclaimPolicy:        &rp,
	}
	return sc
}

//用sc 创建一个 datanode 的 pv

func (r *ReconcileZdyfapi) DataNewPv(m *zdyfv1alpha1.Zdyfapi) *corev1.PersistentVolume {

	dpv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Spec.DataPVName1,
			Labels:      map[string]string{"app": "datanode"},
			Annotations: map[string]string{"volume.beta.kubernetes.io/storage-class": m.Spec.NameSCName},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse(m.Spec.DataPVStorage1),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName:              m.Spec.NameSCName,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				Local: &corev1.LocalVolumeSource{
					Path: "/var/lib/docker/volumes/datanode-1/_data",
				},
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"docker-desktop"},
						}},
					}},
				},
			},
		},
	}
	return dpv
}

func (r *ReconcileZdyfapi) DataNewPvse(m *zdyfv1alpha1.Zdyfapi) *corev1.PersistentVolume {

	dpv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Spec.DataPVName2, //datanode-pv-2
			Labels:      map[string]string{"app": "datanode"},
			Annotations: map[string]string{"volume.beta.kubernetes.io/storage-class": m.Spec.NameSCName},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse(m.Spec.DataPVStorage2),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName:              m.Spec.NameSCName,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				Local: &corev1.LocalVolumeSource{
					Path: "/var/lib/docker/volumes/datanode-2/_data",
				},
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"docker-desktop"},
						}},
					}},
				},
			},
		},
	}
	return dpv
}

func (r *ReconcileZdyfapi) DataNewPvth(m *zdyfv1alpha1.Zdyfapi) *corev1.PersistentVolume {

	dpv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Spec.DataPVName3, //datanode-pv-2
			Labels:      map[string]string{"app": "datanode"},
			Annotations: map[string]string{"volume.beta.kubernetes.io/storage-class": m.Spec.NameSCName},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse(m.Spec.DataPVStorage2),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName:              m.Spec.NameSCName,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				Local: &corev1.LocalVolumeSource{
					Path: "/var/lib/docker/volumes/datanode-2/_data",
				},
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"docker-desktop"},
						}},
					}},
				},
			},
		},
	}
	return dpv
}

func (r *ReconcileZdyfapi) DataNewPvfo(m *zdyfv1alpha1.Zdyfapi) *corev1.PersistentVolume {

	dpv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Spec.DataPVName4, //datanode-pv-2
			Labels:      map[string]string{"app": "datanode"},
			Annotations: map[string]string{"volume.beta.kubernetes.io/storage-class": m.Spec.NameSCName},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse(m.Spec.DataPVStorage2),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName:              m.Spec.NameSCName,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				Local: &corev1.LocalVolumeSource{
					Path: "/var/lib/docker/volumes/datanode-2/_data",
				},
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"docker-desktop"},
						}},
					}},
				},
			},
		},
	}
	return dpv
}

func (r *ReconcileZdyfapi) DataNewPvfi(m *zdyfv1alpha1.Zdyfapi) *corev1.PersistentVolume {

	dpv := &corev1.PersistentVolume{
		TypeMeta: metav1.TypeMeta{
			Kind:       "PersistentVolume",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        m.Spec.DataPVName5, //datanode-pv-2
			Labels:      map[string]string{"app": "datanode"},
			Annotations: map[string]string{"volume.beta.kubernetes.io/storage-class": m.Spec.NameSCName},
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(m, schema.GroupVersionKind{
					Group:   zdyfv1alpha1.SchemeGroupVersion.Group,
					Version: zdyfv1alpha1.SchemeGroupVersion.Version,
					Kind:    "Zdyfapi",
				}),
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: corev1.ResourceList{
				"storage": resource.MustParse(m.Spec.DataPVStorage2),
			},
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			StorageClassName:              m.Spec.NameSCName,
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimDelete,
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				Local: &corev1.LocalVolumeSource{
					Path: "/var/lib/docker/volumes/datanode-2/_data",
				},
			},
			NodeAffinity: &corev1.VolumeNodeAffinity{
				Required: &corev1.NodeSelector{
					NodeSelectorTerms: []corev1.NodeSelectorTerm{{
						MatchExpressions: []corev1.NodeSelectorRequirement{{
							Key:      "kubernetes.io/hostname",
							Operator: corev1.NodeSelectorOpIn,
							Values:   []string{"docker-desktop"},
						}},
					}},
				},
			},
		},
	}
	return dpv
}
