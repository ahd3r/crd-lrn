package main

import (
	"context"
	"log"
	"os"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	l "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Message struct { // type name should be same as Kind name of CRD
	metav1.TypeMeta   `json:",inline"` // consist of kind and apiVersion, inline
	metav1.ObjectMeta `json:"metadata,omitempty"` // omitempty
	Spec              struct {
		Text string `json:"text"`
	} `json:"spec"`
}

func (msg *Message) DeepCopyObject() runtime.Object {
	out := new(Message)
	*out = *msg
	return out
}

type MessageList struct {
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []Message `json:"items"`
}

func (msg *MessageList) DeepCopyObject() runtime.Object {
    out := new(MessageList)
    *out = *msg
    return out
}

type MessageReconciler struct {
	client.Client
}

var count = 0
func (r *MessageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := l.FromContext(ctx)
	logger.Info("Reconcile is triggered")
	var msg Message
    if err := r.Get(ctx, req.NamespacedName, &msg); err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
        return ctrl.Result{}, err
    }
	count++
	logger.Info("Params", "count", count, "message", msg.Spec.Text)
	if !msg.ObjectMeta.DeletionTimestamp.IsZero() {
		msg.SetFinalizers([]string{})
		if err := r.Update(ctx, &msg); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *MessageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&Message{}).
		Complete(r)
}

func main() {
	scheme := runtime.NewScheme()
	gvk := schema.GroupVersion{Group: "init.com", Version: "v100"}
	scheme.AddKnownTypes(gvk, &Message{}, &MessageList{})
	metav1.AddToGroupVersion(scheme, gvk)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true)))
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme: scheme,
	})
	if err != nil {
		log.Fatalf("unable to start manager: %v\n", err)
		os.Exit(1)
	}

	reconciler := &MessageReconciler{Client: mgr.GetClient()}
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
