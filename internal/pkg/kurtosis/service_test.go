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
	exists, err := doesEnclaveExist(context.TODO(), kurtosisContext, targetEnclaveName)
	if exists {
		err = destroyEnclave(context.TODO(), kurtosisContext, targetEnclaveName)
		if err != nil {
			return err
		}
		time.Sleep(kubernetesOverheadDuration)
	}
	return nil
}

func TestCreateNewEnclave(t *testing.T) {
	targetEnclave := "example-devnet-0"
	kurtosisConfig, err := getKurtosisConfig(exampleDevnet0)
	require.NoError(t, err)

	err = forceKillDevnet(targetEnclave)
	require.NoError(t, err)

	service, err := NewService(context.TODO(), kurtosisConfig, kurtosisPackageID, targetEnclave)
	err = service.ForceCreateNewEnclave(context.TODO())
	require.NoError(t, err)

	err = forceKillDevnet(targetEnclave)
	require.NoError(t, err)
}

func Test_StartNonExistingDevnet(t *testing.T) {
	targetEnclave := "example-devnet-0"
	kurtosisConfig, err := getKurtosisConfig(exampleDevnet0)
	require.NoError(t, err)

	err = forceKillDevnet(targetEnclave)
	require.NoError(t, err)

	service, err := NewService(context.TODO(), kurtosisConfig, kurtosisPackageID, targetEnclave)
	err = service.prepareNewEnclaveAndStartDevnet(context.TODO())
	require.NoError(t, err)

	err = forceKillDevnet(targetEnclave)
	require.NoError(t, err)
}

func TestGetTopologyFromRunningEnclave(t *testing.T) {
	runningEnclaveName := "example-devnet-1"
	runningEnclaveConfigPath := exampleDevnet1
	runningEnclaveConfig, err := getKurtosisConfig(runningEnclaveConfigPath)
	require.NoError(t, err)
	matchingConfigTopology, err := ComposeTopologyFromConfig(runningEnclaveConfig)
	require.NoError(t, err)
	kurtosisContext, err := getKurtosisContext()
	require.NoError(t, err)
	err = forceKillDevnet(runningEnclaveName)
	require.NoError(t, err)
	service, err := NewService(context.TODO(), runningEnclaveConfig, kurtosisPackageID, runningEnclaveName)
	require.NoError(t, err)
	err = service.prepareNewEnclaveAndStartDevnet(context.TODO())
	require.NoError(t, err)
	runningEnclaveContext, err := getEnclaveContext(context.TODO(), kurtosisContext, runningEnclaveName)
	topology, err := ComposeTopologyFromRunningEnclave(context.TODO(), runningEnclaveContext)
	require.NoError(t, err)
	isEqual := topology.IsEqual(matchingConfigTopology)
	if !isEqual {
		t.Fatal("expected equal topologies")
	}
	// destroy the test enclave
	err = forceKillDevnet(runningEnclaveName)
	require.NoError(t, err)
}

func TestAttachingToDifferentEnclave(t *testing.T) {
	runningEnclaveName := "example-devnet-1"
	runningEnclaveConfigPath := exampleDevnet1
	runningEnclaveConfig, err := getKurtosisConfig(runningEnclaveConfigPath)
	require.NoError(t, err)
	differentEnclaveConfig, err := getKurtosisConfig(exampleDevnet0)
	require.NoError(t, err)

	seperateConfigTopology, err := ComposeTopologyFromConfig(differentEnclaveConfig)
	require.NoError(t, err)
	kurtosisContext, err := getKurtosisContext()
	require.NoError(t, err)
	service, err := NewService(context.TODO(), runningEnclaveConfig, kurtosisPackageID, runningEnclaveName)
	require.NoError(t, err)
	err = service.prepareNewEnclaveAndStartDevnet(context.TODO())
	require.NoError(t, err)
	runningEnclaveContext, err := getEnclaveContext(context.TODO(), kurtosisContext, runningEnclaveName)
	topology, err := ComposeTopologyFromRunningEnclave(context.TODO(), runningEnclaveContext)
	require.NoError(t, err)
	isEqual := topology.IsEqual(seperateConfigTopology)
	if !isEqual {
		t.Fatal("expected equal topologies")
	}
	// destroy the test enclave
	err = forceKillDevnet(runningEnclaveName)
	require.NoError(t, err)
}
