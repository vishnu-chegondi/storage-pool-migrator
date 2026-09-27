package controllers

import (
	"context"
	"fmt"
	"os"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func GetVolumeClaimSpec(ctx context.Context, r *StoragePoolReconciler, pvcNAME string, namespace string) (v1.PersistentVolumeClaimSpec, error) {
	var pvc v1.PersistentVolumeClaim
	if err := r.Get(ctx, types.NamespacedName{Name: pvcNAME, Namespace: namespace}, &pvc); err != nil {
		return v1.PersistentVolumeClaimSpec{}, err
	}
	return pvc.Spec, nil
}

func DeleteVolumeClaims(ctx context.Context, r *StoragePoolReconciler, pvcNAME string, namespace string) error {
	var pvc v1.PersistentVolumeClaim
	if err := r.Get(ctx, types.NamespacedName{Name: pvcNAME, Namespace: namespace}, &pvc); err != nil {
		return err
	}

	if err := r.Delete(ctx, &pvc); err != nil {
		return err
	}
	return nil
}

func CreateNewVolumeClaims(ctx context.Context, r *StoragePoolReconciler, pvcNAME string, namespace string, volumeClaimSpec v1.PersistentVolumeClaimSpec) error {
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
			StorageClassName: new(os.Getenv("NEW_STORAGE_CLASS")), // TODO: SetUp the NEW_STORAGE_CLASS environment variable in the deployment manifest
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
