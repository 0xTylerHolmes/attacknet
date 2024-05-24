package main

import (
	attacknet "attacknet/cmd"
	"attacknet/cmd/internal/pkg/planner"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

var (
	// The plan to use for building a chaos experiment
	plannerConfigPathFlag = &cli.StringFlag{
		Name:     "planner-config-path",
		Aliases:  []string{"i"},
		Usage:    "the path for the planner config, this is used to generate the chaos experiments",
		Required: true,
	}

	outputExperimentFlag = &cli.StringFlag{
		Name:     "experiment-output",
		Aliases:  []string{"o"},
		Usage:    "the output for the generated experiment",
		Required: true,
	}
	loggingVerbosityFlag = &cli.StringFlag{
		Name:     "verbosity",
		Usage:    "logging verbosity",
		Required: false,
		Value:    "info",
	}
)

var verbosityLevelMap = map[string]log.Level{
	"info":  log.InfoLevel,
	"debug": log.DebugLevel,
	"warn":  log.WarnLevel,
	"trace": log.TraceLevel,
	"fatal": log.FatalLevel,
	"panic": log.PanicLevel,
}

func getVerbostiy(verbosityLevel string) (*log.Level, error) {
	level, ok := verbosityLevelMap[strings.ToLower(verbosityLevel)]
	if !ok {
		return nil, errors.New(fmt.Sprintf("unknown log level: %s", verbosityLevel))
	}
	return &level, nil
}

func setVerbosity(ctx *cli.Context) error {
	verbosityLevel, err := getVerbostiy(ctx.String(loggingVerbosityFlag.Name))
	if err != nil {
		return err
	}
	log.SetReportCaller(false)
	formatter := &log.TextFormatter{
		TimestampFormat: "02-01-2006 15:04:05", // the "time" field configuratiom
		FullTimestamp:   true,
	}
	log.SetFormatter(formatter)
	log.SetLevel(*verbosityLevel)
	return nil
}

func getPlannerConfig(filePath string) (*planner.Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var config planner.Config
	err = yaml.Unmarshal(data, &config)
	return &config, err
}

func main() {
	app := &cli.App{
		Name:        "experiment-builder",
		Version:     "v0.9", //TODO: is this correct?
		Description: "Builds an Experiment file for attacknet to use",
		Flags: []cli.Flag{
			outputExperimentFlag,
			plannerConfigPathFlag,
			loggingVerbosityFlag,
		},
		Action: func(ctx *cli.Context) error {
			err := setVerbosity(ctx)
			if err != nil {
				return err
			}
			plannerConfig, err := getPlannerConfig(ctx.String(plannerConfigPathFlag.Name))
			if err != nil {
				return err
			}
			service, err := planner.NewBuilder(plannerConfig)

			chaosConfig, networkConfig, err := service.BuildPlan()
			if err != nil {
				return err
			}
			experiment := &attacknet.Config{
				EnclaveName:       "test-plan",
				EnclaveNamespace:  "kt-test-plan",
				KurtosisPackageID: "github.com/kurtosis-tech/ethereum-package",
				KurtosisConfig:    networkConfig,
				ChaosConfig:       chaosConfig,
			}
			outputFile := ctx.String(outputExperimentFlag.Name)
			out, err := yaml.Marshal(experiment)
			if err != nil {
				return nil
			}
			return os.WriteFile(outputFile, out, 0666)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
