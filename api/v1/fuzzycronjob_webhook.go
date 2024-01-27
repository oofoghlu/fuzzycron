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
	"oofoghlu/fuzzycron/internal/utils"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// log is for logging in this package.
var fuzzycronjoblog = logf.Log.WithName("fuzzycronjob-resource")

// SetupWebhookWithManager will setup the manager to manage the webhooks
func (r *FuzzyCronJob) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// TODO(user): EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-batch-oofoghlu-v1-fuzzycronjob,mutating=false,failurePolicy=fail,sideEffects=None,groups=batch.oofoghlu,resources=fuzzycronjobs,verbs=create;update,versions=v1,name=vfuzzycronjob.kb.io,admissionReviewVersions=v1

var _ webhook.Validator = &FuzzyCronJob{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *FuzzyCronJob) ValidateCreate() (admission.Warnings, error) {
	fuzzycronjoblog.Info("validate create", "name", r.Name)

	_, err := utils.EvalCrontab(r.Spec.Schedule, r.Namespace+r.Name)
	if err != nil {
		err = field.Invalid(field.NewPath("spec").Child("schedule"), r.Spec.Schedule, err.Error())
	}
	return nil, err
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *FuzzyCronJob) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	fuzzycronjoblog.Info("validate update", "name", r.Name)

	_, err := utils.EvalCrontab(r.Spec.Schedule, r.Namespace+r.Name)
	if err != nil {
		err = field.Invalid(field.NewPath("spec").Child("schedule"), r.Spec.Schedule, err.Error())
	}
	return nil, err
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *FuzzyCronJob) ValidateDelete() (admission.Warnings, error) {
	fuzzycronjoblog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil, nil
}
