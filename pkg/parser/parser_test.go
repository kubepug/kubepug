package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

func TestPopulateKubeAPIs(t *testing.T) {
	mockcontentvalid := `
	{
		"definitions": {
			"io.k8s.api.admissionregistration.v1.MutatingWebhookConfiguration": {
				"description": "MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object.",
				"x-kubernetes-group-version-kind": [
					  {
						  "group": "admissionregistration.k8s.io",
						  "kind": "MutatingWebhookConfiguration",
						  "version": "v1"
					}
				  ]
			},
			"io.k8s.api.extensions.v1beta1.Ingress": {
				"description": "Ingress is a collection of rules that allow inbound connections to reach the endpoints defined by a backend. An Ingress can be configured to give services externally-reachable urls, load balance traffic, terminate SSL, offer name based virtual hosting etc. DEPRECATED - This group version of Ingress is deprecated by networking.k8s.io/v1beta1 Ingress. See the release notes for more information.",
				"x-kubernetes-group-version-kind": [
					  {
						"group": "extensions",
						"kind": "Ingress",
						"version": "v1beta1"
					}
				  ]
			},
			"io.k8s.api.core.v1.Pod": {
				"description": "Pod is a collection of containers that can run on a host. This resource is created by clients and scheduled onto hosts.",
				"x-kubernetes-group-version-kind": [
					  {
						"group": "",
						"kind": "Pod",
						"version": "v1"
					}
				  ]
			}
		}
	}`

	mockcontentinvalidjson := `
	{
		"definitions": { {
			"io.k8s.api.admissionregistration.v1.MutatingWebhookConfiguration": {
				"description": "MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object.",
				"x-kubernetes-group-version-kind": [
					  {
						  "group": "admissionregistration.k8s.io",
						  "kind": "MutatingWebhookConfiguration",
						  "version": "v1"
					}
				  ]
			},
		}
	}`

	mockcontentemptydescription := `
	{
		"definitions": {
			"io.k8s.api.extensions.v1beta1.Ingress": {
				"description": "Ingress is a collection of rules that allow inbound connections to reach the endpoints defined by a backend. An Ingress can be configured to give services externally-reachable urls, load balance traffic, terminate SSL, offer name based virtual hosting etc. DEPRECATED - This group version of Ingress is deprecated by networking.k8s.io/v1beta1 Ingress. See the release notes for more information.",
				"x-kubernetes-group-version-kind": [
					  {
						"group": "extensions",
						"kind": "Ingress",
						"version": "v1beta1"
					}
				  ]
			},
			"io.k8s.api.core.v1.Pod": {
				"description": "",
				"x-kubernetes-group-version-kind": [
					  {
						"group": "",
						"kind": "Pod",
						"version": "v1"
					}
				  ]
			}
		}
	}`

	tests := map[string]struct {
		KubeAPIs    KubernetesAPIs
		swaggerfile string
		mockcontent string
		expectederr string
	}{
		"valid APIs found": {
			KubeAPIs: KubernetesAPIs{
				"admissionregistration.k8s.io/v1/MutatingWebhookConfiguration": {
					Description: "MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object.",
					Group:       "admissionregistration.k8s.io",
					Kind:        "MutatingWebhookConfiguration",
					Version:     "v1",
					Name:        "",
					Deprecated:  false,
				},
				"extensions/v1beta1/Ingress": {
					Description: "Ingress is a collection of rules that allow inbound connections to reach the endpoints defined by a backend. An Ingress can be configured to give services externally-reachable urls, load balance traffic, terminate SSL, offer name based virtual hosting etc. DEPRECATED - This group version of Ingress is deprecated by networking.k8s.io/v1beta1 Ingress. See the release notes for more information.",
					Group:       "extensions",
					Kind:        "Ingress",
					Version:     "v1beta1",
					Name:        "",
					Deprecated:  true,
				},
				"v1/Pod": {
					Description: "Pod is a collection of containers that can run on a host. This resource is created by clients and scheduled onto hosts.",
					Group:       "",
					Kind:        "Pod",
					Version:     "v1",
					Name:        "",
					Deprecated:  false,
				},
			},
			swaggerfile: "/tmp/test1.json",
			mockcontent: mockcontentvalid,
			expectederr: "",
		},
		"invalid JSON found": {
			KubeAPIs:    KubernetesAPIs{},
			swaggerfile: "/tmp/invalidtest1.json",
			mockcontent: mockcontentinvalidjson,
			expectederr: "error parsing the JSON, file might be invalid: invalid character '{' looking for beginning of object key string",
		},
		"some empty objects because of empty description": {
			KubeAPIs: KubernetesAPIs{
				"extensions/v1beta1/Ingress": {
					Description: "Ingress is a collection of rules that allow inbound connections to reach the endpoints defined by a backend. An Ingress can be configured to give services externally-reachable urls, load balance traffic, terminate SSL, offer name based virtual hosting etc. DEPRECATED - This group version of Ingress is deprecated by networking.k8s.io/v1beta1 Ingress. See the release notes for more information.",
					Group:       "extensions",
					Kind:        "Ingress",
					Version:     "v1beta1",
					Name:        "",
					Deprecated:  true,
				},
			},
			swaggerfile: "/tmp/emptydescriptions.json",
			mockcontent: mockcontentemptydescription,
			expectederr: "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			filecontent := []byte(tc.mockcontent)
			err := writeFile(filecontent, tc.swaggerfile)
			defer os.Remove(tc.swaggerfile)
			if err != nil {
				t.Errorf("unexpected error creating temporary file: %v", err)
			}

			o := KubernetesAPIs{}
			err = o.PopulateKubeAPIMap(tc.swaggerfile)
			if err != nil && err.Error() != tc.expectederr {
				t.Errorf("Failed to populate the map: Got %v exoected %v", err, tc.expectederr)
			}

			eq := reflect.DeepEqual(o, tc.KubeAPIs)
			if !eq {
				prettyExpected, err := json.MarshalIndent(tc.KubeAPIs, "", "")
				if err != nil {
					t.Errorf("unexpected error creating temporary file: %v", err)
				}

				prettyGot, err := json.MarshalIndent(o, "", "")
				if err != nil {
					t.Errorf("unexpected error creating temporary file: %v", err)
				}

				t.Errorf("Maps are not equivalent, got %s, expected %s", prettyGot, prettyExpected)
			}
		})
	}
}

func writeFile(filecontent []byte, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("error while creating mock file %s", file)
	}
	defer f.Close()

	_, err = f.Write(filecontent)
	if err != nil {
		return fmt.Errorf("error while writing to file %s", file)
	}

	return nil
}
