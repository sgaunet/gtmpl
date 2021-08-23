package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"github.com/sgaunet/gtmpl/gitlabRequest"
	"gopkg.in/yaml.v2"
)

const GitlabApiVersion = "v4"

type projectVariable struct {
	Key               string `yaml:"key"`
	Value             string `yaml:"value"`
	Protected         bool   `yaml:"protected"`
	Variable_type     string `yaml:"variable_type"`
	Masked            bool   `yaml:"masked"`
	Environment_scope string `yaml:"environment_scope"`
}

type projectVariableJSON struct {
	Key               string `json:"key"`
	Value             string `json:"value"`
	Protected         bool   `json:"protected"`
	Variable_type     string `json:"variable_type"`
	Masked            bool   `json:"masked"`
	Environment_scope string `json:"environment_scope"`
}

type configTmpl struct {
	Vars []projectVariable `yaml:"vars"`
}

func configTmplDir(tmpl string) string {
	return fmt.Sprintf("%s%s.gtmpl%s%s", os.Getenv("HOME"), string(os.PathSeparator), string(os.PathSeparator), tmpl)
}

func configTmplPath(tmpl string) string {
	return fmt.Sprintf("%s%v.gtmpl%s%s%sconfig.yaml", os.Getenv("HOME"), string(os.PathSeparator), string(os.PathSeparator), tmpl, string(os.PathSeparator))
}

func configTmplPathData(tmpl string) string {
	return fmt.Sprintf("%s%v.gtmpl%s%s%sdata", os.Getenv("HOME"), string(os.PathSeparator), string(os.PathSeparator), tmpl, string(os.PathSeparator))
}

func ExistTemplate(tmpl string) bool {
	tmplDir := configTmplDir(tmpl)
	fmt.Println(tmplDir)
	statDir, err := os.Stat(tmplDir)
	if err != nil {
		return false
	}
	if statDir.IsDir() {
		// Check also if config.yaml exists
		tmplConfig := configTmplPath(tmpl)
		fmt.Println(tmplConfig)
		f, err := os.Open(tmplConfig)
		if err != nil {
			return false
		} else {
			f.Close()
			return true
		}
	}
	return false
}

func LoadConfigFile(filename string) (configTmpl, error) {
	var yamlConfig configTmpl

	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading YAML file: %s\n", err)
		return yamlConfig, err
	}

	err = yaml.Unmarshal(yamlFile, &yamlConfig)
	if err != nil {
		fmt.Printf("Error parsing YAML file: %s\n", err)
		return yamlConfig, err
	}

	return yamlConfig, err
}

func createVariable(projectID int, v projectVariableJSON) error {
	url := fmt.Sprintf("%s/api/%s/projects/%d/variables", os.Getenv("GITLAB_URI"), GitlabApiVersion, projectID)
	json_data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", os.Getenv("GITLAB_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))

	return err
}

func updateVariable(projectID int, v projectVariableJSON) error {
	url := fmt.Sprintf("%s/api/%s/projects/%d/variables/%s", os.Getenv("GITLAB_URI"), GitlabApiVersion, projectID, v.Key)
	json_data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(json_data))
	if err != nil {
		return err
	}
	req.Header.Set("PRIVATE-TOKEN", os.Getenv("GITLAB_TOKEN"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	fmt.Println(string(body))

	return err
}

func getVariables(projectID int) ([]projectVariableJSON, error) {
	_, res, err := gitlabRequest.Request(fmt.Sprintf("/projects/%d/variables", projectID))
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	fmt.Println("----------------", string(res))
	var lstVars []projectVariableJSON
	err = json.Unmarshal(res, &lstVars)

	return lstVars, err
}

func isKeyAlreadyCreated(v projectVariableJSON, projectVars []projectVariableJSON) bool {
	for _, variable := range projectVars {
		if v.Key == variable.Key {
			return true
		}
	}
	return false
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil {
		err = os.Mkdir(dst, 0750)
		if err != nil {
			return err
		}
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}
