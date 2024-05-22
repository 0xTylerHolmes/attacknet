package attacknet

import (
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"testing"
)

var attacknetConfigPath = "experiments/clock-skew-test.yaml"

func TestRunExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	config, err := ReadAttacknetConfig(attacknetConfigPath)
	require.NoError(t, err)
	attacknetService, err := NewService(context.TODO(), config)
	require.NoError(t, err)
	err = attacknetService.StartExperiment(context.TODO())
	require.NoError(t, err)

}
