package dailyclose

import (
	_ "embed"
	"flag"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type Options struct {
	LogFormat string
	LogLevel  string
}

//go:embed templates/.goreleaser.yaml
var embeddedTemplate string

const outputFileName = ".goreleaser.yaml"

func Execute() int {
	options := parseArgs()

	logger, err := getLogger(options.LogLevel, options.LogFormat)
	if err != nil {
		slog.Error("getLogger", "error", err)
		return 1
	}

	slog.SetDefault(logger)

	err = run(options)
	if err != nil {
		slog.Error("run failed", "error", err)
		return 1
	}
	return 0
}

func parseArgs() Options {
	options := Options{}

	flag.StringVar(&options.LogLevel, "log-level", "info", "Log level (debug, info, warn, error), default: info")
	flag.StringVar(&options.LogFormat, "log-format", "text", "Log format (text or json)")

	flag.Parse()

	return options
}

func run(options Options) error {
	filename := ".goreleaser.yaml"
	_, err := os.Stat(filename)
	if err == nil {
		slog.Info("file exists, quitting early to prevent overwriting", "file", filename)
		return nil
	}

	tmpl, err := template.New("script").Parse(embeddedTemplate)
	if err != nil {
		slog.Error("Error creating template:", "error", err)
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	data := struct {
		Files []string
		Cwd   string
	}{
		Cwd: filepath.Base(cwd),
	}

	var scriptBuilder strings.Builder
	err = tmpl.Execute(&scriptBuilder, data)
	if err != nil {
		slog.Error("Error executing template:", "error", err)
		return err
	}

	file, err := os.Create(outputFileName)
	if err != nil {
		slog.Error("Error creating file:", "error", err)
		return err
	}
	defer file.Close()

	_, err = file.WriteString(scriptBuilder.String())
	if err != nil {
		slog.Error("Error writing to file:", "error", err)
		return err
	}

	return nil
}
