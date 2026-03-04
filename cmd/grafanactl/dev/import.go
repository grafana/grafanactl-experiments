package dev

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/grafana/grafana-foundation-sdk/go/cog/plugins"
	"github.com/grafana/grafana-foundation-sdk/go/dashboard"
	"github.com/grafana/grafana-foundation-sdk/go/dashboardv2beta1"
	cmdconfig "github.com/grafana/grafanactl/cmd/grafanactl/config"
	cmdio "github.com/grafana/grafanactl/cmd/grafanactl/io"
	"github.com/grafana/grafanactl/cmd/grafanactl/resources"
	model "github.com/grafana/grafanactl/internal/resources"
	"github.com/huandu/xstrings"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

//nolint:gochecknoglobals
var convertersMap = map[string]resourceConverter{
	"Dashboard.dashboard.grafana.app/v0alpha1": dashboardv1Converter,
	"Dashboard.dashboard.grafana.app/v1beta1":  dashboardv1Converter,
	"Dashboard.dashboard.grafana.app/v2beta1":  dashboardv2Converter,
}

type importOpts struct {
	Path string
}

func (opts *importOpts) setup(flags *pflag.FlagSet) {
	flags.StringVarP(&opts.Path, "path", "p", "imported", "Import path.")
}

func importCmd() *cobra.Command {
	configOpts := &cmdconfig.Options{}
	opts := &importOpts{}

	cmd := &cobra.Command{
		Use:   "import [RESOURCE_SELECTOR]...",
		Args:  cobra.ArbitraryArgs,
		Short: "Import resources from Grafana and converts them to code",
		Long:  "TODO.",
		Example: `
	# TODO
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()

			cfg, err := configOpts.LoadRESTConfig(ctx)
			if err != nil {
				return err
			}

			res, err := resources.FetchResources(ctx, resources.FetchRequest{
				Config:      cfg,
				StopOnError: true,
			}, args)
			if err != nil {
				return err
			}

			plugins.RegisterDefaultPlugins()

			imported := 0
			err = res.Resources.ForEach(func(resource *model.Resource) error {
				if err := convertResource(opts.Path, resource); err != nil {
					resourceId := fmt.Sprintf("%s.%s", resource.Kind(), resource.Name())
					cmdio.Info(cmd.OutOrStdout(), "Skipping resource '%s': %s", resourceId, err)
					return nil
				}

				imported += 1
				return nil
			})
			if err != nil {
				return err
			}

			cmdio.Success(cmd.OutOrStdout(), "Imported %d resources in %s", imported, opts.Path)

			return nil
		},
	}

	opts.setup(cmd.Flags())
	configOpts.BindFlags(cmd.Flags())

	return cmd
}

type resourceConverter func(resource *model.Resource) (string, string, error)

func convertResource(destinationRoot string, resource *model.Resource) error {
	tmpl, err := template.New("").Option("missingkey=error").ParseFS(templatesFS, "templates/import/*.tmpl")
	if err != nil {
		return err
	}

	gvk := resource.GroupVersionKind()
	converterKey := fmt.Sprintf("%s.%s", gvk.Kind, gvk.GroupVersion().String())

	converter, ok := convertersMap[converterKey]
	if !ok {
		return fmt.Errorf("no converter found for %s", converterKey)
	}

	converted, sdkPkg, err := converter(resource)
	if err != nil {
		return err
	}

	convertedFile := filepath.Join(destinationRoot, xstrings.ToSnakeCase(resource.Name())) + ".go"

	if err := ensureDirectory(filepath.Dir(convertedFile)); err != nil {
		return err
	}

	fileHandle, err := os.OpenFile(convertedFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fileHandle.Close()

	err = tmpl.ExecuteTemplate(fileHandle, "resource.go.tmpl", map[string]any{
		"Package":          strings.ToLower(resource.Kind()),
		"GroupVersion":     gvk.GroupVersion().String(),
		"Kind":             resource.Kind(),
		"Name":             resource.Name(),
		"SDKPackage":       sdkPkg,
		"FuncName":         xstrings.ToCamelCase(resource.Name()),
		"ConvertedBuilder": converted,
	})
	if err != nil {
		return err
	}

	return nil
}

func dashboardv1Converter(resource *model.Resource) (string, string, error) {
	spec, err := resource.Spec()
	if err != nil {
		return "", "", err
	}

	marshalled, err := json.Marshal(spec)
	if err != nil {
		return "", "", err
	}

	object := dashboard.Dashboard{}
	if err = json.Unmarshal(marshalled, &object); err != nil {
		return "", "", err
	}

	return dashboard.DashboardConverter(object), "dashboard", nil
}

func dashboardv2Converter(resource *model.Resource) (string, string, error) {
	spec, err := resource.Spec()
	if err != nil {
		return "", "", err
	}

	marshalled, err := json.Marshal(spec)
	if err != nil {
		return "", "", err
	}

	object := dashboardv2beta1.Dashboard{}
	if err = json.Unmarshal(marshalled, &object); err != nil {
		return "", "", err
	}

	return dashboardv2beta1.DashboardConverter(object), "dashboardv2beta1", nil
}
