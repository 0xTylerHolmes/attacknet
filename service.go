package attacknet

import (
	"attacknet/cmd/internal/pkg/chaos"
	"attacknet/cmd/internal/pkg/kurtosis"
	"attacknet/cmd/pkg/artifacts"
	"attacknet/cmd/pkg/health"
	"attacknet/cmd/pkg/test_executor"
	"context"
	log "github.com/sirupsen/logrus"
	"time"
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

	return &Service{
		Config:          config,
		KurtosisService: kurtosisService,
		ChaosService:    chaosService,
	}, nil
}

// SetupEnclaveForExperiment handles all the logic for preparing an enclave for the provided experiment.
func (s *Service) SetupEnclaveForExperiment(ctx context.Context) error {
	// check if the enclave exists, if not go ahead and create one.
	enclaveExists, err := s.KurtosisService.TargetEnclaveExists(ctx, s.Config.KurtosisConfig.EnclaveName)
	if err != nil {
		return err
	}
	if !enclaveExists {
		log.Infof("no enclave exists for the specified experiment, creating a new enclave and launching the devnet")
		//create enclave and start it
		err = s.KurtosisService.CreateEnclave(ctx)
		if err != nil {
			return err
		}
		return s.KurtosisService.StartNetwork(ctx)
	}
	// if the enclave exists check if we need to restart it, if not just attach
	if s.Config.ChaosConfig.StartNewDevnet {
		log.Infof("Enclave: %s already exists, experiment calls for a fresh devnet so restarting the devnet", s.Config.KurtosisConfig.EnclaveName)
		return s.KurtosisService.RestartDevnet(ctx)
	}
	// we are just attaching to the devnet make sure the enclave experiment is already running
	devnetRunning, err := s.KurtosisService.IsDevnetRunning(ctx)
	if err != nil {
		return err
	}

	if !devnetRunning {
		log.Warnf("The denvet we are trying to attach to does not appear to have started. Starting the devnet")
		return s.KurtosisService.StartNetwork(ctx)
	}

	log.Infof("Detected a running devnet. We are attached and ready to run an experiment")
	return nil
}

func (s *Service) StartTestSuite(ctx context.Context) error {
	err := s.SetupEnclaveForExperiment(ctx)
	if err != nil {
		return err
	}
	// standby for timer
	log.Infof(
		"Waiting %d seconds before starting fault injection",
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

	//TODO: do we really need to destroy the enclave?
	//return s.KurtosisService.DestroyEnclave(ctx)
	return nil
}
