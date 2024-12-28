package main

import (
	"context"
	"os"

	"github.com/spf13/pflag"
	"go.opentelemetry.io/otel/attribute"
)

func readFlags(ctx context.Context, args []string) (*Options, error) {
	ctx, span := tr.Start(ctx, "read_flags")
	defer span.End()

	flags := pflag.NewFlagSet("all", pflag.ContinueOnError)
	addr := flags.String("nomad-addr", os.Getenv("NOMAD_ADDR"), "e.g. https://localhost:4646")
	topics := flags.StringSlice("topics", []string{}, "e.g. job:redis")
	namespace := flags.String("namespace", "", "the namespace to filter to")
	all := flags.Bool("all", false, "read all events as far back as Nomad can give")
	failFast := flags.Bool("fail-fast", false, "stop when a handler encounters an error")
	version := flags.Bool("version", false, "print the version number and exit")
	verbose := flags.Bool("verbose", false, "print about events seen")

	if err := flags.Parse(args); err != nil {
		return nil, err
	}

	span.SetAttributes(
		attribute.String("nomad.addr", *addr),
		attribute.StringSlice("topics", *topics),
		attribute.String("namespace", *namespace),
		attribute.Bool("failfast", *failFast),
	)

	return &Options{
		NomadAddr: *addr,
		Topics:    *topics,
		Namespace: *namespace,
		All:       *all,
		FailFast:  *failFast,
		Verbose:   *verbose,
		Version:   *version,
	}, nil
}

type Options struct {
	NomadAddr string
	Topics    []string
	Namespace string
	All       bool
	FailFast  bool
	Verbose   bool
	Version   bool
}
