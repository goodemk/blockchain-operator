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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	racecoursev1alpha1 "github.com/mgoode/racecourse-operator/api/v1alpha1"
)

// Creates a Deployment spec for Racecourse instances
func (r *RacecourseReconciler) buildDeployment(racecourse *racecoursev1alpha1.Racecourse) *appsv1.Deployment {
	replicas := int32(2)
	if racecourse.Spec.Replicas != nil {
		replicas = *racecourse.Spec.Replicas
	}

	repository := "racecourse"
	if racecourse.Spec.Image.Repository != "" {
		repository = racecourse.Spec.Image.Repository
	}

	tag := "0.0.1"
	if racecourse.Spec.Image.Tag != "" {
		tag = racecourse.Spec.Image.Tag
	}

	pullPolicy := corev1.PullIfNotPresent
	if racecourse.Spec.Image.PullPolicy != "" {
		pullPolicy = racecourse.Spec.Image.PullPolicy
	}

	labels := labelsForRacecourse(racecourse.Name)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      racecourse.Name,
			Namespace: racecourse.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:            "racecourse",
							Image:           repository + ":" + tag,
							ImagePullPolicy: pullPolicy,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									ContainerPort: 3000,
									Protocol:      corev1.ProtocolTCP,
								},
							},
							Env: []corev1.EnvVar{
								{
									Name:  "APPLICATION_PORT",
									Value: "3000",
								},
								{
									Name: "SIGNER_URL",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: racecourse.Name + "-config",
											},
											Key: "signer-url",
										},
									},
								},
								{
									Name: "CONTRACT_ADDRESS",
									ValueFrom: &corev1.EnvVarSource{
										ConfigMapKeyRef: &corev1.ConfigMapKeySelector{
											LocalObjectReference: corev1.LocalObjectReference{
												Name: racecourse.Name + "-config",
											},
											Key: "contract-address",
										},
									},
								},
							},
							LivenessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(3000),
									},
								},
								InitialDelaySeconds: 30,
								PeriodSeconds:       10,
								TimeoutSeconds:      5,
								FailureThreshold:    3,
							},
							ReadinessProbe: &corev1.Probe{
								ProbeHandler: corev1.ProbeHandler{
									HTTPGet: &corev1.HTTPGetAction{
										Path: "/",
										Port: intstr.FromInt(3000),
									},
								},
								InitialDelaySeconds: 10,
								PeriodSeconds:       5,
								TimeoutSeconds:      3,
								FailureThreshold:    3,
							},
							Resources: racecourse.Spec.Resources,
						},
					},
				},
			},
		},
	}

	return deployment
}

// Creates an Ingress spec for Racecourse instances
func (r *RacecourseReconciler) buildIngress(racecourse *racecoursev1alpha1.Racecourse) *networkingv1.Ingress {
	className := "nginx"
	if racecourse.Spec.Ingress.ClassName != "" {
		className = racecourse.Spec.Ingress.ClassName
	}

	host := "racecourse.local"
	if racecourse.Spec.Ingress.Host != "" {
		host = racecourse.Spec.Ingress.Host
	}

	// annotations for websocket support
	annotations := map[string]string{
		"nginx.ingress.kubernetes.io/websocket-services":     racecourse.Name,
		"nginx.ingress.kubernetes.io/proxy-read-timeout":     "3600",
		"nginx.ingress.kubernetes.io/proxy-send-timeout":     "3600",
		"nginx.ingress.kubernetes.io/affinity":               "cookie",
		"nginx.ingress.kubernetes.io/session-cookie-name":    "racecourse-session",
		"nginx.ingress.kubernetes.io/session-cookie-expires": "172800",
		"nginx.ingress.kubernetes.io/session-cookie-max-age": "172800",
	}

	// Merge in user provided annotations
	for key, value := range racecourse.Spec.Ingress.Annotations {
		annotations[key] = value
	}

	pathTypePrefix := networkingv1.PathTypePrefix

	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:        racecourse.Name,
			Namespace:   racecourse.Namespace,
			Labels:      labelsForRacecourse(racecourse.Name),
			Annotations: annotations,
		},
		Spec: networkingv1.IngressSpec{
			IngressClassName: &className,
			Rules: []networkingv1.IngressRule{
				{
					Host: host,
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Path:     racecourse.Spec.Ingress.Path,
									PathType: &pathTypePrefix,
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: racecourse.Name,
											Port: networkingv1.ServiceBackendPort{
												Number: 3000,
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return ingress
}
