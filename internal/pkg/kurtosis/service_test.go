package kurtosis

import (
	"context"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
	"time"
)

// TODO create a test for attaching to the wrong enclave
var kurtosisPackageID = "github.com/kurtosis-tech/ethereum-package"

var exampleDevnet0 string = "testdata/example-devnet-0.yaml"
var exampleDevnet1 string = "testdata/example-devnet-1.yaml"

func getKurtosisConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}
	var kurtosisConfig Config
	err = yaml.Unmarshal(data, &kurtosisConfig)
	return &kurtosisConfig, err
}

func forceKillDevnet(targetEnclaveName string) error {
	kurtosisContext, err := getKurtosisContext()
	if err != nil {
		return err
	}
	err = destroyEnclave(context.TODO(), kurtosisContext, targetEnclaveName)
	if err != nil {
		time.Sleep(30 * time.Second)
	}
	return nil
}

func TestCreateNewEnclave(t *testing.T) {
	targetEnclave := "example-devnet-0"
	kurtosisConfig, err := getKurtosisConfig(exampleDevnet0)
	require.NoError(t, err)
	kurtosisContext, err := getKurtosisContext()
	require.NoError(t, err)
	isRunning, err := doesEnclaveExist(context.TODO(), kurtosisContext, targetEnclave)
	require.NoError(t, err)
	if isRunning {
		t.Logf("enclave was already running. Killing it and then waiting 30 seconds.")
		err := destroyEnclave(context.TODO(), kurtosisContext, targetEnclave)
		require.NoError(t, err)
		// there is a lot of lag time for kubernetes
		time.Sleep(30 * time.Second)
	}
	t.Logf("Trying to create a new Kurtosis Service")
	service, err := NewService(context.TODO(), kurtosisConfig, kurtosisPackageID, targetEnclave)
	require.NoError(t, err)
	isRunning, err = doesEnclaveExist(context.TODO(), kurtosisContext, targetEnclave)
	require.NoError(t, err)
	if !isRunning {
		t.Fatal("failed to find the running enclave we created")
	}
	// success
	err = service.Destroy(context.TODO())
	require.NoError(t, err)
}

func Test_StartDevnet(t *testing.T) {
	targetEnclaveName := "example-devnet-1"
	configPath := exampleDevnet1
	err := forceKillDevnet(targetEnclaveName)
	require.NoError(t, err)
	kurtosisConfig, err := getKurtosisConfig(configPath)
	service, err := NewService(context.TODO(), kurtosisConfig, kurtosisPackageID, targetEnclaveName)
	require.NoError(t, err)
	err = service.StartDevnet(context.TODO())
	require.NoError(t, err)
}

func TestGetTopologyFromRunningEnclave(t *testing.T) {
	targetEnclaveName := "example-devnet-1"
	configPath := exampleDevnet1
	kurtosisConfig, err := getKurtosisConfig(configPath)
	require.NoError(t, err)
	matchingConfigTopology, err := ComposeTopologyFromConfig(kurtosisConfig)
	require.NoError(t, err)
	service, err := NewService(context.TODO(), kurtosisConfig, kurtosisPackageID, targetEnclaveName)
	require.NoError(t, err)
	topology, err := service.ComposeTopologyFromRunningEnclave(context.TODO())
	require.NoError(t, err)
	isEqual := topology.IsEqual(matchingConfigTopology)
	if !isEqual {
		t.Fatal("expected equal topologies")
	}
}

func TestAttachingToDifferentEnclave(t *testing.T) {
	targetEnclaveName := "example-devnet-1"
	configPath := exampleDevnet1
	kurtosisConfig, err := getKurtosisConfig(configPath)
	require.NoError(t, err)
	mismatchedConfig, err := getKurtosisConfig(exampleDevnet0)
	require.NoError(t, err)
	mismatchedTopology, err := ComposeTopologyFromConfig(mismatchedConfig)
	require.NoError(t, err)
	service, err := NewService(context.TODO(), kurtosisConfig, kurtosisPackageID, targetEnclaveName)
	require.NoError(t, err)
	topology, err := service.ComposeTopologyFromRunningEnclave(context.TODO())
	require.NoError(t, err)
	isEqual := topology.IsEqual(mismatchedTopology)
	if isEqual {
		t.Fatal("expected unequal topologies")
	}
}
