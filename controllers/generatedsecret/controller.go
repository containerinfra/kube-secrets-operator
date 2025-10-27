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

package generatedsecret

import (
	"context"
	"fmt"
	"time"

	generatedsecretv1 "github.com/containerinfra/kube-secrets-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/client-go/util/retry"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	finalizerName = "generatedsecret.k8s.containerinfra.com/finalizer"
)

var (
	RequeueAfterErrorDuration = 10 * time.Second
)

// GeneratedSecretReconciler reconciles a GeneratedSecret object
type GeneratedSecretReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=apps.k8s.containerinfra.com,resources=generatedsecrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps.k8s.containerinfra.com,resources=generatedsecrets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=apps.k8s.containerinfra.com,resources=generatedsecrets/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups="",resources=events,verbs=create;patch

func (r *GeneratedSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var generatedSecret generatedsecretv1.GeneratedSecret
	if err := r.Get(ctx, req.NamespacedName, &generatedSecret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !generatedSecret.ObjectMeta.DeletionTimestamp.IsZero() {
		if controllerutil.ContainsFinalizer(&generatedSecret, finalizerName) {
			if err := r.cleanup(ctx, &generatedSecret); err != nil {
				r.Recorder.Event(&generatedSecret, corev1.EventTypeWarning, "CleanupFailed", fmt.Sprintf("failed to cleanup generated secrets: %s", err.Error()))
				return ctrl.Result{}, err
			}
			if controllerutil.RemoveFinalizer(&generatedSecret, finalizerName) {
				if err := r.Update(ctx, &generatedSecret); err != nil {
					return ctrl.Result{}, err
				}
			}
		}
		r.Recorder.Event(&generatedSecret, corev1.EventTypeNormal, "CleanupSucceeded", "Generated secrets have been cleaned up")
		return ctrl.Result{}, nil
	}

	// The object is not being deleted, so if it does not have our finalizer,
	// then lets add the finalizer and update the object. This is equivalent  to registering our finalizer.
	if controllerutil.AddFinalizer(&generatedSecret, finalizerName) {
		if err := r.Update(ctx, &generatedSecret); err != nil {
			return ctrl.Result{}, err
		}
	}

	err := r.reconcileGeneratedSecrets(ctx, generatedSecret)
	if err != nil {
		return ctrl.Result{RequeueAfter: 1 * time.Minute}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GeneratedSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&generatedsecretv1.GeneratedSecret{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return true
			},
			DeleteFunc: func(e event.DeleteEvent) bool {
				return false
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldGeneration := e.ObjectOld.GetGeneration()
				newGeneration := e.ObjectNew.GetGeneration()
				return oldGeneration != newGeneration
			},
			GenericFunc: func(e event.GenericEvent) bool {
				return false
			},
		}).
		Complete(r)
}

func (r *GeneratedSecretReconciler) updateStatusOrRetry(ctx context.Context, generatedSecret *generatedsecretv1.GeneratedSecret) error {
	return retry.OnError(retry.DefaultRetry, func(err error) bool {
		return true
	}, func() error {
		// Fetch the latest version of the GeneratedSecret object
		latest := &generatedsecretv1.GeneratedSecret{}
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: generatedSecret.Namespace, Name: generatedSecret.Name}, latest); err != nil {
			return fmt.Errorf("failed to fetch latest GeneratedSecret: %w", err)
		}

		// Update the status of the latest object
		latest.Status = generatedSecret.Status
		if err := r.Client.Status().Update(ctx, latest); err != nil {
			return fmt.Errorf("failed to update GeneratedSecret status: %w", err)
		}

		// Refresh the GeneratedSecret object to avoid conflicts
		if err := r.Client.Get(ctx, types.NamespacedName{Namespace: generatedSecret.Namespace, Name: generatedSecret.Name}, generatedSecret); err != nil {
			return fmt.Errorf("failed to re-fetch GeneratedSecret: %w", err)
		}
		return nil
	})
}
