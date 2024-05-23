package plan

import (
	"attacknet/cmd/internal/pkg/chaos"
	"attacknet/cmd/internal/pkg/kurtosis"
	planNetwork "attacknet/cmd/pkg/plan/network"
	"attacknet/cmd/pkg/plan/suite"
)

func BuildPlan(config *PlannerConfig) (*chaos.Config, *kurtosis.Config, error) {

	nodes, err := planNetwork.ComposeNetworkTopology(
		config.Topology,
		config.FaultConfig.TargetClient,
		config.ExecutionClients,
		config.ConsensusClients,
	)
	if err != nil {
		return nil, nil, err
	}

	isExecTarget := config.IsTargetExecutionClient()
	// exclude the bootnode from test targeting
	potentialNodesUnderTest := nodes[1:]
	tests, err := suite.ComposeTestSuite(config.FaultConfig, isExecTarget, potentialNodesUnderTest)
	if err != nil {
		return nil, nil, err
	}

	chaosConfig := &chaos.Config{
		Tests: tests,
	}

	networkConfig, err := SerializeNetworkTopology(nodes, &config.GenesisParams)
	if err != nil {
		return nil, nil, err
	}
	return chaosConfig, networkConfig, nil
}
