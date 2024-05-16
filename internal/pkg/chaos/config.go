package chaos

import "time"

type AttacknetConfig struct {
	GrafanaPodName             string `yaml:"grafanaPodName"`
	GrafanaPodPort             string `yaml:"grafanaPodPort"`
	AllowPostFaultInspection   bool   `yaml:"allowPostFaultInspection"`
	WaitBeforeInjectionSeconds uint32 `yaml:"waitBeforeInjectionSeconds"`
	ReuseDevnetBetweenRuns     bool   `yaml:"reuseDevnetBetweenRuns"`
	ExistingDevnetNamespace    string `yaml:"existingDevnetNamespace"`
}

type HarnessConfig struct {
	NetworkType       string `yaml:"networkType"`
	NetworkPackage    string `yaml:"networkPackage"`
	NetworkConfigPath string `yaml:"networkConfig"`
}

type HarnessConfigParsed struct {
	NetworkType    string
	NetworkPackage string
	NetworkConfig  []byte
}

type Config struct {
	Tests []SuiteTest `yaml:"tests"`
	// StartNewDevnet specifies whether the experiment should be run on a fresh devnet
	StartNewDevnet bool `yaml:"start_new_devnet"`
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

//type Config struct {
//	AttacknetConfig AttacknetConfig  `yaml:"attacknetConfig"`
//	HarnessConfig   HarnessConfig    `yaml:"harnessConfig"`
//	TestConfig      Config `yaml:"testConfig"`
//}

//type ConfigParsed struct {
//	AttacknetConfig AttacknetConfig
//	HarnessConfig   HarnessConfigParsed
//	TestConfig      Config
//}

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
