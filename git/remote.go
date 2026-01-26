package git

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/grokify/releasekit/run"
)

// RemoteInfo represents parsed remote URL information.
type RemoteInfo struct {
	Host  string // github.com
	Owner string // grokify
	Repo  string // releasekit
	URL   string // https://github.com/grokify/releasekit
}

// sshRe matches SSH remote URLs like git@github.com:owner/repo.git
var sshRe = regexp.MustCompile(`^[\w-]+@([\w.-]+):([\w._-]+)/([\w._-]+?)(?:\.git)?$`)

// httpsRe matches HTTPS remote URLs like https://github.com/owner/repo.git
var httpsRe = regexp.MustCompile(`^https?://([\w.-]+)/([\w._-]+)/([\w._-]+?)(?:\.git)?$`)

// ParseRemoteURL parses an SSH or HTTPS remote URL into structured info.
func ParseRemoteURL(rawURL string) (*RemoteInfo, error) {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return nil, fmt.Errorf("empty remote URL")
	}

	if matches := sshRe.FindStringSubmatch(rawURL); matches != nil {
		return &RemoteInfo{
			Host:  matches[1],
			Owner: matches[2],
			Repo:  matches[3],
			URL:   fmt.Sprintf("https://%s/%s/%s", matches[1], matches[2], matches[3]),
		}, nil
	}

	if matches := httpsRe.FindStringSubmatch(rawURL); matches != nil {
		return &RemoteInfo{
			Host:  matches[1],
			Owner: matches[2],
			Repo:  matches[3],
			URL:   fmt.Sprintf("https://%s/%s/%s", matches[1], matches[2], matches[3]),
		}, nil
	}

	return nil, fmt.Errorf("unrecognized remote URL format: %q", rawURL)
}

// RemoteURL returns the raw URL for the named remote.
func (r *Repo) RemoteURL(name string) (string, error) {
	cmd := run.Git(r.Dir, "remote", "get-url", name)
	result := r.Runner.Run(context.Background(), cmd)
	if !result.Passed() {
		return "", fmt.Errorf("git remote get-url %s: %s", name, result.Output())
	}
	return strings.TrimSpace(result.Stdout), nil
}

// ParseRemote fetches and parses the named remote URL.
func (r *Repo) ParseRemote(name string) (*RemoteInfo, error) {
	url, err := r.RemoteURL(name)
	if err != nil {
		return nil, err
	}
	return ParseRemoteURL(url)
}
