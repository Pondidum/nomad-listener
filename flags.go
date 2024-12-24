package main

import (
	"context"
	"net/url"
	"os"

	"github.com/spf13/pflag"
)

func readFlags(ctx context.Context, args []string) (*Options, error) {

	flags := pflag.NewFlagSet("all", pflag.ContinueOnError)
	addr := flags.String("nomad-addr", os.Getenv("NOMAD_ADDR"), "e.g. https://localhost:4646")
	topics := flags.StringSlice("topics", []string{}, "e.g. job:redis")
	namespace := flags.String("namespace", "", "the namespace to filter to")
	failFast := flags.Bool("fail-fast", false, "stop when a handler encounters an error")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	nomadAddr, err := url.Parse(*addr)
	if err != nil {
		return nil, err
	}

	nomadAddr = nomadAddr.JoinPath("v1/event/stream")
	query := nomadAddr.Query()

	if namespace != nil && *namespace != "" {
		query.Add("namespace", *namespace)
	}
	for _, filter := range *topics {
		query.Add("topic", filter)
	}

	nomadAddr.RawQuery = query.Encode()

	return &Options{
		StreamAddr: nomadAddr.String(),
		FailFast:   *failFast,
	}, nil
}

type Options struct {
	StreamAddr string
	FailFast   bool
}
