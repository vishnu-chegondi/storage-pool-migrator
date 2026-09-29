package controllers

import (
	"context"

	"k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type StoragePoolReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func ValidateMigrationRequired(deployment *v1.Deployment) bool {
	// Check if the deployment has no required annotations
	if deployment.Annotations == nil || deployment.Annotations["storagepool.codeamenity.in/enabled"] != "true" {
		return false
	}

	// Check if the deployment has no volume claims
	if deployment.Spec.Template.Spec.Volumes != nil {
		return false
	}

	// Check if the deployment is already migrated to use new node pools and storage classes
	if deployment.Status.Conditions != nil {
		for _, condition := range deployment.Status.Conditions {
			if condition.Type == "StoragePoolMigration" && condition.Reason == "MigrationCompleted" && condition.Status == corev1.ConditionTrue {
				return false
			}
		}
	}

	return true
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

	// Validate if the deployment requires migration
	if migrationRequired := ValidateMigrationRequired(&deployment); migrationRequired == false {
		return ctrl.Result{}, nil
	}

	// Scale down the deployment to 0 replicas
	if err := ScaleDeployment(ctx, r, &deployment, 0); err != nil {
		if statusErr := UpdateFailedCondition(ctx, "DeploymentScaleDown", "Failed to scale down deployment to 0 replicas"); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}
	if statusErr := UpdateSucceededCondition(ctx, "DeploymentScaleDown", "Deployment scaled down to 0 replicas"); statusErr != nil {
		return ctrl.Result{}, statusErr
	}

	// Create the VolumeSnapshotClass if it does not exist
	if err := CreateIfNotVolumeSnapshotClass(ctx, r); err != nil {
		if statusErr := UpdateFailedCondition(ctx, "CreateVolumeSnapshotClass", "Failed to create VolumeSnapshotClass"); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}
	if statusErr := UpdateSucceededCondition(ctx, "CreateVolumeSnapshotClass", "VolumeSnapshotClass created successfully"); statusErr != nil {
		return ctrl.Result{}, statusErr
	}

	// Create the StorageClass if it does not exist
	if err := CheckCreateStorageClass(ctx, r); err != nil {
		if statusErr := UpdateFailedCondition(ctx, "CreateStorageClass", "Failed to create StorageClass"); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}
	if statusErr := UpdateSucceededCondition(ctx, "CreateStorageClass", "StorageClass created successfully"); statusErr != nil {
		return ctrl.Result{}, statusErr
	}

	// Migrate the volume claims to use the new storage class
	if err := MigrateDeploymentVolumes(ctx, r, &deployment); err != nil {
		if statusErr := UpdateFailedCondition(ctx, "MigrateDeploymentVolumes", "Failed to migrate deployment volumes"); statusErr != nil {
			return ctrl.Result{}, statusErr
		}
		return ctrl.Result{}, err
	}
	if statusErr := UpdateSucceededCondition(ctx, "MigrateDeploymentVolumes", "Deployment volumes migrated successfully"); statusErr != nil {
		return ctrl.Result{}, statusErr
	}

	// Update the deployment with the new tolerations to match the new node pool taints
	if err := UpdateDeploymentWithTolerations(ctx, r, &deployment); err != nil {
		return ctrl.Result{}, err
	}

	if err = UpdateSucceededCondition(ctx, "MigrationCompleted", "Migration completed successfully"); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *StoragePoolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).For(&v1.Deployment{}).
		Complete(r)
}
