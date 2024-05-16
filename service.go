package attacknet

import (
	"attacknet/cmd/internal/pkg/chaos"
	"attacknet/cmd/internal/pkg/kurtosis"
	"attacknet/cmd/pkg/test_executor"
)

// Config contains all the information required to run an attacknet experiment
type Config struct {
	KurtosisConfig *kurtosis.Config `yaml:"kurtosis_config"`
	// The chaos tests to be injected into the running enclave
	//ChaosConfig *chaos.Config `yaml:"chaos_config"`
}

// Service the manager service is responsible for interacting with enclaves and performing chaos experiments
type Service struct {
	Config          *Config
	KurtosisService *kurtosis.Service
	ChaosService    *chaos.Service
	TestExecutor    *test_executor.TestExecutor
}
