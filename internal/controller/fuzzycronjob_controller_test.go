package controller

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	fuzzycronjobv1 "oofoghlu/fuzzycron/api/v1"
)

var _ = Describe("Fuzzycronjob Controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		CronjobName      = "test-cronjob"
		CronjobNamespace = "default"
		JobName          = "test-job"

		timeout = time.Second * 10
		// duration = time.Second * 10
		interval = time.Millisecond * 250
	)
	var fuzzyCronJob *fuzzycronjobv1.FuzzyCronJob
	var ctx context.Context

	BeforeEach(func() {
		fuzzyCronJob = &fuzzycronjobv1.FuzzyCronJob{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "batch.oofoghlu/v1",
				Kind:       "FuzzyCronJob",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      CronjobName,
				Namespace: CronjobNamespace,
			},
			Spec: fuzzycronjobv1.FuzzyCronJobSpec{
				Schedule: "1 * * * *",
				CronJob: fuzzycronjobv1.CronJobSpec{
					JobTemplate: batchv1.JobTemplateSpec{
						Spec: batchv1.JobSpec{
							// For simplicity, we only fill out the required fields.
							Template: v1.PodTemplateSpec{
								Spec: v1.PodSpec{
									// For simplicity, we only fill out the required fields.
									Containers: []v1.Container{
										{
											Name:  "test-container",
											Image: "test-image",
										},
									},
									RestartPolicy: v1.RestartPolicyOnFailure,
								},
							},
						},
					},
				},
			},
		}
		ctx = context.Background()
	})

	Context("When creating FuzzyCronJob", func() {
		It("Should be created successfully", func() {
			Expect(k8sClient.Create(ctx, fuzzyCronJob)).Should(Succeed())
			fuzzycronjobLookupKey := types.NamespacedName{Name: CronjobName, Namespace: CronjobNamespace}
			createdFuzzyCronjob := &fuzzycronjobv1.FuzzyCronJob{}

			// We'll need to retry getting this newly created FuzzyCronJob, given that creation may not immediately happen.
			By("By creating a new FuzzyCronJob successfully")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, fuzzycronjobLookupKey, createdFuzzyCronjob)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			cronjobLookupKey := types.NamespacedName{Name: CronjobName, Namespace: CronjobNamespace}
			createdCronjob := &batchv1.CronJob{}

			// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
			By("By creating a new CronJob successfully")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			// Status should be updated.
			By("By Updating the FuzzyCronJob Status with the evaulated schedule successfully")
			Eventually(func() bool {
				err := k8sClient.Get(ctx, fuzzycronjobLookupKey, createdFuzzyCronjob)
				return err == nil && createdFuzzyCronjob.Status.Schedule == "1 * * * *"
			}, timeout, interval).Should(BeTrue())
		})

		When("Evaluating a hashed schedule", func() {
			It("Should be evaluated and created successfully", func() {
				fuzzyCronJob.ObjectMeta.Name = "hashed-cron-job"
				fuzzyCronJob.Spec.Schedule = "* H * * *"
				Expect(k8sClient.Create(ctx, fuzzyCronJob)).Should(Succeed())
				fuzzycronjobLookupKey := types.NamespacedName{Name: "hashed-cron-job", Namespace: CronjobNamespace}
				createdFuzzyCronjob := &fuzzycronjobv1.FuzzyCronJob{}

				By("By Updating the FuzzyCronJob Status with the evaulated schedule successfully")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, fuzzycronjobLookupKey, createdFuzzyCronjob)
					return err == nil && createdFuzzyCronjob.Status.Schedule == "* 19 * * *"
				}, timeout, interval).Should(BeTrue())

				cronjobLookupKey := types.NamespacedName{Name: "hashed-cron-job", Namespace: CronjobNamespace}
				createdCronjob := &batchv1.CronJob{}

				// We'll need to retry getting this newly created CronJob, given that creation may not immediately happen.
				By("By creating a new CronJob successfully")
				Eventually(func() bool {
					err := k8sClient.Get(ctx, cronjobLookupKey, createdCronjob)
					return err == nil && createdFuzzyCronjob.Status.Schedule == createdCronjob.Spec.Schedule
				}, timeout, interval).Should(BeTrue())
			})
		})
	})

	Context("When updating FuzzyCronJob", func() {
		It("Should be updated successfully", func() {
			fuzzycronjobLookupKey := types.NamespacedName{Name: CronjobName, Namespace: CronjobNamespace}
			updatedFuzzyCronjob := &fuzzycronjobv1.FuzzyCronJob{}
			k8sClient.Get(ctx, fuzzycronjobLookupKey, updatedFuzzyCronjob)
			updatedFuzzyCronjob.Spec.Schedule = "2 * * * *"
			Expect(k8sClient.Update(ctx, updatedFuzzyCronjob)).Should(Succeed())

			cronjobLookupKey := types.NamespacedName{Name: CronjobName, Namespace: CronjobNamespace}
			updatedCronjob := &batchv1.CronJob{}

			// Updating fcj should also update cj and fcj status
			Eventually(func() bool {
				err := k8sClient.Get(ctx, cronjobLookupKey, updatedCronjob)
				if err == nil {
					return true
				}
				err = k8sClient.Get(ctx, fuzzycronjobLookupKey, updatedFuzzyCronjob)
				return err == nil && updatedCronjob.Spec.Schedule == "2 * * * *" && updatedFuzzyCronjob.Status.Schedule == "2 * * * *"
			}, timeout, interval).Should(BeTrue())
		})
	})
})
