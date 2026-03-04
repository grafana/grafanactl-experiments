package dev

import (
	"embed"

	"github.com/spf13/cobra"
)

//go:embed templates/import/*.tmpl templates/scaffold/*.tmpl templates/scaffold/internal/*/*.tmpl templates/scaffold/.github/workflows/*.tmpl
var templatesFS embed.FS

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Manage Grafana resources as code",
		Long:  "TODO.",
	}

	cmd.AddCommand(importCmd())
	cmd.AddCommand(scaffoldCmd())

	return cmd
}
