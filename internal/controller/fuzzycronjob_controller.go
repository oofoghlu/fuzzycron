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

	kbatch "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	batchv1 "oofoghlu/fuzzycron/api/v1"
)

// FuzzyCronJobReconciler reconciles a FuzzyCronJob object
type FuzzyCronJobReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=batch.oofoghlu,resources=fuzzycronjobs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=batch.oofoghlu,resources=fuzzycronjobs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=batch.oofoghlu,resources=fuzzycronjobs/finalizers,verbs=update
//+kubebuilder:rbac:groups=batch,resources=cronjobs,verbs=get;list;watch;create;update;patch;delete

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
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO get list of cronjobs associated. via cronJobOwnerKey.
	var childCronJobs kbatch.CronJobList
	if err := r.List(ctx, &childCronJobs, client.InNamespace(req.Namespace), client.MatchingFields{cronJobOwnerKey: req.Name}); err != nil {
		log.Error(err, "unable to list child CronJobs")
		return ctrl.Result{}, err
	}

	// TODO handle creates
	constructCronJobForFuzzyCronJob := func(fuzzyCronJob *batchv1.FuzzyCronJob) (*kbatch.CronJob, error) {
		// We want job names for a given nominal start time to have a deterministic name to avoid the same job being created twice
		cronjob := &kbatch.CronJob{
			ObjectMeta: metav1.ObjectMeta{
				Labels:      make(map[string]string),
				Annotations: make(map[string]string),
				Name:        fuzzyCronJob.Name,
				Namespace:   fuzzyCronJob.Namespace,
			},
			Spec: *fuzzyCronJob.Spec.CronJob.DeepCopy(),
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

	if len(childCronJobs.Items) == 0 {
		cronjob, err := constructCronJobForFuzzyCronJob(&fuzzyCronJob)
		if err != nil {
			log.Error(err, "unable to construct cronjob from template")
			// don't bother requeuing until we get a change to the spec
			return ctrl.Result{}, nil
		}

		if err := r.Create(ctx, cronjob); err != nil {
			log.Error(err, "unable to create CronJob for FuzzyCronJob", "cronjob", cronjob)
			return ctrl.Result{}, err
		}

		log.V(1).Info("created CronJob for FuzzyCronJob run", "cronjob", cronjob)
	}

	// TODO handle updates

	// TODO handle deletes

	// TODO handle cronjob out of sync

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
