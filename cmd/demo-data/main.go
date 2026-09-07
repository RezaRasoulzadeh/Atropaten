package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"Atropaten/internal/demo"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "generate":
		generate(os.Args[2:])
	case "reset":
		reset(os.Args[2:])
	case "clean":
		reset(os.Args[2:])
	default:
		usage()
		os.Exit(2)
	}
}

func generate(args []string) {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	root := fs.String("root", filepath.Join(os.TempDir(), "atropaten-demo"), "isolated demo root")
	seed := fs.Int64("seed", demo.DefaultSeed, "fixed data seed")
	date := fs.String("reference-date", demo.DefaultReferenceDate, "fixed reference date YYYY-MM-DD")
	confirm := fs.Bool("confirm-demo", false, "explicitly confirm development demo generation")
	_ = fs.Parse(args)
	if !*confirm {
		fatal("generation is opt-in; pass --confirm-demo")
	}
	summary, err := demo.Generate(context.Background(), demo.Options{Root: *root, Seed: *seed, ReferenceDate: *date})
	if err != nil {
		fatal(err.Error())
	}
	printJSON(summary)
}

func reset(args []string) {
	fs := flag.NewFlagSet("reset", flag.ExitOnError)
	root := fs.String("root", filepath.Join(os.TempDir(), "atropaten-demo"), "isolated demo root")
	confirm := fs.Bool("confirm-reset-demo", false, "explicitly confirm reset of a marked demo root")
	_ = fs.Parse(args)
	if !*confirm {
		fatal("reset is opt-in; pass --confirm-reset-demo")
	}
	if err := demo.Reset(*root); err != nil {
		fatal(err.Error())
	}
	fmt.Printf("removed isolated demo root %s\n", *root)
}

func printJSON(v any)      { b, _ := json.MarshalIndent(v, "", "  "); fmt.Println(string(b)) }
func fatal(message string) { fmt.Fprintln(os.Stderr, "demo-data:", message); os.Exit(1) }
func usage() {
	fmt.Fprintln(os.Stderr, "Usage: go run ./cmd/demo-data <generate|reset|clean> [flags]")
	fmt.Fprintln(os.Stderr, "Generation and reset require explicit confirmation flags.")
}
