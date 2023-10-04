/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	//"os"
	//"os/user"
	//"path/filepath"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/yaml"

	fivegv1alpha1 "github.com/Lim-Ho-Keun/udm-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

// UDMReconciler reconciles a UDM object
type UDMReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=fiveg.kt.com,resources=udms,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=fiveg.kt.com,resources=udms/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=fiveg.kt.com,resources=udms/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the UDM object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *UDMReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	log.Info("")

	instance := &fivegv1alpha1.UDM{}
	err := r.Get(context.TODO(), req.NamespacedName, instance)

	fmt.Println("start111111111111111111111111111111111")
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	fmt.Println("start22222222222222222222222222222222")
	err = r.ensureLatestCommonConfigMap(instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	fmt.Println("start33333333333333333333333333333333")
	err = r.ensureLatestStatefulset(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.ensureLatestService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UDMReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&fivegv1alpha1.UDM{}).
		Complete(r)
}

func (r *UDMReconciler) ensureLatestCommonConfigMap(instance *fivegv1alpha1.UDM) error {
	configMap := newCommonConfigMap(instance)

	if err := controllerutil.SetControllerReference(instance, configMap, r.Scheme); err != nil {
		return err
	}

	foundMap := &corev1.ConfigMap{}
	fmt.Println("11111111111111111")
	//fmt.Println(configMap)
	err := r.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, foundMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.TODO(), configMap)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (r *UDMReconciler) ensureLatestStatefulset(instance *fivegv1alpha1.UDM) error {
	statefulSet := newStatefulSet(instance)

	if err := controllerutil.SetControllerReference(instance, statefulSet, r.Scheme); err != nil {
		return err
	}

	foundsts := &appsv1.StatefulSet{}
	fmt.Println("22222222" + statefulSet.Namespace)
	err := r.Get(context.TODO(), types.NamespacedName{Name: statefulSet.Name, Namespace: statefulSet.Namespace}, foundsts)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.TODO(), statefulSet)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func (r *UDMReconciler) ensureLatestService(instance *fivegv1alpha1.UDM) error {
	service := newService(instance)

	if err := controllerutil.SetControllerReference(instance, service, r.Scheme); err != nil {
		return err
	}

	foundService := &corev1.Service{}
	fmt.Println("33333333" + service.Namespace)
	err := r.Get(context.TODO(), types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, foundService)
	if err != nil && errors.IsNotFound(err) {
		err = r.Create(context.TODO(), service)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	return nil
}

func newCommonConfigMap(cr *fivegv1alpha1.UDM) *corev1.ConfigMap {
	var err error
	var bs []byte

	fmt.Println("ddddddddddddddddddddddddddddddddddddd")
	/*
		currentUser, err := user.Current()
		if err != nil {
		}

		name := currentUser.Name
		id := currentUser.Uid

		fmt.Printf("name is: %s and id is: %s\n", name, id)
	*/
	/*
		dirpath, _ := os.Getwd()
		fmt.Println("aaaaaaaaaaaaaaaaaaa")
		fmt.Println(dirpath)

		fmt.Println("bbbbbbbbbbbbbbbbbbb")

		files, err := ioutil.ReadDir(dirpath)
		if err != nil {
		}

		for _, t := range files {
			fmt.Println(t.Name())
			if err != nil {
			}
		}

		cmfilepath := filepath.Join(dirpath, "udm-common-configmap.yaml")
		fmt.Println(cmfilepath)
	*/
	/*
		fmt.Println("ccccccccccccccccccc")

		if err := os.Mkdir("udm", os.ModePerm); err != nil {
			//log.Fatal(err)
			fmt.Println("make directory")
		}

		f, err := os.Create(cmfilepath)

		defer f.Close()

		_, err2 := f.WriteString("apiVersion: v1 kind: ConfigMap metadata: name: commondb-config namespace: 5gc-udm data: odbc.ini: | [Mariadb_Udmconf_master] Driver=MariaDB ODBC 3.0 Driver DATABASE=UDM_CFG DESCRIPTION=MariaDB via ODBC SERVER= commondb-sts-0.commondb.5gc-commondb-sts.svc.cluster.local UID=udm PASSWORD=udm PORT=3306 [Mariadb_Udmconf_slave] Driver=MariaDB ODBC 3.0 Driver DATABASE=UDM_CFG DESCRIPTION=MariaDB via ODBC SERVER= commondb-sts-1.commondb.5gc-commondb-sts.svc.cluster.local UID=udm PASSWORD=udm PORT=3306")

		if err2 != nil {
		}
	*/
	/*
		fmt.Println("ccccccccccccccccccc")
		files, err = ioutil.ReadDir(dirpath)
		if err != nil {
		}

		for _, t := range files {
			fmt.Println(t.Name())
			if err != nil {
			}
		}
	*/
	{
		//bs, err = ioutil.ReadFile("/root/5gcore/udm/udm-common-configmap.yaml")
		bs, err = ioutil.ReadFile("/udm-common-configmap.yaml")
	}
	var configmap corev1.ConfigMap
	err = yaml.Unmarshal(bs, &configmap)
	//fmt.Println("testestestse")
	//fmt.Println(configmap)
	if err != nil {
		// handle err
	}
	return &configmap
}

func newStatefulSet(cr *fivegv1alpha1.UDM) *appsv1.StatefulSet {
	var err error
	var bs []byte
	{
		bs, err = ioutil.ReadFile("/udm-statefulset.yaml")
		//bs, err = ioutil.ReadFile("../udm/udm-statefulset.yaml")
	}
	var statefulSet appsv1.StatefulSet
	err = yaml.Unmarshal(bs, &statefulSet)

	if err != nil {
		// handle err
	}
	return &statefulSet
}

func newService(cr *fivegv1alpha1.UDM) *corev1.Service {
	var err error
	var bs []byte
	{
		bs, err = ioutil.ReadFile("/udm-service.yaml")
		//bs, err = ioutil.ReadFile("../udm/udm-service.yaml")
	}
	var service corev1.Service
	err = yaml.Unmarshal(bs, &service)

	if err != nil {
		// handle err
	}
	return &service
}
