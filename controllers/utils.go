package controllers

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetMigrationConfigMap(ctx context.Context, r *StoragePoolReconciler) (corev1.ConfigMap, error) {
	migrationCONFIGMAP := corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKey{Name: "sp-migration-config", Namespace: "default"}, &migrationCONFIGMAP); err != nil {
		// Handle error
		return corev1.ConfigMap{}, err
	}
	return migrationCONFIGMAP, nil
}
