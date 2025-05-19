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
	"github.com/ollama/ollama/api"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

const defaultConfigPath = ".config/agency.yaml"

//go:embed testdata/agency.yaml
var defaultConfig []byte

func getTerminalWidth() int {
	width, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		return 80 // Fallback width
	}
	return max(width-20, 80)
}

func generate(ctx context.Context, model, role, instruct, contexts string) (string, error) {

	prompt := "<CONTEXT>\n" + contexts + "\n</CONTEXT>"
	prompt += "\n\nROLE: " + role
	prompt += "\n\nINSTRUCTION: " + instruct
	prompt += "\n\nRESPONSE:"
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return "", err
	}
	var response strings.Builder
	stream := true
	req := &api.GenerateRequest{
		Prompt: prompt,
		Stream: &stream,
		Model:  model,
	}
	if err := client.Generate(ctx, req, func(gr api.GenerateResponse) error {
		fmt.Print(gr.Response)
		response.WriteString(gr.Response)
		return nil
	}); err != nil {
		return "", err
	}
	return response.String(), nil

}

type Agent struct {
	Name        string `yaml:"name"`
	Model       string `yaml:"model"`
	Role        string `yaml:"role"`
	Instruction string `yaml:"instruction"`
	Interactive bool   `yaml:"-"`
}

func (a Agent) interact(ctx context.Context, info ...string) (string, error) {
	contexts, err := readContext(info)
	if err != nil {
		return "", err
	}
	model := a.Model
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
	return generate(ctx, model, role, instruct, contexts)
}

func (a Agent) do(ctx context.Context, info ...string) (string, error) {
	model := "deepseek-r1"
	if a.Model != "" {
		model = a.Model
	}
	if a.Interactive {
		return a.interact(ctx, info...)
	}
	contexts, err := readContext(info)
	if err != nil {
		return "", err
	}
	return generate(ctx, model, a.Role, a.Instruction, contexts)
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

type Agency []*Agent

func (a Agency) get(name string) (*Agent, error) {
	names := make([]string, len(a))
	for i, agent := range a {
		names[i] = agent.Name
		if strings.ToLower(agent.Name) == strings.ToLower(name) {
			return agent, nil
		}
	}
	return nil, fmt.Errorf("no matching agent for %q in [%s]", name, strings.Join(names, ", "))
}

func (a Agency) Dispatch(ctx context.Context, name string, interactive bool, info ...string) error {
	agent, err := a.get(name)
	if err != nil {
		return err
	}
	agent.Interactive = interactive
	return agent.Do(ctx, info...)
}

func loadConfig(path string) (Agency, error) {
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

func loadDefaultConfig() (Agency, error) {
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
	var isInteractive bool
	flag.BoolVar(&isInteractive, "i", false, "toggle interactive mode")
	flag.Parse()
	agents, err := loadDefaultConfig()
	if err != nil {
		log.Fatal(err)
	}
	args := flag.Args()
	if !isInteractive && len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <agent-name> <context>\n", os.Args[0])
		os.Exit(1)
	}
	if err := agents.Dispatch(context.Background(), args[0], isInteractive, args[1:]...); err != nil {
		log.Fatal(err)
	}
}
