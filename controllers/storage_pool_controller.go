package controllers

import (
	"context"

	"k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type StoragePoolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *StoragePoolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Fetch the Deployment instance
	logger := log.FromContext(ctx)
	var deployment v1.Deployment
	var err error
	if err := r.Get(ctx, req.NamespacedName, &deployment); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var originalREPLICAS = *deployment.Spec.Replicas
	defer func() {
		if scaleError := ScaleDeployment(ctx, r, &deployment, originalREPLICAS); scaleError != nil {
			logger.Error(scaleError, "Failed to scale deployment back to original replicas: ", originalREPLICAS, "Deployment: ", deployment.Name)
			err = scaleError
		}
	}()

	// Check if the deployment has no required annotations
	if deployment.Annotations == nil || deployment.Annotations["storagepool.codeamenity.in/enabled"] != "true" {
		return ctrl.Result{}, nil
	}

	// Check if the deployment has no volume claims
	if deployment.Spec.Template.Spec.Volumes != nil {
		return ctrl.Result{}, nil
	}

	// Scale down the deployment to 0 replicas
	if err := ScaleDeployment(ctx, r, &deployment, 0); err != nil {
		return ctrl.Result{}, err
	}

	// Create the VolumeSnapshotClass if it does not exist
	if err := CreateIfNotVolumeSnapshotClass(ctx, r); err != nil {
		return ctrl.Result{}, err
	}

	// Create the StorageClass if it does not exist
	if err := CheckCreateStorageClass(ctx, r); err != nil {
		return ctrl.Result{}, err
	}

	// Create volume snapshots for each volume claim
	for _, volume := range deployment.Spec.Template.Spec.Volumes {
		var volumeClaimSpec coreV1.PersistentVolumeClaimSpec
		// Create a snapshot for the volume
		if volume.PersistentVolumeClaim == nil {
			continue
		}
		if err := CreateVolumeSnapshot(ctx, r, volume.PersistentVolumeClaim.ClaimName, deployment.Namespace); err != nil {
			return ctrl.Result{}, err
		}

		volumeClaimSpec, err := GetVolumeClaimSpec(ctx, r, volume.PersistentVolumeClaim.ClaimName, deployment.Namespace)
		if err != nil {
			return ctrl.Result{}, err
		}

		if err := DeleteOrRenameVolumeClaims(ctx, r, volume.PersistentVolumeClaim.ClaimName, deployment.Namespace); err != nil {
			return ctrl.Result{}, err
		}

		if err := CreateNewVolumeClaims(ctx, r, volume.PersistentVolumeClaim.ClaimName, deployment.Namespace, volumeClaimSpec); err != nil {
			return ctrl.Result{}, err
		}
	}

	// Update the deployment with the new tolerations
	if err := UpdateDeploymentWithTolerations(ctx, r, &deployment); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, err
}

func (r *StoragePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&v1.Deployment{}).
		Complete(r)
}
