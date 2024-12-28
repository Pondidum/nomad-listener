package main

import (
	"bytes"
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

func TestIndentationWriter(t *testing.T) {

	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "this is a test",
			expected: "    this is a test",
		},
		{
			input:    "this is a test\n",
			expected: "    this is a test\n",
		},
		{
			input:    "this\nis a test",
			expected: "    this\n    is a test",
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			buf := &bytes.Buffer{}
			writer := NewIndenter("    ", buf)
			writer.Write([]byte(tc.input))

			if buf.String() != tc.expected {
				t.Logf("Expect: '%s'", tc.expected)
				t.Logf("Actual: '%s'", buf.String())
				t.Error("Didn't match")
			}
		})
	}

}
