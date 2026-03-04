package dev

import (
	"embed"

	"github.com/spf13/cobra"
)

//go:embed templates/scaffold/*.tmpl templates/scaffold/internal/*/*.tmpl templates/scaffold/.github/workflows/*.tmpl
var templatesFS embed.FS

func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "TODO",
		Long:  "TODO.",
	}

	cmd.AddCommand(scaffoldCmd())

	return cmd
}
