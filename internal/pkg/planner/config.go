package planner

import (
	"attacknet/cmd/internal/pkg/kurtosis"
	"attacknet/cmd/internal/pkg/network"
	"attacknet/cmd/pkg/plan/suite"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/stacktrace"
	"math"
)

var (
	DefaultTargetNodeMultiplier = 1
)

type Config struct {
	ExecutionClients      []*ExecutionClientVersion `yaml:"execution"`
	ConsensusClients      []*ConsensusClientVersion `yaml:"consensus"`
	TargetNetworkTopology TargetNetworkTopology     `yaml:"target_network_topology"`
	GenesisParams         kurtosis.GenesisConfig    `yaml:"network_params"`
	//TODO do we really need this?
	KurtosisPackage     string                          `yaml:"kurtosis_package"`
	KubernetesNamespace string                          `yaml:"kubernetes_namespace"`
	FaultConfig         suite.PlannerFaultConfiguration `yaml:"fault_config"`
}

func (c *Config) isExecutionClientTest() bool {
	return network.ExecutionClientType(c.FaultConfig.TargetClient).IsValid()
}

func (c *Config) isConsensusClientTest() bool {
	return network.ConsensusClientType(c.FaultConfig.TargetClient).IsValid()
}

// returns the target node multiplier to use for generating network topologies
func (c *Config) getTargetNodeMultiplier() int {
	if c.TargetNetworkTopology.TargetNodeMultiplier != nil {
		return int(*c.TargetNetworkTopology.TargetNodeMultiplier)
	}
	return DefaultTargetNodeMultiplier
}

// TargetNetworkTopology can be used to define specific shapes of output kurtosis configurations.
type TargetNetworkTopology struct {
	// TargetNodeMultiplier if set this will enforce the target clients will represent at least this percent of the network.
	TargetsAsPercentOfNetwork *float32 `yaml:"target_as_percent_of_network"`
	// TargetNodeMultiplier if set we will include at least this many nodes containing the target client from the fault config
	TargetNodeMultiplier *uint `yaml:"target_node_multiplier"`
}

type ExecutionClientVersion struct {
	Type  network.ExecutionClientType `yaml:"type"`
	Image *string                     `yaml:"el_image"`
}

type ConsensusClientVersion struct {
	Type           network.ConsensusClientType  `yaml:"type"`
	BeaconImage    *string                      `yaml:"cl_image"`
	ValidatorType  *network.ConsensusClientType `yaml:"vc_type"`
	ValidatorImage *string                      `yaml:"vc_image"`
	HasSidecar     *bool                        `yaml:"has_sidecar"`
}

// checks the fault parameters to ensure they are supported
func validatePlannerFaultConfiguration(config *Config) error {
	// fault type
	_, ok := suite.FaultTypes[config.FaultConfig.FaultType]
	if !ok {
		return stacktrace.NewError("the fault type '%s' is not supported. Supported faults: %v", config.FaultConfig.FaultType, suite.FaultTypesList)
	}

	// targeting dimensions
	for _, spec := range config.FaultConfig.TargetingDimensions {
		_, ok := suite.TargetingSpecs[spec]
		if !ok {
			return stacktrace.NewError("the fault targeting dimension %s is not supported. Supported dimensions: %v", spec, suite.TargetingSpecList)
		}
	}

	// attack size dimensions
	for _, attackSize := range config.FaultConfig.AttackSizeDimensions {
		_, ok := suite.AttackSizes[attackSize]
		if !ok {
			return stacktrace.NewError("the attack size dimension %s is not supported. Supported dimensions: %v", attackSize, suite.AttackSizesList)
		}
	}

	// target client
	if !network.ExecutionClientType(config.FaultConfig.TargetClient).IsValid() && !network.ConsensusClientType(config.FaultConfig.TargetClient).IsValid() {
		return stacktrace.NewError("the target client is not a valid execution or consensus client type")
	}

	// check that the target client has a defined version
	if network.IsValidExecutionClientType(config.FaultConfig.TargetClient) {
		//ensure we have a definition
		_, err := config.getExecutionClientVersionForType(network.ExecutionClientType(config.FaultConfig.TargetClient))
		if err != nil {
			return stacktrace.NewError(err.Error())
		}
	} else {
		_, err := config.getConsensusClientVersionForType(network.ConsensusClientType(config.FaultConfig.TargetClient))
		if err != nil {
			return stacktrace.NewError(err.Error())
		}
	}
	return nil
}

// CalculateTargetNetworkSize calculates the network size based on the target multiplier and the target as percent of network parameters
// errors if the target network percent is below 0 or above 100.
func (c *Config) CalculateTargetNetworkSize() (int, error) {
	numTestNodes := 0
	if c.isConsensusClientTest() {
		numTestNodes = c.getTargetNodeMultiplier() * len(c.ConsensusClients)
	} else {
		numTestNodes = c.getTargetNodeMultiplier() * len(c.ExecutionClients)
	}

	if c.TargetNetworkTopology.TargetsAsPercentOfNetwork == nil {
		return numTestNodes, nil
	}
	// add nodes to conform to target percentages
	targetPercentage := *c.TargetNetworkTopology.TargetsAsPercentOfNetwork
	if targetPercentage > 1.0 || targetPercentage < 0 {
		return -1, errors.New(fmt.Sprintf("invalid value: (%02f) for targets_as_percent_of_network, must be >=0 and < 1", targetPercentage))
	}

	targetNodes := float32(numTestNodes) / targetPercentage

	return int(math.Ceil(float64(targetNodes))), nil
}

func (c *Config) getExecutionClientVersionForType(clientType network.ExecutionClientType) (*ExecutionClientVersion, error) {
	for _, elVersion := range c.ExecutionClients {
		if elVersion.Type == clientType {
			return elVersion, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("no version found for execution client type: %s", clientType))
}

func (c *Config) getConsensusClientVersionForType(clientType network.ConsensusClientType) (*ConsensusClientVersion, error) {
	for _, clVersion := range c.ConsensusClients {
		if clVersion.Type == clientType {
			return clVersion, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("no version found for consensus client type: %s", clientType))
}
