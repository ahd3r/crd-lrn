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
	metav1.TypeMeta   `json:",inline"` // consist of kind and apiVersion
	metav1.ObjectMeta `json:"metadata,omitempty"` // consist of all possible fields in metadata
	Spec              struct {
		Text string `json:"text"`
	} `json:"spec"`
}

func (msg *Message) DeepCopyObject() runtime.Object { // required method in order to add to known types (scheme.AddKnownTypes)
	out := new(Message)
	*out = *msg
	return out
}

type MessageList struct { // extra struct in addition to the above one in order to add new type of resource to cluster
    metav1.TypeMeta `json:",inline"`
    metav1.ListMeta `json:"metadata,omitempty"`
    Items           []Message `json:"items"`
}

func (msg *MessageList) DeepCopyObject() runtime.Object {
    out := new(MessageList)
    *out = *msg
    return out
}

type MessageReconciler struct { // reconcile initialization
	client.Client
}

var count = 0
func (r *MessageReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) { // the exact fuction gets to call on resource change
	logger := l.FromContext(ctx) // build in logger can not be used
	logger.Info("Reconcile is triggered")
	var msg Message
    if err := r.Get(ctx, req.NamespacedName, &msg); err != nil { // get resource that triggers reconcile
		if errors.IsNotFound(err) { // avoid error on deletion, since delete also triggers reconcile and no resource is found on execution
			return ctrl.Result{}, nil
		}
        return ctrl.Result{}, err // if reconcile returns error, kubernetes tries to call it again 4 more times (5 times total)
    }
	count++
	logger.Info("Params", "count", count, "message", msg.Spec.Text)
	if !msg.ObjectMeta.DeletionTimestamp.IsZero() { // deleteTimestamp gets propogated on `kubectl delete ...` run, if finalizers set to be non empty array, just to be able to get resource for the last time in reconcile and then delete it
		msg.SetFinalizers([]string{}) // clearing finalizers
		if err := r.Update(ctx, &msg); err != nil { // saving updated value, which clears finalizers, which leads to immidiate and complete resource deletion. Triggers reconcile one more time, but falls into NotFound error and returns None
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

func (r *MessageReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&Message{}). // kind type that will be watched
		Complete(r)
}

func main() {
	scheme := runtime.NewScheme()
	gvk := schema.GroupVersion{Group: "init.com", Version: "v100"} // new crd group initialization
	scheme.AddKnownTypes(gvk, &Message{}, &MessageList{}) // new kinds initialization
	metav1.AddToGroupVersion(scheme, gvk)

	ctrl.SetLogger(zap.New(zap.UseDevMode(true))) // another logger can not be used
	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{ // new watcher initialization
		Scheme: scheme,
	})
	if err != nil {
		log.Fatalf("unable to start manager: %v\n", err)
		os.Exit(1)
	}

	reconciler := &MessageReconciler{Client: mgr.GetClient()}
	if err := reconciler.SetupWithManager(mgr); err != nil { // register created reconcile into new watcher
		log.Fatalf("unable to set up controller: %v\n", err)
		os.Exit(1)
	}

	log.Println("Controller starting...")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil { // start watcher
		log.Fatalf("problem running manager: %v\n", err)
		os.Exit(1)
	}
}
