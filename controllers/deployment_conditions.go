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
}

func NewStoragePoolMigrationCondition(deployment *v1.Deployment, storagePoolReconciler *StoragePoolReconciler) *StoragePoolMigrationCondition {
	return &StoragePoolMigrationCondition{
		Deployment:            deployment,
		StoragePoolReconciler: storagePoolReconciler,
	}
}

func (s *StoragePoolMigrationCondition) IsFailed() bool {
	return s.Status == corev1.ConditionFalse
}

func (s *StoragePoolMigrationCondition) IsSucceeded() bool {
	return s.Status == corev1.ConditionTrue
}

func (s *StoragePoolMigrationCondition) UpdateCondition(ctx context.Context, condition *v1.DeploymentCondition) error {
	for i, c := range s.Deployment.Status.Conditions {
		if c.Type == condition.Type {
			s.Deployment.Status.Conditions[i] = *condition
			return nil
		}
	}
	s.Deployment.Status.Conditions = append(s.Deployment.Status.Conditions, *condition)
	if err := s.StoragePoolReconciler.Status().Update(ctx, s.Deployment); err != nil {
		return err
	}
	return nil
}

func (s *StoragePoolMigrationCondition) UpdateFailedCondition(ctx context.Context, reason string, msg string) error {
	s.Type = "StoragePoolMigration"
	s.Status = corev1.ConditionFalse
	s.LastUpdateTime = metav1.Now()
	s.Reason = reason
	s.Message = msg
	if err := s.UpdateCondition(ctx, &s.DeploymentCondition); err != nil {
		return err
	}
	return nil
}

func (s *StoragePoolMigrationCondition) UpdateSucceededCondition(ctx context.Context, reason string, msg string) error {
	s.Type = "StoragePoolMigration"
	s.Status = corev1.ConditionTrue
	s.LastUpdateTime = metav1.Now()
	s.Reason = reason
	s.Message = msg
	if err := s.UpdateCondition(ctx, &s.DeploymentCondition); err != nil {
		return err
	}
	return nil
}
