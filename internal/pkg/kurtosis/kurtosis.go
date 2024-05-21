package kurtosis

import (
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/starlark_run_config"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/kurtosis-tech/stacktrace"
	"github.com/sirupsen/logrus"
	"regexp"
	"strings"
)

func destroyEnclave(ctx context.Context, kurtosisContext *kurtosis_context.KurtosisContext, targetEnclaveName string) error {
	return kurtosisContext.DestroyEnclave(ctx, targetEnclaveName)
}

// createEnclave create a production enclave with the provided target name. Returns the enclave context
func createEnclave(ctx context.Context, kurtosisContext *kurtosis_context.KurtosisContext, targetEnclaveName string) (*enclaves.EnclaveContext, error) {
	return kurtosisContext.CreateProductionEnclave(ctx, targetEnclaveName)
}

// getKurtosisContext fetches the context from the local kurtosis engine
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

// from a running enclave extract all the service IDs that correspond to nodes
func getViableNodeServiceIDs(ctx context.Context, enclaveContext *enclaves.EnclaveContext) ([]string, error) {
	regexExpression := "(el|cl|vc)-\\d-\\w+-\\w+"
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
	logrus.Infof("------------ EXECUTING PACKAGE ---------------")
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
				logrus.Infof("[Kurtosis] %s", progressMsgs[i])
			}
			progressIndex = len(progressMsgs)
		}

		info := t.GetInfo()
		if info != nil {
			logrus.Infof("[Kurtosis] %s", info.InfoMessage)
		}

		warn := t.GetWarning()
		if warn != nil {
			logrus.Warnf("[Kurtosis] %s", warn.WarningMessage)
		}

		e := t.GetError()
		if e != nil {
			logrus.Errorf("[Kurtosis] %s", e.String())
			return stacktrace.Propagate(errors.New("kurtosis deployment failed during execution"), "%s", e.String())
		}

		insRes := t.GetInstructionResult()
		if insRes != nil {
			logrus.Infof("[Kurtosis] %s", insRes.SerializedInstructionResult)
		}

		finishRes := t.GetRunFinishedEvent()
		if finishRes != nil {
			logrus.Infof("[Kurtosis] %s", finishRes.GetSerializedOutput())
			if finishRes.IsRunSuccessful {
				logrus.Info("[Kurtosis] Devnet genesis successful. Passing back to Attacknet")
				return nil
			} else {
				logrus.Error("[Kurtosis] There was an error during genesis.")
				return stacktrace.Propagate(errors.New("kurtosis deployment failed"), "%s", finishRes.GetSerializedOutput())
			}
		}
	}
}
