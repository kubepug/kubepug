package swaggerstore

import (
	"context"
	"fmt"
	"os"
	"testing"

	apis "github.com/rikatz/kubepug/pkg/apis/v1alpha1"

	"github.com/rikatz/kubepug/pkg/errors"
	"github.com/stretchr/testify/require"
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

const (
	testdatalocation = "../../../test/testdata/swagger"
)

var k8sversions = []struct {
	version string
}{
	{version: "v1.19.5"},
	{version: "v1.23.4"},
	{version: "v1.27.2"},
}

func BenchmarkSwaggerStore(b *testing.B) {
	for _, v := range k8sversions {
		b.Run(fmt.Sprintf("version_%s", v.version), func(b *testing.B) {
			swaggerfile := fmt.Sprintf("%s/swagger-%s.json", testdatalocation, v.version)
			for i := 0; i < b.N; i++ {
				v, err := NewSwaggerStoreFromFile(swaggerfile)
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
		_, err := newInternalDatabase([]byte(mockcontentinvalidjson))
		require.Error(t, err)
	})

	t.Run("with valid json file", func(t *testing.T) {
		v, err := newInternalDatabase([]byte(mockcontentvalid))
		require.NoError(t, err)

		require.Equal(t, v["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1beta1"].DeprecationVersion, "true")
		require.Equal(t,
			"MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object. Deprecated in favor of v1",
			v["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1beta1"].Description)
		require.Equal(t, v["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1"].DeprecationVersion, "false")

		require.Equal(t,
			"Namespace provides a scope for Names. Use of multiple namespaces is optional.",
			v[apis.CoreAPI]["Namespace"]["v1"].Description)
	})
}

func TestNewSwaggerStoreFromFile(t *testing.T) {
	t.Run("with invalid file path should return an error", func(t *testing.T) {
		_, err := NewSwaggerStoreFromFile("/xpto/blabla/123")
		require.Error(t, err)
	})

	t.Run("with valid path should be able to parse it", func(t *testing.T) {
		tmp, err := os.MkdirTemp("", "")
		require.NoError(t, err)
		err = os.WriteFile(tmp+"/testfile", []byte(mockcontentvalid), 0o600)
		require.NoError(t, err)
		v, err := NewSwaggerStoreFromFile(tmp + "/testfile")
		require.NoError(t, err)
		require.Equal(t, v.db["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1beta1"].DeprecationVersion, "true")
		require.Equal(t,
			"MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object. Deprecated in favor of v1",
			v.db["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1beta1"].Description)
		require.Equal(t, v.db["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1"].DeprecationVersion, "false")

		require.Equal(t,
			"Namespace provides a scope for Names. Use of multiple namespaces is optional.",
			v.db[apis.CoreAPI]["Namespace"]["v1"].Description)
	})
}

func TestNewSwaggerStoreFromBytes(t *testing.T) {
	t.Run("with invalid bytes should return an error", func(t *testing.T) {
		_, err := NewSwaggerStoreFromBytes([]byte(mockcontentinvalidjson))
		require.Error(t, err)
	})

	t.Run("with valid bytes should be able to parse it", func(t *testing.T) {
		v, err := NewSwaggerStoreFromBytes([]byte(mockcontentvalid))
		require.NoError(t, err)
		require.Equal(t, v.db["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1beta1"].DeprecationVersion, "true")
		require.Equal(t,
			"MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object. Deprecated in favor of v1",
			v.db["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1beta1"].Description)
		require.Equal(t, v.db["admissionregistration.k8s.io"]["MutatingWebhookConfiguration"]["v1"].DeprecationVersion, "false")

		require.Equal(t,
			"Namespace provides a scope for Names. Use of multiple namespaces is optional.",
			v.db[apis.CoreAPI]["Namespace"]["v1"].Description)

		results, err := v.GetAPIDefinition(context.TODO(), "admissionregistration.k8s.io", "v1beta1", "MutatingWebhookConfiguration")
		require.NoError(t, err)
		require.Equal(t, internalStatusVersion, results.DeprecationVersion)
		require.Equal(t,
			"MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object. Deprecated in favor of v1",
			results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "admissionregistration.k8s.io", "v1beta1", "MutatingWebhookConfigurationAAA")
		require.ErrorIs(t, err, errors.ErrAPINotFound)
		require.Equal(t, internalStatusVersion, results.DeletedVersion)
		require.Equal(t, "", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "admissionregistration1.k8s.io", "v1beta1", "MutatingWebhookConfiguration")
		require.ErrorIs(t, err, errors.ErrAPINotFound)
		require.Equal(t, internalStatusVersion, results.DeletedVersion)
		require.Equal(t, "", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "admissionregistration.k8s.io", "v1alpha1", "MutatingWebhookConfiguration")
		require.ErrorIs(t, err, errors.ErrAPINotFound)
		require.Equal(t, internalStatusVersion, results.DeletedVersion)
		require.Equal(t, "", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "admissionregistration.k8s.io", "v1", "MutatingWebhookConfiguration")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "MutatingWebhookConfiguration describes the configuration of and admission webhook that accept or reject and may change the object.", results.Description)

		results, err = v.GetAPIDefinition(context.TODO(), "", "v1", "Pod")
		require.NoError(t, err)
		require.Equal(t, "", results.DeletedVersion)
		require.Equal(t, "", results.DeprecationVersion)
		require.Equal(t, "Pod is a collection of containers that can run on a host. This resource is created by clients and scheduled onto hosts.", results.Description)
	})
}
