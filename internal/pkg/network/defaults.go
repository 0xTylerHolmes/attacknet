package network

import (
	"errors"
	"fmt"
)

// TODO validator sidecars
const (
	DefaultBeaconLighthouseImage = "sigp/lighthouse:latest"
	DefaultBeaconTekuImage       = "consensys/teku:latest"
	DefaultBeaconNimbusImage     = "statusim/nimbus-eth2:multiarch-latest"
	DefaultBeaconPrysmImage      = "gcr.io/prysmaticlabs/prysm/beacon-chain:latest"
	DefaultBeaconLodestarImage   = "chainsafe/lodestar:latest"
	DefaultBeaconGrandineImage   = "ethpandaops/grandine:master"

	DefaultExecutionGethImage       = "ethereum/client-go:latest"
	DefaultExecutionErigonImage     = "ethpandaops/erigon:devel"
	DefaultExecutionNethermindImage = "nethermindeth/nethermind:master"
	DefaultExecutionBesuImage       = "hyperledger/besu:latest"
	DefaultExecutionRethImage       = "ghcr.io/paradigmxyz/reth"
	DefaultExecutionEthereumjsImage = "ethpandaops/ethereumjs:master"
	DefaultExecutionNimbusImage     = "ethpandaops/nimbus-eth1:master"

	//TODO figure out which do not have sidecar support
	DefaultValidatorLighthouseImage = "sigp/lighthouse:latest"
	DefaultValidatorLodestarImage   = "chainsafe/lodestar:latest"
	DefaultValidatorNimbusImage     = "statusim/nimbus-validator-client:multiarch-latest"
	DefaultValidatorPrysmImage      = "gcr.io/prysmaticlabs/prysm/validator:latest"
	DefaultValidatorTekuImage       = "consensys/teku:latest"
	DefaultValidatorGrandineImage   = "ethpandaops/grandine:master"

	//TODO
	DefaultGethMaxMem       = 1024 // # 1GB
	DefaultGethMaxCPU       = 1000 //  # 1 core
	DefaultErigonMaxMem     = 1024 //  # 1GB
	DefaultErigonMaxCPU     = 1000 //  # 1 core
	DefaultNethermindMaxMem = 1024 //  # 1GB
	DefaultNethermindMaxCPU = 1000 //  # 1 core
	DefaultBesuMaxMem       = 1024 //  # 1GB
	DefaultBesuMaxCPU       = 1000 //  # 1 core
	DefaultRethMaxMem       = 1024 //  # 1GB
	DefaultRethMaxCPU       = 1000 //  # 1 core
	DefaultEthereumjsMaxMem = 1024 //  # 1GB
	DefaultEthereumjsMaxCPU = 1000 //  # 1 core
	DefaultNimbusEth1MaxMem = 1024 //  # 1GB
	DefaultNimbusEth1MaxCPU = 1000 //  # 1 core
	DefaultPrysmMaxMem      = 1024 //  # 1GB
	DefaultPrysmMaxCPU      = 1000 //  # 1 core
	DefaultLighthouseMaxMem = 1024 //  # 1GB
	DefaultLighthouseMaxCPU = 1000 //  # 1 core
	DefaultTekuMaxMem       = 2048 //  # 2GB
	DefaultTekuMaxCPU       = 1000 //  # 1 core
	DefaultNimbusMaxMem     = 1024 //  # 1GB
	DefaultNimbusMaxCPU     = 1000 //  # 1 core
	DefaultLodestarMaxMem   = 2048 //  # 2GB
	DefaultLodestarMaxCPU   = 1000 //  # 1 core
	DefaultGrandineMaxMem   = 2048 //  # 2GB
	DefaultGrandineMaxCPU   = 1000 //  # 1 core
)

// GetDefaultBeaconImage get the default kurtosis image for the specifed consensus beacon client
func GetDefaultBeaconImage(beaconType string) (string, error) {
	if beaconType == "prysm" {
		return DefaultBeaconPrysmImage, nil
	}
	if beaconType == "lighthouse" {
		return DefaultBeaconLighthouseImage, nil
	}
	if beaconType == "teku" {
		return DefaultBeaconTekuImage, nil
	}
	if beaconType == "nimbus" {
		return DefaultBeaconNimbusImage, nil
	}
	if beaconType == "lodestar" {
		return DefaultBeaconLodestarImage, nil
	}
	if beaconType == "grandine" {
		return DefaultBeaconGrandineImage, nil
	}
	return "", errors.New(fmt.Sprintf("no default image for the specified beacon type: %s", beaconType))
}

// GetDefaultValidatorImage get the default kurtosis image for the specifed consensus beacon client
func GetDefaultValidatorImage(validatorType string) (string, error) {
	if validatorType == "prysm" {
		return DefaultValidatorPrysmImage, nil
	}
	if validatorType == "lighthouse" {
		return DefaultValidatorLighthouseImage, nil
	}
	if validatorType == "teku" {
		return DefaultValidatorTekuImage, nil
	}
	if validatorType == "nimbus" {
		return DefaultValidatorNimbusImage, nil
	}
	if validatorType == "lodestar" {
		return DefaultValidatorLodestarImage, nil
	}
	if validatorType == "grandine" {
		return DefaultValidatorGrandineImage, nil
	}
	return "", errors.New(fmt.Sprintf("no default image for the specified validator type: %s", validatorType))
}

// GetDefaultExecutionImage get the default kurtosis image for the specifed execution client
func GetDefaultExecutionImage(executionType string) (string, error) {
	if executionType == "geth" {
		return DefaultExecutionGethImage, nil
	}
	if executionType == "erigon" {
		return DefaultExecutionErigonImage, nil
	}
	if executionType == "nethermind" {
		return DefaultExecutionNethermindImage, nil
	}
	if executionType == "besu" {
		return DefaultExecutionBesuImage, nil
	}
	if executionType == "reth" {
		return DefaultExecutionRethImage, nil
	}
	if executionType == "nimbus" {
		return DefaultExecutionNimbusImage, nil
	}
	if executionType == "ethereumjs" {
		return DefaultExecutionEthereumjsImage, nil
	}
	return "", errors.New(fmt.Sprintf("no default image for the specified execution type: %s", executionType))

}
