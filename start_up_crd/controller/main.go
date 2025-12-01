package main

import (
	"context"
	"log"
	"os"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	l "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type NginxStart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              struct {
		NodePort int32 `json:"text"`
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

func (r *NginxStartReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := l.FromContext(ctx)
	logger.Info("Reconcile is triggered")
	var ns NginxStart
	logger.Info(req.NamespacedName.String())
    if err := r.Get(ctx, req.NamespacedName, &ns); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
        return ctrl.Result{}, err
    }
	if !ns.ObjectMeta.DeletionTimestamp.IsZero() {
		ns.SetFinalizers([]string{})
		if err := r.Update(ctx, &ns); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	if err := r.generateRecourse(ctx, ns.Spec.NodePort); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *NginxStartReconciler) generateRecourse(ctx context.Context, nodePort int32) error {
	replicas := int32(1)
	deploymentResource := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
            Name:      "nginx-deployment-from-crd-cc",
            Namespace: "default",
			Labels: map[string]string{
				"manged": "cc",
			},
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
							Name: "nginx",
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
		return err;
	}
	serviceResource := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-nginx-service-from-crd-cc",
			Namespace: "default",
			Labels: map[string]string{
				"manged": "cc",
			},
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{"app": "nginx"},
			Type: "NodePort",
			Ports: []corev1.ServicePort{
				{
					Protocol: "TCP",
					Port: 3200,
					TargetPort: intstr.FromInt(80),
					NodePort: nodePort,
				},
			},
		},
	}
	if err := r.Create(ctx, &serviceResource); err != nil {
		return err;
	}
	return nil;
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
