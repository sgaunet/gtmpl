package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

func configTmplDir(tmpl string) string {
	return fmt.Sprintf("%s%s.gtmpl%s%s", os.Getenv("HOME"), string(os.PathSeparator), string(os.PathSeparator), tmpl)
}

func ExistTemplate(tmpl string) bool {
	tmplDir := configTmplDir(tmpl)
	// fmt.Println(tmplDir)
	statDir, err := os.Stat(tmplDir)
	if err != nil {
		return false
	}
	if statDir.IsDir() {
		return true
	}
	return false
}

// CopyDir recursively copies a directory tree, attempting to preserve permissions.
// Source directory must exist, destination directory must exist.
// Symlinks are ignored and skipped.
func CopyDir(src string, dst string, overwrite bool) (err error) {
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
		log.Infoln("Create directory:", dst)
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath, overwrite)
			if err != nil {
				return err
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = CopyFile(srcPath, dstPath, overwrite)
			if err != nil {
				return err
			}
		}
	}

	return err
}

// CopyFile copies the contents of the file named src to the file named
// by dst. The file will be created if it does not already exist. If the
// destination file exists, all it's contents will be replaced by the contents
// of the source file. The file mode will be copied from the source and
// the copied data is synced/flushed to stable storage.
func CopyFile(src, dst string, overwrite bool) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	_, err = os.Stat(dst)
	if err == nil {
		// File exists
		if !overwrite {
			log.Infof("File %s exists, won't overwrite it.\n", dst)
			return
		}
	}

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	log.Infof("Copy file %s", src)
	log.Infof("       to %s", dst)
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
