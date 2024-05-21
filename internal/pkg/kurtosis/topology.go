package kurtosis

import (
	"attacknet/cmd/internal/pkg/network"
	"context"
	"errors"
	"fmt"
	"github.com/kurtosis-tech/kurtosis/api/golang/core/lib/services"
	"strconv"
	"strings"
)

// ComposeTopology using the supplied kurtosis config file generate the expected network topology
func ComposeTopologyFromConfig(config *Config) (*network.Topology, error) {
	var topologyNodes []*network.Node
	currNode := 0
	for _, participant := range config.Participants {
		for nodeNum := 0; nodeNum < participant.Count; nodeNum++ {
			currNode += 1
			node, err := composeTopologyNodeFromConfig(currNode, config.NetParams.NumValKeysPerNode, participant)
			if err != nil {
				return nil, err
			}
			topologyNodes = append(topologyNodes, node)
		}

	}
	return &network.Topology{Nodes: topologyNodes}, nil
}

func composeTopologyNodeFromConfig(ndx int, validatorsPerNode int, participant *Participant) (*network.Node, error) {

	consensusClient, err := composeConsensusClientFromParticipant(participant)
	if err != nil {
		return nil, err
	}

	executionClient, err := composeExecutionClientFromParticipant(participant)
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

func composeConsensusClientFromParticipant(participant *Participant) (*network.ConsensusClient, error) {
	consensusClient := &network.ConsensusClient{
		Type: participant.ClClientType,
		// TODO
		//SidecarCpuRequired:    0,
		//SidecarMemoryRequired: 0,
	}
	if participant.ClClientImage == nil {
		image, err := network.GetDefaultBeaconImage(participant.ClClientType)
		if err != nil {
			return nil, err
		}
		consensusClient.Image = image
	}

	if participant.CLSeparateVC == nil {
		// default to false
		consensusClient.HasValidatorSidecar = false
	} else {
		consensusClient.HasValidatorSidecar = *participant.CLSeparateVC
	}

	//TODO do we really need this for basic topology information
	//if participant.CLExtraLabels == nil {
	//	consensusClient.ExtraLabels = make(map[string]string)
	//} else {
	//	consensusClient.ExtraLabels = participant.CLExtraLabels
	//}

	//TODO do we really need this for basic topology information
	//consensusClient.CpuRequired = 1000
	//consensusClient.MemoryRequired = 1024
	//consensusClient.SidecarCpuRequired = 1000
	//consensusClient.SidecarMemoryRequired = 1024

	// populate validator sidecar information
	if consensusClient.HasValidatorSidecar {
		// check if there is a validator type to use
		if participant.CLValidatorType != nil {
			image, err := network.GetDefaultValidatorImage(*participant.CLValidatorType)
			if err != nil {
				return nil, err
			}
			consensusClient.ValidatorImage = image
		} else {
			// no validator type specified use the beacon type
			image, err := network.GetDefaultValidatorImage(consensusClient.Type)
			if err != nil {
				return nil, err
			}
			consensusClient.ValidatorImage = image
		}
		//TODO do we really need this for basic topology information
		//if participant.VCExtraLabels == nil {
		//	consensusClient.ValidatorExtraLabels = make(map[string]string)
		//} else {
		//	consensusClient.ValidatorExtraLabels = participant.VCExtraLabels
		//}
	}

	return consensusClient, nil
}

func composeExecutionClientFromParticipant(participant *Participant) (*network.ExecutionClient, error) {
	executionClient := &network.ExecutionClient{
		Type: participant.ElClientType,
	}
	if participant.ElClientImage == nil {
		image, err := network.GetDefaultExecutionImage(participant.ElClientType)
		if err != nil {
			return nil, err
		}
		executionClient.Image = image
	} else {
		executionClient.Image = *participant.ElClientImage
	}

	//TODO do we really need this for basic topology information
	//if participant.ELExtraLabels == nil {
	//	executionClient.ExtraLabels = make(map[string]string)
	//} else {
	//	executionClient.ExtraLabels = participant.ELExtraLabels
	//}
	//TODO do we really need this for basic topology information
	//executionClient.CpuRequired = 1000
	//executionClient.MemoryRequired = 1024

	return executionClient, nil
}

// hacky workaround until we can retrieve the starlark_run_config from a running enclave.
func (s *Service) ComposeTopologyFromRunningEnclave(ctx context.Context) (*network.Topology, error) {
	isRunning, err := s.IsDevnetRunning(ctx)
	if err != nil {
		return nil, err
	}
	if !isRunning {
		return nil, errors.New("devnet is not running in the target enclave")
	}

	viableNodeServiceIDs, err := getViableNodeServiceIDs(ctx, s.enclaveContext)
	if err != nil {
		return nil, err
	}

	cls := make(map[int]*network.ConsensusClient)
	els := make(map[int]*network.ExecutionClient)
	vcs := make(map[int]*network.ValidatorClient)

	for _, serviceName := range viableNodeServiceIDs {
		if serviceName[:2] == "el" {
			service, err := s.enclaveContext.GetServiceContext(serviceName)
			if err != nil {
				return nil, err
			}
			executionClient, ndx, err := composeExecutionClientFromService(service)
			els[ndx] = executionClient
			continue
		}
		if serviceName[:2] == "cl" {
			service, err := s.enclaveContext.GetServiceContext(serviceName)
			if err != nil {
				return nil, err
			}
			consensusClient, ndx, err := composeConsensusClientFromService(service)
			cls[ndx] = consensusClient
			continue
		}
		if serviceName[:2] == "vc" {
			service, err := s.enclaveContext.GetServiceContext(serviceName)
			if err != nil {
				return nil, err
			}
			validatorClient, ndx, err := composeValidatorClientFromService(service)
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
			//Image:               "", //undeterminable
			//HasValidatorSidecar: false,
			//ValidatorImage:      "", //undeterminable
		}
		vc, ok := vcs[ndx]
		if !ok {
			consensusClient.HasValidatorSidecar = false
		} else {
			consensusClient.HasValidatorSidecar = true
			consensusClient.ValidatorType = vc.Type
		}
		node.Consensus = consensusClient
		nodes = append(nodes, node)
	}
	return &network.Topology{Nodes: nodes}, nil
}

// returns a pointer to a basic network.ExecutionClient and the associated node number
func composeExecutionClientFromService(service *services.ServiceContext) (*network.ExecutionClient, int, error) {
	serviceName := string(service.GetServiceName())
	split := strings.Split(serviceName, "-")
	nodeNumber, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, -1, err
	}
	return &network.ExecutionClient{
		Type:  split[2],
		Image: "", //non-determinable
	}, nodeNumber, nil
}

func composeConsensusClientFromService(service *services.ServiceContext) (*network.ConsensusClient, int, error) {
	serviceName := string(service.GetServiceName())
	split := strings.Split(serviceName, "-")
	nodeNumber, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, -1, err
	}
	return &network.ConsensusClient{
		Type:  split[2],
		Image: "", //non-determinable
		//HasValidatorSidecar: false,
		//ValidatorImage:      "",
	}, nodeNumber, nil
}

// TODO multiple validator clients
func composeValidatorClientFromService(service *services.ServiceContext) (*network.ValidatorClient, int, error) {
	serviceName := string(service.GetServiceName())
	split := strings.Split(serviceName, "-")
	nodeNumber, err := strconv.Atoi(split[1])
	if err != nil {
		return nil, -1, err
	}
	return &network.ValidatorClient{
		Type:  split[2],
		Image: "", //non-determinable
		//HasValidatorSidecar: false,
		//ValidatorImage:      "",
	}, nodeNumber, nil
}
