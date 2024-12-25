package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
)

func main() {
	if err := runMain(context.Background(), os.Args[:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMain(ctx context.Context, args []string) error {

	opts, err := readFlags(ctx, args)
	if err != nil {
		return err
	}

	addr, err := buildUrl(ctx, opts)
	if err != nil {
		return err
	}

	resp, err := http.Get(addr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	handlers, err := scanHandlers(ctx)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return err // make this restart the consumption later
		}

		events := &EventsLine{}
		if err := json.Unmarshal(line, events); err != nil {
			return err
		}

		for _, event := range events.Events {
			if err := processEvent(ctx, handlers, event); err != nil {
				if opts.FailFast {
					return err
				}

				fmt.Println("Error:", err.Error())
			}
		}
	}

}

func buildUrl(ctx context.Context, opt *Options) (string, error) {

	nomadAddr, err := url.Parse(opt.NomadAddr)
	if err != nil {
		return "", err
	}

	nomadAddr = nomadAddr.JoinPath("v1/event/stream")
	query := nomadAddr.Query()

	if opt.Namespace != "" {
		query.Add("namespace", opt.Namespace)
	}
	for _, filter := range opt.Topics {
		query.Add("topic", filter)
	}

	nomadAddr.RawQuery = query.Encode()

	return nomadAddr.String(), nil
}

func processEvent(ctx context.Context, handlers map[string]string, event json.RawMessage) error {
	eventKey, err := getEventKey(event)
	if err != nil {
		return err
	}

	handler, found := handlers[eventKey]
	if !found {
		return nil
	}

	fmt.Println("-->", handler)

	tmp, err := os.CreateTemp("", "nomad-event-*.json")
	if err != nil {
		return err
	}
	defer tmp.Close() // just in case we error before the real .Close() call

	if _, err := tmp.Write(event); err != nil {
		return err
	}
	tmp.Close()

	cmd := exec.CommandContext(ctx, handler, tmp.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func scanHandlers(ctx context.Context) (map[string]string, error) {
	entries, err := os.ReadDir("handlers")
	if err != nil {
		return nil, err
	}

	handlers := make(map[string]string, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		handlers[entry.Name()] = path.Join("handlers", entry.Name())
	}

	return handlers, nil
}

func getEventKey(raw json.RawMessage) (string, error) {
	event := &Event{}
	if err := json.Unmarshal(raw, event); err != nil {
		return "", err
	}

	return fmt.Sprintf("%s-%s", event.Topic, event.Type), nil
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
