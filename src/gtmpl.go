package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/sgaunet/gtmpl/gitlabRequest"
	"gopkg.in/ini.v1"
)

type whoami struct {
	Id int `json:id`
}

type projects struct {
	Items []project
}

type project struct {
	Id            int    `json:"id"`
	Name          string `json:"name"`
	SshUrlToRepo  string `json:"ssh_url_to_repo"`
	HttpUrlToRepo string `json:"http_url_to_repo"`
}

func main() {
	// Arguments
	var tmpl string
	// Parameters treatment
	flag.StringVar(&tmpl, "t", "default", "Template Name")
	flag.Parse()

	if len(tmpl) == 0 {
		fmt.Println("Template name cannot be empty")
		os.Exit(1)
	}

	if !ExistTemplate(tmpl) {
		fmt.Printf("Template named %s does not exist. Exit\n", tmpl)
		os.Exit(1)
	}

	if len(os.Getenv("GITLAB_TOKEN")) == 0 {
		fmt.Println("Set GITLAB_TOKEN environment variable")
		os.Exit(1)
	}

	if len(os.Getenv("GITLAB_URI")) == 0 {
		os.Setenv("GITLAB_URI", "https://gitlab.com")
	}

	// log.Debugf("GITLAB_TOKEN=%s\n", os.Getenv("GITLAB_TOKEN"))
	// log.Debugf("GITLAB_URI=%s\n", os.Getenv("GITLAB_URI"))

	gitFolder, err := findGitRepository()

	if err != nil {
		fmt.Fprintf(os.Stderr, "Folder .git not found")
		os.Exit(1)
	}

	remoteOrigin := GetRemoteOrigin(gitFolder + string(os.PathSeparator) + ".git" + string(os.PathSeparator) + "config")

	_, res, err := gitlabRequest.Request("user")
	if err != nil {
		fmt.Println("...")
		os.Exit(1)
	}

	fmt.Println(string(res))
	var w whoami
	err = json.Unmarshal(res, &w)
	fmt.Println("ID", w.Id)
	if err != nil {
		fmt.Println("...")
		os.Exit(1)
	}

	id := strconv.Itoa(w.Id)
	fmt.Println("users/" + id + "/projects")
	// _, res, err = gitlabRequest.Request("users/" + id + "/projects?pagination=keyset&per_page=100&order_by=id&sort=asc")

	project, err := findProject(remoteOrigin)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("===>", project.SshUrlToRepo)
	cfg, err := LoadConfigFile(configTmplPath(tmpl))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("getVariables")
	projectVars, err := getVariables(project.Id)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("projectVars=", projectVars)

	for _, v := range cfg.Vars {
		fmt.Println(v.Key)
		var newV projectVariableJSON
		newV.Environment_scope = v.Environment_scope
		newV.Key = v.Key
		newV.Masked = v.Masked
		newV.Value = v.Value
		newV.Variable_type = v.Variable_type
		if !isKeyAlreadyCreated(newV, projectVars) {

			createVariable(project.Id, newV)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		} else {
			updateVariable(project.Id, newV)
		}
	}

	fmt.Printf("Need to copy %s to %s\n", configTmplPathData(tmpl), gitFolder)
	err = CopyDir(configTmplPathData(tmpl), gitFolder)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func findProject(remoteOrigin string) (project, error) {
	projectName := filepath.Base(remoteOrigin)
	projectName = strings.ReplaceAll(projectName, ".git", "")
	fmt.Println("Try to find ", projectName)

	_, res, err := gitlabRequest.Request("search?scope=projects&search=" + projectName)
	if err != nil {
		fmt.Println("...")
		os.Exit(1)
	}

	var p []project
	err = json.Unmarshal(res, &p)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for _, project := range p {
		fmt.Println(project.Name)
		fmt.Println(project.Id)
		fmt.Println(project.HttpUrlToRepo)
		fmt.Println(project.SshUrlToRepo)

		if project.SshUrlToRepo == remoteOrigin {
			return project, err
		}
	}
	return project{}, errors.New("Project not found")
}

func findGitRepository() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for cwd != "/" {
		fmt.Println(cwd)
		stat, err := os.Stat(cwd + string(os.PathSeparator) + ".git")
		if err == nil {
			if stat.IsDir() {
				return cwd, err
			}
		}
		cwd = filepath.Dir(cwd)
	}
	return "", errors.New(".git not found")
}

func GetRemoteOrigin(gitConfigFile string) string {
	cfg, err := ini.Load(gitConfigFile)
	if err != nil {
		fmt.Printf("Fail to read file: %v", err)
		os.Exit(1)
	}

	url := cfg.Section("remote \"origin\"").Key("url").String()
	fmt.Println("url:", url)
	return url
}
