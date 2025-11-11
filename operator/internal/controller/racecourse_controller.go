/*
Copyright 2025.

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

package controller

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	racecoursev1alpha1 "github.com/mgoode/racecourse-operator/api/v1alpha1"
)

// RacecourseReconciler reconciles a Racecourse object
type RacecourseReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=racecourse.kaleido.io,resources=racecourses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=racecourse.kaleido.io,resources=racecourses/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=racecourse.kaleido.io,resources=racecourses/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=configmaps,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=networking.k8s.io,resources=ingresses,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

func (r *RacecourseReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	racecourse := &racecoursev1alpha1.Racecourse{}
	if err := r.Get(ctx, req.NamespacedName, racecourse); err != nil {
		if errors.IsNotFound(err) {

			log.Info("Racecourse resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		log.Error(err, "Failed to get Racecourse")
		return ctrl.Result{}, err
	}

	log.Info("Reconciling Racecourse", "name", racecourse.Name, "namespace", racecourse.Namespace)

	if err := r.reconcileConfigMap(ctx, racecourse); err != nil {
		log.Error(err, "Failed to reconcile ConfigMap")
		return ctrl.Result{}, err
	}

	if err := r.reconcileService(ctx, racecourse); err != nil {
		log.Error(err, "Failed to reconcile Service")
		return ctrl.Result{}, err
	}

	if err := r.reconcileDeployment(ctx, racecourse); err != nil {
		log.Error(err, "Failed to reconcile Deployment")
		return ctrl.Result{}, err
	}

	if racecourse.Spec.Ingress.Enabled {
		if err := r.reconcileIngress(ctx, racecourse); err != nil {
			log.Error(err, "Failed to reconcile Ingress")
			return ctrl.Result{}, err
		}
	}

	if err := r.updateStatus(ctx, racecourse); err != nil {
		log.Error(err, "Failed to update status")
		return ctrl.Result{}, err
	}

	log.Info("Successfully reconciled Racecourse")
	return ctrl.Result{}, nil
}

func (r *RacecourseReconciler) reconcileConfigMap(ctx context.Context, racecourse *racecoursev1alpha1.Racecourse) error {
	log := log.FromContext(ctx)

	walletNamespace := racecourse.Spec.WalletService.Namespace
	if walletNamespace == "" {
		walletNamespace = racecourse.Namespace
	}
	walletPort := racecourse.Spec.WalletService.Port
	if walletPort == 0 {
		walletPort = 8545
	}

	walletURL := fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		racecourse.Spec.WalletService.Name,
		walletNamespace,
		walletPort,
	)

	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      racecourse.Name + "-config",
			Namespace: racecourse.Namespace,
		},
		Data: map[string]string{
			"signer-url":       walletURL,
			"contract-address": racecourse.Spec.ContractAddress,
		},
	}

	if err := controllerutil.SetControllerReference(racecourse, configMap, r.Scheme); err != nil {
		return err
	}

	found := &corev1.ConfigMap{}
	err := r.Get(ctx, types.NamespacedName{Name: configMap.Name, Namespace: configMap.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating ConfigMap", "name", configMap.Name)
		return r.Create(ctx, configMap)
	} else if err != nil {
		return err
	}

	if found.Data["signer-url"] != configMap.Data["signer-url"] ||
		found.Data["contract-address"] != configMap.Data["contract-address"] {
		log.Info("Updating ConfigMap", "name", configMap.Name)
		found.Data = configMap.Data
		return r.Update(ctx, found)
	}

	return nil
}

func (r *RacecourseReconciler) reconcileService(ctx context.Context, racecourse *racecoursev1alpha1.Racecourse) error {
	log := log.FromContext(ctx)

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      racecourse.Name,
			Namespace: racecourse.Namespace,
			Labels:    labelsForRacecourse(racecourse.Name),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeClusterIP,
			Selector: labelsForRacecourse(racecourse.Name),
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       3000,
					TargetPort: intstr.FromInt(3000),
				},
			},
		},
	}

	if err := controllerutil.SetControllerReference(racecourse, service, r.Scheme); err != nil {
		return err
	}

	found := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: service.Name, Namespace: service.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Service", "name", service.Name)
		return r.Create(ctx, service)
	} else if err != nil {
		return err
	}

	found.Labels = service.Labels
	found.Spec.Selector = service.Spec.Selector
	found.Spec.Ports = service.Spec.Ports
	log.Info("Updating Service", "name", service.Name)
	return r.Update(ctx, found)
}

func (r *RacecourseReconciler) reconcileDeployment(ctx context.Context, racecourse *racecoursev1alpha1.Racecourse) error {
	log := log.FromContext(ctx)

	deployment := r.buildDeployment(racecourse)

	if err := controllerutil.SetControllerReference(racecourse, deployment, r.Scheme); err != nil {
		return err
	}

	found := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: deployment.Name, Namespace: deployment.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Deployment", "name", deployment.Name)
		return r.Create(ctx, deployment)
	} else if err != nil {
		return err
	}

	found.Spec = deployment.Spec
	log.Info("Updating Deployment", "name", deployment.Name)
	return r.Update(ctx, found)
}

func (r *RacecourseReconciler) reconcileIngress(ctx context.Context, racecourse *racecoursev1alpha1.Racecourse) error {
	log := log.FromContext(ctx)

	ingress := r.buildIngress(racecourse)

	if err := controllerutil.SetControllerReference(racecourse, ingress, r.Scheme); err != nil {
		return err
	}

	found := &networkingv1.Ingress{}
	err := r.Get(ctx, types.NamespacedName{Name: ingress.Name, Namespace: ingress.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Ingress", "name", ingress.Name)
		return r.Create(ctx, ingress)
	} else if err != nil {
		return err
	}

	found.Annotations = ingress.Annotations
	found.Spec = ingress.Spec
	log.Info("Updating Ingress", "name", ingress.Name)
	return r.Update(ctx, found)
}

func (r *RacecourseReconciler) updateStatus(ctx context.Context, racecourse *racecoursev1alpha1.Racecourse) error {
	log := log.FromContext(ctx)

	deployment := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: racecourse.Name, Namespace: racecourse.Namespace}, deployment)
	if err != nil {
		if errors.IsNotFound(err) {
			racecourse.Status.Phase = racecoursev1alpha1.RacecoursePhasePending
			racecourse.Status.DeploymentReady = false
			racecourse.Status.AvailableReplicas = 0
		} else {
			return err
		}
	} else {
		racecourse.Status.AvailableReplicas = deployment.Status.AvailableReplicas

		if deployment.Status.AvailableReplicas > 0 {
			racecourse.Status.Phase = racecoursev1alpha1.RacecoursePhaseRunning
			racecourse.Status.DeploymentReady = true
		} else {
			racecourse.Status.Phase = racecoursev1alpha1.RacecoursePhasePending
			racecourse.Status.DeploymentReady = false
		}
	}

	walletNamespace := racecourse.Spec.WalletService.Namespace
	if walletNamespace == "" {
		walletNamespace = racecourse.Namespace
	}
	walletPort := racecourse.Spec.WalletService.Port
	if walletPort == 0 {
		walletPort = 8545
	}
	racecourse.Status.WalletServiceEndpoint = fmt.Sprintf("http://%s.%s.svc.cluster.local:%d",
		racecourse.Spec.WalletService.Name,
		walletNamespace,
		walletPort,
	)

	if racecourse.Spec.Ingress.Enabled && racecourse.Spec.Ingress.Host != "" {
		racecourse.Status.URL = fmt.Sprintf("http://%s", racecourse.Spec.Ingress.Host)
	}

	log.Info("Updating status", "phase", racecourse.Status.Phase, "replicas", racecourse.Status.AvailableReplicas)
	return r.Status().Update(ctx, racecourse)
}

func (r *RacecourseReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&racecoursev1alpha1.Racecourse{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&networkingv1.Ingress{}).
		Complete(r)
}

// Get labels for selecting Racecourse resources
func labelsForRacecourse(name string) map[string]string {
	return map[string]string{
		"app":        "racecourse",
		"racecourse": name,
	}
}
