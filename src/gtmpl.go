package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sgaunet/gtmpl/gitlabRequest"
	log "github.com/sirupsen/logrus"
	"gopkg.in/ini.v1"
)

func initTrace(debugLevel string) {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})
	// log.SetFormatter(&log.TextFormatter{
	// 	DisableColors: true,
	// 	FullTimestamp: true,
	// })

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	switch debugLevel {
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	default:
		log.SetLevel(log.DebugLevel)
	}
}

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

var version string = "development"

func printVersion() {
	fmt.Println(version)
}

func main() {
	// Arguments
	var tmpl, debugLevel, projectDir string
	var vOption, overwriteOption bool
	// Parameters treatment
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.BoolVar(&overwriteOption, "f", false, "Overwrite files")
	flag.StringVar(&tmpl, "t", "default", "Template Name")
	flag.StringVar(&projectDir, "p", "", "Project directory")
	flag.StringVar(&debugLevel, "d", "info", "debuglevel : debug/info/warn/error")
	flag.Parse()

	if vOption {
		printVersion()
		os.Exit(0)
	}

	initTrace(debugLevel)

	if len(tmpl) == 0 {
		log.Errorln("Template name cannot be empty")
		os.Exit(1)
	}

	if !ExistTemplate(tmpl) {
		log.Errorf("Template named %s does not exist. Exit\n", tmpl)
		os.Exit(1)
	}

	if len(os.Getenv("GITLAB_TOKEN")) == 0 {
		log.Errorf("Set GITLAB_TOKEN environment variable")
		os.Exit(1)
	}

	if len(os.Getenv("GITLAB_URI")) == 0 {
		os.Setenv("GITLAB_URI", "https://gitlab.com")
	}

	// log.Debugf("GITLAB_TOKEN=%s\n", os.Getenv("GITLAB_TOKEN"))
	// log.Debugf("GITLAB_URI=%s\n", os.Getenv("GITLAB_URI"))
	if projectDir != "" {
		err := os.Chdir(projectDir)
		if err != nil {
			log.Errorf(err.Error())
			os.Exit(1)
		}
	}

	gitFolder, err := findGitRepository()

	if err != nil {
		log.Errorf("Folder .git not found")
		os.Exit(1)
	}

	log.Infof("Will copy %s to %s\n", configTmplPathData(tmpl), gitFolder)
	err = CopyDir(configTmplPathData(tmpl), gitFolder, overwriteOption)
	if err != nil {
		log.Errorln(err.Error())
		os.Exit(1)
	}

	remoteOrigin := GetRemoteOrigin(gitFolder + string(os.PathSeparator) + ".git" + string(os.PathSeparator) + "config")

	// Infos on gitlab user
	_, res, err := gitlabRequest.Request("user")
	if err != nil {
		log.Errorln(err.Error())
		os.Exit(1)
	}

	if !doesConfigFileExists(tmpl) {
		log.Infoln("No config.yaml file in this template (no CI vars to init in gitlab project)")
		os.Exit(0)
	}
	cfg, err := LoadConfigFile(configTmplPath(tmpl))
	if err != nil {
		log.Errorln(err.Error())
		os.Exit(0)
	}

	// log.Debugln(string(res))
	var w whoami
	err = json.Unmarshal(res, &w)
	if err != nil {
		log.Errorln(err.Error())
		os.Exit(1)
	}
	log.Debugln("ID", w.Id)

	if len(cfg.Vars) == 0 {
		log.Infoln("No CI vars to init")
	} else {
		project, err := findProject(remoteOrigin)
		if err != nil {
			log.Warnln(err.Error())
		} else {
			log.Infoln("Project found: ", project.SshUrlToRepo)
			// Get project vars
			projectVars, err := getVariables(project.Id)
			if err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}

			log.Debugln("projectVars=", projectVars)

			for _, v := range cfg.Vars {
				log.Debugln(v.Key)
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
		}
	}
}

func findProject(remoteOrigin string) (project, error) {
	projectName := filepath.Base(remoteOrigin)
	projectName = strings.ReplaceAll(projectName, ".git", "")
	log.Infof("Try to find project %s in %s\n", projectName, os.Getenv("GITLAB_URI"))

	_, res, err := gitlabRequest.Request("search?scope=projects&search=" + projectName)
	if err != nil {
		log.Errorln(err.Error())
		os.Exit(1)
	}

	var p []project
	err = json.Unmarshal(res, &p)
	if err != nil {
		log.Errorln(err.Error())
		os.Exit(1)
	}

	for _, project := range p {
		log.Debugln(project.Name)
		log.Debugln(project.Id)
		log.Debugln(project.HttpUrlToRepo)
		log.Debugln(project.SshUrlToRepo)

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
		log.Debugln(cwd)
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
		log.Errorf("Fail to read file: %v", err)
		os.Exit(1)
	}

	url := cfg.Section("remote \"origin\"").Key("url").String()
	log.Debugln("url:", url)
	return url
}
