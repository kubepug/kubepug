/*
Copyright 2023 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package deprecations

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/gengo/types"
)

var mockType = &types.Type{
	CommentLines: []string{
		"RandomType defines a random structure in Kubernetes",
		"It should be used just when you need something different than 42",
	},
	SecondClosestCommentLines: []string{},
}

func Test_extractKubeVersionTag(t *testing.T) {
	tests := []struct {
		name        string
		tagName     string
		tagComments []string
		wantValue   *tagValue
		wantMajor   int
		wantMinor   int
		wantErr     bool
	}{
		{
			name:    "not found tag should generate an error",
			tagName: "someVersionTag:version",
			tagComments: []string{
				"+someOtherTag:version=1.5",
			},
			wantValue: nil,
			wantErr:   true,
		},
		{
			name:    "found tag should return correctly",
			tagName: "someVersionTag:version",
			tagComments: []string{
				"+someVersionTag:version=1.5",
			},
			wantValue: &tagValue{
				value: "1.5",
			},
			wantMajor: 1,
			wantMinor: 5,
			wantErr:   false,
		},
		{
			name:    "multiple declarations of same tag should return an error",
			tagName: "someVersionTag:version",
			tagComments: []string{
				"+someVersionTag:version=1.5",
				"+someVersionTag:version=v1.7",
			},
			wantValue: nil,
			wantErr:   true,
		},
		/*{
			name:    "multiple values on same tag should return an error",
			tagName: "someVersionTag:version",
			tagComments: []string{
				"+someVersionTag:version=1.5,something",
			},
			wantValue: nil,
			wantErr:   true,
		},*/
		{
			name:    "wrong tag major value should return an error",
			tagName: "someVersionTag:version",
			tagComments: []string{
				"+someVersionTag:version=.5",
			},
			wantErr: true,
		},
		{
			name:    "wrong tag minor value should return an error",
			tagName: "someVersionTag:version",
			tagComments: []string{
				"+someVersionTag:version=1.",
			},
			wantErr: true,
		},
		{
			name:    "wrong tag format should return an error",
			tagName: "someVersionTag:version",
			tagComments: []string{
				"+someVersionTag:version=1.5.7",
			},
			wantErr: true,
		},
		{
			name:    "wrong tag major int value should return an error",
			tagName: "someVersionTag:version",
			tagComments: []string{
				"+someVersionTag:version=blah.5",
			},
			wantErr: true,
		},
		{
			name:    "wrong tag minor int value should return an error",
			tagName: "someVersionTag:version",
			tagComments: []string{
				"+someVersionTag:version=1.blah",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockType.SecondClosestCommentLines = tt.tagComments
			gotTag, gotMajor, gotMinor, err := extractKubeVersionTag(tt.tagName, mockType)
			if (err != nil) != tt.wantErr {
				t.Errorf("extractKubeVersionTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(gotTag, tt.wantValue) {
				t.Errorf("extractKubeVersionTag() got = %v, want %v", gotTag, tt.wantValue)
			}
			if gotMajor != tt.wantMajor {
				t.Errorf("extractKubeVersionTag() got1 = %v, want %v", gotMajor, tt.wantMajor)
			}
			if gotMinor != tt.wantMinor {
				t.Errorf("extractKubeVersionTag() got2 = %v, want %v", gotMinor, tt.wantMinor)
			}
		})
	}
}

func TestCommonFunctions(t *testing.T) {
	t.Run("should return common Namespace system", func(t *testing.T) {
		namerOutput := NameSystems()
		require.Len(t, namerOutput, 2)
		require.NotNil(t, namerOutput["public"])
		require.NotNil(t, namerOutput["raw"])
		require.Equal(t, DefaultNameSystem(), "public")
	})

	t.Run("should create an APIRegistry correctly", func(t *testing.T) {
		reg := NewAPIRegistry()
		require.Len(t, reg.registry, 0)
		require.Len(t, reg.Registry(), 0)

		g := reg.NewDeprecatedDefinitionsGen("bla", "k8s.io/api/bla/v1beta1", "bla", "v1beta1")
		namerOutput := g.Namers(nil)
		require.Len(t, namerOutput, 3)
		require.NotNil(t, namerOutput["public"])
		require.NotNil(t, namerOutput["raw"])
		require.NotNil(t, namerOutput["intrapackage"])

		t.Run("should return correct isOtherPackage", func(t *testing.T) {
			g := reg.NewDeprecatedDefinitionsGen("bla", "k8s.io/api/bla/v1beta1", "bla", "v1beta1")
			gInner, ok := g.(*genDeprecatedDefinitions)
			require.True(t, ok)
			require.False(t, gInner.isOtherPackage("k8s.io/api/bla/v1beta1"))
			require.False(t, gInner.isOtherPackage(`bla "k8s.io/api/bla/v1beta1"`))
			require.True(t, gInner.isOtherPackage("k8s.io/api/bla/v1beta2"))
		})

		t.Run("Should fail with bad tagged API", func(t *testing.T) {
			newTestMock := *mockType
			newTestMock.SecondClosestCommentLines = []string{
				"+k8s:prerelease-lifecycle-gen:introduced=1.1",
				"+k8s:prerelease-lifecycle-gen:deprecated=1.8",
				"+k8s:prerelease-lifecycle-gen:deprecated=1.16",
				"+k8s:prerelease-lifecycle-gen:replacement=apps,v1,Deployment",
			}
			err := g.GenerateType(nil, &newTestMock, nil)
			require.Error(t, err)
			require.Len(t, reg.Registry(), 0)
		})

		t.Run("Should generate type correctly", func(t *testing.T) {
			newTestMock := *mockType
			newTestMock.SecondClosestCommentLines = []string{
				"+k8s:prerelease-lifecycle-gen:introduced=1.1",
				"+k8s:prerelease-lifecycle-gen:deprecated=1.8",
				"+k8s:prerelease-lifecycle-gen:removed=1.16",
				"+k8s:prerelease-lifecycle-gen:replacement=apps,v1,Deployment",
			}
			err := g.GenerateType(nil, &newTestMock, nil)
			require.NoError(t, err)
			require.Len(t, reg.Registry(), 1)
			require.Equal(t, reg.Registry()[0].DeprecatedVersion, Version{VersionMajor: 1, VersionMinor: 8})
			require.Equal(t, reg.Registry()[0].Replacement, GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"})
		})
	})
}

func TestTagParsing(t *testing.T) {
	t.Run("should extract tags correctly", func(t *testing.T) {
		g := &genDeprecatedDefinitions{}
		newTestMock := *mockType
		newTestMock.SecondClosestCommentLines = []string{
			"+k8s:prerelease-lifecycle-gen:introduced=1.1",
			"+k8s:prerelease-lifecycle-gen:deprecated=1.8",
			"+k8s:prerelease-lifecycle-gen:removed=1.16",
			"+k8s:prerelease-lifecycle-gen:replacement=apps,v1,Deployment",
		}

		r, err := g.argsFromType(nil, &newTestMock)
		require.NoError(t, err)
		require.Equal(t, r.IntroducedVersion, Version{VersionMajor: 1, VersionMinor: 1})
		require.Equal(t, r.DeprecatedVersion, Version{VersionMajor: 1, VersionMinor: 8})
		require.Equal(t, r.RemovedVersion, Version{VersionMajor: 1, VersionMinor: 16})
		require.Equal(t, r.Replacement, GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"})
	})

	t.Run("should fail on introduced duplicated tags", func(t *testing.T) {
		g := &genDeprecatedDefinitions{}
		newTestMock := *mockType
		newTestMock.SecondClosestCommentLines = []string{
			"+k8s:prerelease-lifecycle-gen:introduced=1.1",
			"+k8s:prerelease-lifecycle-gen:introduced=1.8",
		}

		r, err := g.argsFromType(nil, &newTestMock)
		require.Error(t, err)
		require.Nil(t, r)
	})

	t.Run("should fail on deprecated tags", func(t *testing.T) {
		g := &genDeprecatedDefinitions{}
		newTestMock := *mockType
		newTestMock.SecondClosestCommentLines = []string{
			"+k8s:prerelease-lifecycle-gen:deprecated=1.1",
			"+k8s:prerelease-lifecycle-gen:deprecated=1.8",
		}

		r, err := g.argsFromType(nil, &newTestMock)
		require.Error(t, err)
		require.Nil(t, r)
	})
}

func TestReplacementTag(t *testing.T) {
	t.Run("should fail on duplicated replacement tags", func(t *testing.T) {
		g := &genDeprecatedDefinitions{}
		newTestMock := *mockType
		newTestMock.SecondClosestCommentLines = []string{
			"+k8s:prerelease-lifecycle-gen:replacement=apps,v1,Deployment",
			"+k8s:prerelease-lifecycle-gen:replacement=apps,v2,Deployment",
		}

		r, err := g.argsFromType(nil, &newTestMock)
		require.Error(t, err)
		require.Nil(t, r)
	})

	t.Run("should fail on bad group replacement tags", func(t *testing.T) {
		g := &genDeprecatedDefinitions{}
		newTestMock := *mockType
		newTestMock.SecondClosestCommentLines = []string{
			"+k8s:prerelease-lifecycle-gen:replacement=ApPs,v1,Deployment",
		}

		r, err := g.argsFromType(nil, &newTestMock)
		require.Error(t, err)
		require.Nil(t, r)
	})

	t.Run("should not find replacement tags but pass anyway", func(t *testing.T) {
		g := &genDeprecatedDefinitions{}
		newTestMock := *mockType
		newTestMock.SecondClosestCommentLines = []string{
			"+k8s:prerelease-lifecycle-gen:introduced=1.1",
			"+k8s:prerelease-lifecycle-gen:deprecated=1.8",
			"+k8s:prerelease-lifecycle-gen:removed=1.16",
		}

		r, err := g.argsFromType(nil, &newTestMock)
		require.NoError(t, err)
		require.Equal(t, r.IntroducedVersion, Version{VersionMajor: 1, VersionMinor: 1})
		require.Equal(t, r.DeprecatedVersion, Version{VersionMajor: 1, VersionMinor: 8})
		require.Equal(t, r.RemovedVersion, Version{VersionMajor: 1, VersionMinor: 16})
	})
}

func Test_isAPIType(t *testing.T) {
	tests := []struct {
		name string
		t    *types.Type
		want bool
	}{
		{
			name: "private name is not apytype",
			want: false,
			t: &types.Type{
				Name: types.Name{
					Name: "notpublic",
				},
			},
		},
		{
			name: "non struct is not apytype",
			want: false,
			t: &types.Type{
				Name: types.Name{
					Name: "Public",
				},
				Kind: types.Slice,
			},
		},
		{
			name: "contains member type",
			want: true,
			t: &types.Type{
				Name: types.Name{
					Name: "Public",
				},
				Kind: types.Struct,
				Members: []types.Member{
					{
						Embedded: true,
						Name:     "TypeMeta",
					},
				},
			},
		},

		{
			name: "contains no type",
			want: false,
			t: &types.Type{
				Name: types.Name{
					Name: "Public",
				},
				Kind: types.Struct,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isAPIType(tt.t); got != tt.want {
				t.Errorf("isAPIType() = %v, want %v", got, tt.want)
			}
		})
	}
}
