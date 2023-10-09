// -*- mode:go;mode:go-playground -*-
// snippet of code @ 2023-07-06 09:37:17

// === Go Playground ===
// Execute the snippet with:                 Ctl-Return
// Provide custom arguments to compile with: Alt-Return
// Other useful commands:
// - remove the snippet completely with its dir and all files: (go-playground-rm)
// - upload the current buffer to playground.golang.org:       (go-playground-upload)

package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"text/template"

	_ "embed"

	"github.com/google/go-github/v55/github"
	gh "github.com/google/go-github/v55/github"
	"gopkg.in/yaml.v2"
)

//go:embed pipelinerun.yaml.go.tmpl
var templateContent []byte

type AutoGenerate struct {
	configs       map[string]Config
	ghc           *github.Client
	owner, repo   string
	files_in_repo []string
}

type Params struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Workspace struct {
	Disabled bool   `yaml:"disabled,omitempty"`
	Name     string `yaml:"name,omitempty"`
}
type Task struct {
	Name      string    `yaml:"name"`
	Params    []Params  `yaml:"params,omitempty"`
	Workspace Workspace `yaml:"workspace,omitempty"`
	RunAfter  []string  `yaml:"runAfter,omitempty"`
}

type Config struct {
	Name    string `yaml:"name"`
	Tasks   []Task `yaml:"tasks"`
	Pattern string `yaml:"pattern,omitempty"`
}

func (ag *AutoGenerate) New(filename string) error {
	if _, err := os.Stat(filename); err != nil {
		return fmt.Errorf("file %s not found", filename)
	}
	// open file
	content, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s", filename)
	}
	if err := yaml.Unmarshal(content, &ag.configs); err != nil {
		return fmt.Errorf("failed to parse yaml file %s: %w", filename, err)
	}
	return nil
}

func (ag *AutoGenerate) GetAllFilesInRepo(ctx context.Context) ([]string, error) {
	ret := []string{}
	info, _, err := ag.ghc.Repositories.Get(ctx, ag.owner, ag.repo)
	if err != nil {
		return ret, err
	}
	tree, _, err := ag.ghc.Git.GetTree(ctx, ag.owner, ag.repo, info.GetDefaultBranch(), true)
	if err != nil {
		return ret, err
	}
	for _, entry := range tree.Entries {
		ret = append(ret, entry.GetPath())
	}
	return ret, nil
}

func (ag *AutoGenerate) GetTasks() ([]string, error) {
	var tasks []string
	for k, config := range ag.configs {
		if k == "file_match" {
			fptasks, err := ag.GetFilePatternTasks(context.Background(), config)
			if err != nil {
				// TODO: handle error in main
				return []string{}, fmt.Errorf("Error getting file pattern tasks: %w", err)
			}
			for _, task := range fptasks {
				tasks = append(tasks, task)
			}
			continue
		}
		for _, task := range config.Tasks {
			tasks = append(tasks, task.Name)
		}
	}
	return tasks, nil
}

func (ag *AutoGenerate) GetFilePatternTasks(ctx context.Context, config Config) ([]string, error) {
	var ret []string
	if ag.files_in_repo == nil {
		var err error
		if ag.files_in_repo, err = ag.GetAllFilesInRepo(ctx); err != nil {
			return ret, fmt.Errorf("Error getting all files in repo: %w", err)
		}
	}

	reg, err := regexp.Compile(config.Pattern)
	if err != nil {
		return ret, err
	}
	matched := false
	for _, file := range ag.files_in_repo {
		if reg.MatchString(file) {
			matched = true
			break
		}
	}
	if !matched {
		return ret, nil
	}

	for _, task := range config.Tasks {
		ret = append(ret, task.Name)
	}
	return ret, nil
}

func (ag *AutoGenerate) Output(configs map[string]Config) (string, error) {
	funcMap := template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
	}
	tmpl, err := template.New("pipelineRun").Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	all_tasks, err := ag.GetTasks()
	if err != nil {
		return "", fmt.Errorf("failed to get tasks: %w", err)
	}
	var outputBuffer bytes.Buffer
	data := map[string]interface{}{
		"Configs": configs,
		"Tasks":   all_tasks,
	}
	if err := tmpl.Execute(&outputBuffer, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return outputBuffer.String(), nil
}

func detect(ownerRepo []string) (string, error) {
	ctx := context.Background()
	ghC := gh.NewClient(nil)
	detectLanguages, _, err := ghC.Repositories.ListLanguages(ctx, ownerRepo[0], ownerRepo[1])
	if err != nil {
		return "", err
	}

	ag := &AutoGenerate{ghc: ghC, owner: ownerRepo[0], repo: ownerRepo[1]}
	if err := ag.New("tknautogenerate.yaml"); err != nil {
		return "", err
	}

	configs := map[string]Config{}
	for k := range detectLanguages {
		kl := strings.ToLower(k)
		if c, ok := (ag.configs)[kl]; ok {
			kn := kl
			if c.Name != "" {
				kn = c.Name
			}
			configs[kn] = (ag.configs)[kl]
		}
	}
	for _, config := range ag.configs {
		if config.Pattern == "" {
			continue
		}

		fptasks, err := ag.GetFilePatternTasks(ctx, config)
		if err != nil {
			return "", fmt.Errorf("Error getting file pattern tasks: %w", err)
		}
		if config.Name == "" {
			return "", fmt.Errorf("file_match config on pattern: %s should have a Name", config.Pattern)
		}
		if len(fptasks) != 0 {
			configs[config.Name] = config
		}
	}

	return ag.Output(configs)
}

func main() {
	// parse yaml file and generate configs
	_, err := os.Stat("tknautogenerate.yaml")
	if err != nil {
		fmt.Println("tknautogenerate.yaml not found")
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("usage: tknautogenerate <owner> <repo>")
		return
	}
	ownerRepo := strings.Split(os.Args[1], "/")
	output, err := detect(ownerRepo)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(output)
}
