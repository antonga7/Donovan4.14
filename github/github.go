package github

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type User struct {
	Login string `json:"login"`
}

type Milestone struct {
	Title  string `json:"title"`
	Number int    `json:"number"`
}

type Issue struct {
	Title     string     `json:"title"`
	Number    int        `json:"number"`
	HTMLURL   string     `json:"html_url"`
	Body      string     `json:"body"`
	User      *User      `json:"user"`
	Milestone *Milestone `json:"milestone"`
}

type IssuesSearchResult struct {
	TotalCount int      `json:"total_count"`
	Items      []*Issue `json:"items"`
}

func SearchIssues(terms []string, page int) (*IssuesSearchResult, error) {
	query := "repo:golang/go " + strings.Join(terms, " ")

	apiURL := fmt.Sprintf("https://api.github.com/search/issues?q=%s&page=%d", url.QueryEscape(query), page)

	resp, err := http.Get(apiURL)
	if err != nil {
		return nil, err
	}

	fmt.Println("URL:", apiURL)
	fmt.Println("STATUS:", resp.Status)

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {

		return nil, fmt.Errorf("bad response: %s", resp.Status)
	}

	var result IssuesSearchResult

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
