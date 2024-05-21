package kurtosis

import (
	"attacknet/cmd/internal/pkg/network"
	"context"
	"errors"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	log "github.com/sirupsen/logrus"
)

type Service struct {
	enclaveName string `yaml:"enclave_name"`
	//enclaveNamespace  string `yaml:"enclave_namespace"` //TODO is this necessary
	kurtosisPackageID string `yaml:"kurtosis_package_id"`

	config          *Config
	configTopology  *network.Topology
	kurtosisContext *kurtosis_context.KurtosisContext
	enclaveContext  *enclaves.EnclaveContext

	devnetRunning bool // whether the devnet is running in the enclave
}

// NewService creates a new kurtosis service. If the target enclave already exists we attach to it, if not we create the enclave.
func NewService(ctx context.Context, config *Config, kurtosisPackageID string, targetEnclaveName string) (*Service, error) {
	log.Infof("Creating a new kurtosis service.")
	kurtosisContext, err := getKurtosisContext()
	if err != nil {
		return nil, err
	}

	service := &Service{
		enclaveName:       targetEnclaveName,
		kurtosisPackageID: kurtosisPackageID,
		config:            config,
		kurtosisContext:   kurtosisContext,
		enclaveContext:    nil,
	}

	configTopology, err := ComposeTopologyFromConfig(config)
	if err != nil {
		return nil, err
	}
	service.configTopology = configTopology

	// check if the target enclave exists
	log.Debugf("checking if an enclave with the target name: %s exists.", targetEnclaveName)
	enclaveExists, err := doesEnclaveExist(ctx, kurtosisContext, targetEnclaveName)
	if err != nil {
		// unrecoverable
		return nil, err
	}

	if enclaveExists {
		log.Infof("target enclave does exist. Checking to see if the devnet is running.")
		err = service.AttachToRunningContext(ctx)
		if err != nil {
			return nil, err
		}

		return service, nil
	}

	// enclave does not exist, create it but don't start it.
	//enclaveContext, err := createEnclave(ctx, service.kurtosisContext, targetEnclaveName)
	log.Infof("target enclave: %s does not exist, creating it.", targetEnclaveName)
	err = service.CreateEnclave(ctx)
	if err != nil {
		return nil, err
	}
	service.devnetRunning = false
	log.Infof("target enclave: %s created.", targetEnclaveName)

	return service, nil
}

// DoesTargetEnclaveExist checks whether the target enclave for this service exists.
func (s *Service) DoesTargetEnclaveExist(ctx context.Context) (bool, error) {
	runningEnclaves, err := s.kurtosisContext.GetEnclaves(ctx)
	if err != nil {
		return false, err
	}

	for enclaveName := range runningEnclaves.GetEnclavesByName() {
		if enclaveName == s.enclaveName {
			return true, nil
		}
	}
	return false, nil
}

// Destroy destroys the target enclave
func (e *Service) Destroy(ctx context.Context) error {
	return destroyEnclave(ctx, e.kurtosisContext, e.enclaveName)
}

//// IsDevnetRunning checks to see if there are running services within the target enclave
//func (s *Service) IsDevnetRunning(ctx context.Context) (bool, error) {
//	services, err := s.enclaveContext.GetServices()
//	if err != nil {
//		return false, err
//	}
//	return len(services) > 0, nil
//}

func (s *Service) CreateEnclave(ctx context.Context) error {
	enclaveContext, err := s.kurtosisContext.CreateProductionEnclave(ctx, s.enclaveName)
	if err != nil {
		return err
	}
	//update the enclave context
	s.enclaveContext = enclaveContext
	return nil
}

// AttachToRunningContext attaches to an already running context, returns an error if it doesn't exist or if there is an internal error
func (s *Service) AttachToRunningContext(ctx context.Context) error {
	enclaveContext, err := s.kurtosisContext.GetEnclaveContext(ctx, s.enclaveName)
	if err != nil {
		return err
	}

	devnetRunning, err := isDevnetRunning(ctx, enclaveContext)
	if err != nil {
		return err
	}
	if !devnetRunning {
		log.Infof("The devnet we are attaching to has not started. Starting it now.")
		err = startDevnet(ctx, enclaveContext, s.kurtosisPackageID, s.config)
		s.enclaveContext = enclaveContext
		return nil
	}

	// the devnet in the enclave is already running, is it the expected devnet for the provided configuration file?
	isExpectedDevnet, err := isExpectedDevnetRunning(ctx, enclaveContext, s.config)
	if isExpectedDevnet {
		s.enclaveContext = enclaveContext
		return nil
	}

	return errors.New("the running devnet does not match the specified configuration file")
}

func (s *Service) StartDevnet(ctx context.Context) error {
	return startDevnet(ctx, s.enclaveContext, s.kurtosisPackageID, s.config)
}
