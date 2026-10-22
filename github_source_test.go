package selfupdate

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"sync/atomic"
	"testing"
	"time"

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

// fakeGitHubReleases serves a repository's release lists, one per request and the last one
// from then on, and its latest release: the JSON latest answered with status latestStatus.
// It returns how many times the list was asked for.
func fakeGitHubReleases(t *testing.T, lists []string, latestStatus int, latest string) (*GitHubSource, *atomic.Int32) {
	t.Helper()
	retryDelay := githubListRetryDelay
	githubListRetryDelay = 0
	t.Cleanup(func() { githubListRetryDelay = retryDelay })

	listed := &atomic.Int32{}
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/owner/repo/releases", func(w http.ResponseWriter, _ *http.Request) {
		n := int(listed.Add(1))
		_, _ = w.Write([]byte(lists[min(n, len(lists))-1]))
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
	return &GitHubSource{api: client}, listed
}

const (
	completeRelease  = `{"id":1,"tag_name":"v1.0.0","assets":[{"id":10,"name":"repo_linux_amd64.tar.gz"}]}`
	latestRelease    = `{"id":2,"tag_name":"v1.1.0","assets":[{"id":20,"name":"repo_linux_amd64.tar.gz"}]}`
	uploadingRelease = `{"id":2,"tag_name":"v1.1.0","assets":[]}`
	// published after latestRelease, as a fix of an older line
	backportRelease = `{"id":3,"tag_name":"v1.0.1","assets":[{"id":30,"name":"repo_linux_amd64.tar.gz"}]}`
)

func detectLatestLinuxAMD64(t *testing.T, source Source) (*Release, bool, error) {
	t.Helper()
	updater, err := NewUpdater(Config{Source: source, OS: "linux", Arch: "amd64"})
	require.NoError(t, err)
	return updater.DetectLatest(context.Background(), ParseSlug("owner/repo"))
}

// GitHub intermittently answers the release list with 200 and no releases.
// The latest release is the most recently published one, not the highest version.
func TestGitHubDetectsTheHighestReleaseWhenTheListIsEmptyOnce(t *testing.T) {
	source, listed := fakeGitHubReleases(t, []string{`[]`, `[` + backportRelease + `,` + latestRelease + `]`}, http.StatusOK, backportRelease)

	rel, found, err := detectLatestLinuxAMD64(t, source)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "1.1.0", rel.Version())
	assert.Equal(t, int32(2), listed.Load())
}

func TestGitHubDetectsTheLatestReleaseWhenTheListStaysEmpty(t *testing.T) {
	source, listed := fakeGitHubReleases(t, []string{`[]`}, http.StatusOK, latestRelease)

	rel, found, err := detectLatestLinuxAMD64(t, source)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, "1.1.0", rel.Version())
	assert.Equal(t, int64(20), rel.AssetID)
	assert.Equal(t, 1+githubListRetries, int(listed.Load()))
}

func TestGitHubListReleasesStopsRetryingWhenCancelled(t *testing.T) {
	source, _ := fakeGitHubReleases(t, []string{`[]`}, http.StatusOK, latestRelease)
	githubListRetryDelay = time.Hour
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := source.ListReleases(ctx, ParseSlug("owner/repo"))
	assert.ErrorIs(t, err, context.DeadlineExceeded)
}

// The latest release's assets are uploaded after it is published.
func TestGitHubDetectsTheNewestListedReleaseWhileTheLatestLacksAssets(t *testing.T) {
	source, _ := fakeGitHubReleases(t, []string{`[` + completeRelease + `]`}, http.StatusOK, uploadingRelease)

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
			source, _ := fakeGitHubReleases(t, []string{tc.list}, tc.latestStatus, tc.latest)

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
