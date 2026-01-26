package git

import "testing"

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *RemoteInfo
		wantErr bool
	}{
		{
			"SSH standard",
			"git@github.com:grokify/releasekit.git",
			&RemoteInfo{Host: "github.com", Owner: "grokify", Repo: "releasekit", URL: "https://github.com/grokify/releasekit"},
			false,
		},
		{
			"SSH without .git",
			"git@github.com:grokify/releasekit",
			&RemoteInfo{Host: "github.com", Owner: "grokify", Repo: "releasekit", URL: "https://github.com/grokify/releasekit"},
			false,
		},
		{
			"HTTPS standard",
			"https://github.com/grokify/releasekit.git",
			&RemoteInfo{Host: "github.com", Owner: "grokify", Repo: "releasekit", URL: "https://github.com/grokify/releasekit"},
			false,
		},
		{
			"HTTPS without .git",
			"https://github.com/grokify/releasekit",
			&RemoteInfo{Host: "github.com", Owner: "grokify", Repo: "releasekit", URL: "https://github.com/grokify/releasekit"},
			false,
		},
		{
			"HTTP",
			"http://github.com/grokify/releasekit",
			&RemoteInfo{Host: "github.com", Owner: "grokify", Repo: "releasekit", URL: "https://github.com/grokify/releasekit"},
			false,
		},
		{
			"GitLab SSH",
			"git@gitlab.com:user/project.git",
			&RemoteInfo{Host: "gitlab.com", Owner: "user", Repo: "project", URL: "https://gitlab.com/user/project"},
			false,
		},
		{
			"empty",
			"",
			nil,
			true,
		},
		{
			"invalid",
			"not-a-url",
			nil,
			true,
		},
		{
			"owner with dots",
			"git@github.com:my.org/my.repo.git",
			&RemoteInfo{Host: "github.com", Owner: "my.org", Repo: "my.repo", URL: "https://github.com/my.org/my.repo"},
			false,
		},
		{
			"owner with hyphens",
			"https://github.com/my-org/my-repo.git",
			&RemoteInfo{Host: "github.com", Owner: "my-org", Repo: "my-repo", URL: "https://github.com/my-org/my-repo"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRemoteURL(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got.Host != tt.want.Host {
				t.Errorf("Host = %q, want %q", got.Host, tt.want.Host)
			}
			if got.Owner != tt.want.Owner {
				t.Errorf("Owner = %q, want %q", got.Owner, tt.want.Owner)
			}
			if got.Repo != tt.want.Repo {
				t.Errorf("Repo = %q, want %q", got.Repo, tt.want.Repo)
			}
			if got.URL != tt.want.URL {
				t.Errorf("URL = %q, want %q", got.URL, tt.want.URL)
			}
		})
	}
}
