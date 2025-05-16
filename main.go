package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ollama/ollama/api"
	"gopkg.in/yaml.v3"
)

const defaultConfigPath = ".config/agency.yaml"

func readFiles(paths []string) (string, error) {
	contexts := make([]string, len(paths))
	for _, path := range paths {
		files, err := filepath.Glob(path)
		if err != nil {
			return "", err
		}
		for _, file := range files {
			b, err := os.ReadFile(file)
			if err != nil {
				return "", err
			}
			contexts = append(contexts, string(b))
		}

	}
	return strings.Join(contexts, "\n"), nil
}

type Agent struct {
	Name        string `yaml:"name"`
	Model       string `yaml:"model"`
	Role        string `yaml:"role"`
	Instruction string `yaml:"instruction"`
}

func (a Agent) Do(ctx context.Context, info ...string) error {
	model := "deepseek-r1"
	if a.Model != "" {
		model = a.Model
	}
	contexts, err := readFiles(info)
	prompt := "<CONTEXT>\n" + contexts + "\n</CONTEXT>"
	prompt += "\n\nROLE: " + a.Role
	prompt += "\n\nINSTRUCTION: " + a.Instruction
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return err
	}
	stream := true
	req := &api.GenerateRequest{
		Prompt: prompt,
		Stream: &stream,
		Model:  model,
	}
	return client.Generate(ctx, req, func(gr api.GenerateResponse) error {
		fmt.Print(gr.Response)
		return nil
	})
}

type Agency []Agent

func (a Agency) Dispatch(ctx context.Context, name string, info ...string) error {
	for _, agent := range a {
		if strings.ToLower(agent.Name) == strings.ToLower(name) {
			if err := agent.Do(ctx, info...); err != nil {
				return err
			}
			return nil
		}
	}
	return fmt.Errorf("no matching agent for %q", name)
}

func loadConfig(path string) (Agency, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	agents := []Agent{}
	if err := yaml.Unmarshal(b, &agents); err != nil {
		return nil, err
	}
	return agents, nil
}

func loadDefaultConfig() (Agency, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	return loadConfig(filepath.Join(dir, defaultConfigPath))
}

func main() {
	agents, err := loadDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}
	args := os.Args[1:]
	if len(args) < 2 {
		log.Fatal("agent and context are not provided.")
	}
	if err := agents.Dispatch(context.Background(), args[0], args[1:]...); err != nil {
		log.Fatal(err)
	}
}
