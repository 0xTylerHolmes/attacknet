package kurtosis

import (
	"attacknet/cmd/internal/pkg/network"
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/enclaves"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"math"
	"strconv"
	"strings"
)

//TODO: we currently don't test the validators for each service, do we need to do this?

// TopologyFromConfig using a kurtosis ethereum-package configuration file build a network topology.
// This is the PREFERRED method of generating a topology.
func TopologyFromConfig(config *Config) (*network.Topology, error) {
	var topologyNodes []*network.Node
	currNode := 0
	for _, participant := range config.Participants {
		for nodeNum := 0; nodeNum < participant.Count; nodeNum++ {
			currNode += 1
			node, err := topologyNodeFromConfig(currNode, config.NetParams.NumValKeysPerNode, participant)
			if err != nil {
				return nil, err
			}
			topologyNodes = append(topologyNodes, node)
		}

	}
	return &network.Topology{Nodes: topologyNodes}, nil
}

func ParticipantsFromToplogy(topology *network.Topology) []*Participant {
	var participants []*Participant
	for _, node := range topology.Nodes {
		p := &Participant{
			ElClientType:  node.Execution.Type,
			ClClientType:  node.Consensus.Type,
			Count:         1,
			ElClientImage: node.Execution.Image,
			ELExtraLabels: node.Execution.ExtraLabels,
			ClClientImage: node.Consensus.Image,
			CLExtraLabels: node.Consensus.ExtraLabels,
			CLSeparateVC:  &node.Consensus.HasValidatorSidecar,
			// TODO
			ElMinCpu:     node.Execution.CpuRequired,
			ElMaxCpu:     nil,
			ElMinMemory:  node.Execution.MemoryRequired,
			ElMaxMemory:  nil,
			ClMinCpu:     node.Consensus.CpuRequired,
			ClMaxCpu:     nil,
			ClMinMemory:  node.Consensus.MemoryRequired,
			ClMaxMemory:  nil,
			ValMaxCpu:    nil,
			ValMaxMemory: nil,
		}

		if node.Consensus.ValidatorClient != nil && node.Consensus.HasValidatorSidecar {
			p.CLValidatorType = node.Consensus.ValidatorClient.Type
			p.CLValidatorImage = node.Consensus.ValidatorClient.Image
			p.VCExtraLabels = node.Consensus.ValidatorClient.ExtraLabels
			p.ValMinCpu = node.Consensus.ValidatorClient.CpuRequired
			p.ValMinMemory = node.Consensus.ValidatorClient.MemoryRequired
		}
		participants = append(participants, p)
	}
	return participants
}

func topologyNodeFromConfig(ndx int, validatorsPerNode int, participant *Participant) (*network.Node, error) {

	consensusClient, err := ConsensusClientFromParticipant(participant)
	if err != nil {
		return nil, err
	}

	executionClient, err := ExecutionClientFromParticipant(participant)
	if err != nil {
		return nil, err
	}

	return &network.Node{
		Index:          ndx,
		Execution:      executionClient,
		Consensus:      consensusClient,
		ConsensusVotes: validatorsPerNode,
	}, nil
}

func ExecutionClientFromParticipant(participant *Participant) (*network.ExecutionClient, error) {
	executionClient := &network.ExecutionClient{
		Type:           participant.ElClientType,
		Image:          participant.ElClientImage,
		CpuRequired:    participant.ElMinCpu,
		MemoryRequired: participant.ElMinMemory,
		ExtraLabels:    participant.ELExtraLabels,
	}
	// let the network implementation fix any issues
	err := network.UpdateExecutionClientWithDefaults(executionClient)
	return executionClient, err
}

// ConsensusClientFromParticipant given a participant generate a network.ConsensusClient,
// also adds the default extra labels if they do not exist.
func ConsensusClientFromParticipant(participant *Participant) (*network.ConsensusClient, error) {
	consensusClient := &network.ConsensusClient{
		Type:            participant.ClClientType,
		Image:           participant.ClClientImage,
		ExtraLabels:     participant.CLExtraLabels,
		CpuRequired:     participant.ClMinCpu,
		MemoryRequired:  participant.ClMinMemory,
		ValidatorClient: nil,
	}

	// default to false
	if participant.CLSeparateVC == nil {
		consensusClient.HasValidatorSidecar = false
	} else {
		consensusClient.HasValidatorSidecar = *participant.CLSeparateVC
	}

	if consensusClient.HasValidatorSidecar {
		validatorClient := &network.ValidatorClient{
			Type:           participant.CLValidatorType,
			Image:          participant.CLValidatorImage,
			ExtraLabels:    participant.VCExtraLabels,
			CpuRequired:    participant.ValMinCpu,
			MemoryRequired: participant.ValMinMemory,
		}
		consensusClient.ValidatorClient = validatorClient
	}

	err := network.UpdateConsensusClientWithDefaults(consensusClient)
	return consensusClient, err
}

// TopologyFromRunningEnclave create a Topology of the running enclave we are attaching to.
// DO NOT USE THIS if you have access to the original configuration file
// hacky workaround until we can retrieve the starlark_run_config from a running enclave.
func TopologyFromRunningEnclave(ctx context.Context, enclaveContext *enclaves.EnclaveContext) (*network.Topology, error) {
	isRunning, err := hasEnclaveStarted(enclaveContext)
	if err != nil {
		return nil, err
	}
	if !isRunning {
		return nil, errors.New("failed to compose the topology of the target enclave, the devnet is not running and must be started first")
	}

	viableNodeServiceIDs, err := getViableNodeServiceIDs(ctx, enclaveContext)
	if err != nil {
		return nil, err
	}

	cls := make(map[int]*network.ConsensusClient)
	els := make(map[int]*network.ExecutionClient)
	vcs := make(map[int]*network.ValidatorClient)

	for _, serviceName := range viableNodeServiceIDs {
		if serviceName[:2] == "el" {
			service, err := enclaveContext.GetServiceContext(serviceName)
			if err != nil {
				return nil, err
			}
			executionClient, ndx, err := executionClientFromService(service)
			els[ndx] = executionClient
			continue
		}
		if serviceName[:2] == "cl" {
			service, err := enclaveContext.GetServiceContext(serviceName)
			if err != nil {
				return nil, err
			}
			consensusClient, ndx, err := consensusClientFromService(service)
			cls[ndx] = consensusClient
			continue
		}
		if serviceName[:2] == "vc" {
			service, err := enclaveContext.GetServiceContext(serviceName)
			if err != nil {
				return nil, err
			}
			validatorClient, ndx, err := validatorClientFromService(service)
			//TODO multiple validator clients per cl (requires updated ethereum-package)
			vcs[ndx] = validatorClient
			continue
		}
		return nil, errors.New(fmt.Sprintf("Got unexpected service name format: %s expected cl/el/vc prefix.", serviceName))
	}
	// each node has one el so iterate over these service ids to create the nodes
	var nodes []*network.Node
	for ndx, el := range els {
		node := &network.Node{
			Index:     ndx,
			Execution: el,
			//Consensus:      nil,
			//ConsensusVotes: ,
		}
		cl, ok := cls[ndx]
		if !ok {
			return nil, errors.New(fmt.Sprintf("no matching cl node found for el: %s with ndx: %d", el, ndx))
		}
		consensusClient := &network.ConsensusClient{
			Type: cl.Type,
			//BeaconImage:               "", //undeterminable
			//HasValidatorSidecar: false,
			//ValidatorImage:      "", //undeterminable
		}
		vc, ok := vcs[ndx]
		if !ok {
			consensusClient.HasValidatorSidecar = false
		} else {
			consensusClient.HasValidatorSidecar = true
			consensusClient.ValidatorClient = &network.ValidatorClient{
				Type:           vc.Type,
				Image:          nil, // undeterminable
				ExtraLabels:    nil, // undeterminable
				CpuRequired:    nil, // undeterminable
				MemoryRequired: nil, // undeterminable
			}
		}
		node.Consensus = consensusClient
		nodes = append(nodes, node)
	}
	return &network.Topology{Nodes: nodes}, nil
}

// using the service context of a running enclave create an ExecutionClient object
func executionClientFromService(service *services.ServiceContext) (*network.ExecutionClient, int, error) {
	serviceName := string(service.GetServiceName())
	split := strings.Split(serviceName, "-")
	nodeNumber, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, -1, err
	}
	return &network.ExecutionClient{
		Type:  network.ExecutionClientType(split[2]),
		Image: nil, //non-determinable
	}, nodeNumber, nil
}

// using the service context in a running enclave create a ConsensusClient object
func consensusClientFromService(service *services.ServiceContext) (*network.ConsensusClient, int, error) {
	serviceName := string(service.GetServiceName())
	split := strings.Split(serviceName, "-")
	nodeNumber, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, -1, err
	}
	return &network.ConsensusClient{
		Type: network.ConsensusClientType(split[2]),
		//Image: "", //non-determinable
		//HasValidatorSidecar: false,
		//ValidatorImage:      "",
	}, nodeNumber, nil
}

// using a service context in a running enclave create a ValidatorClient object
func validatorClientFromService(service *services.ServiceContext) (*network.ValidatorClient, int, error) {
	serviceName := string(service.GetServiceName())
	split := strings.Split(serviceName, "-")
	nodeNumber, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, -1, err
	}
	sideCarType := network.ConsensusClientType(split[2])
	return &network.ValidatorClient{
		Type: &sideCarType,
		//Image: "", //non-determinable
		//HasValidatorSidecar: false,
		//ValidatorImage:      "",
	}, nodeNumber, nil
}

// TODO: pr into kurtosis to make numbering standard
func GetNodeExecutionServiceId(node *network.Node, numNodes int) string {
	format := fmt.Sprintf("%%0%dd", int(math.Ceil(math.Log10(float64(numNodes)))))
	numStr := fmt.Sprintf(format, node.Index)
	return fmt.Sprintf("el-%s-%s-%s", numStr, node.Execution.Type, node.Consensus.Type)
}

func GetNodeConsensusServiceId(node *network.Node, numNodes int) string {
	format := fmt.Sprintf("%%0%dd", int(math.Ceil(math.Log10(float64(numNodes)))))
	numStr := fmt.Sprintf(format, node.Index)
	return fmt.Sprintf("cl-%s-%s-%s", numStr, node.Consensus.Type, node.Execution.Type)
}

// note this will just generate one
func GetNodeValidatorServiceId(node *network.Node, numNodes int) (string, error) {
	if node.Consensus.ValidatorClient == nil {
		return "error", errors.New("cant get node validators service id, node does not have a validator client")
	}
	if node.Consensus.ValidatorClient.Type == nil {
		return "error", errors.New("can't get node validators service id, node type is nil, this should not occur, PLEASE REPORT ISSUE	")
	}
	format := fmt.Sprintf("%%0%dd", int(math.Ceil(math.Log10(float64(numNodes)))))
	numStr := fmt.Sprintf(format, node.Index)
	return fmt.Sprintf("vc-%s-%s-%s", numStr, *node.Consensus.ValidatorClient.Type, node.Execution.Type), nil
}
