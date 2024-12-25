package main

import (
	"context"
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

	return &Options{
		NomadAddr: *addr,
		Topics:    *topics,
		Namespace: *namespace,
		FailFast:  *failFast,
	}, nil
}

type Options struct {
	NomadAddr string
	Topics    []string
	Namespace string
	FailFast  bool
}
