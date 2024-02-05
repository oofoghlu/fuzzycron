package utils

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fuzzy crontab evaluation", func() {
	Context("When evaluating", func() {
		It("Should be evaluated successfully for standard crontab expressions", func() {
			schedule, err := EvalCrontab("* * * * *", "namespace-name")
			Expect(schedule).Should(Equal("* * * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("5 10 * * *", "namespace-name")
			Expect(schedule).Should(Equal("5 10 * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("@hourly", "namespace-name")
			Expect(schedule).Should(Equal("@hourly"))
			Expect(err).Should(BeNil())
		})

		It("Should be evaluated unsuccessfully for non-standard crontab expressions", func() {
			schedule, err := EvalCrontab("* * * * * * *", "namespace-name")
			Expect(schedule).Should(Equal(""))
			Expect(err).ShouldNot(BeNil())

			schedule, err = EvalCrontab("5 144 * * *", "namespace-name")
			Expect(schedule).Should(Equal(""))
			Expect(err).ShouldNot(BeNil())

			schedule, err = EvalCrontab("* * *", "namespace-name")
			Expect(schedule).Should(Equal(""))
			Expect(err).ShouldNot(BeNil())
		})

		It("Should be evaluated successfully for hashed expressions", func() {
			schedule, err := EvalCrontab("H * * * *", "namespace-name")
			Expect(schedule).Should(Equal("20 * * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H * * *", "namespace-name")
			Expect(schedule).Should(Equal("20 20 * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H * * *", "namespace-name-2")
			Expect(schedule).Should(Equal("27 3 * * *"))
			Expect(err).Should(BeNil())
		})
	})
})
