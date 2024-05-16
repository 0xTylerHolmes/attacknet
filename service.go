package attacknet

import (
	"attacknet/cmd/internal/pkg/chaos"
	"attacknet/cmd/internal/pkg/kurtosis"
	"attacknet/cmd/pkg/test_executor"
	"context"
)

// Config contains all the information required to run an attacknet experiment
type Config struct {
	KurtosisConfig *kurtosis.Config `yaml:"kurtosis_config"`
	// The chaos tests to be injected into the running enclave
	ChaosConfig *chaos.Config `yaml:"chaos_config"`
}

// Service the manager service is responsible for interacting with enclaves and performing chaos experiments
type Service struct {
	Config          *Config
	KurtosisService *kurtosis.Service
	ChaosService    *chaos.Service
	TestExecutor    *test_executor.TestExecutor
}

// NewService create a new attacknet service to manage chaos experiments
func NewService(ctx context.Context, config *Config) (*Service, error) {
	kurtosisService, err := kurtosis.NewService(ctx, config.KurtosisConfig)
	if err != nil {
		return nil, err
	}
	chaosService, err := chaos.NewService(config.KurtosisConfig.EnclaveNamespace)
	if err != nil {
		return nil, err
	}

	//TODO: figure out where to handle the list of tests. for testing just use the first one.
	test := config.ChaosConfig.Tests[0]
	experimentRunner := test_executor.CreateTestExecutor(chaosService.ChaosClient, test)

	return &Service{
		Config:          config,
		KurtosisService: kurtosisService,
		ChaosService:    chaosService,
		TestExecutor:    experimentRunner,
	}, nil
}
