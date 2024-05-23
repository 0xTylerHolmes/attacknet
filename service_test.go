package attacknet

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

var (
	attacknetConfigPath = "experiments/clock-skew-test.yaml"

	exampleClockSkewTest        = "experiments/examples/attacknet-configs/clock-skew.yaml"
	exampleCPUStressTest        = "experiments/examples/attacknet-configs/cpu-stress.yaml"
	exampleNetworkSplitTest     = "experiments/examples/attacknet-configs/network-split.yaml"
	exampleMemoryStressTest     = "experiments/examples/attacknet-configs/memory-stress.yaml"
	exampleNetworkLatencyTest   = "experiments/examples/attacknet-configs/network-latency.yaml"
	examplePacketCorruptionTest = "experiments/examples/attacknet-configs/packet-corruption.yaml"
	examplePacketDropTest       = "experiments/examples/attacknet-configs/packet-drop.yaml"
	examplePodRestartTest       = "experiments/examples/attacknet-configs/pod-restart.yaml"
)

func TestRunExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(attacknetConfigPath)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)
}

func TestExampleClockSkewExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(exampleClockSkewTest)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)
}

func TestExampleCPUStressExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(exampleCPUStressTest)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)
}

func TestExampleNetworkSplitExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(exampleNetworkSplitTest)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)
}

func TestExampleMemoryStressExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(exampleMemoryStressTest)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)
}

func TestExampleNetworkLatencyExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(exampleNetworkLatencyTest)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)
}

func TestExamplePacketCorruptionExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(examplePacketCorruptionTest)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)
}

func TestExamplePacketDropExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(examplePacketDropTest)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)
}

func TestExamplePodRestartExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(examplePodRestartTest)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)
}
