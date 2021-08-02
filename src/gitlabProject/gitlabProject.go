package gitlabProject

import (
	"encoding/json"
	"fmt"

	"github.com/sgaunet/gtmpl/gitlabRequest"
)

func New(projectID int) (res gitlabProject, err error) {
	var project gitlabProject
	url := fmt.Sprintf("projects/%d", projectID)
	_, body, err := gitlabRequest.Request(url)
	if err != nil {
		return res, err
	}
	if err := json.Unmarshal(body, &project); err != nil {
		return res, err
	}
	return project, err
}

func (p gitlabProject) GetID() int {
	return p.Id
}

func (p gitlabProject) GetName() string {
	return p.Name
}

// https://docs.gitlab.com/ee/api/project_level_variables.html
