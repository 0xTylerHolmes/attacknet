package main

import (
	attacknet "attacknet/cmd"
	"context"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
	"os"
)

//var CLI struct {
//	Init struct {
//		Force bool   `arg:"force" optional:"" default:"false" name:"force" help:"Overwrite existing project."`
//		Path  string `arg:"" optional:"" type:"existingdir" name:"path" help:"Path to initialize project on. Defaults to current working directory."`
//	} `cmd:"" help:"Initialize an attacknet project"`
//	Start struct {
//		Suite string `arg:"" name:"suite name" help:"The test suite to run. These are located in ./test-suites."`
//	} `cmd:"" help:"Run a specified test suite"`
//	Plan struct {
//		Name string `arg:"" optional:"" name:"name" help:"The name of the test suite to be generated."`
//		Path string `arg:"" optional:"" type:"existingfile" name:"path" help:"Location of the planner configuration."`
//	} `cmd:"" help:"Construct an attacknet suite for a client"`
//	// Explore struct{} `cmd:"" help:"Run in exploration mode"`
//}

var (
	experimentPathFlag = &cli.StringFlag{
		Name:     "experiment-path",
		Aliases:  []string{"i"},
		Usage:    "the path for the experiment to run",
		Required: true,
	}
)

// runAttacknetExperiment run the attacknet service and start the experiment
func runAttacknetExperiment(config *attacknet.Config) error {
	ctx, cancelCtxFunc := context.WithCancel(context.Background())
	defer cancelCtxFunc()
	service, err := attacknet.NewService(ctx, config)
	if err != nil {
		return err
	}
	return service.StartTestSuite(ctx)
}

func main() {
	app := &cli.App{
		Name:        "attacknet",
		Version:     "v0.9", //TODO: is this correct?
		Description: "Performs chaos experiments inside of kurtosis",
		Flags: []cli.Flag{
			experimentPathFlag,
		},
		Action: func(ctx *cli.Context) error {
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

//func main() {
//	// todo: use flag for arg parse
//
//	c := kong.Parse(&CLI)
//
//	b := c.Command()
//	switch b {
//
//	case "start <suite name>":
//		ctx, cancelCtxFunc := context.WithCancel(context.Background())
//		defer cancelCtxFunc()
//		cfg, err := project.LoadSuiteConfigFromName(CLI.Start.Suite)
//		if err != nil {
//			log.Fatal(err)
//		}
//		err = attacknet.StartTestSuite(ctx, cfg)
//		if err != nil {
//			log.Fatal(err)
//			os.Exit(1)
//		}
//	//TODO: reimplement plan with new configs and put into a different executable
//	//case "plan <name> <path>":
//	//	config, err := plan.LoadPlannerConfigFromPath(CLI.Plan.Path)
//	//	if err != nil {
//	//		log.Fatal(err)
//	//		os.Exit(1)
//	//	}
//	//	err = plan.BuildPlan(CLI.Plan.Name, config)
//	//	if err != nil {
//	//		log.Fatal(err)
//	//		os.Exit(1)
//	//	}
//	/*
//		case "explore":
//			topo, err := plan.LoadPlannerConfigFromPath("planner-configs/network-latency-reth.yaml")
//			if err != nil {
//				log.Fatal(err)
//			}
//			suiteCfg, err := project.LoadSuiteConfigFromName("plan/network-latency-reth")
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			f, err := os.ReadFile("webhook")
//			if err != nil {
//				log.Fatal(err)
//			}
//			w := string(f)
//			client, err := webhook.NewWithURL(w)
//			if err != nil {
//				log.Fatal(err)
//			}
//
//			err = exploration.StartExploration(topo, suiteCfg)
//			if err != nil {
//				message, err := client.CreateContent(fmt.Sprintf("attacknet run completed with error  %s", err.Error()))
//				if err != nil {
//					log.Fatal(err)
//				}
//				_ = message
//				log.Fatal(err)
//			}
//
//			message, err := client.CreateContent("attacknet run completed with error ")
//			if err != nil {
//				log.Fatal(err)
//			}
//			_ = message
//			os.Exit(1)
//	*/
//	default:
//		log.Fatal("unrecognized arguments")
//	}
//}
