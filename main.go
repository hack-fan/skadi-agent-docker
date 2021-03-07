package main

import (
	"context"
	"encoding/json"
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
		warning, err := update(service)
		if err != nil {
			log.Error(e)
			return "", err
		}
		if warning != "" {
			log.Warnf("service [%s] update warning:\n%s", service, warning)
			return fmt.Sprintf("update service [%s] successful with warnings:\n%s", service, warning), nil
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
func update(service string) (string, error) {
	resp, _, err := cli.ServiceInspectWithRaw(ctx, service, types.ServiceInspectOptions{})
	if err != nil {
		return "", fmt.Errorf("check service [%s] failed: %w", service, err)
	}
	// if you start service from cli, the image will be locked to a version, change it to tag only
	image := resp.Spec.TaskTemplate.ContainerSpec.Image
	i := strings.Index(image, "@")
	if i > 0 {
		resp.Spec.TaskTemplate.ContainerSpec.Image = image[0:i]
	}
	respLog, _ := json.MarshalIndent(resp, "", "    ")
	log.Debugf("service: %s", string(respLog))
	// this field must greater than prev state to take effect
	resp.Spec.TaskTemplate.ForceUpdate += 1
	// then force update
	res, err := cli.ServiceUpdate(ctx, service, resp.Version, resp.Spec, types.ServiceUpdateOptions{})
	if err != nil {
		return "", fmt.Errorf("update service [%s] failed: %w", service, err)
	}
	if len(res.Warnings) > 0 {
		warning := strings.Join(res.Warnings, "\n")
		return warning, nil
	}
	return "", nil
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
	agent := skadigo.New(settings.Token, settings.Server, handler, &skadigo.Options{
		Logger: log,
	})
	log.Info("Skadi agent start")
	agent.Start()
}
