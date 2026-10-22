package selfupdate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"

	"github.com/google/go-github/v86/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubTokenEnv(t *testing.T) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("because $GITHUB_TOKEN is not set")
	}

	if _, err := NewGitHubSource(GitHubConfig{}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
	if _, err := NewGitHubSource(GitHubConfig{APIToken: token}); err != nil {
		t.Error("Failed to initialize GitHub source with API token config")
	}
}

func TestGitHubTokenIsNotSet(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")

	if _, err := NewGitHubSource(GitHubConfig{}); err != nil {
		t.Error("Failed to initialize GitHub source with empty config")
	}
}

func TestGitHubEnterpriseClientInvalidURL(t *testing.T) {
	_, err := NewGitHubSource(GitHubConfig{APIToken: "my_token", EnterpriseBaseURL: ":this is not a URL"})
	if err == nil {
		t.Fatal("Invalid URL should raise an error")
	}
}

func TestGitHubEnterpriseClientValidURL(t *testing.T) {
	_, err := NewGitHubSource(GitHubConfig{APIToken: "my_token", EnterpriseBaseURL: "http://localhost"})
	if err != nil {
		t.Fatal("Failed to initialize GitHub source with valid URL")
	}
}

func TestGitHubListReleasesContextCancelled(t *testing.T) {
	source, err := NewGitHubSource(GitHubConfig{})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.ListReleases(ctx, ParseSlug("creativeprojects/resticprofile"))
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGitHubDownloadReleaseAssetContextCancelled(t *testing.T) {
	source, err := NewGitHubSource(GitHubConfig{})
	require.NoError(t, err)

	ctx, cancelFn := context.WithCancel(context.Background())
	cancelFn()

	_, err = source.DownloadReleaseAsset(ctx, &Release{repository: ParseSlug("creativeprojects/resticprofile")}, 11)
	assert.ErrorIs(t, err, context.Canceled)
}

func TestGitHubDownloadReleaseAssetWithNilRelease(t *testing.T) {
	source, err := NewGitHubSource(GitHubConfig{})
	require.NoError(t, err)

	_, err = source.DownloadReleaseAsset(context.Background(), nil, 11)
	assert.ErrorIs(t, err, ErrInvalidRelease)
}

// fakeGitHubReleases serves a repository's release list and its latest release:
// latest is the JSON answered with status latestStatus.
func fakeGitHubReleases(t *testing.T, list string, latestStatus int, latest string) *GitHubSource {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(list))
	})
	mux.HandleFunc("/repos/owner/repo/releases/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(latestStatus)
		_, _ = w.Write([]byte(latest))
	})
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	client := github.NewClient(nil)
	baseURL, err := url.Parse(server.URL + "/")
	require.NoError(t, err)
	client.BaseURL = baseURL
	return &GitHubSource{api: client}
}

const (
	completeRelease  = `{"id":1,"tag_name":"v1.0.0","assets":[{"id":10,"name":"repo_linux_amd64.tar.gz"}]}`
	latestRelease    = `{"id":2,"tag_name":"v1.1.0","assets":[{"id":20,"name":"repo_linux_amd64.tar.gz"}]}`
	uploadingRelease = `{"id":2,"tag_name":"v1.1.0","assets":[]}`
)

func detectLatestLinuxAMD64(t *testing.T, source Source) (*Release, bool, error) {
	t.Helper()
	updater, err := NewUpdater(Config{Source: source, OS: "linux", Arch: "amd64"})
	require.NoError(t, err)
	return updater.DetectLatest(context.Background(), ParseSlug("owner/repo"))
}

// GitHub intermittently answers the release list with 200 and no releases.
func TestGitHubDetectsTheLatestReleaseWhenTheListIsEmpty(t *testing.T) {
	source := fakeGitHubReleases(t, `[]`, http.StatusOK, latestRelease)

	rel, found, err := detectLatestLinuxAMD64(t, source)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "1.1.0", rel.Version())
	assert.Equal(t, int64(20), rel.AssetID)
}

// The latest release's assets are uploaded after it is published.
func TestGitHubDetectsTheNewestListedReleaseWhileTheLatestLacksAssets(t *testing.T) {
	source := fakeGitHubReleases(t, `[`+completeRelease+`]`, http.StatusOK, uploadingRelease)

	rel, found, err := detectLatestLinuxAMD64(t, source)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "1.0.0", rel.Version())
}

func TestGitHubListReleasesWithTheLatestRelease(t *testing.T) {
	for _, tc := range []struct {
		name         string
		list         string
		latestStatus int
		latest       string
		want         []string
		wantErr      bool
	}{
		{"latest added", `[` + completeRelease + `]`, http.StatusOK, latestRelease, []string{"v1.0.0", "v1.1.0"}, false},
		{"latest listed once", `[` + latestRelease + `,` + completeRelease + `]`, http.StatusOK, latestRelease, []string{"v1.1.0", "v1.0.0"}, false},
		{"no release yet", `[]`, http.StatusNotFound, `{"message":"Not Found"}`, []string{}, false},
		{"latest fails, the list stands", `[` + completeRelease + `]`, http.StatusBadGateway, `{}`, []string{"v1.0.0"}, false},
		{"latest fails, nothing listed", `[]`, http.StatusBadGateway, `{}`, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := fakeGitHubReleases(t, tc.list, tc.latestStatus, tc.latest)

			rels, err := source.ListReleases(context.Background(), ParseSlug("owner/repo"))
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			tags := []string{}
			for _, rel := range rels {
				tags = append(tags, rel.GetTagName())
			}
			assert.Equal(t, tc.want, tags)
		})
	}
}
