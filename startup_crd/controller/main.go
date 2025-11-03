package main

import (
	"context"
	"log"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	l "sigs.k8s.io/controller-runtime/pkg/log"
)

type NginxStart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              struct {
		ExternalLBPort int `json:"externalLBPort"`
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

type NginxReconciler struct {
	client.Client
}

func (r *NginxReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := l.FromContext(ctx)
	logger.Info("Reconcile is triggered")
	return ctrl.Result{}, nil
}

func (r *NginxReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&NginxStart{}).
		Complete(r)
}

func main() {
	scheme := runtime.NewScheme()
	gvk := schema.GroupVersion{Group: "init.com", Version: "v100"}
	scheme.AddKnownTypes(gvk, &NginxStart{}, &NginxStartList{})
	metav1.AddToGroupVersion(scheme, gvk)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatalf("unable to start manager: %v\n", err)
		os.Exit(1)
	}

	reconciler := &NginxReconciler{Client: mgr.GetClient()}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to set up controller: %v\n", err)
		os.Exit(1)
	}

	log.Println("✅ Controller starting...")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("problem running manager: %v\n", err)
		os.Exit(1)
	}
}
