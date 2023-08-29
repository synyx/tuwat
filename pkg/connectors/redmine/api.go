package redmine

type response struct {
	Issues     []issue `json:"issues"`
	TotalCount int     `json:"total_count"`
	Limit      int     `json:"limit"`
}
type idField struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
type issue struct {
	Id          int     `json:"id"`
	Project     idField `json:"project"`
	Tracker     idField `json:"tracker"`
	Status      idField `json:"status"`
	Priority    idField `json:"priority"`
	Author      idField `json:"author"`
	AssignedTo  idField `json:"assigned_to"`
	Subject     string  `json:"subject"`
	Description string  `json:"description"`
	StartDate   string  `json:"start_date"`
	DueDate     string  `json:"due_date"`
	IsPrivate   bool    `json:"is_private"`
	CreatedOn   string  `json:"created_on"`
	UpdatedOn   string  `json:"updated_on"`
}
