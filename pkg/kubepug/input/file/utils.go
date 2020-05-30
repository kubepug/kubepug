package fileinput

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"

	log "github.com/sirupsen/logrus"

	"github.com/rikatz/kubepug/pkg/results"
)

// FileStruct defines a type that will receive a common format of objects regardless of the input format (yaml, json)
type FileStruct struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace,omitempty"`
	}
}

// FileItems is a map which key is the Group/Version/Kind from K8s and value are the items found in
// the input files
type FileItems map[string][]results.Item

// GetFileItems converts a bunch of input files into a map of Items
func GetFileItems(location string) (fileItems FileItems) {
	var filesInfo []os.FileInfo

	fileItems = make(FileItems)
	// First we get the list of files

	fileLocation, err := os.Stat(location)
	if os.IsNotExist(err) {
		log.Fatalf("input location %s does not exist", location)
	}

	if fileLocation.IsDir() {
		filesInfo, err = ioutil.ReadDir(location)
		if err != nil {
			log.Fatalf("error to read input location %s. Error: %v", location, err)
		}
	} else {
		filesInfo = append(filesInfo, fileLocation)
	}

	// Then we loop each of them and feed the fileItems struct
	for _, file := range filesInfo {
		fileItems.yamlToMap(file, location, fileLocation.IsDir())
	}
	return fileItems
}

// Yaml to Map takes a YAML and insert its items into the FileItems Map
func (fileItems FileItems) yamlToMap(file os.FileInfo, location string, isDir bool) {

	if isDir {
		location = fmt.Sprintf("%s/%s", location, file.Name())
	}

	yamlFile, err := ioutil.ReadFile(location)
	if err != nil {
		log.Warningf("Failed to read file %s: %v. Skipping to next file", location, err)
		return
	}

	dec := yaml.NewDecoder(bytes.NewReader(yamlFile))

	for {
		var obj FileStruct
		if err = dec.Decode(&obj); err != nil && err != io.EOF {
			log.Warningf("Found invalid yaml: %s: %v. Skipping to next", location, err)
			break
		}
		if err == io.EOF {
			return
		}

		var group, version, objIndex string

		gv := strings.Split(obj.APIVersion, "/")
		if len(gv) > 1 {
			group = gv[0]
			version = gv[1]
			objIndex = fmt.Sprintf("%s/%s/%s", group, version, obj.Kind)
		} else {
			group = ""
			version = gv[0]
			objIndex = fmt.Sprintf("%s/%s", version, obj.Kind)
		}
		if version == "" || obj.Kind == "" {
			log.Warningf("YAML file does not contain apiVersion or Kind: %s  Skipping to next", location)
			break
		}
		item := results.Item{
			ObjectName: obj.Metadata.Name,
			Namespace:  obj.Metadata.Namespace,
			Location:   location,
			Scope:      "OBJECT",
		}

		if items, ok := fileItems[objIndex]; !ok {
			fileItems[objIndex] = []results.Item{item}
		} else {
			items = append(items, item)
			fileItems[objIndex] = items
		}
	}
}
