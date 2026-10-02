package controllers

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	storageV1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func CheckCreateStorageClass(ctx context.Context, r *StoragePoolReconciler) error {
	storageCLASSNAME, err := GetMigrationConfigMapDataValue(ctx, r, "NEW_STORAGE_CLASS")
	if err != nil {
		return err
	}

	storageClass := storageV1.StorageClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "storage.k8s.io/v1",
			Kind:       "StorageClass",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: storageCLASSNAME,
		},
		Provisioner: "pd.csi.storage.gke.io",
		Parameters: map[string]string{
			"type":                             "hyperdisk-balanced",
			"provisioned-iops-on-create":       "3000",  // Minimum is 3,000 baseline
			"provisioned-throughput-on-create": "140Mi", // Minimum is 140MiB/s baseline
		},
		VolumeBindingMode:    new(storageV1.VolumeBindingWaitForFirstConsumer),
		AllowVolumeExpansion: new(true),
	}
	if err := r.Create(ctx, &storageClass); err != nil {
		return err
	}

	return nil
}

func GetVolumeClaimSpec(ctx context.Context, r *StoragePoolReconciler, pvcNAME string, namespace string) (v1.PersistentVolumeClaimSpec, error) {
	var pvc v1.PersistentVolumeClaim
	if err := r.Get(ctx, types.NamespacedName{Name: pvcNAME, Namespace: namespace}, &pvc); err != nil {
		return v1.PersistentVolumeClaimSpec{}, err
	}
	return pvc.Spec, nil
}

func DeleteOrRenameVolumeClaims(ctx context.Context, r *StoragePoolReconciler, pvcNAME string, namespace string) error {
	var pvc v1.PersistentVolumeClaim
	deleteVolumeClaim, err := GetMigrationConfigMapDataValue(ctx, r, "DELETE_OLD_VOLUME_CLAIMS")
	if err != nil {
		return err
	}

	if err := r.Get(ctx, types.NamespacedName{Name: pvcNAME, Namespace: namespace}, &pvc); err != nil {
		return err
	}

	if deleteVolumeClaim == "true" {
		if err := r.Delete(ctx, &pvc); err != nil {
			return err
		}
	} else {
		pvc.Name = pvc.Name + "-old"
		if err := r.Update(ctx, &pvc); err != nil {
			return err
		}
	}
	return nil
}

func CreateNewVolumeClaims(ctx context.Context, r *StoragePoolReconciler, pvcNAME string, namespace string, volumeClaimSpec v1.PersistentVolumeClaimSpec) error {
	storageCLASSNAME, err := GetMigrationConfigMapDataValue(ctx, r, "NEW_STORAGE_CLASS")
	if err != nil {
		return err
	}

	var pvc v1.PersistentVolumeClaim
	if err := r.Get(ctx, types.NamespacedName{Name: pvcNAME, Namespace: namespace}, &pvc); err == nil {
		return fmt.Errorf("VolumeClaim %s already exists", pvcNAME)
	}

	pvc = v1.PersistentVolumeClaim{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "PersistentVolumeClaim",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcNAME,
			Namespace: namespace,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources:        volumeClaimSpec.Resources,
			StorageClassName: &storageCLASSNAME,
			DataSource: &v1.TypedLocalObjectReference{
				Kind:     "VolumeSnapshot",
				Name:     pvcNAME + "-snapshot",
				APIGroup: new("snapshot.storage.k8s.io"),
			},
		},
	}
	if err := r.Create(ctx, &pvc); err != nil {
		return err
	}

	return nil
}
