package main

import (
	attacknet "attacknet/cmd"
	"context"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
)

var (
	experimentPathFlag = &cli.StringFlag{
		Name:     "experiment-path",
		Aliases:  []string{"i"},
		Usage:    "the path for the experiment to run",
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

// runAttacknetExperiment run the attacknet service and start the experiment
func runAttacknetExperiment(config *attacknet.Config) error {
	ctx, cancelCtxFunc := context.WithCancel(context.Background())
	defer cancelCtxFunc()
	service, err := attacknet.NewService(ctx, config)
	if err != nil {
		return err
	}
	return service.StartExperiment(ctx)
}

func main() {
	app := &cli.App{
		Name:        "attacknet",
		Version:     "v0.9", //TODO: is this correct?
		Description: "Performs chaos experiments inside of kurtosis",
		Flags: []cli.Flag{
			experimentPathFlag,
			loggingVerbosityFlag,
		},
		Action: func(ctx *cli.Context) error {
			err := setVerbosity(ctx)
			if err != nil {
				return err
			}
			log.Debugf("getting attacknet config from file: %s", experimentPathFlag.Name)
			var attacknetConfig attacknet.Config
			path := ctx.String(experimentPathFlag.Name)
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			err = yaml.Unmarshal(data, &attacknetConfig)
			if err != nil {
				return err
			}
			return runAttacknetExperiment(&attacknetConfig)
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
