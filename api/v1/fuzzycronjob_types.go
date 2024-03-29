/*
Copyright 2024.

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

package v1

import (
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// Copied from https://pkg.go.dev/k8s.io/api@v0.28.3/batch/v1#CronJobSpec with Schedule removed.
type CronJobSpec struct {
	TimeZone *string `json:"timeZone,omitempty" protobuf:"bytes,8,opt,name=timeZone"`

	StartingDeadlineSeconds *int64 `json:"startingDeadlineSeconds,omitempty" protobuf:"varint,2,opt,name=startingDeadlineSeconds"`

	ConcurrencyPolicy batchv1.ConcurrencyPolicy `json:"concurrencyPolicy,omitempty" protobuf:"bytes,3,opt,name=concurrencyPolicy,casttype=ConcurrencyPolicy"`

	Suspend *bool `json:"suspend,omitempty" protobuf:"varint,4,opt,name=suspend"`

	JobTemplate batchv1.JobTemplateSpec `json:"jobTemplate" protobuf:"bytes,5,opt,name=jobTemplate"`

	SuccessfulJobsHistoryLimit *int32 `json:"successfulJobsHistoryLimit,omitempty" protobuf:"varint,6,opt,name=successfulJobsHistoryLimit"`

	FailedJobsHistoryLimit *int32 `json:"failedJobsHistoryLimit,omitempty" protobuf:"varint,7,opt,name=failedJobsHistoryLimit"`
}

// FuzzyCronJobSpec defines the desired state of FuzzyCronJob
type FuzzyCronJobSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of FuzzyCronJob. Edit fuzzycronjob_types.go to remove/update
	Foo string `json:"foo,omitempty"`

	Schedule string `json:"schedule"`

	CronJob CronJobSpec `json:"cronJob"`
}

// FuzzyCronJobStatus defines the observed state of FuzzyCronJob
type FuzzyCronJobStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Schedule string `json:"schedule,omitempty"`
}

//+kubebuilder:resource:shortName=fcj
//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Fuzzy Schedule",type="string",JSONPath=`.spec.schedule`
//+kubebuilder:printcolumn:name="Schedule",type="string",JSONPath=`.status.schedule`

// FuzzyCronJob is the Schema for the fuzzycronjobs API
type FuzzyCronJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   FuzzyCronJobSpec   `json:"spec,omitempty"`
	Status FuzzyCronJobStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// FuzzyCronJobList contains a list of FuzzyCronJob
type FuzzyCronJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []FuzzyCronJob `json:"items"`
}

func init() {
	SchemeBuilder.Register(&FuzzyCronJob{}, &FuzzyCronJobList{})
}
