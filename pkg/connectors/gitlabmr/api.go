package gitlabmr

type Alert struct {
	Title      string     `json:"title"`
	Labels     []string   `json:"labels"`
	UpdatedAt  string     `json:"updated_at"`
	Author     Person     `json:"author"`
	Assignee   Person     `json:"assignee"`
	WebUrl     string     `json:"web_url"`
	References References `json:"references"`
	Milestone  Milestone  `json:"milestone"`
}

type Person struct {
	Name string `json:"name"`
}

type References struct {
	Short    string `json:"short"`
	Relative string `json:"relative"`
	Full     string `json:"full"`
}

type Milestone struct {
	Title string `json:"title"`
}
