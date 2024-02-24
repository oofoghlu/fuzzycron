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
			Expect(schedule).Should(Equal("40 * * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H * * *", "namespace-name")
			Expect(schedule).Should(Equal("40 15 * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H * * *", "namespace-name-2")
			Expect(schedule).Should(Equal("25 14 * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H H H H", "namespace-name-3")
			Expect(schedule).Should(Equal("20 19 11 6 4"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H H H H", "namespace-name-4")
			Expect(schedule).Should(Equal("43 20 16 7 3"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H H H H", "namespace-name-5")
			Expect(schedule).Should(Equal("14 1 11 4 2"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H H H H", "another-namespace-name")
			Expect(schedule).Should(Equal("14 1 1 4 2"))
			Expect(err).Should(BeNil())
		})

		It("Should be evaluated successfully for hashed expressions steps", func() {
			schedule, err := EvalCrontab("H/3 * * * *", "namespace-name-step")
			Expect(schedule).Should(Equal("2/3 * * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H/5 * * *", "namespace-name-step")
			Expect(schedule).Should(Equal("47 3/5 * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H * H/2 *", "namespace-name-step-2")
			Expect(schedule).Should(Equal("18 17 * 1/2 *"))
			Expect(err).Should(BeNil())
		})

		It("Should be evaluated successfully for hashed expressions steps with offset for lower bounds", func() {
			By("Never going out of their step bounds for any hash number")
			schedule, err := EvalCrontab("* * H/2 H/2 *", "namespace-name-step")
			Expect(schedule).Should(Equal("* * 1/2 1/2 *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("* * H/2 H/2 *", "namespace-name-step-2")
			Expect(schedule).Should(Equal("* * 1/2 1/2 *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("* * H/2 H/2 *", "namespace-name-step-3")
			Expect(schedule).Should(Equal("* * 1/2 1/2 *"))
			Expect(err).Should(BeNil())
		})

		It("Should be evaluated successfully for hashed expressions ranges", func() {
			schedule, err := EvalCrontab("H(0-4)/5 * * * *", "namespace-name-step")
			Expect(schedule).Should(Equal("3/5 * * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H(5-15)/20 H/5 * * *", "namespace-name-step")
			Expect(schedule).Should(Equal("12/20 3/5 * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H H * H/2 *", "namespace-name-step-2")
			Expect(schedule).Should(Equal("18 17 * 1/2 *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H(0-5) * * * *", "namespace-name-range")
			Expect(schedule).Should(Equal("4 * * * *"))
			Expect(err).Should(BeNil())

			schedule, err = EvalCrontab("H(5-15) H(1-4) H(2-6) H(1-11) H(3-7)", "namespace-name-range")
			Expect(schedule).Should(Equal("9 3 4 2 3"))
			Expect(err).Should(BeNil())
		})
	})
})
