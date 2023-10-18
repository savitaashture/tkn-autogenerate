package tknautogenerate

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/google/go-github/v55/github"
)

func Detect(cli *CliStruct) (string, error) {
	ownerRepo := strings.Split(cli.OwnerRepo, "/")
	ctx := context.Background()
	ghC := gh.NewClient(nil)
	if cli.Token != "" {
		ghC = ghC.WithAuthToken(cli.Token)
	}
	detectLanguages, _, err := ghC.Repositories.ListLanguages(ctx, ownerRepo[0], ownerRepo[1])
	if err != nil {
		return "", err
	}

	ag := &AutoGenerate{ghc: ghC, owner: ownerRepo[0], repo: ownerRepo[1], cli: cli}
	if err := ag.New(cli.autoGenerateYaml); err != nil {
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
	for k, config := range ag.configs {
		if config.Pattern == "" {
			continue
		}
		fptasks, err := ag.GetFilePatternTasks(ctx, config)
		if err != nil {
			return "", fmt.Errorf("Error getting file pattern tasks: %w", err)
		}
		if config.Name != "" {
			k = config.Name
		}
		if len(fptasks) != 0 {
			configs[k] = config
		}
	}

	return ag.Output(configs)
}
