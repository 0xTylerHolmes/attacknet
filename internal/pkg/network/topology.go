package network

import (
	"fmt"
)

type Topology struct {
	Nodes []*Node
}

type Node struct {
	Index          int // due to the way the ethereum-package assigns indexes we use strings. The index could be 9 or 09 depending on size
	Execution      *ExecutionClient
	Consensus      *ConsensusClient
	ConsensusVotes int
}

func (n *Node) ToString() string {
	return fmt.Sprintf("#%d %s/%s", n.Index, n.Execution.Type, n.Consensus.Type)
}

// IsEqual only compares the names due to issues with getting more detailed information from running kurtosis enclave
func (t *Topology) IsEqual(t2 *Topology) bool {
	var foundMatch bool = false
	if len(t.Nodes) != len(t2.Nodes) {
		return false
	}
	for _, node := range t.Nodes {
		foundMatch = false
		for _, node2 := range t2.Nodes {

			// hacky workaround until we can get startlark run config from running enclave
			if node.ToString() == node2.ToString() {
				foundMatch = true
				continue
			}
		}
		if !foundMatch {
			return false
		}
	}
	return true
}
