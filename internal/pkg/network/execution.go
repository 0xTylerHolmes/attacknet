package network

import (
	"errors"
	"fmt"
	"slices"
)

var (
	GethClient       ExecutionClientType = "geth"
	ErigonClient     ExecutionClientType = "erigon"
	NethermindClient ExecutionClientType = "nethermind"
	BesuClient       ExecutionClientType = "besu"
	RethClient       ExecutionClientType = "reth"
	EthereumjsClient ExecutionClientType = "ethereumjs"
	NimbusEth1Client ExecutionClientType = "nimbus-eth1"

	ValidExecutionClientTypes = []ExecutionClientType{
		GethClient,
		ErigonClient,
		NethermindClient,
		BesuClient,
		RethClient,
		EthereumjsClient,
		NimbusEth1Client,
	}

	defaultExecutionImageMap = map[ExecutionClientType]string{
		GethClient:       "ethereum/client-go:latest",
		ErigonClient:     "ethpandaops/erigon:devel",
		NethermindClient: "nethermindeth/nethermind:master",
		BesuClient:       "hyperledger/besu:latest",
		RethClient:       "ghcr.io/paradigmxyz/reth",
		EthereumjsClient: "ethpandaops/ethereumjs:master",
		NimbusEth1Client: "ethpandaops/nimbus-eth1:master",
	}

	defaultExecutionMinMem = map[ExecutionClientType]int{
		GethClient:       1024,
		ErigonClient:     1024,
		NethermindClient: 1024,
		BesuClient:       1024,
		RethClient:       1024,
		EthereumjsClient: 1024,
		NimbusEth1Client: 1024,
	}

	defaultExecutionMinCPU = map[ExecutionClientType]int{
		GethClient:       1000,
		ErigonClient:     1000,
		NethermindClient: 1000,
		BesuClient:       1000,
		RethClient:       1000,
		EthereumjsClient: 1000,
		NimbusEth1Client: 1000,
	}
)

type ExecutionClientType string

type ExecutionClient struct {
	Type  ExecutionClientType
	Image *string

	ExtraLabels    map[string]string
	CpuRequired    *int
	MemoryRequired *int
}

func (e ExecutionClientType) IsValid() bool {
	return slices.Contains(ValidExecutionClientTypes, e)
}

// IsValidExecutionClientType returns whether the target string maps to a valid ExecutionClientType
func IsValidExecutionClientType(clientType string) bool {
	return ExecutionClientType(clientType).IsValid()
}

func GetDefaultExecutionImage(clientType ExecutionClientType) (string, error) {
	image, ok := defaultExecutionImageMap[clientType]
	if !ok {
		return "", errors.New(fmt.Sprintf("no default image foun d for execution client type: %s", clientType))
	}
	return image, nil
}

// getExecutionClientDefaultLabels creates some basic extra labels that can be used for more targeted fault generation
func getExecutionClientDefaultLabels(client ExecutionClientType) map[string]string {
	return map[string]string{
		"ethereum-package.service-type": "execution-client",
		"ethereum-package.client-type":  string(client),
	}
}

// UpdateExecutionClientWithDefaults takes the partially filled execution client and updates the fields with the default values
func UpdateExecutionClientWithDefaults(client *ExecutionClient) error {
	if client.Image == nil {
		image, err := GetDefaultExecutionImage(client.Type)
		if err != nil {
			return err
		}
		client.Image = &image
	}

	extraLabels := getExecutionClientDefaultLabels(client.Type)
	if client.ExtraLabels == nil {
		client.ExtraLabels = make(map[string]string)
	}
	for key, value := range extraLabels {
		_, isSet := client.ExtraLabels[key]
		if !isSet {
			client.ExtraLabels[key] = value
		}
	}

	if client.MemoryRequired == nil {
		minMem, ok := defaultExecutionMinMem[client.Type]
		if !ok {
			return errors.New(fmt.Sprintf("no default minimum memory value found for execution client type: %s", client.Type))
		}
		client.MemoryRequired = &minMem
	}

	if client.CpuRequired == nil {
		minCpu, ok := defaultExecutionMinCPU[client.Type]
		if !ok {
			return errors.New(fmt.Sprintf("no default minimum cpu value found for exeuction client type: %s", client.Type))
		}
		client.CpuRequired = &minCpu
	}

	return nil
}
