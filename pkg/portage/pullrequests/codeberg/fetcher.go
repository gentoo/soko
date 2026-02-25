// SPDX-License-Identifier: GPL-2.0-only
package codeberg

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"iter"
	"log/slog"
	"net/http"
	"net/url"
	"soko/pkg/config"
	"soko/pkg/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

func FetchPullRequests() iter.Seq[models.PullRequestProvider] {
	const pageSize = 50
	client, err := NewClient("https://codeberg.org")
	if err != nil {
		slog.Error("Failed to create Codeberg client", slog.Any("err", err))
		return func(yield func(models.PullRequestProvider) bool) {}
	}

	return func(yield func(models.PullRequestProvider) bool) {
		for page := 1; ; page++ {
			prs, err := client.listOpenPulls(page, pageSize)
			if err != nil {
				slog.Error("Failed to list open pulls", slog.Int("page", page), slog.Any("err", err))
				return
			}
			if len(prs) == 0 {
				return
			}

			// Fetch CI status for all PRs in this page concurrently, up to fetchConcurrency at a time.
			type ciResult struct {
				state string
				link  string
			}
			results := make([]ciResult, len(prs))
			files := make([][]apiPRFile, len(prs))

			var wg sync.WaitGroup
			for index, pr := range prs {
				if sha := pr.Head.Sha; sha != "" {
					wg.Go(func() {
						s, link, err := client.getLatestCIStatus(sha)
						if err != nil {
							slog.Error("Failed to get latest CI status", slog.Int("pr", int(pr.Number)), slog.String("sha", sha), slog.Any("err", err))
							return
						}
						results[index] = ciResult{state: s, link: link}
					})
				}
				wg.Go(func() {
					const pageSize = 50
					var allFiles []apiPRFile
					for page := 1; ; page++ {
						files, err := client.listPRFiles(int(pr.Number), page, pageSize)
						if err != nil {
							slog.Error("Failed to list PR files", slog.Int("pr", int(pr.Number)), slog.Int("page", page), slog.Any("err", err))
							break
						}
						allFiles = append(allFiles, files...)
						if len(files) < pageSize {
							break
						}
					}
					files[index] = allFiles
				})
			}
			wg.Wait()

			for i, pr := range prs {
				if !yield(&codebergPRProvider{
					prPayload: pr,
					files:     files[i],
					ciState:   results[i].state,
					ciLink:    results[i].link,
				}) {
					return
				}
			}

			if len(prs) < pageSize {
				return
			}
		}
	}
}

type codebergPRProvider struct {
	prPayload apiPullRequest
	files     []apiPRFile
	ciState   string
	ciLink    string
}

func (p *codebergPRProvider) ToPullRequest() *models.PullRequest {
	pr := p.prPayload

	labels := make([]models.PullRequestLabel, len(pr.Labels))
	for i, l := range pr.Labels {
		labels[i] = models.PullRequestLabel{Name: l.Name, Color: l.Color}
	}

	return &models.PullRequest{
		Id:          "codeberg/" + strconv.FormatInt(pr.Number, 10),
		Closed:      strings.EqualFold(pr.State, "closed"),
		Url:         pr.HTMLURL,
		Title:       pr.Title,
		CreatedAt:   pr.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   pr.UpdatedAt.Format(time.RFC3339),
		CiState:     p.ciState,
		CiStateLink: p.ciLink,
		Labels:      labels,
		Comments:    pr.Comments,
		Author:      pr.User.Login,
	}
}

func (p *codebergPRProvider) GetFiles() iter.Seq[string] {
	return func(yield func(string) bool) {
		for _, f := range p.files {
			if f.Filename != "" {
				if !yield(f.Filename) {
					return
				}
			}
		}
	}
}

type client struct {
	Token string
	HTTP  *http.Client
	Rate  *rate.Limiter
}

func NewClient(baseURL string) (*client, error) {
	token := config.CodebergAPIToken()
	if token == "" {
		return nil, fmt.Errorf("API token is required")
	}
	return &client{
		Token: token,
		Rate:  rate.NewLimiter(rate.Every(100*time.Millisecond), 10),
		HTTP: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *client) newRequest(method, path string, q url.Values) (*http.Request, error) {
	url := "https://codeberg.org" + path + "?" + q.Encode()
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "token "+c.Token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", config.UserAgent())
	return req, nil
}

func (c *client) doJSON(req *http.Request, out any) error {
	_ = c.Rate.Wait(context.Background())
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 128))
		return fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(b)))
	}
	if out == nil {
		return nil
	}
	dec := json.NewDecoder(resp.Body)
	return dec.Decode(out)
}

func (c *client) listOpenPulls(page, limit int) ([]apiPullRequest, error) {
	req, err := c.newRequest(http.MethodGet, "/api/v1/repos/gentoo/gentoo/pulls", url.Values{
		"state": []string{"open"},
		"page":  []string{strconv.Itoa(page)},
		"limit": []string{strconv.Itoa(limit)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create request for listing open pulls: %w", err)
	}

	var out []apiPullRequest
	if err := c.doJSON(req, &out); err != nil {
		return nil, fmt.Errorf("failed to list open pulls: %w", err)
	}
	return out, nil
}

func (c *client) listPRFiles(prNumber int, page, limit int) ([]apiPRFile, error) {
	path := fmt.Sprintf("/api/v1/repos/gentoo/gentoo/pulls/%d/files", prNumber)
	req, err := c.newRequest(http.MethodGet, path, url.Values{
		"page":  []string{strconv.Itoa(page)},
		"limit": []string{strconv.Itoa(limit)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create request for listing pr files: %w", err)
	}

	var out []apiPRFile
	if err := c.doJSON(req, &out); err != nil {
		return nil, fmt.Errorf("failed to list pr files: %w", err)
	}
	return out, nil
}

func (c *client) listCommitStatuses(sha string, page, limit int) ([]apiCommitStatus, error) {
	path := fmt.Sprintf("/api/v1/repos/gentoo/gentoo/statuses/%s", url.PathEscape(sha))
	req, err := c.newRequest(http.MethodGet, path, url.Values{
		"page":  []string{strconv.Itoa(page)},
		"limit": []string{strconv.Itoa(limit)},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create request for listing commit statuses: %w", err)
	}

	var out []apiCommitStatus
	if err := c.doJSON(req, &out); err != nil {
		return nil, fmt.Errorf("failed to list commit statuses: %w", err)
	}
	return out, nil
}

func (c *client) getLatestCIStatus(sha string) (string, string, error) {
	const pageSize = 50

	var newest *apiCommitStatus

	for page := 1; ; page++ {
		statuses, err := c.listCommitStatuses(sha, page, pageSize)
		if err != nil {
			return "", "", err
		}
		if len(statuses) == 0 {
			break
		}

		for _, s := range statuses {
			if s.Context != "gentoo-ci" {
				continue
			}
			// pick newest by updated/created time
			t := s.UpdatedAt
			if t.IsZero() {
				t = s.CreatedAt
			}
			if newest == nil {
				cp := s
				newest = &cp
				continue
			}
			nt := newest.UpdatedAt
			if nt.IsZero() {
				nt = newest.CreatedAt
			}
			if t.After(nt) {
				cp := s
				newest = &cp
			}
		}

		if len(statuses) < pageSize {
			break
		}
	}

	if newest == nil {
		return "", "", nil
	}
	return strings.ToUpper(newest.Status), newest.TargetURL, nil
}
