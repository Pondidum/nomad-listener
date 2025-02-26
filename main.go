package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tr trace.Tracer

func main() {
	ctx := withSignals(context.Background())

	tracerProvider := configureTelemetry(ctx)
	defer tracerProvider.Shutdown(context.Background())

	tr = tracerProvider.Tracer("nomad-listener")

	if err := runMain(ctx, os.Args[:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMain(ctx context.Context, args []string) error {
	ctx, span := tr.Start(ctx, "main")
	defer span.End()

	opts, err := readFlags(ctx, args)
	if err != nil {
		return err
	}

	if opts.Version {
		fmt.Println(version())
		return nil
	}

	handlers, err := scanHandlers(ctx)
	if err != nil {
		return err
	}

	if err := readEvents(ctx, opts, withTrace(withEventProcessor(opts, handlers))); err != nil {
		return err
	}

	return nil
}

type EventHandler = func(ctx context.Context, event json.RawMessage) error

func withTrace(other EventHandler) EventHandler {
	return func(ctx context.Context, event json.RawMessage) error {
		ctx, span := tr.Start(ctx, "on_event", trace.WithNewRoot())
		defer span.End()

		return other(ctx, event)
	}
}

func withEventProcessor(opts *Options, handlers map[string]string) EventHandler {

	return func(ctx context.Context, rawEvent json.RawMessage) error {
		ctx, span := tr.Start(ctx, "process_event")
		defer span.End()

		event := &Event{}
		if err := json.Unmarshal(rawEvent, event); err != nil {
			return traceError(span, err)
		}

		eventKey := strings.ToLower(fmt.Sprintf("%s-%s", event.Topic, event.Type))
		handler, found := handlers[eventKey]

		if opts.Verbose {
			fmt.Println("-->", event.Topic, event.Type)
			if found {
				fmt.Println("   ", "Running", handler)
			}
		} else if found {
			fmt.Println("-->", "Running", handler)
		}

		span.SetAttributes(
			attribute.String("event.key", eventKey),
			attribute.Bool("handler.found", found),
		)
		if !found {
			return nil
		}

		tmp, err := os.CreateTemp("", "nomad-event-*.json")
		if err != nil {
			return traceError(span, err)
		}
		defer tmp.Close() // just in case we error before the real .Close() call

		span.SetAttributes(attribute.String("event.path", tmp.Name()))

		if _, err := tmp.Write(rawEvent); err != nil {
			return traceError(span, err)
		}
		if err := tmp.Close(); err != nil {
			return traceError(span, err)
		}

		span.SetAttributes(attribute.Bool("event.written", true))

		defer func() {
			// don't care if this fails really
			err := os.Remove(tmp.Name())
			span.SetAttributes(attribute.Bool("event.cleaned", err == nil))
		}()

		cmd := exec.CommandContext(ctx, handler, tmp.Name())
		cmd.Stdin = os.Stdin
		cmd.Stdout = NewIndenter("    ", os.Stdout)
		cmd.Stderr = NewIndenter("    ", os.Stderr)
		cmd.Env = append(os.Environ(), "TRACEPARENT="+traceParent(span))

		if err := cmd.Run(); err != nil {
			return traceError(span, err)
		}

		span.SetStatus(codes.Ok, "")
		return nil
	}
}

func readEvents(ctx context.Context, opts *Options, onEvent EventHandler) error {
	ctx, span := tr.Start(ctx, "read_events")
	defer span.End()

	addr, err := buildUrl(ctx, opts)
	if err != nil {
		return err
	}

	span.SetAttributes(attribute.String("nomad.addr", addr.String()))
	fmt.Printf("    Listening to events from %s\n", addr.String())

	resp, err := http.Get(addr.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)
	for {
		select {
		case <-ctx.Done():
			return nil

		default:
			line, err := reader.ReadBytes('\n')
			if err != nil {
				return err // make this restart the consumption later
			}

			events := &EventsLine{}
			if err := json.Unmarshal(line, events); err != nil {
				return err
			}

			for _, event := range events.Events {
				if err := onEvent(ctx, event); err != nil {
					if opts.FailFast {
						return err
					}

					fmt.Println("Error:", err.Error())
				}
			}
		}

	}
}

func buildUrl(ctx context.Context, opt *Options) (*url.URL, error) {

	nomadAddr, err := url.Parse(opt.NomadAddr)
	if err != nil {
		return nil, err
	}

	eventsUrl := nomadAddr.JoinPath("v1/event/stream")
	query := eventsUrl.Query()

	if !opt.All {
		res, err := http.Get(nomadAddr.JoinPath("/v1/jobs").String())
		if err != nil {
			return nil, err
		}
		res.Body.Close()
		if currentIndex := res.Header.Get("X-Nomad-Index"); currentIndex != "" {
			query.Add("index", currentIndex)
		}
	}

	if opt.Namespace != "" {
		query.Add("namespace", opt.Namespace)
	}
	for _, filter := range opt.Topics {
		query.Add("topic", filter)
	}

	eventsUrl.RawQuery = query.Encode()

	return eventsUrl, nil
}

func scanHandlers(ctx context.Context) (map[string]string, error) {
	ctx, span := tr.Start(ctx, "scan_handlers")
	defer span.End()

	entries, err := os.ReadDir("handlers")
	if os.IsNotExist(err) {
		entries = []fs.DirEntry{}
	} else if err != nil {
		return nil, traceError(span, err)
	}

	handlers := make(map[string]string, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		key := strings.ToLower(entry.Name())
		location := path.Join("handlers", entry.Name())

		info, err := entry.Info()
		if err != nil {
			return nil, traceError(span, err)
		}

		if !isExecutable(info.Mode()) {
			fmt.Printf("--> Warning, %s is not executable, and will fail to run\n", location)
			continue
		}

		handlers[key] = location
	}

	span.SetAttributes(
		attribute.Int("entries.total", len(entries)),
		attribute.Int("entries.handlers", len(handlers)),
	)

	fmt.Printf("    Found %v event handlers\n", len(handlers))

	return handlers, nil
}

type EventsLine struct {
	Index  int
	Events []json.RawMessage
}

type Event struct {
	Topic     string
	Type      string
	Namespace string
	Index     int
	Key       string
}
