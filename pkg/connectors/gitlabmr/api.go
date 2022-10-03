package gitlabmr

type mergeRequest struct {
	Title      string     `json:"title"`
	Labels     []string   `json:"labels"`
	UpdatedAt  string     `json:"updated_at"`
	Author     person     `json:"author"`
	Assignee   person     `json:"assignee"`
	WebUrl     string     `json:"web_url"`
	References references `json:"references"`
	Milestone  milestone  `json:"milestone"`
}

type person struct {
	Name string `json:"name"`
}

type references struct {
	Short    string `json:"short"`
	Relative string `json:"relative"`
	Full     string `json:"full"`
}

type milestone struct {
	Title string `json:"title"`
}
