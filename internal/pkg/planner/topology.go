package planner

import (
	"attacknet/cmd/internal/pkg/network"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
)

// ComposeNetworkTopology creates an entire network topology and stores it in the builder object.
func (b *Builder) ComposeNetworkTopology() error {
	if len(b.nodes) > 0 {
		return errors.New(fmt.Sprintf("calling ComposetNetworkTopology() on a builder with > 0 nodes is undefined, already have %d nodes", len(b.nodes)))
	}

	log.Debugf("creating base test with all pairings against the target client: %s", b.config.FaultConfig.TargetClient)
	if b.config.isExecutionClientTest() {
		b.populateBuilderWithAllConsensusClientNodePairings(b.targetExecutionClients)
	}

	if b.config.isConsensusClientTest() {
		b.populateBuilderWithAllExecutionClientPairints(b.targetConsensusClients)
	}
	// TODO: random tests go here

	// add extra nodes if needed.
	if len(b.nodes) < b.targetNetworkSize {
		if b.config.isConsensusClientTest() {
			// random nodes should exclude the client
			return b.populateBuilderWithRandomNodes(b.targetNetworkSize-len(b.nodes), []network.ExecutionClientType{}, []network.ConsensusClientType{network.ConsensusClientType(b.config.FaultConfig.TargetClient)})
		}
		if b.config.isExecutionClientTest() {
			return b.populateBuilderWithRandomNodes(b.targetNetworkSize-len(b.nodes), []network.ExecutionClientType{network.ExecutionClientType(b.config.FaultConfig.TargetClient)}, []network.ConsensusClientType{})
		}
		// TODO: random tests go here
	}

	return nil
}

// ConsensusClientFromVersion given a ConsensusClientVersion create a full populated ConsensusClient
func ConsensusClientFromVersion(version *ConsensusClientVersion) (*network.ConsensusClient, error) {
	consensusClient := &network.ConsensusClient{
		Type:        version.Type,
		Image:       version.BeaconImage,
		ExtraLabels: nil,
	}

	if version.HasSidecar != nil && *version.HasSidecar {
		consensusClient.HasValidatorSidecar = true
		consensusClient.ValidatorClient = &network.ValidatorClient{
			Type:           version.ValidatorType,
			Image:          version.ValidatorImage,
			ExtraLabels:    nil,
			CpuRequired:    nil,
			MemoryRequired: nil,
		}
	}

	err := network.UpdateConsensusClientWithDefaults(consensusClient)

	return consensusClient, err
}

// ExecutionClientFromVersion given an ExecutionClientVersion create a fully populated ExecutionClient
func ExecutionClientFromVersion(version *ExecutionClientVersion) (*network.ExecutionClient, error) {
	executionClient := &network.ExecutionClient{
		Type:           version.Type,
		Image:          version.Image,
		ExtraLabels:    nil,
		CpuRequired:    nil,
		MemoryRequired: nil,
	}

	err := network.UpdateExecutionClientWithDefaults(executionClient)
	return executionClient, err
}
