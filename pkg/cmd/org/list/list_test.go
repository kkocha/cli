package list

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/MakeNowJust/heredoc"
	"github.com/cli/cli/v2/internal/config"
	"github.com/cli/cli/v2/pkg/cmdutil"
	"github.com/cli/cli/v2/pkg/httpmock"
	"github.com/cli/cli/v2/pkg/iostreams"
	"github.com/google/shlex"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_NewCmdList(t *testing.T) {
	tests := []struct {
		name     string
		args     string
		want     ListOptions
		wantsErr string
	}{
		{
			name: "no arguments",
			args: "",
			want: ListOptions{
				LimitResults: 30,
			},
		},
		{
			name: "with limit",
			args: "--limit 101",
			want: ListOptions{
				LimitResults: 101,
			},
		},
		{
			name:     "invalid limit",
			args:     "-L 0",
			wantsErr: "invalid value for --limit: 0",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, _, _ := iostreams.Test()

			f := &cmdutil.Factory{
				IOStreams: ios,
			}

			var opts *ListOptions
			cmd := NewCmdList(f, func(o *ListOptions) error {
				opts = o
				return nil
			})

			argv, err := shlex.Split(tt.args)
			require.NoError(t, err)
			cmd.SetArgs(argv)
			cmd.SetIn(&bytes.Buffer{})
			cmd.SetOut(io.Discard)
			cmd.SetErr(io.Discard)

			_, err = cmd.ExecuteC()
			if tt.wantsErr != "" {
				require.EqualError(t, err, tt.wantsErr)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.want.LimitResults, opts.LimitResults)
		})
	}
}

func Test_listRun(t *testing.T) {
	tests := []struct {
		name        string
		opts        ListOptions
		wantsErr    string
		wantsStdout string
		wantsStderr string
	}{
		{
			name: "list organizations",
			opts: ListOptions{
				LimitResults: 30,
				Config: func() (config.Config, error) {
					cfg := &config.ConfigMock{}
					cfg.AuthenticationFunc = func() *config.AuthConfig {
						authCfg := &config.AuthConfig{}
						authCfg.SetHosts([]string{})
						return authCfg
					}
					return cfg, nil
				},
			},
			wantsStdout: heredoc.Doc(`
				cli
				github
				myorganization
			`),
			wantsStderr: ``,
		},
		{
			name: "list one organization with limit",
			opts: ListOptions{
				LimitResults: 1,
				Config: func() (config.Config, error) {
					cfg := &config.ConfigMock{}
					cfg.AuthenticationFunc = func() *config.AuthConfig {
						authCfg := &config.AuthConfig{}
						authCfg.SetHosts([]string{})
						return authCfg
					}
					return cfg, nil
				},
			},
			wantsStdout: heredoc.Doc(`
				cli
			`),
			wantsStderr: ``,
		},
		{
			name: "no organizations",
			opts: ListOptions{
				LimitResults: 30,
				Config: func() (config.Config, error) {
					cfg := &config.ConfigMock{}
					cfg.AuthenticationFunc = func() *config.AuthConfig {
						authCfg := &config.AuthConfig{}
						authCfg.SetHosts([]string{})
						return authCfg
					}
					return cfg, nil
				},
			},
			wantsStdout: ``,
			wantsErr:    `no organizations found`,
			wantsStderr: ``,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ios, _, stdout, stderr := iostreams.Test()

			fakeHTTP := &httpmock.Registry{}

			stringResponse := `
				{
					"data": {
						"viewer": {
							"organizations": {
								"nodes": [
									{
										"login": "cli"
									},
									{
										"login": "github"
									},
									{
										"login": "myorganization"
									}
								]
				} } } }`

			if tt.wantsErr == "no organizations found" {
				stringResponse = `
					{
						"data": {
							"viewer": {
								"organizations": {
									"nodes": []
					} } } }`
			}

			fakeHTTP.Register(httpmock.GraphQL(`OrganizationSearch\b`), httpmock.StringResponse(stringResponse))

			tt.opts.IO = ios
			tt.opts.HttpClient = func() (*http.Client, error) {
				return &http.Client{Transport: fakeHTTP}, nil
			}

			err := listRun(&tt.opts)

			if tt.wantsErr != "" {
				require.EqualError(t, err, tt.wantsErr)
				return
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, tt.wantsStdout, stdout.String())
			assert.Equal(t, tt.wantsStderr, stderr.String())
		})
	}

}
