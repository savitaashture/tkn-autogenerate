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
	"html/template"
	"log"
	"os"
	"strings"

	_ "embed"

	gh "github.com/google/go-github/v55/github"
	"gopkg.in/yaml.v2"
)

//go:embed pipelinerun.yaml.go.tmpl
var templateContent []byte

// type genConfig struct {
// 	Tasks     []string   `yaml:"tasks"`
// 	Params    []genParam `yaml:"params"`
// 	Workspace bool       `yaml:"workspace"`
// }

// type genParam struct {
// 	Name  string `yaml:"name"`
// 	Value string `yaml:"value"`
// }

type AutoGenerate map[string]Config

type Params struct {
	Name  string `yaml:"name"`
	Value string `yaml:"value"`
}

type Task struct {
	Name      string   `yaml:"name"`
	Params    []Params `yaml:"params,omitempty"`
	Workspace bool     `yaml:"workspace,omitempty"`
}

type Config struct {
	Tasks []Task `yaml:"tasks"`
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
	if err := yaml.Unmarshal(content, &ag); err != nil {
		return fmt.Errorf("failed to parse yaml file %s", filename)
	}
	return nil
}

func (ag *AutoGenerate) GetTasks() []string {
	var tasks []string
	for _, config := range *ag {
		for _, task := range config.Tasks {
			tasks = append(tasks, task.Name)
		}
	}
	return tasks
}

func (ag *AutoGenerate) Output(configs map[string]Config) (string, error) {
	funcMap := template.FuncMap{
		"add": func(a int, b int) int {
			return a + b
		},
	}
	tmpl, err := template.New("yamltemplates").Funcs(funcMap).Parse(string(templateContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var outputBuffer bytes.Buffer
	data := map[string]interface{}{
		"Configs": configs,
		"Tasks":   ag.GetTasks(),
	}
	if err := tmpl.Execute(&outputBuffer, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return outputBuffer.String(), nil
}

func main() {
	// parse yaml file and generate configs
	_, err := os.Stat("tknautogenerate.yaml")
	if err != nil {
		fmt.Println("tknautogenerate.yaml not found")
		return
	}
	ag := &AutoGenerate{}
	if err := ag.New("tknautogenerate.yaml"); err != nil {
		log.Fatal(err)
		return
	}

	if len(os.Args) < 2 {
		fmt.Println("usage: tknautogenerate <owner> <repo>")
		return
	}
	ctx := context.Background()
	ghC := gh.NewClient(nil)
	ownerRepo := strings.Split(os.Args[1], "/")
	ghAuto, _, err := ghC.Repositories.ListLanguages(ctx, ownerRepo[0], ownerRepo[1])
	if err != nil {
		log.Fatal(err)
	}

	configs := map[string]Config{}
	for k := range ghAuto {
		kl := strings.ToLower(k)
		if _, ok := (*ag)[kl]; ok {
			configs[kl] = (*ag)[kl]
		}
	}
	output, err := ag.Output(configs)
	if err != nil {
		log.Fatal(err)
		return
	}
	fmt.Println(output)
}
