package utils

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)

const baseURL = "https://raw.githubusercontent.com/kubernetes/kubernetes"
const fileURL = "api/openapi-spec/swagger.json"

// From https://golangcode.com/download-a-file-from-a-url/ which was easier than create :P
func downloadFile(filename, url string) error {
	// Get the data
	log.Debugf("Downloading file from %s", url)
	resp, err := http.Get(url)
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

// DownloadSwaggerFile checks whether a swagger.json file needs to be downloaded,
// download the file and returns the location to be used
func DownloadSwaggerFile(version, swaggerdir string, force bool) (filename string, err error) {

	if swaggerdir == "" {
		swaggerdir, err = ioutil.TempDir("", "kubepug")
		if err != nil {
			return "", err
		}
	}
	dir, err := os.Stat(swaggerdir)
	if os.IsNotExist(err) || !dir.IsDir() {
		return "", fmt.Errorf("directory %s does not exist or is already created as a file", swaggerdir)
	}

	filename = fmt.Sprintf("%s/swagger-%s.json", swaggerdir, version)
	fileExists, err := os.Stat(filename)

	if os.IsNotExist(err) || (force && !fileExists.IsDir()) {
		log.Infof("File does not exist or download is forced, downloading the file")
		url := fmt.Sprintf("%s/%s/%s", baseURL, version, fileURL)
		err := downloadFile(filename, url)
		if err != nil {
			return "", err
		}
		return filename, nil
	}

	if fileExists.IsDir() {
		return "", fmt.Errorf("%s already exists as a directory", filename)
	}

	return filename, nil
}
