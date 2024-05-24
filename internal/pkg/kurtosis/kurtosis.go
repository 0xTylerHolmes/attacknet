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
	"regexp"
	"strings"
)

// wrappers around basic kurtosis functionality to be used by tests and the kurtosis Service

func destroyEnclave(ctx context.Context, kurtosisContext *kurtosis_context.KurtosisContext, targetEnclaveName string) error {
	log.Tracef("destroyEnclave: destroying the target enclave: %s", targetEnclaveName)
	return kurtosisContext.DestroyEnclave(ctx, targetEnclaveName)
}

// createEnclave create a production enclave with the provided target name. Returns the enclave context
func createEnclave(ctx context.Context, kurtosisContext *kurtosis_context.KurtosisContext, targetEnclaveName string) (*enclaves.EnclaveContext, error) {
	log.Tracef("createEnclave: creating target enclave: %s", targetEnclaveName)
	return kurtosisContext.CreateProductionEnclave(ctx, targetEnclaveName)
}

func GetEnclaveContext(ctx context.Context, kurtosisContext *kurtosis_context.KurtosisContext, targetEnclaveName string) (*enclaves.EnclaveContext, error) {
	log.Tracef("GetEnclaveContext: getting context for target enclave: %s", targetEnclaveName)
	return kurtosisContext.GetEnclaveContext(ctx, targetEnclaveName)
}

// GetKurtosisContext fetches the context from the local kurtosis engine
func GetKurtosisContext() (*kurtosis_context.KurtosisContext, error) {
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

// from a running enclave extract all the service IDs that correspond to nodes
func getViableNodeServiceIDs(ctx context.Context, enclaveContext *enclaves.EnclaveContext) ([]string, error) {
	regexExpression := "(el|cl|vc)-\\d+-\\w+-\\w+"
	var matchingServices []string

	services, err := enclaveContext.GetServices()
	if err != nil {
		return nil, err
	}
	for name, _ := range services {
		match, err := regexp.MatchString(regexExpression, string(name))
		if err != nil {
			return nil, err
		}
		if match {
			matchingServices = append(matchingServices, string(name))
		}
	}
	return matchingServices, nil
}

// doesEnclaveExist check if the target enclave exists, error is not recoverable
func doesEnclaveExist(ctx context.Context, kurtosisContext *kurtosis_context.KurtosisContext, targetEnclaveName string) (bool, error) {
	log.Tracef("doesEnclaveExist: searching for target enclave: %s", targetEnclaveName)
	runningEnclaves, err := kurtosisContext.GetEnclaves(ctx)
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

// starts the devnet
func startDevnet(ctx context.Context, enclaveContext *enclaves.EnclaveContext, kurtosisPackageID string, kurtosisConfig *Config) error {
	devnetRunning, err := hasEnclaveStarted(enclaveContext)
	if err != nil {
		return err
	}
	if devnetRunning {
		return errors.New(fmt.Sprintf("can't create devnet in enclave: %s, it has already started.", enclaveContext.GetEnclaveName()))
	}
	log.Infof("------------ EXECUTING PACKAGE ---------------")
	cfg := &starlark_run_config.StarlarkRunConfig{
		SerializedParams: kurtosisConfig.String(),
	}
	a, _, err := enclaveContext.RunStarlarkRemotePackage(ctx, kurtosisPackageID, cfg)
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

// hasEnclaveStarted checks if there are running services within the enclave
func hasEnclaveStarted(enclaveContext *enclaves.EnclaveContext) (bool, error) {
	services, err := enclaveContext.GetServices()
	if err != nil {
		return false, err
	}
	return len(services) > 0, nil
}

// isExpectedDevnetRunning checks if the devnet specified by service config is running in the target enclave
func isExpectedDevnetRunning(ctx context.Context, config *Config, enclaveContext *enclaves.EnclaveContext) (bool, error) {
	configTopology, err := TopologyFromConfig(config)
	if err != nil {
		return false, err
	}
	runningTopology, err := TopologyFromRunningEnclave(ctx, enclaveContext)
	if err != nil {
		return false, err
	}
	if err != nil {
		return false, err
	}
	// return whether the running enclave is the expected enclave
	return configTopology.IsEqual(runningTopology), nil
}
