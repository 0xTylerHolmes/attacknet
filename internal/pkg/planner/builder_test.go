package planner

import (
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

func getPlannerConfig(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(data, &config)
	return &config, err
}

func TestPlannerRunThrough(t *testing.T) {
	log.SetLevel(log.TraceLevel)
	filePath := "../../../planner-configs/clock-skew.yaml"
	plannerConfig, err := getPlannerConfig(filePath)
	require.NoError(t, err)
	service, err := NewBuilder(plannerConfig)

	_, _, err = service.BuildPlan()
	require.NoError(t, err)

}
