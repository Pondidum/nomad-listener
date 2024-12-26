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
	failFast := flags.Bool("fail-fast", false, "stop when a handler encounters an error")
	version := flags.Bool("version", false, "print the version number and exit")

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
		FailFast:  *failFast,
		Version:   *version,
	}, nil
}

type Options struct {
	NomadAddr string
	Topics    []string
	Namespace string
	FailFast  bool
	Version   bool
}
