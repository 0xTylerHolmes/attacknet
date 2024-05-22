package chaos

import "time"

type Config struct {
	Tests []SuiteTest `yaml:"tests"`
	// StartNewDevnet specifies whether the experiment should be run on a fresh devnet
	StartNewDevnet bool          `yaml:"start_new_devnet"`
	ChaosDelay     time.Duration `yaml:"chaos_delay"`
}

type HealthCheckConfig struct {
	EnableChecks bool           `yaml:"enableChecks"`
	GracePeriod  *time.Duration `yaml:"gracePeriod"`
}

type SuiteTest struct {
	TestName     string            `yaml:"testName"`
	PlanSteps    []PlanStep        `yaml:"planSteps"`
	HealthConfig HealthCheckConfig `yaml:"health"`
}

type StepType string

const (
	InvalidStepType        StepType = ""
	InjectFault            StepType = "injectFault"
	WaitForFaultCompletion StepType = "waitForFaultCompletion"
	WaitForDuration        StepType = "waitForDuration"
	// note: we'll have to think hard about how chaosSessions determine dead pods if we allow inter-step health checks.
	// WaitForHealthChecks    StepType = "waitForHealthChecks"
)

type PlanStep struct {
	StepType        StepType               `yaml:"stepType"`
	StepDescription string                 `yaml:"description"`
	Spec            map[string]interface{} `yaml:",inline"`
}
