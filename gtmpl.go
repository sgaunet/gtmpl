package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
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

var version string = "development"

func printVersion() {
	fmt.Println(version)
}

func main() {
	var tmpl, debugLevel, projectDir string
	var vOption, overwriteOption bool
	var err error
	// Parameters treatment
	flag.BoolVar(&vOption, "v", false, "Get version")
	flag.BoolVar(&overwriteOption, "f", false, "Overwrite files")
	flag.StringVar(&tmpl, "t", "", "Template Name")
	flag.StringVar(&projectDir, "p", "", "Project directory")
	flag.StringVar(&debugLevel, "d", "info", "debuglevel : debug/info/warn/error")
	flag.Parse()

	if vOption {
		printVersion()
		os.Exit(0)
	}

	initTrace(debugLevel)

	if len(tmpl) == 0 {
		log.Infoln("No template specified, will list templates")
		err := listTemplates()
		if err != nil {
			log.Errorln(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if !ExistTemplate(tmpl) {
		log.Errorf("Template named %s does not exist. Exit\n", tmpl)
		os.Exit(1)
	}

	if projectDir != "" {
		err := os.Chdir(projectDir)
		if err != nil {
			log.Errorf(err.Error())
			os.Exit(1)
		}
	}

	// gitFolder, err := findGitRepository()

	// if err != nil {
	// 	log.Errorf("Folder .git not found")
	// 	os.Exit(1)
	// }

	log.Infof("Will copy %s to %s\n", configTmplDir(tmpl), ".")
	err = CopyDir(configTmplDir(tmpl), ".", overwriteOption)
	if err != nil {
		log.Errorln(err.Error())
		os.Exit(1)
	}

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

func listTemplates() error {
	entries, err := ioutil.ReadDir(os.Getenv("HOME") + string(os.PathSeparator) + ".gtmpl")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != ".git" {
			log.Infoln("Template :", entry.Name())
		}
	}
	return err
}
