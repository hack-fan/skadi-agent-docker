package main

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/hack-fan/skadigo"
	"go.uber.org/zap"

	"github.com/hack-fan/config"
)

var cli *client.Client
var log *zap.SugaredLogger
var ctx = context.Background()

type Settings struct {
	Debug  bool `default:"false"`
	Token  string
	Server string `default:"https://api.letserver.run"`
	// default service name to restart
	Default string
}

func handler(msg string) (string, error) {
	resp, _, err := cli.ServiceInspectWithRaw(ctx, "server_api", types.ServiceInspectOptions{})
	if err != nil {
		panic(err)
	}
	return resp.Spec.Name, nil
}

func main() {
	var err error
	// load config
	var settings = new(Settings)
	config.MustLoad(settings)

	// logger
	var logger *zap.Logger
	if settings.Debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // nolint
	log = logger.Sugar()

	// docker cli
	cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// skadi agent
	agent := skadigo.New(settings.Token, settings.Server, handler, nil)
	log.Info("Skadi agent start")
	agent.Start()
}
