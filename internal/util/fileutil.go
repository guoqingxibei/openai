package util

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

func DownloadFile(url string, path string) (err error) {
	start := time.Now()
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()

	err = os.MkdirAll(filepath.Dir(path), os.ModePerm)
	if err != nil {
		return
	}

	file, err := os.Create(path)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return
	}

	slog.Info(fmt.Sprintf("[DownloadFileAPI] Duration: %dms, path: %s, fileUrl: %s",
		int(time.Since(start).Milliseconds()),
		path,
		url,
	))
	return
}

func DownloadFileInto(url string, destDir string) (path string, err error) {
	fileName, err := extractFileName(url)
	if err != nil {
		return
	}

	path = destDir + "/" + fileName
	err = DownloadFile(url, path)
	if err != nil {
		return
	}

	return
}

func extractFileName(fileUrl string) (fileName string, err error) {
	parsedURL, err := url.Parse(fileUrl)
	if err != nil {
		return
	}

	fileName = filepath.Base(parsedURL.Path)
	return
}
