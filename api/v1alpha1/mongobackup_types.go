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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BackupDestination defines the Destination to store backups
type BackupDestination struct {
	// Destination folder to save the backup
	Path string `json:"path"`
}

// MongoBackupSpec defines the desired state of MongoBackup
type MongoBackupSpec struct {

	// Mongodb Host to connect to
	Host string `json:"host"`

	// Mongodb database to take backup from
	Database string `json:"database"`

	// The schedule in Cron format, see https://en.wikipedia.org/wiki/Cron.
	Schedule string `json:"schedule"`

	// Rclone configuration to save the backup
	RcloneConfig string `json:"rcloneConfig"`

	// Destination to save the backup
	Destination BackupDestination `json:"destination"`
}

// MongoBackupStatus defines the observed state of MongoBackup
type MongoBackupStatus struct {
	// Information when the last time the backup job was done
	LastRun *metav1.Time `json:"lastRun,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Host",type=string,JSONPath=`.spec.host`
//+kubebuilder:printcolumn:name="Database",type=string,JSONPath=`.spec.database`
//+kubebuilder:printcolumn:name="Schedule",type=string,JSONPath=`.spec.schedule`
//+kubebuilder:printcolumn:name="Destination",type=string,JSONPath=`.spec.destination.path`
//+kubebuilder:printcolumn:name="LastRun",type=string,JSONPath=`.status.lastRun`

// MongoBackup is the Schema for the mongobackups API
type MongoBackup struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MongoBackupSpec   `json:"spec,omitempty"`
	Status MongoBackupStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// MongoBackupList contains a list of MongoBackup
type MongoBackupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MongoBackup `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MongoBackup{}, &MongoBackupList{})
}
