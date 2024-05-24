package network

import (
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"slices"
)

var (
	LighthouseClient ConsensusClientType = "lighthouse"
	TekuClient       ConsensusClientType = "teku"
	NimbusClient     ConsensusClientType = "nimbus"
	PrysmClient      ConsensusClientType = "prysm"
	LodestarClient   ConsensusClientType = "lodestar"
	GrandineClient   ConsensusClientType = "grandine"

	ValidConsensusClientTypes = []ConsensusClientType{
		LighthouseClient,
		TekuClient,
		NimbusClient,
		PrysmClient,
		LodestarClient,
		GrandineClient,
	}

	defaultBeaconImageMap = map[ConsensusClientType]string{
		LighthouseClient: "sigp/lighthouse:latest",
		TekuClient:       "consensys/teku:latest",
		NimbusClient:     "statusim/nimbus-eth2:multiarch-latest",
		PrysmClient:      "gcr.io/prysmaticlabs/prysm/beacon-chain:latest",
		LodestarClient:   "chainsafe/lodestar:latest",
		GrandineClient:   "ethpandaops/grandine:master",
	}

	defaultValidatorImageMap = map[ConsensusClientType]string{
		LighthouseClient: "sigp/lighthouse:latest",
		LodestarClient:   "chainsafe/lodestar:latest",
		NimbusClient:     "statusim/nimbus-validator-client:multiarch-latest",
		PrysmClient:      "gcr.io/prysmaticlabs/prysm/validator:latest",
		TekuClient:       "consensys/teku:latest",
		GrandineClient:   "ethpandaops/grandine:master",
	}

	DefaultBeaconMinCPU = map[ConsensusClientType]int{
		LighthouseClient: 1000,
		TekuClient:       1000,
		NimbusClient:     1000,
		PrysmClient:      1000,
		LodestarClient:   1000,
		GrandineClient:   1000,
	}

	DefaultBeaconMinMem = map[ConsensusClientType]int{
		LighthouseClient: 1024,
		TekuClient:       1024,
		NimbusClient:     1024,
		PrysmClient:      1024,
		LodestarClient:   1024,
		GrandineClient:   1024,
	}

	DefaultValidatorMinMem = map[ConsensusClientType]int{
		LighthouseClient: 1024,
		TekuClient:       1024,
		NimbusClient:     1024,
		PrysmClient:      1024,
		LodestarClient:   1024,
		GrandineClient:   1024,
	}

	DefaultValidatorMinCPU = map[ConsensusClientType]int{
		LighthouseClient: 1000,
		TekuClient:       1000,
		NimbusClient:     1000,
		PrysmClient:      1000,
		LodestarClient:   1000,
		GrandineClient:   1000,
	}

	// TODO map these out for better stress testing.

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

type ConsensusClientType string

type ConsensusClient struct {
	// Type is the only real required field
	Type           ConsensusClientType
	Image          *string
	ExtraLabels    map[string]string
	CpuRequired    *int
	MemoryRequired *int
	// HasValidatorSidecar should be optional but with edge cases on clients
	// who support it's easier to just assume we know
	HasValidatorSidecar bool
	ValidatorClient     *ValidatorClient
}

type ValidatorClient struct {
	// if nil we use the consensus client type
	Type  *ConsensusClientType
	Image *string

	ExtraLabels    map[string]string
	CpuRequired    *int
	MemoryRequired *int
}

func (c ConsensusClientType) IsValid() bool {
	return slices.Contains(ValidConsensusClientTypes, c)
}

// IsValidConsensusClientType returns whether the target string maps to a valid ConsensusClientType
func IsValidConsensusClientType(clientType string) bool {
	return ConsensusClientType(clientType).IsValid()
}

//// GetDefaultBeaconImage get the default kurtosis image for the specifed consensus beacon client
//func GetDefaultBeaconImage(beaconType ConsensusClientType) (string, error) {
//	value, ok := defaultBeaconImageMap[beaconType]
//	if !ok {
//		return "", errors.New(fmt.Sprintf("no default image for the specified beacon type: %s", beaconType))
//	}
//	return value, nil
//}
//
//// GetDefaultValidatorImage get the default kurtosis image for the specifed consensus beacon client
//func GetDefaultValidatorImage(validatorType ConsensusClientType) (string, error) {
//	value, ok := defaultValidatorImageMap[validatorType]
//	if !ok {
//		return "", errors.New(fmt.Sprintf("no default image for the specified validator type: %s", validatorType))
//	}
//	return value, nil
//}

func IsValidConsensusClient(client string) bool {
	_, ok := defaultBeaconImageMap[ConsensusClientType(client)]
	return ok
}

// getConsensusClientDefaultLabels creates some basic extra labels that can be used for more targeted fault generation
func getConsensusClientDefaultLabels(client ConsensusClientType) map[string]string {
	return map[string]string{
		"ethereum-package.service-type": "consensus-client",
		"ethereum-package.client-type":  string(client),
	}
}

// getValidatorClientDefaultLabels creates some basic extra labels that can be used for more targeted fault generation
func getValidatorClientDefaultLabels(client ConsensusClientType) map[string]string {
	return map[string]string{
		"ethereum-package.service-type": "validator-client",
		"ethereum-package.client-type":  string(client),
	}
}

// UpdateConsensusClientWithDefaults populates the consensus client object with all default information
func UpdateConsensusClientWithDefaults(client *ConsensusClient) error {
	err := handleDefaultBeaconFields(client)
	if err != nil {
		return err
	}
	err = handleDefaultValidatorFields(client)
	if err != nil {
		return err
	}
	return nil
}

func handleDefaultBeaconFields(client *ConsensusClient) error {
	if client.Image == nil {
		image, ok := defaultBeaconImageMap[client.Type]
		if !ok {
			return errors.New(fmt.Sprintf("no default beacon image found for client type: %s", client.Type))
		}
		client.Image = &image
	}

	if client.ExtraLabels == nil {
		client.ExtraLabels = make(map[string]string)
	}

	for key, value := range getConsensusClientDefaultLabels(client.Type) {
		_, ok := client.ExtraLabels[key]
		if !ok {
			client.ExtraLabels[key] = value
		}
	}
	// for now set minimum memory/CPU TODO address where we should handle this
	if client.CpuRequired == nil {
		minCPU, ok := DefaultBeaconMinCPU[client.Type]
		if !ok {
			return errors.New(fmt.Sprintf("unable to find default minimum CPU for beacon client type: %s", client.Type))
		}
		client.CpuRequired = &minCPU
	}

	if client.MemoryRequired == nil {
		minMem, ok := DefaultBeaconMinMem[client.Type]
		if !ok {
			return errors.New(fmt.Sprintf("unable to find default minimum memeory for beacon client type: %s", client.Type))
		}
		client.MemoryRequired = &minMem
	}

	return nil
}

func handleDefaultValidatorFields(client *ConsensusClient) error {
	// default is false
	if !client.HasValidatorSidecar {
		// panic if we have a sidecar, this is definetly an error
		if client.ValidatorClient != nil {
			return errors.New("got client HasValidatorSideCar = false, but ValidatorClient != nil")
		}
		log.Tracef("no validator sidecar specified for the client, skipping adding the default values")
		return nil
	}

	if client.ValidatorClient == nil {
		client.ValidatorClient = &ValidatorClient{
			Type:           &client.Type,
			Image:          nil,
			ExtraLabels:    nil,
			CpuRequired:    nil,
			MemoryRequired: nil,
		}
	}

	log.Tracef("populating validator default vaules for the client")
	if client.ValidatorClient.Type == nil {
		client.ValidatorClient.Type = &client.Type
	}
	if client.ValidatorClient.Image == nil {
		image, ok := defaultValidatorImageMap[*client.ValidatorClient.Type]
		if !ok {
			//should not occur
			return errors.New(fmt.Sprintf("unable to find default validator image for client type: %s", *client.ValidatorClient.Type))
		}
		client.ValidatorClient.Image = &image
	}
	if client.ValidatorClient.MemoryRequired == nil {
		mem, ok := DefaultValidatorMinMem[*client.ValidatorClient.Type]
		if !ok {
			//should not occur
			return errors.New(fmt.Sprintf("unable to find default validator min mem client type: %s", *client.ValidatorClient.Type))
		}
		client.ValidatorClient.MemoryRequired = &mem
	}

	if client.ValidatorClient.CpuRequired == nil {
		cpu, ok := DefaultValidatorMinCPU[*client.ValidatorClient.Type]
		if !ok {
			//should not occur
			return errors.New(fmt.Sprintf("unable to find default validator min CPU client type: %s", *client.ValidatorClient.Type))
		}
		client.ValidatorClient.CpuRequired = &cpu
	}

	if client.ValidatorClient.ExtraLabels == nil {
		client.ValidatorClient.ExtraLabels = make(map[string]string)
	}

	extraLabels := getValidatorClientDefaultLabels(client.Type)
	for key, value := range extraLabels {
		_, ok := client.ValidatorClient.ExtraLabels[key]
		if !ok {
			client.ValidatorClient.ExtraLabels[key] = value
		}
	}
	return nil
}
