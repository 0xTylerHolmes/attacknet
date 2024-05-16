package kurtosis

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	log "github.com/sirupsen/logrus"
	"strings"
)

// Service handles the interaction with the enclave using the kurtosis engine API
type Service struct {
	// Kubernetes namespace
	Namespace string

	config          *Config
	kurtosisContext *kurtosis_context.KurtosisContext
	enclaveContext  *enclaves.EnclaveContext
	//reuseDevnetBetweenRuns bool
}

// isTargetEnclaveRunning checks if the desired enclave we want to test is already running in kurtosis
func (s *Service) isTargetEnclaveRunning(ctx context.Context, targetEnclaveName string) (bool, error) {
	runningEnclaves, err := s.kurtosisContext.GetEnclaves(ctx)
	if err != nil {
		return false, err
	}

	for enclaveName := range runningEnclaves.GetEnclavesByName() {
		if enclaveName == targetEnclaveName {
			return true, nil
		}
	}
	return false, nil
}

// Destroy tears down the enclave we are using
func (e *Service) Destroy(ctx context.Context) error {
	return e.kurtosisContext.DestroyEnclave(ctx, e.enclaveContext.GetEnclaveName())
}

//
//// pass-thru func. Figure out how to remove eventually.
//func (e *Service) RunStarlarkRemotePackageBlocking(
//	ctx context.Context,
//	packageId string,
//	cfg *starlark_run_config.StarlarkRunConfig,
//) (*enclaves.StarlarkRunResult, error) {
//	return e.enclaveContext.RunStarlarkRemotePackageBlocking(ctx, packageId, cfg)
//}
//
//// pass-thru func. Figure out how to remove eventually.
//func (e *Service) RunStarlarkRemotePackage(
//	ctx context.Context,
//	packageRootPath string,
//	runConfig *starlark_run_config.StarlarkRunConfig,
//) (chan *kurtosis_core_rpc_api_bindings.StarlarkRunResponseLine, context.CancelFunc, error) {
//	return e.enclaveContext.RunStarlarkRemotePackage(ctx, packageRootPath, runConfig)
//}

func getKurtosisContext() (*kurtosis_context.KurtosisContext, error) {
	kurtosisCtx, err := kurtosis_context.NewKurtosisContextFromLocalEngine()
	if err != nil {
		if strings.Contains(err.Error(), "connect: connection refused") {
			return nil, fmt.Errorf("could not connect to the Kurtosis engine. Be sure the engine is running using `kurtosis engine status` or `kurtosis engine start`. You might also need to start the gateway using `kurtosis gateway` - %w", err)
		} else {
			return nil, err
		}
	}
	return kurtosisCtx, nil
}

//func getEnclaveName(namespace string) string {
//	var enclaveName string
//	if namespace != "" {
//		enclaveName = namespace[3:]
//	} else {
//		enclaveName = fmt.Sprintf("attacknet-%d", time.Now().Unix())
//	}
//	return enclaveName
//}

func isErrorNoEnclaveFound(err error) bool {
	rootCause := stacktrace.RootCause(err)
	if strings.Contains(rootCause.Error(), "Couldn't find an enclave for identifier") {
		return true
	} else {
		return false
	}
}

// NewService creates a new kurtosis service to interact with
func NewService(ctx context.Context, config *Config) (*Service, error) {
	// get the kurtosis context
	kurtosisContext, err := getKurtosisContext()
	if err != nil {
		return nil, err
	}

	var enclaveContext *enclaves.EnclaveContext

	// first check for existing enclave
	enclaveContext, err = kurtosisContext.GetEnclaveContext(ctx, config.EnclaveName)
	// either there was an error or the enclave does not exist
	if err != nil {
		// check if no enclave found,
		if !isErrorNoEnclaveFound(err) {
			// unexpected error
			return nil, err
		}

		log.Infof("No existing kurtosis enclave by the name of %s was found. Creating a new one.", config.EnclaveName)
		enclaveContext, err = kurtosisContext.CreateProductionEnclave(ctx, config.EnclaveName)
		if err != nil {
			return nil, err
		}
		return &Service{
			Namespace:       config.EnclaveNamespace,
			config:          config,
			kurtosisContext: kurtosisContext,
			enclaveContext:  enclaveContext,
		}, nil
	}

	log.Infof("An existing enclave was found with the name %s, but ReuseDevnetBetweenRuns is set to false. Todo: add tear-down logic here.", config.EnclaveName)
	return &Service{
		Namespace:       config.EnclaveNamespace,
		kurtosisContext: kurtosisContext,
		enclaveContext:  enclaveContext,
		config:          config,
	}, nil
}

// StartNetwork starts the ethereum-package devnet using the Service's config
func (s *Service) StartNetwork(ctx context.Context) error {
	log.Infof("------------ EXECUTING PACKAGE ---------------")
	cfg := &starlark_run_config.StarlarkRunConfig{
		SerializedParams: s.config.NetworkConfig.String(),
	}
	a, _, err := s.enclaveContext.RunStarlarkRemotePackage(ctx, s.config.KurtosisPackageID, cfg)
	if err != nil {
		return stacktrace.Propagate(err, "error running Starklark script")
	}

	// todo: clean this up when we decide to add log filtering
	progressIndex := 0
	for {
		t := <-a
		progress := t.GetProgressInfo()
		if progress != nil {
			progressMsgs := progress.CurrentStepInfo
			for i := progressIndex; i < len(progressMsgs); i++ {
				log.Infof("[Kurtosis] %s", progressMsgs[i])
			}
			progressIndex = len(progressMsgs)
		}

		info := t.GetInfo()
		if info != nil {
			log.Infof("[Kurtosis] %s", info.InfoMessage)
		}

		warn := t.GetWarning()
		if warn != nil {
			log.Warnf("[Kurtosis] %s", warn.WarningMessage)
		}

		e := t.GetError()
		if e != nil {
			log.Errorf("[Kurtosis] %s", e.String())
			return stacktrace.Propagate(errors.New("kurtosis deployment failed during execution"), "%s", e.String())
		}

		insRes := t.GetInstructionResult()
		if insRes != nil {
			log.Infof("[Kurtosis] %s", insRes.SerializedInstructionResult)
		}

		finishRes := t.GetRunFinishedEvent()
		if finishRes != nil {
			log.Infof("[Kurtosis] %s", finishRes.GetSerializedOutput())
			if finishRes.IsRunSuccessful {
				log.Info("[Kurtosis] Devnet genesis successful. Passing back to Attacknet")
				return nil
			} else {
				log.Error("[Kurtosis] There was an error during genesis.")
				return stacktrace.Propagate(errors.New("kurtosis deployment failed"), "%s", finishRes.GetSerializedOutput())
			}
		}
	}
}
