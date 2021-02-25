package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func main() {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	resp, raw, err := cli.ServiceInspectWithRaw(ctx, "server_api", types.ServiceInspectOptions{})
	if err != nil {
		panic(err)
	}

	data, err := json.Marshal(resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(data)
	fmt.Println(raw)

	time.Sleep(time.Minute)
}
