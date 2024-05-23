package kurtosis

import (
	"attacknet/cmd/internal/pkg/network"
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	log "github.com/sirupsen/logrus"
	"time"
)

var (
	// TODO it would be nice if the kubernetes service got a signal and could kill it faster
	kubernetesOverheadDuration = time.Second * 60
)

type Service struct {
	enclaveName       string `yaml:"enclave_name"`
	kurtosisPackageID string `yaml:"kurtosis_package_id"`

	config          *Config
	configTopology  *network.Topology
	kurtosisContext *kurtosis_context.KurtosisContext
}

// NewService creates a new kurtosis service. This service stores all information about the enclave of interest and the ethereum-package topology.
func NewService(ctx context.Context, config *Config, kurtosisPackageID string, targetEnclaveName string) (*Service, error) {
	log.Infof("Creating a new kurtosis service.")
	kurtosisContext, err := GetKurtosisContext()
	if err != nil {
		return nil, err
	}

	configTopology, err := ComposeTopologyFromConfig(config)
	if err != nil {
		return nil, err
	}

	return &Service{
		enclaveName:       targetEnclaveName,
		kurtosisPackageID: kurtosisPackageID,
		config:            config,
		kurtosisContext:   kurtosisContext,
		configTopology:    configTopology,
	}, nil
}

func (s *Service) ForceCreateNewEnclave(ctx context.Context) error {
	enclaveExists, err := doesEnclaveExist(ctx, s.kurtosisContext, s.enclaveName)
	if err != nil {
		return err
	}

	if !enclaveExists {
		_, err = createEnclave(ctx, s.kurtosisContext, s.enclaveName)
		if err != nil {
			return err
		}
		return nil
	}

	err = destroyEnclave(ctx, s.kurtosisContext, s.enclaveName)
	if err != nil {
		return err
	}
	time.Sleep(kubernetesOverheadDuration)
	_, err = createEnclave(ctx, s.kurtosisContext, s.enclaveName)
	return err
}

// PrepareEnclave prepares the enclave for attacknet to use.
func (s *Service) PrepareEnclave(ctx context.Context, restartDevnet bool) error {
	// go ahead and tear down the enclave and start the devnet
	enclaveExists, err := doesEnclaveExist(ctx, s.kurtosisContext, s.enclaveName)
	if err != nil {
		return err
	}

	if restartDevnet {

		if !enclaveExists {
			return s.prepareNewEnclaveAndStartDevnet(ctx)
		}

		// enclave exists destroy and recreate
		err = destroyEnclave(ctx, s.kurtosisContext, s.enclaveName)
		if err != nil {
			return err
		}
		time.Sleep(kubernetesOverheadDuration)
		return s.prepareNewEnclaveAndStartDevnet(ctx)
	}

	if enclaveExists {
		return s.verifyRunningDevnet(ctx)
	}

	return s.prepareNewEnclaveAndStartDevnet(ctx)
}

// prepareNewEnclaveAndStartDevnet prepare a new enclave and start the devnet. errors on internal error and if the enclave already exists
func (s *Service) prepareNewEnclaveAndStartDevnet(ctx context.Context) error {
	exists, err := doesEnclaveExist(ctx, s.kurtosisContext, s.enclaveName)
	if err != nil {
		return err
	}

	if exists {
		return errors.New(fmt.Sprintf("cant create a new enclave; enclave %s already exists", s.enclaveName))
	}

	enclaveContext, err := createEnclave(ctx, s.kurtosisContext, s.enclaveName)
	if err != nil {
		return err
	}
	return startDevnet(ctx, enclaveContext, s.kurtosisPackageID, s.config)
}

// verifyRunningDevnet verifies that the running devnet is the one specified by the config file.
// errors if the topologies don't match or there is an internal error.
func (s *Service) verifyRunningDevnet(ctx context.Context) error {
	enclaveContext, err := GetEnclaveContext(ctx, s.kurtosisContext, s.enclaveName)
	// verify devnet running
	running, err := hasEnclaveStarted(enclaveContext)
	if err != nil {
		return err
	}
	if !running {
		return errors.New(fmt.Sprintf("can't attach to running devnet in enclave: %s, devnet was not running", s.enclaveName))
	}

	isExpectedDevnet, err := isExpectedDevnetRunning(ctx, s.config, enclaveContext)
	if err != nil {
		return err
	}
	if !isExpectedDevnet {
		return errors.New(fmt.Sprintf("the running devnet in enclave: %s has a different topology than the one specified in the config file", s.enclaveName))
	}

	return nil
}

// ForceRestartDevnet force restarts the devnet, if it doesn't exist it will be created, if it is running it will be restarted
func (s *Service) ForceRestartDevnet(ctx context.Context) error {
	// check to see if the enclave exists.
	log.Infof("Attempting to restart the devnet.")
	enclaveExists, err := doesEnclaveExist(ctx, s.kurtosisContext, s.enclaveName)
	if err != nil {
		return err
	}

	// create non-existing enclave
	if !enclaveExists {
		log.Infof("the devnet we are trying to start belongs to a enclave that hasn't been created. Creating it now.")
		return s.prepareNewEnclaveAndStartDevnet(ctx)
	}

	err = destroyEnclave(ctx, s.kurtosisContext, s.enclaveName)
	if err != nil {
		return err
	}
	time.Sleep(kubernetesOverheadDuration)
	return s.prepareNewEnclaveAndStartDevnet(ctx)
}
