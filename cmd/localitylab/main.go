package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"example.com/rfc-go-labs/lab"
)

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string, input io.Reader, output io.Writer) error {
	flags := flag.NewFlagSet("localitylab", flag.ContinueOnError)
	html := flags.Bool("html", false, "render a static HTML report")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("usage: localitylab [-html] < scenario.json")
	}
	const maxInput = 1 << 20
	data, err := io.ReadAll(io.LimitReader(input, maxInput+1))
	if err != nil {
		return err
	}
	if len(data) > maxInput {
		return fmt.Errorf("input exceeds 1 MiB")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var scenario lab.Scenario
	if err := decoder.Decode(&scenario); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("expected exactly one JSON object")
	}
	report, err := lab.Compare(scenario)
	if err != nil {
		return err
	}
	if *html {
		return page.Execute(output, report)
	}
	encoder := json.NewEncoder(output)
	encoder.SetIndent("", "  ")
	return encoder.Encode(report)
}
