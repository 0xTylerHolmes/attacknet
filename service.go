package attacknet

import (
	"attacknet/cmd/internal/pkg/artifacts"
	"attacknet/cmd/internal/pkg/chaos"
	"attacknet/cmd/internal/pkg/health"
	"attacknet/cmd/internal/pkg/kurtosis"
	"attacknet/cmd/internal/pkg/test_executor"
	"context"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"time"
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

func ReadAttacknetConfig(configFilePath string) (*Config, error) {
	data, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	var attacknetConfig Config
	err = yaml.Unmarshal(data, &attacknetConfig)
	return &attacknetConfig, err
}

// NewService create a new attacknet service to manage chaos experiments
func NewService(ctx context.Context, config *Config) (*Service, error) {
	log.Trace("creating the kurtosis service")
	kurtosisService, err := kurtosis.NewService(ctx, config.KurtosisConfig, config.KurtosisPackageID, config.EnclaveName)
	if err != nil {
		return nil, err
	}
	log.Trace("creating the chaos service")
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

func (s *Service) StartExperiment(ctx context.Context) error {
	log.Infof("Preparing the envlave for attacknet experiment. (restarting devnet = %v)", s.Config.ChaosConfig.StartNewDevnet)
	err := s.KurtosisService.PrepareEnclave(ctx, s.Config.ChaosConfig.StartNewDevnet)
	if err != nil {
		return err
	}

	// standby for timer
	log.Infof(
		"Waiting %f seconds before starting fault injection",
		s.Config.ChaosConfig.ChaosDelay.Seconds(),
	)
	time.Sleep(s.Config.ChaosConfig.ChaosDelay)

	log.Infof("Running %d tests", len(s.Config.ChaosConfig.Tests))

	var testArtifacts []*artifacts.TestArtifact

	for i, test := range s.Config.ChaosConfig.Tests {
		log.Infof("Running test (%d/%d): '%s'", i+1, len(s.Config.ChaosConfig.Tests), test.TestName)
		executor := test_executor.CreateTestExecutor(s.ChaosService.ChaosClient, test)

		err = executor.RunTestPlan(ctx)
		if err != nil {
			log.Errorf("Error while running test #%d", i+1)
			return err
		} else {
			log.Infof("Test #%d steps completed.", i+1)
		}

		if test.HealthConfig.EnableChecks {
			log.Info("Starting health checks")
			podsUnderTest, err := executor.GetPodsUnderTest()
			if err != nil {
				return err
			}

			hc, err := health.BuildHealthChecker(s.ChaosService.KubeClient, podsUnderTest, test.HealthConfig)
			if err != nil {
				return err
			}
			results, err := hc.RunChecks(ctx)
			if err != nil {
				return err
			}
			testArtifact := artifacts.BuildTestArtifact(results, podsUnderTest, test)
			testArtifacts = append(testArtifacts, testArtifact)
			if !testArtifact.TestPassed {
				log.Warn("Some health checks failed. Stopping test suite.")
				break
			}
		} else {
			log.Info("Skipping health checks")
		}
	}
	err = artifacts.SerializeTestArtifacts(testArtifacts)
	if err != nil {
		return err
	}

	return nil
}
