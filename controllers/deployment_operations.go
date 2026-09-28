package controllers

import (
	"context"
	"time"

	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ScaleDeployment(ctx context.Context, r *StoragePoolReconciler, deployment *v1.Deployment, replicas int32) error {
	*deployment.Spec.Replicas = replicas
	if err := r.Update(ctx, deployment); err != nil {
		return err
	}

	for deployment.Status.ReadyReplicas != 0 {
		// Wait for the deployment to scale down to 0 replicas
		time.Sleep(10 * time.Second)
		if err := r.Get(ctx, client.ObjectKey{Name: deployment.Name, Namespace: deployment.Namespace}, deployment); err != nil {
			return err
		}
	}

	return nil
}

func GetNewTolerations(ctx context.Context, r *StoragePoolReconciler) ([]corev1.Toleration, error) {
	tolerationYAML, err := GetMigrationConfigMapDataValue(ctx, r, "tolerations.yaml")
	if err != nil {
		return []corev1.Toleration{}, err
	}

	var toleration []corev1.Toleration
	if err := yaml.Unmarshal([]byte(tolerationYAML), &toleration); err != nil {
		return []corev1.Toleration{}, err
	}
	return toleration, nil
}

func UpdateDeploymentWithTolerations(ctx context.Context, r *StoragePoolReconciler, deployment *v1.Deployment) error {
	migrationTOLERATIONS, err := GetNewTolerations(ctx, r)
	if err != nil {
		return err
	}
	oldTOLERATIONS := deployment.Spec.Template.Spec.Tolerations
	newTOLERATIONS := append(oldTOLERATIONS, migrationTOLERATIONS...)
	deployment.Spec.Template.Spec.Tolerations = newTOLERATIONS
	if err := r.Update(ctx, deployment); err != nil {
		return err
	}
	return nil
}
