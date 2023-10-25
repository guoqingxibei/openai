package util

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func DownloadFile(url string, fileName string) (err error) {
	start := time.Now()
	response, err := http.Get(url)
	if err != nil {
		return
	}
	defer response.Body.Close()

	err = os.MkdirAll(filepath.Dir(fileName), os.ModePerm)
	if err != nil {
		return
	}

	file, err := os.Create(fileName)
	if err != nil {
		return
	}
	defer file.Close()

	_, err = io.Copy(file, response.Body)
	if err != nil {
		return
	}

	log.Printf("[DownloadFileAPI] Duration: %dms, fileName: %s, imageUrl: %s",
		int(time.Since(start).Milliseconds()),
		fileName,
		url,
	)
	return
}
