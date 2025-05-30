package main

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/glow/v2/ui"
	"github.com/charmbracelet/huh"
	"github.com/google/subcommands"
	"github.com/ollama/ollama/api"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

const (
	defaultConfigPath = ".config/agency.yaml"
	defaultModel      = "deepseek-r1"
)

//go:embed testdata/agency.yaml
var defaultConfig []byte

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Fallback width
	}
	return max(width-20, 80)
}

func chat(ctx context.Context, model, role, instruct, contexts string) (string, error) {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return "", err
	}
	var response strings.Builder
	stream := true
	if err := client.Chat(ctx, &api.ChatRequest{
		Model: model,
		Messages: []api.Message{
			{Role: "System", Content: role},
			{Role: "User", Content: contexts},
			{Role: "User", Content: instruct},
		},
		Stream: &stream,
	}, func(cr api.ChatResponse) error {
		os.Stdout.WriteString(cr.Message.Content)
		response.WriteString(cr.Message.Content)
		return nil
	}); err != nil {
		return "", err
	}
	return response.String(), nil
}

type Agent struct {
	AgentName   string `yaml:"name"`
	Model       string `yaml:"model"`
	Role        string `yaml:"role"`
	Instruction string `yaml:"instruction"`
	Interactive bool   `yaml:"-"`
}

func (a *Agent) Name() string     { return strings.ToLower(a.AgentName) }
func (a *Agent) Synopsis() string { return a.Role }
func (a *Agent) Usage() string    { return "" }
func (a *Agent) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&a.Interactive, "i", false, "interactive mode")
}

func (a Agent) modelOrDefault() string {
	if a.Model != "" {
		return a.Model
	}
	return defaultModel
}

func (a Agent) interact(ctx context.Context, info ...string) (string, error) {
	contexts, err := readContext(info)
	if err != nil {
		return "", err
	}
	model := a.modelOrDefault()
	role := a.Role
	instruct := a.Instruction
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Title("model").Value(&model).Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("model not provided")
				}
				return nil
			}),
			huh.NewText().Title("context").Value(&contexts).Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("context not provided")
				}
				return nil
			}),
			huh.NewText().Title("role").Value(&role).Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("role not provided")
				}
				return nil
			}),
			huh.NewText().Title("instruction").Value(&instruct).Validate(func(s string) error {
				if s == "" {
					return fmt.Errorf("instruction not provided")
				}
				return nil
			}),
		),
	)
	if err := form.WithWidth(getTerminalWidth()).Run(); err != nil {
		return "", err
	}
	return chat(ctx, model, role, instruct, contexts)
}

func (a Agent) do(ctx context.Context, info ...string) (string, error) {
	if a.Interactive || len(info) == 0 {
		return a.interact(ctx, info...)
	}
	contexts, err := readContext(info)
	if err != nil {
		return "", err
	}
	return chat(ctx, a.modelOrDefault(), a.Role, a.Instruction, contexts)
}

func (a Agent) Do(ctx context.Context, info ...string) error {
	response, err := a.do(ctx, info...)
	if err != nil {
		return err
	}
	if md, err := glamour.Render(response, "dark"); err == nil {
		response = md
	}
	if _, err := ui.NewProgram(ui.Config{ShowLineNumbers: true}, response).Run(); err != nil {
		return err
	}
	return nil
}

func (a *Agent) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if err := a.Do(ctx, f.Args()...); err != nil {
		return subcommands.ExitFailure
	}
	return subcommands.ExitSuccess
}

func loadConfig(path string) ([]*Agent, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	agents := []*Agent{}
	if err := yaml.Unmarshal(b, &agents); err != nil {
		return nil, err
	}
	return agents, nil
}

func loadDefaultConfig() ([]*Agent, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(home, defaultConfigPath)
	if _, err := os.Stat(path); err != nil {
		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
		if err := os.WriteFile(path, defaultConfig, 0644); err != nil {
			return nil, err
		}
	}
	return loadConfig(path)
}

func main() {
	subcommands.Register(subcommands.HelpCommand(), "")
	agents, err := loadDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}
	for _, agent := range agents {
		subcommands.Register(agent, "")
	}
	flag.Parse()
	os.Exit(int(subcommands.Execute(context.Background())))
}
