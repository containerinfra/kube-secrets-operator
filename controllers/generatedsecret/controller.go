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
	"time"

	generatedsecretv1 "github.com/containerinfra/kube-secrets-operator/api/v1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
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

func (r *GeneratedSecretReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var generatedSecret generatedsecretv1.GeneratedSecret
	if err := r.Get(ctx, req.NamespacedName, &generatedSecret); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !generatedSecret.Status.Initalized {
		logger.V(1).Info("initalizing generated secret", "name", generatedSecret.GetName())
		err := r.validateSpec(generatedSecret)
		if err != nil {
			r.Recorder.Event(&generatedSecret, corev1.EventTypeWarning, "Initalize failed", err.Error())

			return ctrl.Result{
				RequeueAfter: 1 * time.Minute,
			}, err
		}

		err = r.initalizeGeneratedSecret(ctx, generatedSecret)
		if err != nil {
			r.Recorder.Event(&generatedSecret, corev1.EventTypeWarning, "Initalize failed", err.Error())
			return ctrl.Result{
				RequeueAfter: 1 * time.Minute,
			}, err
		}
		logger.Info("Initalized, secrets have been created....")
		r.Recorder.Event(&generatedSecret, corev1.EventTypeNormal, "Initalized", "Secrets have been initalized")
		return ctrl.Result{}, nil
	}

	logger.Info("Resource has already been initalizing, checking existing status")
	err := r.reconcileExistingSecret(ctx, generatedSecret)
	if err != nil {
		return ctrl.Result{
			RequeueAfter: 1 * time.Minute,
		}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *GeneratedSecretReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&generatedsecretv1.GeneratedSecret{}).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(event event.CreateEvent) bool {
				return true
			},
			DeleteFunc: func(event event.DeleteEvent) bool {
				return true
			},
			UpdateFunc: func(event event.UpdateEvent) bool {
				return true
			},
			GenericFunc: func(event event.GenericEvent) bool {
				return false
			},
		}).
		Complete(r)
}
