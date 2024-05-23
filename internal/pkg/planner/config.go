package planner

import (
	"attacknet/cmd/internal/pkg/kurtosis"
	"attacknet/cmd/pkg/plan/suite"
)

type Config struct {
	ExecutionClients      []ConsensusClientVersion `yaml:"execution"`
	ConsensusClients      []ConsensusClientVersion `yaml:"consensus"`
	TargetNetworkTopology TargetNetworkTopology    `yaml:"target_network_topology"`
	GenesisParams         kurtosis.GenesisConfig   `yaml:"network_params"`
	//TODO do we really need this?
	KurtosisPackage     string                          `yaml:"kurtosis_package"`
	KubernetesNamespace string                          `yaml:"kubernetes_namespace"`
	FaultConfig         suite.PlannerFaultConfiguration `yaml:"fault_config"`
}

// TargetNetworkTopology can be used to define specific shapes of output kurtosis configurations.
type TargetNetworkTopology struct {
	// ExecutionClientUnderTest if included we will prefer this client type when creating node pairings.
	ExecutionClientUnderTest *string `yaml:"el_under_test,omitempty"`
	// ConsensusClientUnderTest if included we will prefer this client type when creating node pairings.
	ConsensusClientUnderTest *string `yaml:"cl_under_test,omitempty"`
	// TargetNodeCount if set this will enforce the target clients will represent at least this percent of the network.
	TargetsAsPercentOfNetwork *float32 `yaml:"targets_as_percent_of_network,omitempty"`
	// TargetNodeCount if set we will include at least this many nodes containing the target client(s)
	TargetNodeCount *uint `yaml:"target_node_count,omitempty"`
	// MaxTotalNodeCount restricts the total amount of nodes in the network.
	MaxTotalNodeCount *uint `yaml:"max_total_node_count"`
}

type ExecutionClientVersion struct {
	Name  string `yaml:"name"`
	Image string `yaml:"el_image"`
}

type ConsensusClientVersion struct {
	Name           string  `yaml:"name"`
	BeaconImage    string  `yaml:"cl_image"`
	ValidatorImage *string `yaml:"vc_image,omitempty"`
	HasSidecar     *bool   `yaml:"has_sidecar,omitempty"`
}

func (c *Config) IsTargetExecutionClient() bool {
	for _, execClient := range c.ExecutionClients {
		if execClient.Name == c.FaultConfig.TargetClient {
			return true
		}
	}
	return false
}

func (c *Config) IsTargetConsensusClient() bool {
	for _, consClient := range c.ConsensusClients {
		if consClient.Name == c.FaultConfig.TargetClient {
			return true
		}
	}
	return false
}
