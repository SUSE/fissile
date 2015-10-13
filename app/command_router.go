package app

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/fatih/color"
)

// CommandRouter will dispatch CLI commands to their relevant functions
func (f *Fissile) CommandRouter(c *cli.Context) {
	paths, err := absolutePathsForFlags(c, "release", "target", "light-opinions", "dark-opinions", "roles-manifest", "compiled-packages")
	if err != nil {
		log.Fatalln(color.RedString("%v", err))
	}
		
	switch {
	case c.Command.FullName() == "release jobs-report":
		f.ListJobs(
			paths["release"],
		)
	case c.Command.FullName() == "release packages-report":
		f.ListPackages(
			paths["release"],
		)
	case c.Command.FullName() == "compilation build-base":
		f.CreateBaseCompilationImage(
			c.String("base-image"),
			c.String("repository"),
		)
	case c.Command.FullName() == "compilation show-base":
		f.ShowBaseImage(
			c.String("base-image"),
			c.String("repository"),
		)
	case c.Command.FullName() == "compilation start":
		f.Compile(
			paths["release"],
			c.String("repository"),
			paths["target"],
			c.Int("workers"),
		)
	case c.Command.FullName() == "configuration report":
		f.ListFullConfiguration(
			paths["release"],
		)
	case c.Command.FullName() == "templates report":
		f.PrintTemplateReport(
			paths["release"],
		)
	case c.Command.FullName() == "configuration generate":
		f.GenerateConfigurationBase(
			paths["release"],
			paths["light-opinions"],
			paths["dark-opinions"],
			paths["target"],
			c.String("prefix"),
			c.String("provider"),
		)
	case c.Command.FullName() == "images create-base":
		f.GenerateBaseDockerImage(
			paths["target"],
			c.String("configgin"),
			c.String("base-image"),
			c.Bool("no-build"),
			c.String("repository"),
		)
	case c.Command.FullName() == "images create-roles":
		f.GenerateRoleImages(
			paths["target"],
			c.String("repository"),
			c.Bool("no-build"),
			paths["release"],
			paths["roles-manifest"],
			paths["compiled-packages"],
			c.String("default-consul-address"),
			c.String("default-config-store-prefix"),
			c.String("version"),
		)
	case c.Command.FullName() == "images list-roles":
		f.ListRoleImages(
			c.String("repository"),
			paths["release"],
			paths["roles-manifest"],
			c.String("version"),
			c.Bool("docker-only"),
			c.Bool("with-sizes"),
		)
	}
}

func absolutePathsForFlags(c *cli.Context, flagNames ...string) (map[string]string, error) {
	absolutePaths := map[string]string{}
	for _, flagName := range(flagNames) {
		if !c.IsSet(flagName) {
			continue
		}
		path, err := filepath.Abs(c.String(flagName))
		if err != nil {
			return nil, fmt.Errorf("Error getting absolute path for option %s, path %s: %v",
				flagName, c.String(flagName), err)
		}
		absolutePaths[flagName] = path
	}
	return absolutePaths, nil
}
