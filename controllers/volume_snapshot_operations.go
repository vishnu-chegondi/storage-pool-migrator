package controllers

import (
	"context"
	"os"

	snapshotv1 "github.com/kubernetes-csi/external-snapshotter/client/v8/apis/volumesnapshot/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func CreateIfNotVolumeSnapshotClass(ctx context.Context, r *StoragePoolReconciler) error {
	volumeSnapshotClass := snapshotv1.VolumeSnapshotClass{}
	namespacedName := types.NamespacedName{
		Name: os.Getenv("STORAGE_SNAPSHOT_CLASS"),
	}
	if err := r.Get(ctx, namespacedName, &volumeSnapshotClass); err == nil {
		return nil
	}

	// If the VolumeSnapshotClass does not exist, create it
	volumeSnapshotClass = snapshotv1.VolumeSnapshotClass{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "snapshot.storage.k8s.io/v1",
			Kind:       "VolumeSnapshotClass"},
		ObjectMeta: metav1.ObjectMeta{
			Name: os.Getenv("STORAGE_SNAPSHOT_CLASS"), // TODO: SetUp the STORAGE_SNAPSHOT_CLASS environment variable in the deployment manifest
		},
		Driver:         "pd.csi.storage.gke.io",
		DeletionPolicy: snapshotv1.VolumeSnapshotContentRetain,
	}
	if err := r.Create(ctx, &volumeSnapshotClass); err != nil {
		return err
	}
	return nil
}

func CreateVolumeSnapshot(ctx context.Context, r *StoragePoolReconciler, pvcNAME string, namespace string) error {
	volumeSnapshot := snapshotv1.VolumeSnapshot{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pvcNAME + "-snapshot",
			Namespace: namespace,
		},
		Spec: snapshotv1.VolumeSnapshotSpec{
			Source: snapshotv1.VolumeSnapshotSource{
				PersistentVolumeClaimName: &pvcNAME,
			},
			VolumeSnapshotClassName: func() *string {
				v := os.Getenv("STORAGE_SNAPSHOT_CLASS")
				return &v
			}(),
		},
	}
	if err := r.Create(ctx, &volumeSnapshot); err != nil {
		return err
	}
	return nil
}
