package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"example.com/rfc-go-labs/lab"
)

func TestSampleOutputs(t *testing.T) {
	input, err := os.ReadFile("../../examples/backlog.json")
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err := run(nil, bytes.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	var got lab.Report
	if err := json.Unmarshal(output.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Schedules) != 2 || got.Schedules[0].Ticks != 16 || got.Schedules[1].Ticks != 7 {
		t.Fatalf("unexpected sample: %+v", got)
	}
	output.Reset()
	if err := run([]string{"-html"}, bytes.NewReader(input), &output); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(output.String(), "<th scope=\"row\">bounded locality</th><td>1</td><td>7</td>") {
		t.Fatal("HTML is missing the expected locality summary")
	}
}

func TestInvalidInput(t *testing.T) {
	for _, input := range []string{
		`{"resident":"b","swap_cost":3,"max_burst":2,"messages":["a"],"extra":1}`,
		`{"resident":"b","swap_cost":3,"max_burst":2,"messages":["a"]} {}`,
		`{"resident":"b","max_burst":2,"messages":["?"]}`,
		strings.Repeat(" ", (1<<20)+1),
	} {
		var output bytes.Buffer
		if err := run(nil, strings.NewReader(input), &output); err == nil || output.Len() != 0 {
			t.Fatalf("input of %d bytes produced %d bytes, error %v", len(input), output.Len(), err)
		}
	}
}
