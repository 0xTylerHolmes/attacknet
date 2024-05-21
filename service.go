package attacknet

import (
	"attacknet/cmd/internal/pkg/chaos"
	"attacknet/cmd/internal/pkg/kurtosis"
	"context"
)

// Config contains all the information required to run an attacknet experiment
type Config struct {
	EnclaveName       string `yaml:"enclave_name"`
	EnclaveNamespace  string `yaml:"enclave_namespace"`
	KurtosisPackageID string `yaml:"kurtosis_package_id"`

	KurtosisConfig *kurtosis.Config `yaml:"kurtosis_config"`
	// The chaos tests to be injected into the running enclave
	ChaosConfig *chaos.Config `yaml:"chaos_config"`
}

// Service the manager service is responsible for interacting with enclaves and performing chaos experiments
type Service struct {
	Config          *Config
	KurtosisService *kurtosis.Service
	ChaosService    *chaos.Service
}

// NewService create a new attacknet service to manage chaos experiments
func NewService(ctx context.Context, config *Config) (*Service, error) {
	kurtosisService, err := kurtosis.NewService(ctx, config.KurtosisConfig, config.KurtosisPackageID, config.EnclaveName)
	if err != nil {
		return nil, err
	}
	chaosService, err := chaos.NewService(config.EnclaveNamespace)
	if err != nil {
		return nil, err
	}

	return &Service{
		Config:          config,
		KurtosisService: kurtosisService,
		ChaosService:    chaosService,
	}, nil
}
