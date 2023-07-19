package utils

import (
	"fmt"
	"io"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

const (
	baseURL       = "https://raw.githubusercontent.com/kubernetes/kubernetes"
	fileURL       = "api/openapi-spec/swagger.json"
	generatedJSON = "https://deprecations.k8s.church/src/data.json"
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
		return fmt.Errorf("could not download the swagger file %s", url)
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

// DownloadSwaggerFile checks whether a swagger.json file needs to be downloaded,
// download the file and returns the location to be used
func DownloadSwaggerFile(version, swaggerdir string, force bool) (filename string, err error) {
	if swaggerdir == "" {
		swaggerdir, err = os.MkdirTemp("", "kubepug")
		if err != nil {
			return "", err
		}
	}

	filename = fmt.Sprintf("%s/swagger-%s.json", swaggerdir, version)
	fileExists, err := os.Stat(filename)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}

	if err == nil && fileExists.IsDir() {
		return "", fmt.Errorf("%s already exists as a directory", filename)
	}

	if os.IsNotExist(err) || force {
		log.Infof("File does not exist or download is forced, downloading the file")
		url := fmt.Sprintf("%s/%s/%s", baseURL, version, fileURL)
		err := downloadFile(filename, url)
		if err != nil {
			return "", err
		}
	}

	return filename, nil
}
