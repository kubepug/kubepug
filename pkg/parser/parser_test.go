package parser

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	testdatalocation = "../../test/testdata/swagger"
)

var (
	mockcontentvalid = `
	{
		"definitions": {
			"io.k8s.api.core.v1.Namespace": {
				"description": "Namespace provides a scope for Names. Use of multiple namespaces is optional.",
				"x-kubernetes-group-version-kind": [
				  {
					"group": "",
					"kind": "Namespace",
					"version": "v1"
				  }
				]
			},
			"io.k8s.api.admissionregistration.v1beta1.MutatingWebhookConfiguration": {
				"description": "MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object. Deprecated in favor of v1",
				"x-kubernetes-group-version-kind": [
					  {
						  "group": "admissionregistration.k8s.io",
						  "kind": "MutatingWebhookConfiguration",
						  "version": "v1beta1"
					}
				  ]
			},
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

	mockcontentinvalidjson = `
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
)

var k8sversions = []struct {
	version string
}{
	{version: "v1.19.5"},
	{version: "v1.23.4"},
	{version: "v1.27.2"},
}

func BenchmarkNewStructParser(b *testing.B) {
	for _, v := range k8sversions {
		b.Run(fmt.Sprintf("version_%s", v.version), func(b *testing.B) {
			swaggerfile := fmt.Sprintf("%s/swagger-%s.json", testdatalocation, v.version)
			for i := 0; i < b.N; i++ {
				v, err := NewAPIGroupsFromSwaggerFile(swaggerfile)
				if err != nil {
					b.Error(err)
				}
				if v == nil {
					b.Error("b shouldn't be null")
				}
			}
		})
	}
}

func TestPopulateStruct(t *testing.T) {
	t.Run("with invalid json file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "pugtmp")
		require.NoError(t, err)
		_, err = tmpFile.Write([]byte(mockcontentinvalidjson))
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())
		defer os.Remove(tmpFile.Name())
		_, err = NewAPIGroupsFromSwaggerFile(tmpFile.Name())
		require.Error(t, err)
	})

	t.Run("with valid json file", func(t *testing.T) {
		tmpFile, err := os.CreateTemp("", "pugtmp")
		require.NoError(t, err)
		_, err = tmpFile.Write([]byte(mockcontentvalid))
		require.NoError(t, err)
		require.NoError(t, tmpFile.Close())
		defer os.Remove(tmpFile.Name())
		v, err := NewAPIGroupsFromSwaggerFile(tmpFile.Name())
		require.NoError(t, err)

		require.True(t, v["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1beta1"].Deprecated)
		require.Equal(t,
			"MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object. Deprecated in favor of v1",
			v["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1beta1"].Description)
		require.False(t, v["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1"].Deprecated)

		require.Equal(t,
			"Namespace provides a scope for Names. Use of multiple namespaces is optional.",
			v[CoreAPI]["Namespace"]["v1"].Description)
	})
}
