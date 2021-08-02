package gitlabProject

type GitlabProject interface {
	GetName() string
	GetID() int
}

type gitlabProject struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}
