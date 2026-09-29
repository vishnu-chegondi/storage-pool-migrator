package controllers

import (
	"context"

	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type StoragePoolMigrationCondition struct {
	v1.DeploymentCondition
	Deployment            *v1.Deployment         `json:"deployment,omitempty"`
	StoragePoolReconciler *StoragePoolReconciler `json:"storagePoolReconciler,omitempty"`
	Reason                string                 `json:"reason,omitempty"`
	Msg                   string                 `json:"msg,omitempty"`
}

func UpdateFailedCondition(ctx context.Context, reason string, msg string) error {
	s := &StoragePoolMigrationCondition{}
	s.Type = "StoragePoolMigration"
	s.Status = corev1.ConditionFalse
	s.LastUpdateTime = metav1.Now()
	s.Reason = reason
	s.Message = msg
	if err := UpdateDeploymentCondition(ctx, s.StoragePoolReconciler, s.Deployment, &s.DeploymentCondition); err != nil {
		return err
	}
	return nil
}

func UpdateSucceededCondition(ctx context.Context, reason string, msg string) error {
	s := &StoragePoolMigrationCondition{}
	s.Type = "StoragePoolMigration"
	s.Status = corev1.ConditionTrue
	s.LastUpdateTime = metav1.Now()
	s.Reason = reason
	s.Message = msg
	if err := UpdateDeploymentCondition(ctx, s.StoragePoolReconciler, s.Deployment, &s.DeploymentCondition); err != nil {
		return err
	}
	return nil
}
