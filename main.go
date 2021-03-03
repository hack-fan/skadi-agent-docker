package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"go.uber.org/zap"

	"github.com/hack-fan/config"
	"github.com/hack-fan/skadigo"
)

var cli *client.Client
var log *zap.SugaredLogger
var ctx = context.Background()
var settings = new(Settings)

type Settings struct {
	Debug  bool `default:"false"`
	Token  string
	Server string `default:"https://api.letserver.run"`
	// default service name to update
	Default string
}

// up: update default service
// up <service>: update the service
func handler(msg string) (string, error) {
	log.Infof("job received: %s", msg)
	// default error
	e := fmt.Errorf("unsupported command: %s", msg)
	// parse command
	switch {
	// update
	case strings.HasPrefix(msg, "up"):
		args := strings.Split(msg, " ")
		service := settings.Default
		if len(args) == 1 {
			if settings.Default == "" {
				log.Error("missing default setting")
				return "", errors.New("default service is not defined")
			}
		} else if len(args) == 2 {
			service = strings.TrimSpace(args[1])
		} else {
			log.Error(e)
			return "", e
		}
		err := update(service)
		if err != nil {
			log.Error(e)
			return "", err
		}
		log.Infof("succeeded: %s", msg)
		return fmt.Sprintf("update service [%s] successful", service), nil
	// other
	default:
		log.Error(e)
		return "", e
	}
}

// update docker service
func update(service string) error {
	resp, _, err := cli.ServiceInspectWithRaw(ctx, service, types.ServiceInspectOptions{})
	if err != nil {
		return fmt.Errorf("check service [%s] failed: %w", service, err)
	}
	_, err = cli.ServiceUpdate(ctx, service, resp.Version, resp.Spec, types.ServiceUpdateOptions{})
	if err != nil {
		return fmt.Errorf("update service [%s] failed: %w", service, err)
	}
	return nil
}

func main() {
	var err error
	// load config
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
