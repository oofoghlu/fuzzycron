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

package controller

import (
	"context"
	"fmt"

	batchv1 "oofoghlu/fuzzycron/api/v1"
	"oofoghlu/fuzzycron/internal/metrics"

	"github.com/oofoghlu/fuzzycrontab"

	"github.com/jinzhu/copier"
	kbatch "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// FuzzyCronJobReconciler reconciles a FuzzyCronJob object
type FuzzyCronJobReconciler struct {
	client.Client
	Scheme   *runtime.Scheme
	Recorder record.EventRecorder
}

//+kubebuilder:rbac:groups=batch.oofoghlu,resources=fuzzycronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.oofoghlu,resources=fuzzycronjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.oofoghlu,resources=fuzzycronjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the FuzzyCronJob object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *FuzzyCronJobReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)
	// TODO(user): your logic here
	var fuzzyCronJob batchv1.FuzzyCronJob
	if err := r.Get(ctx, req.NamespacedName, &fuzzyCronJob); err != nil {
		log.Error(err, "unable to fetch FuzzyCronJob")
		// TODO deletes work and propagate but there must be a better way than returning error here every time??
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	var childCronJobs kbatch.CronJobList
	if err := r.List(ctx, &childCronJobs, client.InNamespace(req.Namespace), client.MatchingFields{cronJobOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list child CronJobs")
		return ctrl.Result{}, err
	}

	constructCronJobForFuzzyCronJob := func(fuzzyCronJob *batchv1.FuzzyCronJob, schedule string) (*kbatch.CronJob, error) {
		// We want job names for a given nominal start time to have a deterministic name to avoid the same job being created twice
		cronJobSpec := &kbatch.CronJobSpec{Schedule: schedule}
		copier.Copy(cronJobSpec, fuzzyCronJob.Spec.CronJob)
		cronjob := &kbatch.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
				Name:        fuzzyCronJob.Name,
				Namespace:   fuzzyCronJob.Namespace,
			},
			Spec: *cronJobSpec,
		}
		for k, v := range fuzzyCronJob.Annotations {
			cronjob.Annotations[k] = v
		}
		for k, v := range fuzzyCronJob.Labels {
			cronjob.Labels[k] = v
		}
		if err := ctrl.SetControllerReference(fuzzyCronJob, cronjob, r.Scheme); err != nil {
			return nil, err
		}

		return cronjob, nil
	}

	// Evaluate hashes in crontab expression
	schedule, err := fuzzycrontab.EvalCrontab(fuzzyCronJob.Spec.Schedule, req.Namespace+fuzzyCronJob.Name)
	if err != nil {
		log.Error(err, "unable to construct CronJob schedule from FuzzyCronJob schedule")
		// don't bother requeuing until we get a change to the spec
		return ctrl.Result{}, nil
	}
	cronjob, err := constructCronJobForFuzzyCronJob(&fuzzyCronJob, schedule)
	if err != nil {
		log.Error(err, "unable to construct cronjob from template")
		// don't bother requeuing until we get a change to the spec
		return ctrl.Result{}, nil
	}

	if len(childCronJobs.Items) == 0 {
		if err := r.Create(ctx, cronjob); err != nil {
			metrics.CronCreationError.WithLabelValues(fuzzyCronJob.Name, fuzzyCronJob.Namespace).Inc()
			log.Error(err, "unable to create CronJob for FuzzyCronJob", "cronjob", cronjob)
			r.Recorder.Event(&fuzzyCronJob, "Warning", "FailedCreating",
				fmt.Sprintf("CronJob Resource %s failed to be created from the namespace %s",
					cronjob.Name,
					cronjob.Namespace))
			return ctrl.Result{}, err
		}

		metrics.CronCreation.WithLabelValues(fuzzyCronJob.Name, fuzzyCronJob.Namespace).Inc()
		r.Recorder.Event(&fuzzyCronJob, "Normal", "Creating",
			fmt.Sprintf("CronJob Resource %s is being created from the namespace %s",
				cronjob.Name,
				cronjob.Namespace))
		log.V(1).Info("created CronJob for FuzzyCronJob", "cronjob", cronjob)

	} else {
		// TODO add labels/annotation to equality check to allow for updates to them.
		if !equality.Semantic.DeepDerivative(cronjob.Spec, childCronJobs.Items[0].Spec) {
			if err := r.Update(ctx, cronjob); err != nil {
				metrics.CronUpdateError.WithLabelValues(fuzzyCronJob.Name, fuzzyCronJob.Namespace).Inc()
				log.Error(err, "unable to update CronJob for FuzzyCronJob", "cronjob", cronjob)
				r.Recorder.Event(&fuzzyCronJob, "Warning", "FailedUpdating",
					fmt.Sprintf("CronJob Resource %s failed to be updated from the namespace %s",
						cronjob.Name,
						cronjob.Namespace))
				return ctrl.Result{}, err
			}

			metrics.CronUpdate.WithLabelValues(fuzzyCronJob.Name, fuzzyCronJob.Namespace).Inc()
			r.Recorder.Event(&fuzzyCronJob, "Normal", "Updating",
				fmt.Sprintf("CronJob Resource %s is being updated from the namespace %s",
					cronjob.Name,
					cronjob.Namespace))
			log.V(1).Info("updated CronJob for FuzzyCronJob", "cronjob", cronjob)
		} else {
			log.V(1).Info("no differences in cronjob so skipping update", "cronjob", cronjob)
		}
	}

	if fuzzyCronJob.Status.Schedule != schedule {
		patch := client.MergeFrom(fuzzyCronJob.DeepCopy())
		fuzzyCronJob.Status.Schedule = schedule

		// patch is safer than update here as the resource version isn't guaranteed to match and we don't want to have
		// to fetch the latest version simply to update the status.
		if err := r.Status().Patch(ctx, &fuzzyCronJob, patch); err != nil {

			metrics.FuzzyStatusUpdateError.WithLabelValues(fuzzyCronJob.Name, fuzzyCronJob.Namespace).Inc()
			log.Error(err, "unable to update FuzzyCronJob status")
			return ctrl.Result{}, err
		}

		metrics.FuzzyStatusUpdate.WithLabelValues(fuzzyCronJob.Name, fuzzyCronJob.Namespace).Inc()
		log.V(1).Info("Successfully updated Status", "fuzzycronjob", fuzzyCronJob)
	} else {
		log.V(1).Info("no differences in cronjob status so skipping patch", "fuzzycronjob", fuzzyCronJob.Status)
	}

	return ctrl.Result{}, nil
}

var (
	cronJobOwnerKey = ".metadata.controller"
	apiGVStr        = batchv1.GroupVersion.String()
)

// SetupWithManager sets up the controller with the Manager.
func (r *FuzzyCronJobReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &kbatch.CronJob{}, cronJobOwnerKey, func(rawObj client.Object) []string {
		// grab the cronjob object, extract the owner...
		cronJob := rawObj.(*kbatch.CronJob)
		owner := metav1.GetControllerOf(cronJob)
		if owner == nil {
			return nil
		}
		// ...make sure it's a FuzzyCronJob...
		if owner.APIVersion != apiGVStr || owner.Kind != "FuzzyCronJob" {
			return nil
		}

		// ...and if so, return it
		return []string{owner.Name}
	}); err != nil {
		return err
	}

	return ctrl.NewControllerManagedBy(mgr).
		For(&batchv1.FuzzyCronJob{}).
		Owns(&kbatch.CronJob{}).
		Complete(r)
}
