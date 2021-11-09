package utils

import (
	"fmt"
	"os"
	"testing"
)

func TestDownloadSwaggerFile(t *testing.T) {
	var tmpdir string
	// Github actions does not have a temporary dir :/
	tmpdir = os.Getenv("RUNNER_TEMP")
	if tmpdir == "" {
		tmpdir = "/tmp"
	}

	tests := map[string]struct {
		version    string
		swaggerdir string
		force      bool
		filename   string
		createDir  bool
		expected   string
	}{
		"valid download and directory": {
			version:    "v1.19.0",
			swaggerdir: tmpdir,
			force:      false,
			filename:   fmt.Sprintf("%s/swagger-v1.19.0.json", tmpdir),
			expected:   "",
		},
		"Invalid directory": {
			version:    "v1.19.0",
			swaggerdir: fmt.Sprintf("%s/lalalaldsa", tmpdir),
			force:      false,
			filename:   "",
			expected:   fmt.Sprintf("directory %s/lalalaldsa does not exist or is already created as a file", tmpdir),
		},
		"Tries to use an existing file": {
			version:    "v1.19.0",
			swaggerdir: "/bin/clear",
			force:      false,
			filename:   "",
			expected:   "directory /bin/clear does not exist or is already created as a file",
		},
		"Invalid kubernetes version": {
			version:    "v1.19.xxx",
			swaggerdir: tmpdir,
			force:      false,
			filename:   "",
			expected:   "could not download the swagger file https://raw.githubusercontent.com/kubernetes/kubernetes/v1.19.xxx/api/openapi-spec/swagger.json",
		},
		"Tries to download over an existing directory": {
			version:    "v1.18.0",
			swaggerdir: tmpdir,
			force:      false,
			filename:   "",
			createDir:  true,
			expected:   fmt.Sprintf("%s/swagger-v1.18.0.json already exists as a directory", tmpdir),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			if tc.createDir {
				newdir := fmt.Sprintf("%s/swagger-%s.json", tc.swaggerdir, tc.version)
				if _, err := os.Stat(newdir); os.IsNotExist(err) {
					err := os.Mkdir(newdir, os.ModePerm)
					if err != nil {
						t.Fatalf("Failed to create the temporary directory %s: %v", newdir, err)
					}
				}
			}

			file, err := DownloadSwaggerFile(tc.version, tc.swaggerdir, tc.force)
			if err != nil && err.Error() != tc.expected {
				t.Errorf("unexpected error: got %v and expecting %v", err, tc.expected)
			}

			if tc.filename != file {
				t.Errorf("expected file %s, got file %s", tc.filename, file)
			}
		})
	}
}
