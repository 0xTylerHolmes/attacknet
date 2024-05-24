package planner

import (
	"attacknet/cmd/internal/pkg/chaos"
	"attacknet/cmd/internal/pkg/kurtosis"
	"attacknet/cmd/internal/pkg/network"
	"attacknet/cmd/pkg/plan/suite"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"math/rand"
)

type Builder struct {
	config *Config
	nodes  []*network.Node

	targetNetworkSize       int
	definedExecutionClients []*network.ExecutionClient
	definedConsensusClients []*network.ConsensusClient

	// TODO: these are here for future, currently only support one target client
	targetExecutionClients []*network.ExecutionClient
	targetConsensusClients []*network.ConsensusClient
}

// NewBuilder create a new planner Builder, validates the configuration before proceeding. The planner service is used
// to generate new experiments to run and can also be used to generate kurtosis network configs to create devnets to test.
func NewBuilder(config *Config) (*Builder, error) {
	var definedCls []*network.ConsensusClient
	var definedEls []*network.ExecutionClient
	var targetEls []*network.ExecutionClient
	var targetCLs []*network.ConsensusClient

	err := validatePlannerFaultConfiguration(config)
	if err != nil {
		return nil, err
	}

	for _, elVersion := range config.ExecutionClients {
		el, elErr := ExecutionClientFromVersion(elVersion)
		if elErr != nil {
			return nil, elErr
		}
		definedEls = append(definedEls, el)
	}

	for _, clVersion := range config.ConsensusClients {
		cl, clErr := ConsensusClientFromVersion(clVersion)
		if clErr != nil {
			return nil, clErr
		}
		definedCls = append(definedCls, cl)
	}

	if config.isConsensusClientTest() {
		clVersion, clErr := config.getConsensusClientVersionForType(network.ConsensusClientType(config.FaultConfig.TargetClient))
		if clErr != nil {
			return nil, clErr
		}
		cl, clErr := ConsensusClientFromVersion(clVersion)
		if clErr != nil {
			return nil, clErr
		}
		log.WithFields(log.Fields{
			"consensus-client": fmt.Sprintf("%+v", cl),
		}).Tracef("created target-cl")
		targetCLs = append(targetCLs, cl)
	}

	if config.isExecutionClientTest() {
		elVersion, elErr := config.getExecutionClientVersionForType(network.ExecutionClientType(config.FaultConfig.TargetClient))
		if elErr != nil {
			return nil, elErr
		}
		el, elErr := ExecutionClientFromVersion(elVersion)
		if elErr != nil {
			return nil, elErr
		}
		log.WithFields(log.Fields{
			"execution-client": fmt.Sprintf("%+v", el),
		}).Tracef("created target-el")
		targetEls = append(targetEls, el)
	}

	//calculate the target network size
	targetNetworkSize, err := config.CalculateTargetNetworkSize()
	if err != nil {
		return nil, err
	}
	log.Debugf("builder calculated the target network size to be: %d nodes", targetNetworkSize)

	service := &Builder{
		config:                  config,
		nodes:                   make([]*network.Node, 0),
		targetNetworkSize:       targetNetworkSize,
		definedConsensusClients: definedCls,
		definedExecutionClients: definedEls,
		targetExecutionClients:  targetEls,
		targetConsensusClients:  targetCLs,
	}

	return service, nil
}

// populates the builder with random nodes excluding the supplied el/cl types. errors if there aren't enough type definitions to perform the matchings
func (b *Builder) populateBuilderWithRandomNodes(numNodes int, excludedElTypes []network.ExecutionClientType, excludedClTypes []network.ConsensusClientType) error {
	var randomCl *network.ConsensusClient
	var randomEl *network.ExecutionClient

	if len(b.config.ExecutionClients) <= len(excludedElTypes) {
		return errors.New(fmt.Sprintf("unable to create a random node with the excluded execution clients: %s, not enough defined Els", excludedElTypes))
	}
	if len(b.config.ConsensusClients) <= len(excludedClTypes) {
		return errors.New(fmt.Sprintf("unable to create a random node with the excluded consensus clients: %s, not enough defined Cls", excludedClTypes))
	}

	for i := 0; i < numNodes; i++ {
	clOut:
		for {
			cl := b.definedConsensusClients[rand.Intn(len(b.definedConsensusClients))]
			for _, excludedType := range excludedClTypes {
				if cl.Type == excludedType {
					break clOut
				}
			}
			randomCl = cl
			break
		}
	elOut:
		for {
			el := b.definedExecutionClients[rand.Intn(len(b.definedExecutionClients))]
			for _, excludedType := range excludedElTypes {
				if el.Type == excludedType {
					break elOut
				}
			}
			randomEl = el
			break
		}
		// add the node
		node := &network.Node{
			Index:          len(b.nodes) + 1,
			Consensus:      randomCl,
			Execution:      randomEl,
			ConsensusVotes: b.config.GenesisParams.NumValKeysPerNode,
		}
		b.nodes = append(b.nodes, node)
	}
	return nil
}

// populates the builder with all node
func (b *Builder) populateBuilderWithAllConsensusClientNodePairings(executionClients []*network.ExecutionClient) {
	for _, executionClient := range executionClients {
		for _, cl := range b.definedConsensusClients {
			for i := 0; i < b.config.getTargetNodeMultiplier(); i++ {
				node := &network.Node{
					Index:          len(b.nodes) + 1,
					Execution:      executionClient,
					Consensus:      cl,
					ConsensusVotes: b.config.GenesisParams.NumValKeysPerNode,
				}
				b.nodes = append(b.nodes, node)
			}
		}
	}
}

func (b *Builder) populateBuilderWithAllExecutionClientPairints(consensusClients []*network.ConsensusClient) {
	for _, consensusClient := range consensusClients {
		// create base test
		for _, el := range b.definedExecutionClients {
			for i := 0; i < b.config.getTargetNodeMultiplier(); i++ {
				node := &network.Node{
					Index:          len(b.nodes) + 1,
					Execution:      el,
					Consensus:      consensusClient,
					ConsensusVotes: b.config.GenesisParams.NumValKeysPerNode,
				}
				b.nodes = append(b.nodes, node)
			}
		}
	}
}

func (s *Builder) BuildPlan() (*chaos.Config, *kurtosis.Config, error) {

	err := s.ComposeNetworkTopology()
	if err != nil {
		return nil, nil, err
	}

	topology := &network.Topology{Nodes: s.nodes}

	tests, err := suite.ComposeTestSuite(s.config.FaultConfig, s.config.isExecutionClientTest(), topology.Nodes)
	if err != nil {
		return nil, nil, err
	}

	chaosConfig := &chaos.Config{
		Tests: tests,
	}

	kurtosisConfig := &kurtosis.Config{
		Participants:        kurtosis.ParticipantsFromToplogy(topology),
		NetParams:           s.config.GenesisParams,
		AdditionalServices:  []string{"prometheus_grafana", "dora", "tx_spammer", "blob_spammer", "beacon_metrics_gazer", "el_forkmon"},
		ParallelKeystoreGen: false,
		Persistent:          false,
		DisablePeerScoring:  true,
	}

	return chaosConfig, kurtosisConfig, nil
}
