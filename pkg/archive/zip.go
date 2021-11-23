package archive

import (
	"archive/zip"
	"github.com/sirupsen/logrus"
	"github.com/zhiting-tech/smartassistant/pkg/errors"
	"io"
	"os"
	"path/filepath"
)

func UnZip(dst, src string) (err error) {
	logrus.Printf("unzip file: %s -> %s", src, dst)

	r, err := zip.OpenReader(src)
	if err != nil {
		return
	}
	var extractedFiles []string
	if dst != "" {
		if err = os.MkdirAll(dst, 0755); err != nil {
			return
		}
		extractedFiles = append(extractedFiles, dst)
	}
	for _, file := range r.File {
		logrus.Println(file.Name)
		if err = extractZipFile(file, dst); err != nil {
			return
		}
		extractedFiles = append(extractedFiles, filepath.Join(dst, file.Name))
	}
	return
}
func Zip(dst string, srcs ...string) (err error) {

	fw, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer fw.Close()

	zw := zip.NewWriter(fw)
	defer zw.Close()

	for _, src := range srcs {

		walker := func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			data, err := os.Open(path)
			if err != nil {
				return err
			}
			defer data.Close()

			w, err := zw.Create(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, data)
			logrus.Printf("unzip %s to %s", path, filepath.Join(src, path))
			return err
		}
		if err = filepath.Walk(src, walker); err != nil {
			return
		}
	}
	return
}

func extractZipFile(file *zip.File, dst string) (err error) {

	path := filepath.Join(dst, file.Name)

	if file.FileInfo().IsDir() {
		return os.MkdirAll(path, file.Mode())
	}

	if err = os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		err = errors.Wrap(err, errors.InternalServerErr)
		return
	}

	fw, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_TRUNC, file.Mode())
	if err != nil {
		return
	}
	defer fw.Close()

	fr, err := file.Open()
	if err != nil {
		return
	}
	defer fr.Close()
	if _, err = io.Copy(fw, fr); err != nil {
		return
	}
	return
}
