package main

import (
	"os"
	"testing"
)

func TestIsExecutable(t *testing.T) {
	file, err := os.Stat("handlers/Job-JobRegistered")
	if err != nil {
		t.Error(err)
	}

	if !isExecutable(file.Mode()) {
		t.Errorf("%s should be executable, but was not", file.Name())
	}
}

func TestIsExecutableScan(t *testing.T) {
	entries, _ := os.ReadDir("handlers")
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		info, _ := entry.Info()
		if !isExecutable(info.Mode()) {
			t.Errorf("%s should be executable, but was %v", entry.Name(), entry.Type())
		}
	}
}
