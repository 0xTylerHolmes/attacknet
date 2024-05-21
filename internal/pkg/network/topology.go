package network

import "fmt"

type Topology struct {
	Nodes []*Node
}

type ValidatorClient struct {
	Type  string
	Image string

	//TODO do we really need this for basic topology information
	//ExtraLabels map[string]string
	//CpuRequired int
	//MemoryRequired int
}

type ConsensusClient struct {
	Type                string
	Image               string
	HasValidatorSidecar bool
	ValidatorType       string
	ValidatorImage      string
	//TODO do we really need this for basic topology information
	//ValidatorExtraLabels  map[string]string
	//ExtraLabels           map[string]string
	//CpuRequired           int
	//MemoryRequired        int
	//SidecarCpuRequired    int
	//SidecarMemoryRequired int
}

type ExecutionClient struct {
	Type  string
	Image string
	//TODO do we really need this for basic topology information
	//ExtraLabels    map[string]string
	//CpuRequired    int
	//MemoryRequired int
}

type Node struct {
	Index          int
	Execution      *ExecutionClient
	Consensus      *ConsensusClient
	ConsensusVotes int
}

func (n *Node) ToString() string {
	return fmt.Sprintf("#%d %s/%s", n.Index, n.Execution.Type, n.Consensus.Type)
}

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
