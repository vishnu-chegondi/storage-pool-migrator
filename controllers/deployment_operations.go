package controllers

import (
	"context"
	"time"

	v1 "k8s.io/api/apps/v1"
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
