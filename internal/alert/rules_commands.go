package alert

import (
	"context"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	cmdio "github.com/grafana/grafanactl/cmd/grafanactl/io"
	"github.com/grafana/grafanactl/internal/config"
	"github.com/grafana/grafanactl/internal/format"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// RESTConfigLoader can load a NamespacedRESTConfig from the active context.
type RESTConfigLoader interface {
	LoadRESTConfig(ctx context.Context) (config.NamespacedRESTConfig, error)
}

// rulesCommands returns the rules command group.
func rulesCommands(loader RESTConfigLoader) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage alert rules.",
	}
	cmd.AddCommand(
		newRulesListCommand(loader),
		newRulesGetCommand(loader),
		newRulesStatusCommand(loader),
	)
	return cmd
}

type rulesListOpts struct {
	IO        cmdio.Options
	GroupName string
	FolderUID string
	Status    string
}

func (o *rulesListOpts) setup(flags *pflag.FlagSet) {
	o.IO.RegisterCustomCodec("table", &rulesTableCodec{})
	o.IO.DefaultFormat("table")
	o.IO.BindFlags(flags)
	flags.StringVar(&o.GroupName, "group", "", "Filter by group name")
	flags.StringVar(&o.FolderUID, "folder", "", "Filter by folder UID")
	flags.StringVar(&o.Status, "status", "", "Filter by rule state (firing, pending, inactive)")
}

func newRulesListCommand(loader RESTConfigLoader) *cobra.Command {
	opts := &rulesListOpts{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List alert rules.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.IO.Validate(); err != nil {
				return err
			}

			ctx := cmd.Context()
			restCfg, err := loader.LoadRESTConfig(ctx)
			if err != nil {
				return err
			}

			client, err := NewClient(restCfg)
			if err != nil {
				return err
			}

			resp, err := client.List(ctx, ListOptions{
				GroupName: opts.GroupName,
				FolderUID: opts.FolderUID,
			})
			if err != nil {
				return err
			}

			codec, err := opts.IO.Codec()
			if err != nil {
				return err
			}

			if codec.Format() == "table" {
				var rules []RuleStatus
				for _, g := range resp.Data.Groups {
					for _, r := range g.Rules {
						if opts.Status == "" || r.State == opts.Status {
							rules = append(rules, r)
						}
					}
				}
				return codec.Encode(cmd.OutOrStdout(), rules)
			}

			if opts.Status != "" {
				for i := range resp.Data.Groups {
					var filtered []RuleStatus
					for _, r := range resp.Data.Groups[i].Rules {
						if r.State == opts.Status {
							filtered = append(filtered, r)
						}
					}
					resp.Data.Groups[i].Rules = filtered
				}
			}

			return codec.Encode(cmd.OutOrStdout(), resp.Data.Groups)
		},
	}
	opts.setup(cmd.Flags())
	return cmd
}

type rulesTableCodec struct{}

func (c *rulesTableCodec) Format() format.Format { return "table" }

func (c *rulesTableCodec) Encode(w io.Writer, v any) error {
	rules, ok := v.([]RuleStatus)
	if !ok {
		return errors.New("invalid data type for table codec: expected []RuleStatus")
	}

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "UID\tNAME\tSTATE\tHEALTH\tPAUSED")

	for _, r := range rules {
		paused := "no"
		if r.IsPaused {
			paused = "yes"
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", r.UID, r.Name, r.State, r.Health, paused)
	}

	return tw.Flush()
}

func (c *rulesTableCodec) Decode(r io.Reader, v any) error {
	return errors.New("table format does not support decoding")
}

type rulesGetOpts struct {
	IO cmdio.Options
}

func (o *rulesGetOpts) setup(flags *pflag.FlagSet) {
	o.IO.RegisterCustomCodec("table", &ruleDetailTableCodec{})
	o.IO.DefaultFormat("table")
	o.IO.BindFlags(flags)
}

//nolint:dupl // Similar structure to groups get command is intentional
func newRulesGetCommand(loader RESTConfigLoader) *cobra.Command {
	opts := &rulesGetOpts{}
	cmd := &cobra.Command{
		Use:   "get UID",
		Short: "Get a single alert rule.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.IO.Validate(); err != nil {
				return err
			}

			ctx := cmd.Context()
			uid := args[0]

			restCfg, err := loader.LoadRESTConfig(ctx)
			if err != nil {
				return err
			}

			client, err := NewClient(restCfg)
			if err != nil {
				return err
			}

			rule, err := client.GetRule(ctx, uid)
			if err != nil {
				return err
			}

			codec, err := opts.IO.Codec()
			if err != nil {
				return err
			}

			return codec.Encode(cmd.OutOrStdout(), rule)
		},
	}
	opts.setup(cmd.Flags())
	return cmd
}

// ruleDetailTableCodec renders a single rule as a table row.
type ruleDetailTableCodec struct{}

func (c *ruleDetailTableCodec) Format() format.Format { return "table" }

func (c *ruleDetailTableCodec) Encode(w io.Writer, v any) error {
	rule, ok := v.(*RuleStatus)
	if !ok {
		return errors.New("invalid data type for table codec: expected *RuleStatus")
	}

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "UID\tNAME\tSTATE\tHEALTH\tPAUSED")

	paused := "no"
	if rule.IsPaused {
		paused = "yes"
	}
	fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n", rule.UID, rule.Name, rule.State, rule.Health, paused)

	return tw.Flush()
}

func (c *ruleDetailTableCodec) Decode(r io.Reader, v any) error {
	return errors.New("table format does not support decoding")
}

type rulesStatusOpts struct {
	IO cmdio.Options
}

func (o *rulesStatusOpts) setup(flags *pflag.FlagSet) {
	o.IO.RegisterCustomCodec("table", &rulesStatusTableCodec{})
	o.IO.RegisterCustomCodec("wide", &rulesStatusTableCodec{Wide: true})
	o.IO.DefaultFormat("table")
	o.IO.BindFlags(flags)
}

func newRulesStatusCommand(loader RESTConfigLoader) *cobra.Command {
	opts := &rulesStatusOpts{}
	cmd := &cobra.Command{
		Use:   "status [UID]",
		Short: "Show alert rule status.",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := opts.IO.Validate(); err != nil {
				return err
			}

			ctx := cmd.Context()
			restCfg, err := loader.LoadRESTConfig(ctx)
			if err != nil {
				return err
			}

			client, err := NewClient(restCfg)
			if err != nil {
				return err
			}

			var rules []RuleStatus
			if len(args) == 1 {
				rule, err := client.GetRule(ctx, args[0])
				if err != nil {
					return err
				}
				rules = []RuleStatus{*rule}
			} else {
				resp, err := client.List(ctx, ListOptions{})
				if err != nil {
					return err
				}
				for _, g := range resp.Data.Groups {
					rules = append(rules, g.Rules...)
				}
			}

			if len(rules) == 0 {
				cmdio.Info(cmd.OutOrStdout(), "No alert rules found.")
				return nil
			}

			codec, err := opts.IO.Codec()
			if err != nil {
				return err
			}

			return codec.Encode(cmd.OutOrStdout(), rules)
		},
	}
	opts.setup(cmd.Flags())
	return cmd
}

type rulesStatusTableCodec struct {
	Wide bool
}

func (c *rulesStatusTableCodec) Format() format.Format {
	if c.Wide {
		return "wide"
	}
	return "table"
}

func (c *rulesStatusTableCodec) Encode(w io.Writer, v any) error {
	rules, ok := v.([]RuleStatus)
	if !ok {
		return errors.New("invalid data type for status table codec: expected []RuleStatus")
	}

	tw := tabwriter.NewWriter(w, 0, 4, 2, ' ', 0)

	if c.Wide {
		fmt.Fprintln(tw, "UID\tNAME\tSTATE\tHEALTH\tLAST_EVAL\tEVAL_TIME\tPAUSED\tFOLDER")
	} else {
		fmt.Fprintln(tw, "UID\tNAME\tSTATE\tHEALTH\tLAST_EVAL\tPAUSED")
	}

	for _, r := range rules {
		paused := "no"
		if r.IsPaused {
			paused = "yes"
		}
		lastEval := r.LastEvaluation
		if lastEval == "0001-01-01T00:00:00Z" {
			lastEval = "never"
		}

		if c.Wide {
			evalTime := fmt.Sprintf("%.3fs", r.EvaluationTime)
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				r.UID, r.Name, r.State, r.Health, lastEval, evalTime, paused, r.FolderUID)
		} else {
			fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\t%s\n",
				r.UID, r.Name, r.State, r.Health, lastEval, paused)
		}
	}

	return tw.Flush()
}

func (c *rulesStatusTableCodec) Decode(r io.Reader, v any) error {
	return errors.New("status table codec does not support decoding")
}
