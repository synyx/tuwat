package github

type issue struct {
	User        user        `json:"user"`
	Assignee    user        `json:"assignee"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
	Labels      []label     `json:"labels"`
	HTMLUrl     string      `json:"html_url"`
	Number      int         `json:"number"`
	Title       string      `json:"title"`
	Draft       bool        `json:"draft"`
	PullRequest pullRequest `json:"pull_request"`
}

type user struct {
	Login string `json:"login"`
}

type label struct {
	Name string `json:"name"`
}

type pullRequest struct {
	URL     string `json:"url"`
	HTMLUrl string `json:"html_url"`
}
