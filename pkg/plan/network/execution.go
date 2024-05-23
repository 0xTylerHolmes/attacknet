package network

import (
	"attacknet/cmd/internal/pkg/network"
	"github.com/kurtosis-tech/stacktrace"
)

const defaultElCpu = 768
const defaultElMem = 1024

func composeExecTesterNetwork(nodeMultiplier int, execClient string, consClientList []ClientVersion, execClientMap map[string]ClientVersion) ([]*network.Node, error) {
	// make sure execClient actually exists
	clientUnderTest, ok := execClientMap[execClient]
	if !ok {
		return nil, stacktrace.NewError("unknown execution client %s", execClient)
	}

	// start from 2 because bootnode is index 1
	index := 2
	nodes, err := composeNodesForElTesting(nodeMultiplier, index, clientUnderTest, consClientList)
	return nodes, err
}

func composeNodesForElTesting(nodeMultiplier, index int, execClient ClientVersion, consClientList []ClientVersion) ([]*network.Node, error) {
	var nodes []*network.Node

	for _, consensusClient := range consClientList {
		for i := 0; i < nodeMultiplier; i++ {
			node := buildNode(index, execClient, consensusClient)
			nodes = append(nodes, node)

			index += 1
		}
	}
	return nodes, nil
}

func composeExecutionClient(config ClientVersion) *network.ExecutionClient {
	return &network.ExecutionClient{
		Type:           config.Name,
		Image:          config.Image,
		ExtraLabels:    make(map[string]string),
		CpuRequired:    defaultElCpu,
		MemoryRequired: defaultElMem,
	}
}
