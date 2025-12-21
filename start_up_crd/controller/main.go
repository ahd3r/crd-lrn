package main

import (
	"context"
	"fmt"
	"log"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	l "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type NginxStart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Status            *struct {
		ObservedGeneration *int64 `json:"observedGeneration,omitempty"`
	} `json:"status,omitempty"`
	Spec struct {
		NodePort int32 `json:"nodePort"`
	} `json:"spec"`
}

func (ns *NginxStart) DeepCopyObject() runtime.Object {
	out := new(NginxStart)
	*out = *ns
	return out
}

type NginxStartList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxStart `json:"items"`
}

func (ns *NginxStartList) DeepCopyObject() runtime.Object {
	out := new(NginxStartList)
	*out = *ns
	return out
}

type NginxStartReconciler struct {
	client.Client
}

// Why can't I get nginxstart with kubectl get all -n ns-namespace or kubectl get all --all-namespace? Because kubectl get all doesn't return CR or CRD, it only returns: pods, services, deployments, replicasets, statefulsets, daemonsets, jobs, cronjobs
// what would happen if unchanged version of CRD gets applied? Would it trigger a reconcile? NO, it will be just ignored
// TODO: on child remove - remove parent CR
func (r *NginxStartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := l.FromContext(ctx)
	logger.Info("Reconcile is triggered")
	var ns NginxStart
	if err := r.Get(ctx, req.NamespacedName, &ns); err != nil { // TODO: request the right kind with similar req.NamespacedName, if reconcile is triggered by a child0
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if !ns.ObjectMeta.DeletionTimestamp.IsZero() { // resources will be deleted separatlly
		logger.Info("Remove Children")
		var dep appsv1.Deployment
		depExist := true
		logger.Info("Looking for a Deployment")
		if err := r.Get(ctx, types.NamespacedName{Name: "nginx-deployment-from-crd-cc-" + req.NamespacedName.Name, Namespace: req.NamespacedName.Namespace}, &dep); err != nil {
			depExist = false
			if !errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			logger.Info("Couldn't find a Deployment")
		}
		if depExist {
			logger.Info("Found a Deployment to remove")
			logger.Info("Validating ownership")
			depExist = false
			for _, owner := range dep.GetOwnerReferences() {
				logger.Info("Found owner with next properties", "Name", owner.Name, "Kind", owner.Kind)
				if owner.Kind == "NginxStart" {
					logger.Info("Ownership approved")
					depExist = true
				}
			}
			if depExist {
				logger.Info("Removing Finalizers")
				dep.SetFinalizers([]string{})
				if err := r.Update(ctx, &dep); err != nil { // Would it trigger reconcile? YES, but it will finish this current execution and on the next on it falls under NotFound error that is ignored
					return ctrl.Result{}, err
				}
				logger.Info("Deleting resource")
				if err := r.Delete(ctx, &dep); err != nil { // This will not trigger reconcile, as well as next ones
					return ctrl.Result{}, err
				}
			} else {
				logger.Info("Ownership failed")
			}
		}
		var ser corev1.Service
		serExist := true
		logger.Info("Looking for a Service")
		if err := r.Get(ctx, types.NamespacedName{Name: "my-nginx-service-from-crd-cc-" + req.NamespacedName.Name, Namespace: req.NamespacedName.Namespace}, &ser); err != nil {
			serExist = false
			if !errors.IsNotFound(err) {
				return ctrl.Result{}, err
			}
			logger.Info("Couldn't find a Service")
		}
		if serExist {
			logger.Info("Found a Service to remove")
			logger.Info("Validating ownership")
			serExist = false
			for _, owner := range ser.GetOwnerReferences() {
				logger.Info("Found owner with next properties", "Name", owner.Name, "Kind", owner.Kind)
				if owner.Kind == "NginxStart" {
					serExist = true
					logger.Info("Ownership approved")
				}
			}
			if serExist {
				logger.Info("Removing Finalizers")
				ser.SetFinalizers([]string{})
				if err := r.Update(ctx, &ser); err != nil {
					return ctrl.Result{}, err
				}
				logger.Info("Deleting resource")
				if err := r.Delete(ctx, &ser); err != nil {
					return ctrl.Result{}, err
				}
			} else {
				logger.Info("Ownership failed")
			}
		}
		logger.Info("Remove CR Finalizers")
		ns.SetFinalizers([]string{})
		if err := r.Update(ctx, &ns); err != nil {
			return ctrl.Result{}, err
		}
		logger.Info("Auto Removal of CR")
		return ctrl.Result{}, nil
	}
	logger.Info(fmt.Sprintf("NodePort is %d", ns.Spec.NodePort))

	// TODO: keep child resources in sync with crd val (if updated -> update)
	// TODO: if child resource gets updated, then sync cr
	var dep appsv1.Deployment
	depExist := true
	if err := r.Get(ctx, types.NamespacedName{Name: "nginx-deployment-from-crd-cc-" + req.NamespacedName.Name, Namespace: req.NamespacedName.Namespace}, &dep); err != nil {
		depExist = false
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}
	var ser corev1.Service
	serExist := true
	if err := r.Get(ctx, types.NamespacedName{Name: "my-nginx-service-from-crd-cc-" + req.NamespacedName.Name, Namespace: req.NamespacedName.Namespace}, &ser); err != nil {
		serExist = false
		if !errors.IsNotFound(err) {
			return ctrl.Result{}, err
		}
	}
	if !depExist && !serExist {
		logger.Info("Generate Resources")
		if err := r.generateRecourse(ctx, ns.Spec.NodePort, ns.Namespace, ns.Name); err != nil {
			return ctrl.Result{}, err
		}
	} else {
		depExist = false
		for _, owner := range dep.GetOwnerReferences() {
			if owner.Kind == "NginxStart" {
				depExist = true
			}
		}
		serExist = false
		for _, owner := range ser.GetOwnerReferences() {
			if owner.Kind == "NginxStart" {
				serExist = true
			}
		}
		if depExist && serExist {
			logger.Info("Synchronize Resources")
		} else {
			logger.Info("Resources name conflict")
			return ctrl.Result{}, errors.NewBadRequest("Resources name conflict") // Maybe buggy, in the way that if CR is created, and conflicted resource gets deleted, it takes some time (up to 5 mins) for reconcile to re-run
		}
	}
	return ctrl.Result{}, nil
}

func (r *NginxStartReconciler) generateRecourse(ctx context.Context, nodePort int32, namespace string, name string) error {
	replicas := int32(1)
	deploymentResource := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "nginx-deployment-from-crd-cc-" + name,
			Namespace: namespace,
			Labels: map[string]string{
				"manged": "cc",
			},
			Finalizers: []string {"true.test/finalizer"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"app": "nginx"},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"app": "nginx"},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "nginx",
							Image: "nginx:latest",
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: 80,
								},
							},
						},
					},
				},
			},
		},
	}
	if err := r.Create(ctx, &deploymentResource); err != nil {
		return err
	}
	serviceResource := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-nginx-service-from-crd-cc-" + name,
			Namespace: namespace,
			Labels: map[string]string{
				"manged": "cc",
			},
			Finalizers: []string{"true.test/finalizer"},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "nginx"},
			Type:     "NodePort",
			Ports: []corev1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       3200,
					TargetPort: intstr.FromInt(80),
					NodePort:   nodePort,
				},
			},
		},
	}
	if err := r.Create(ctx, &serviceResource); err != nil {
		return err
	}
	return nil
}

func (r *NginxStartReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&NginxStart{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}

func main() {
	scheme := runtime.NewScheme()
	gvk := schema.GroupVersion{Group: "init.com", Version: "v100"}
	scheme.AddKnownTypes(gvk, &NginxStart{}, &NginxStartList{})
	metav1.AddToGroupVersion(scheme, gvk)
	appsv1.AddToScheme(scheme)
	corev1.AddToScheme(scheme)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatalf("unable to start manager: %v\n", err)
		os.Exit(1)
	}

	reconciler := &NginxStartReconciler{Client: mgr.GetClient()}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to set up controller: %v\n", err)
		os.Exit(1)
	}

	log.Println("Controller starting...")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("problem running manager: %v\n", err)
		os.Exit(1)
	}
}
