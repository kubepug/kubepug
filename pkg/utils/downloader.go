package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

// From https://golangcode.com/download-a-file-from-a-url/ which was easier than create :P
func downloadFile(filename, url string) error {
	// Get the data
	log.Debugf("Downloading file from %s", url)
	resp, err := http.Get(url) //nolint: gosec
	if err != nil {
		return err
	}
	if resp.StatusCode > 305 {
		return fmt.Errorf("could not download the data file %s", url)
	}
	defer resp.Body.Close()

	// Create the file
	log.Debugf("Creating the file %s", filename)
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)

	return err
}

func DownloadGeneratedJSON(urlpath string) (filename string, err error) {
	tmpdir, err := os.MkdirTemp("", "kubepug")
	if err != nil {
		return "", err
	}

	filename = fmt.Sprintf("%s/data.json", tmpdir)
	err = downloadFile(filename, urlpath)
	if err != nil {
		return "", err
	}

	return filename, nil
}
