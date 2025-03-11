package main

import (
	"context"
	"fmt"
	"time"

	"github.com/masa213f/test-statefulset-orphan-delete/pkg/constant"

	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type StatefulSetReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *StatefulSetReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrl.Log.WithName("reconcile").WithValues("name", req.NamespacedName.String())

	var orig appsv1.StatefulSet
	err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, &orig)
	if err != nil {
		if apierrors.IsNotFound(err) {
			log.Info("not found")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, fmt.Errorf("failed to get StatefulSet: %w", err)
	}
	if orig.DeletionTimestamp != nil {
		log.Info("deletionTimestamp exists")
		return ctrl.Result{}, nil
	}

	if orig.Spec.VolumeClaimTemplates[0].Labels == nil {
		log.Info("no need to re-create")
		return ctrl.Result{}, nil
	}
	log.Info("volumeClaimTemplates has changed, delete StatefulSet and try to re-create it")

	// Make a patch before re-creating.
	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(constant.GetAppsV1StatefulSetAC())
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to convert StatefulSet %s/%s to unstructured: %w", constant.Namespace, constant.StatefulSetName, err)
	}
	patch := &unstructured.Unstructured{Object: obj}

	err = r.Delete(ctx, &orig, &client.DeleteOptions{
		PropagationPolicy: ptr.To[metav1.DeletionPropagation](metav1.DeletePropagationOrphan),
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to delete StatefulSet: %w", err)
	}

	// Wait until the StatefulSet resource disappears.
	waitFunc := func(ctx context.Context) (bool, error) {
		err := r.Get(ctx, client.ObjectKey{Namespace: req.Namespace, Name: req.Name}, &appsv1.StatefulSet{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	}
	if err := wait.PollUntilContextTimeout(ctx, time.Millisecond*500, time.Second*5, true, waitFunc); err != nil {
		return ctrl.Result{}, fmt.Errorf("re-creation failed: %w", err)
	}

	err = r.Patch(ctx, patch, client.Apply, &client.PatchOptions{
		FieldManager: constant.FieldManagerController,
		Force:        ptr.To[bool](true),
	})
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to create StatefulSet %w", err)
	}

	log.Info("reconciled StatefulSet")

	return ctrl.Result{}, nil
}

func (r *StatefulSetReconciler) SetupWithManager(mgr ctrl.Manager) error {
	prctFunc := func(o client.Object) bool {
		return o.GetNamespace() == constant.Namespace && o.GetName() == constant.StatefulSetName
	}

	prct := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool { return prctFunc(e.ObjectNew) },
		CreateFunc: func(e event.CreateEvent) bool { return prctFunc(e.Object) },
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.StatefulSet{}, builder.WithPredicates(prct)).
		Complete(r)
}
