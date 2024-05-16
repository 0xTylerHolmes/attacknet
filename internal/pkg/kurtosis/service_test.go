package kurtosis

import (
	"context"
	"github.com/kurtosis-tech/kurtosis/api/golang/engine/lib/kurtosis_context"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
	"time"
)

//TODO the latency introduced with kubernetes makes running these tests in quick succession impossible.
//     we should find a way to wait the right amount of time to ensure these tests are returning valid
//     results when we run them in quick succession.

var exampleKurtosisConfigPath string = "testdata/example-kurtosis-config.yaml"

func getKurtosisConfig() (*Config, error) {
	data, err := os.ReadFile(exampleKurtosisConfigPath)
	if err != nil {
		return nil, err
	}
	var kurtosisConfig Config
	err = yaml.Unmarshal(data, &kurtosisConfig)
	return &kurtosisConfig, err
}

func isTargetEnclaveRunning(ctx context.Context, kurtosisContext *kurtosis_context.KurtosisContext, targetEnclaveName string) (bool, error) {
	runningEnclaves, err := kurtosisContext.GetEnclaves(ctx)
	if err != nil {
		return false, err
	}

	for enclaveName := range runningEnclaves.GetEnclavesByName() {
		if enclaveName == targetEnclaveName {
			return true, nil
		}
	}
	return false, nil
}

func TestCreateNewEnclave(t *testing.T) {
	kurtosisConfig, err := getKurtosisConfig()
	require.NoError(t, err)
	kurtosisContext, err := getKurtosisContext()
	require.NoError(t, err)
	isRunning, err := isTargetEnclaveRunning(context.TODO(), kurtosisContext, kurtosisConfig.EnclaveName)
	require.NoError(t, err)
	if isRunning {
		t.Log("enclave already running destroying it first.")
		err = kurtosisContext.DestroyEnclave(context.TODO(), kurtosisConfig.EnclaveName)
		require.NoError(t, err)
		t.Logf("%s successfully destroyed, waiting for a few seconds and then continuing test", kurtosisConfig.EnclaveName)
		time.Sleep(10 * time.Second)
	}
	service, err := NewService(context.TODO(), kurtosisConfig)
	require.NoError(t, err)
	time.Sleep(10 * time.Second)
	isRunning, err = isTargetEnclaveRunning(context.TODO(), kurtosisContext, kurtosisConfig.EnclaveName)
	if !isRunning {
		t.Fatal("enclave wasn't running when it was expected to be alive")
	}
	err = service.Destroy(context.TODO())
	t.Logf("destroying the created enclave")
	require.NoError(t, err)
}

func TestKurtosisStartNewDevnet(t *testing.T) {
	kurtosisConfig, err := getKurtosisConfig()
	require.NoError(t, err)
	kurtosisContext, err := getKurtosisContext()
	require.NoError(t, err)
	isRunning, err := isTargetEnclaveRunning(context.TODO(), kurtosisContext, kurtosisConfig.EnclaveName)
	require.NoError(t, err)
	if isRunning {
		t.Log("enclave already running destroying it first.")
		err = kurtosisContext.DestroyEnclave(context.TODO(), kurtosisConfig.EnclaveName)
		require.NoError(t, err)
		t.Logf("%s successfully destroyed, waiting for a few seconds and then continuing test", kurtosisConfig.EnclaveName)
		time.Sleep(10 * time.Second)
	}
	service, err := NewService(context.TODO(), kurtosisConfig)
	require.NoError(t, err)
	time.Sleep(10 * time.Second)
	isRunning, err = isTargetEnclaveRunning(context.TODO(), kurtosisContext, kurtosisConfig.EnclaveName)
	if !isRunning {
		t.Fatal("enclave wasn't running when it was expected to be alive")
	}
	err = service.StartNetwork(context.TODO())
	require.NoError(t, err)
	// check that there are some running services
	services, err := service.enclaveContext.GetServices()
	//TODO check that services match the kurtosis config file
	if len(services) == 0 {
		t.Fatal("no services running after starting the devnet")
	}
}
