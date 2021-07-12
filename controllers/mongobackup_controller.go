/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"encoding/base64"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	backupv1alpha1 "github.com/kmjayadeep/majo-operator/api/v1alpha1"
)

// MongoBackupReconciler reconciles a MongoBackup object
type MongoBackupReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=backup.16cloud.online,resources=mongobackups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=backup.16cloud.online,resources=mongobackups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=backup.16cloud.online,resources=mongobackups/finalizers,verbs=update
func (r *MongoBackupReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	mb := &backupv1alpha1.MongoBackup{}
	err := r.Get(ctx, req.NamespacedName, mb)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Error reading object")
		return ctrl.Result{}, err
	}

	se := &corev1.Secret{}
	err = r.Get(ctx, req.NamespacedName, se)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Secret doesn't exist, creating")
			se, err = r.secretForMongoBackup(mb)

			if err != nil {
				logger.Error(err, "Unable to create secret, unrecoverable error")
				return ctrl.Result{}, nil
			}

			if err := r.Create(ctx, se); err != nil {
				logger.Error(err, "Unable to create secret")
				return ctrl.Result{}, err
			}
			logger.Info("secret created successfully!")
			return ctrl.Result{Requeue: true}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Error reading secret, requeuing")
		return ctrl.Result{}, err
	}

	cron := &batchv1beta1.CronJob{}
	err = r.Get(ctx, req.NamespacedName, cron)
	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Backup job doesn't exist, creating")
			cron = r.cronJobForMongoBackup(mb)
			if err := r.Create(ctx, cron); err != nil {
				logger.Error(err, "Unable to create cronjob")
				return ctrl.Result{}, err
			}
			logger.Info("Cronjob created successfully!")
			return ctrl.Result{Requeue: true}, nil
		}
		// Error reading the object - requeue the request.
		logger.Error(err, "Error reading cronjob, requeuing")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MongoBackupReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&backupv1alpha1.MongoBackup{}).
		Complete(r)
}

func (r *MongoBackupReconciler) secretForMongoBackup(mb *backupv1alpha1.MongoBackup) (*corev1.Secret, error) {
	conf, err := base64.RawStdEncoding.DecodeString(mb.Spec.RcloneConfig)

	if err != nil {
		return nil, err
	}

	se := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mb.Name,
			Namespace: mb.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: map[string][]byte{
			"rclone.conf": conf,
		},
	}

	// Set MongoBackup as owner for secret
	ctrl.SetControllerReference(mb, se, r.Scheme)

	return se, nil
}

func (r *MongoBackupReconciler) cronJobForMongoBackup(mb *backupv1alpha1.MongoBackup) *batchv1beta1.CronJob {

	var hLimit int32 = 1

	cron := &batchv1beta1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mb.Name,
			Namespace: mb.Namespace,
		},
		Spec: batchv1beta1.CronJobSpec{
			Schedule:                   mb.Spec.Schedule,
			SuccessfulJobsHistoryLimit: &hLimit,
			FailedJobsHistoryLimit:     &hLimit,
			ConcurrencyPolicy:          batchv1beta1.ForbidConcurrent,
			JobTemplate: batchv1beta1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						Spec: corev1.PodSpec{
							RestartPolicy: corev1.RestartPolicyOnFailure,
							Volumes: []corev1.Volume{
								{
									Name: "temp-volume",
									VolumeSource: corev1.VolumeSource{
										EmptyDir: &corev1.EmptyDirVolumeSource{},
									},
								},
								{
									Name: "rclone-config",
									VolumeSource: corev1.VolumeSource{
										Secret: &corev1.SecretVolumeSource{
											SecretName: mb.Name,
										},
									},
								},
							},
							InitContainers: []corev1.Container{{
								Name:  "backup",
								Image: "istepanov/mongodump:4.4",
								Command: []string{
									"/bin/sh",
									"-c",

									"echo \"running a script\"" +
										"DIR=`date +\"%Y-%m-%d_%T\"`" +
										"DEST=/mongodump/$DIR" +
										"mongodump -h $DB_HOST -d $DB_NAME -o $DEST --gzip || { echo 'mongo backup failed' ; exit 1; }",
								},
								Env: []corev1.EnvVar{
									{
										Name:  "DB_NAME",
										Value: mb.Spec.Database,
									},
									{
										Name:  "DB_HOST",
										Value: mb.Spec.Host,
									},
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										MountPath: "/mongodump",
										Name:      "temp-volume",
									},
								},
							}},
							Containers: []corev1.Container{{
								Name:  "upload",
								Image: "rclone/rclone:1",
								Command: []string{
									"rclone",
									"--config",
									"/config/rclone.conf",
									"copy",
									"/mongodump",
									mb.Spec.Destination.Path,
									"-P",
									"-v",
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										MountPath: "/mongodump",
										Name:      "temp-volume",
									},
									{
										MountPath: "/config",
										Name:      "rclone-config",
									},
								},
							}},
						},
					},
				},
			},
		},
	}

	// Set MongoBackup as owner
	ctrl.SetControllerReference(mb, cron, r.Scheme)

	return cron
}
