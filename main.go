package main

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"
)

func main() {
	if err := runMain(context.Background()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func runMain(ctx context.Context) error {
	addr := os.Getenv("NOMAD_ADDR")
	if addr == "" {
		return errors.New("NOMAD_ADDR must be set")
	}

	nomadAddr, err := url.Parse(addr)
	if err != nil {
		return err
	}

	nomadAddr = nomadAddr.JoinPath("v1/event/stream")

	resp, err := http.Get(nomadAddr.String())
	if err != nil {
		return err
	}
	defer resp.Body.Close()

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
			fmt.Println(time.Now(), event.Topic, event.Type, event.Key, event.Index)
		}
	}

	return nil
}

type EventsLine struct {
	Index  int
	Events []Event
}

type Event struct {
	Topic     string
	Type      string
	Namespace string
	Index     int
	Key       string
}
