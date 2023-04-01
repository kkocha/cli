package list

import (
	"fmt"
	"net/http"

	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/internal/tableprinter"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/spf13/cobra"
)

type ListOptions struct {
	HttpClient   func() (*http.Client, error)
	IO           *iostreams.IOStreams
	Config       func() (config.Config, error)
	LimitResults int
}

func NewCmdList(f *cmdutil.Factory, runF func(*ListOptions) error) *cobra.Command {
	opts := &ListOptions{
		HttpClient: f.HttpClient,
		IO:         f.IOStreams,
		Config:     f.Config,
	}

	cmd := &cobra.Command{
		Use:     "list",
		Short:   "List organizations of which user is a member or admin",
		Aliases: []string{"ls"},
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if opts.LimitResults < 1 {
				return cmdutil.FlagErrorf("invalid value for --limit: %v", opts.LimitResults)
			}

			if runF != nil {
				return runF(opts)
			}
			return listRun(opts)
		},
	}

	cmd.Flags().IntVarP(&opts.LimitResults, "limit", "L", 30, "Maximum number of items to fetch")

	return cmd
}

func listRun(opts *ListOptions) error {
	httpClient, err := opts.HttpClient()
	if err != nil {
		return err
	}

	cfg, err := opts.Config()
	if err != nil {
		return err
	}
	authCfg := cfg.Authentication()
	hostname, _ := authCfg.DefaultHost()

	organizations, err := listOrganizations(httpClient, hostname, opts.LimitResults)
	if err != nil {
		return err
	}

	if len(organizations) == 0 {
		return cmdutil.NewNoResultsError("no organizations found")
	}

	err = opts.IO.StartPager()
	if err != nil {
		fmt.Fprintf(opts.IO.ErrOut, "error starting pager: %v\n", err)
	}
	defer opts.IO.StopPager()

	table := tableprinter.New(opts.IO)
	for _, org := range organizations {
		table.AddField(org.Login)
		table.EndRow()
	}
	err = table.Render()
	if err != nil {
		return err
	}

	return nil
}
