package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GetMigrationConfigMap(ctx context.Context, r *StoragePoolReconciler) (corev1.ConfigMap, error) {
	migrationCONFIGMAP := corev1.ConfigMap{}
	if err := r.Get(ctx, client.ObjectKey{Name: "sp-migration-config", Namespace: "storage-pool-migrator-system"}, &migrationCONFIGMAP); err != nil {
		// Handle error
		return corev1.ConfigMap{}, err
	}
	return migrationCONFIGMAP, nil
}

func GetMigrationConfigMapDataValue(ctx context.Context, r *StoragePoolReconciler, key string) (string, error) {
	migrationCONFIGMAP, err := GetMigrationConfigMap(ctx, r)
	if err != nil {
		return "", err
	}

	value, exists := migrationCONFIGMAP.Data[key]
	if exists == false {
		return "", fmt.Errorf("Key %s not found in ConfigMap sp-migration-config", key)
	}

	return value, nil
}
