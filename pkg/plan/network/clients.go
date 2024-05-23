package network

import (
	"attacknet/cmd/internal/pkg/network"
	"github.com/kurtosis-tech/stacktrace"
)

func buildNode(index int, execConf, consensusConf ClientVersion) *network.Node {
	return &network.Node{
		Index:     index,
		Execution: composeExecutionClient(execConf),
		Consensus: composeConsensusClient(consensusConf),
	}
}

func composeBootnode(bootEl, bootCl string, execClients, consensusClients map[string]ClientVersion) (*network.Node, error) {
	execConf, ok := execClients[bootEl]
	if !ok {
		return nil, stacktrace.NewError("unable to load configuration for exec client %s", bootEl)
	}
	consConf, ok := consensusClients[bootCl]
	if !ok {
		return nil, stacktrace.NewError("unable to load configuration for exec client %s", bootCl)
	}
	return buildNode(1, execConf, consConf), nil
}
