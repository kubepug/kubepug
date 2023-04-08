package fileinput

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/rikatz/kubepug/pkg/results"
	log "github.com/sirupsen/logrus"

	"sigs.k8s.io/yaml"
)

// FileStruct defines a type that will receive a common format of objects regardless of the input format (yaml, json)
type FileStruct struct {
	Kind       string `yaml:"kind"`
	APIVersion string `yaml:"apiVersion"`
	Metadata   struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"metadata"`
	Items []FileStruct `yaml:"items,omitempty"`
}

// FileItems is a map which key is the Group/Version/Kind from K8s and value are the items found in
// the input files
type FileItems map[string][]results.Item

// GetFileItems converts a bunch of input files into a map of Items
func GetFileItems(location string) (fileItems FileItems) {
	fileItems = make(FileItems)
	// First we get the list of files

	if location == "-" {
		fileItems.yamlToMap(nil, "-", false)
		return fileItems
	}

	var filesInfo []os.FileInfo

	fileLocation, err := os.Stat(location)
	if os.IsNotExist(err) {
		log.Fatalf("input location %s does not exist", location)
	}

	if fileLocation.IsDir() {
		entries, err := os.ReadDir(location) // Too lazy to refactor right now :P
		if err != nil {
			log.Fatalf("error to read input location %s. Error: %v", location, err)
		}
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				log.Fatalf("error converting filedir to fileinfo: %s", err)
			}
			filesInfo = append(filesInfo, info)
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

	var err error
	var yamlFiles []byte
	if location == "-" {
		reader := bufio.NewReader(os.Stdin)
		yamlFiles, err = io.ReadAll(reader)
		if err != nil {
			log.Warningf("Unable to read from STDIN: %s", err)
			return
		}
		location = "STDIN" // Changing here just to be beautified in the list
	} else {
		yamlFiles, err = os.ReadFile(location)
		if err != nil {
			log.Warningf("Unable to read manifest file, skipping: %s", err)
			return
		}
	}

	yamlObjects := bytes.Split(yamlFiles, []byte("---"))

	for _, yamlObject := range yamlObjects {
		var obj FileStruct
		err := yaml.Unmarshal(yamlObject, &obj)
		if err != nil {
			log.Warningf("Found invalid yaml: %v. Skipping to next", err)
			continue
		}
		if len(obj.Items) > 0 {
			for item := range obj.Items {
				fileItems.addObject(&obj.Items[item], location)
			}
		} else {
			fileItems.addObject(&obj, location)
		}
	}
}

func (fileItems FileItems) addObject(obj *FileStruct, location string) {
	var group, version, objIndex string

	gv := strings.Split(obj.APIVersion, "/")
	if len(gv) > 1 {
		group = gv[0]
		version = gv[1]
		objIndex = fmt.Sprintf("%s/%s/%s", group, version, obj.Kind)
	} else {
		version = gv[0]
		objIndex = fmt.Sprintf("%s/%s", version, obj.Kind)
	}

	if version == "" || obj.Kind == "" {
		log.Infof("YAML file does not contain apiVersion or Kind: %s  Skipping to next", location)
		return
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
