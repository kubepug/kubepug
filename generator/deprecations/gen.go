/*
Copyright 2020 The Kubernetes Authors.

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
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"k8s.io/gengo/args"
	"k8s.io/gengo/examples/set-gen/sets"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"

	"k8s.io/klog/v2"
)

// This is the comment tag that carries parameters for API status generation.  Because the cadence is fixed, we can predict
// with near certainty when this lifecycle happens as the API is introduced.
const (
	tagEnabledName    = "k8s:prerelease-lifecycle-gen" //nolint: gosec
	introducedTagName = tagEnabledName + ":introduced"
	deprecatedTagName = tagEnabledName + ":deprecated"
	removedTagName    = tagEnabledName + ":removed"

	replacementTagName = tagEnabledName + ":replacement"
)

type APIRegistry struct {
	registry []APIDeprecation
	mu       sync.Mutex
}

func NewAPIRegistry() *APIRegistry {
	regMap := make([]APIDeprecation, 0)
	return &APIRegistry{
		registry: regMap,
		mu:       sync.Mutex{},
	}
}

func (r *APIRegistry) Registry() []APIDeprecation {
	return r.registry
}

// enabledTagValue holds parameters from a tagName tag.
type tagValue struct {
	value string
}

func extractEnabledTypeTag(t *types.Type) (*tagValue, error) {
	comments := append(append([]string{}, t.SecondClosestCommentLines...), t.CommentLines...)
	return extractTag(tagEnabledName, comments)
}

func tagExists(tagName string, t *types.Type) (bool, error) {
	comments := append(append([]string{}, t.SecondClosestCommentLines...), t.CommentLines...)
	rawTag, err := extractTag(tagName, comments)
	if err != nil {
		return false, err
	}
	return rawTag != nil, nil
}

func extractKubeVersionTag(tagName string, t *types.Type) (value *tagValue, majorversion, minorversion int, err error) {
	comments := append(append([]string{}, t.SecondClosestCommentLines...), t.CommentLines...)
	rawTag, err := extractTag(tagName, comments)
	if err != nil {
		return nil, -1, -1, err
	}
	if rawTag == nil || rawTag.value == "" {
		return nil, -1, -1, fmt.Errorf("%v missing %v=Version tag", t, tagName)
	}

	splitValue := strings.Split(rawTag.value, ".")
	if len(splitValue) != 2 || splitValue[0] == "" || splitValue[1] == "" {
		return nil, -1, -1, fmt.Errorf("%v format must match %v=xx.yy tag", t, tagName)
	}
	major, err := strconv.ParseInt(splitValue[0], 10, 32)
	if err != nil {
		return nil, -1, -1, fmt.Errorf("%v format must match %v=xx.yy : %w", t, tagName, err)
	}
	minor, err := strconv.ParseInt(splitValue[1], 10, 32)
	if err != nil {
		return nil, -1, -1, fmt.Errorf("%v format must match %v=xx.yy : %w", t, tagName, err)
	}

	return rawTag, int(major), int(minor), nil
}

func extractIntroducedTag(t *types.Type) (value *tagValue, major, minor int, err error) {
	return extractKubeVersionTag(introducedTagName, t)
}

func extractDeprecatedTag(t *types.Type) (value *tagValue, major, minor int, err error) {
	return extractKubeVersionTag(deprecatedTagName, t)
}

func extractRemovedTag(t *types.Type) (value *tagValue, major, minor int, err error) {
	return extractKubeVersionTag(removedTagName, t)
}

func extractReplacementTag(t *types.Type) (group, version, kind string, hasReplacement bool, err error) {
	comments := append(append([]string{}, t.SecondClosestCommentLines...), t.CommentLines...)

	tagVals := types.ExtractCommentTags("+", comments)[replacementTagName]
	if len(tagVals) == 0 {
		// No match for the tag.
		return "", "", "", false, nil
	}
	// If there are multiple values, abort.
	if len(tagVals) > 1 {
		return "", "", "", false, fmt.Errorf("found %d %s tags: %q", len(tagVals), replacementTagName, tagVals)
	}
	tagValue := tagVals[0]
	parts := strings.Split(tagValue, ",")
	if len(parts) != 3 {
		return "", "", "", false, fmt.Errorf(`%s value must be "<group>,<version>,<kind>", got %q`, replacementTagName, tagValue)
	}
	group, version, kind = parts[0], parts[1], parts[2]
	if version == "" || kind == "" {
		return "", "", "", false, fmt.Errorf(`%s value must be "<group>,<version>,<kind>", got %q`, replacementTagName, tagValue)
	}
	// sanity check the group
	if strings.ToLower(group) != group {
		return "", "", "", false, fmt.Errorf(`replacement group must be all lower-case, got %q`, group)
	}
	// sanity check the version
	if !strings.HasPrefix(version, "v") || strings.ToLower(version) != version {
		return "", "", "", false, fmt.Errorf(`replacement version must start with "v" and be all lower-case, got %q`, version)
	}
	// sanity check the kind
	if strings.ToUpper(kind[:1]) != kind[:1] {
		return "", "", "", false, fmt.Errorf(`replacement kind must start with uppercase-letter, got %q`, kind)
	}
	return group, version, kind, true, nil
}

func extractTag(tagName string, comments []string) (*tagValue, error) {
	tagVals := types.ExtractCommentTags("+", comments)[tagName]
	if tagVals == nil {
		// No match for the tag.
		return nil, nil
	}
	// If there are multiple values, abort.
	if len(tagVals) > 1 {
		return nil, fmt.Errorf("found %d %s tags: %q", len(tagVals), tagName, tagVals)
	}

	// If we got here we are returning something.
	tag := &tagValue{}

	// Get the primary value.
	parts := strings.Split(tagVals[0], ",")
	if len(parts) >= 1 {
		tag.value = parts[0]
	}

	return tag, nil
}

// NameSystems returns the name system used by the generators in this package.
func NameSystems() namer.NameSystems {
	return namer.NameSystems{
		"public": namer.NewPublicNamer(1),
		"raw":    namer.NewRawNamer("", nil),
	}
}

// DefaultNameSystem returns the default name system for ordering the types to be
// processed by the generators in this package.
func DefaultNameSystem() string {
	return "public"
}

// Packages makes the package definition.
func (r *APIRegistry) Packages(context *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
	inputs := sets.NewString(context.Inputs...)
	packages := generator.Packages{}

	for i := range inputs {
		klog.V(5).Infof("Considering pkg %q", i)
		pkg := context.Universe[i]
		if pkg == nil {
			// If the input had no Go files, for example.
			continue
		}

		ptag, err := extractTag(tagEnabledName, pkg.Comments)
		if err != nil {
			klog.Fatalf("%w", err)
		}
		pkgNeedsGeneration := false
		if ptag != nil {
			pkgNeedsGeneration, err = strconv.ParseBool(ptag.value)
			if err != nil {
				klog.Fatalf("Package %v: unsupported %s value: %q :%w", i, tagEnabledName, ptag.value, err)
			}
		}
		if !pkgNeedsGeneration {
			klog.V(5).Infof("  skipping package")
			continue
		}
		klog.V(3).Infof("Generating package %q", pkg.Path)

		// If the pkg-scoped tag says to generate, we can skip scanning types.
		if !pkgNeedsGeneration {
			// If the pkg-scoped tag did not exist, scan all types for one that
			// explicitly wants generation.
			for _, t := range pkg.Types {
				klog.V(5).Infof("  considering type %q", t.Name.String())
				ttag, err := extractEnabledTypeTag(t)
				if err != nil {
					klog.Fatalf("%w", err)
				}
				if ttag != nil && ttag.value == "true" {
					klog.V(5).Infof("    tag=true")
					if !isAPIType(t) {
						klog.Fatalf("Type %v requests deepcopy generation but is not copyable", t)
					}
					pkgNeedsGeneration = true
					break
				}
			}
		}

		if pkgNeedsGeneration {
			path := pkg.Path
			// if the source path is within a /vendor/ directory (for example,
			// k8s.io/kubernetes/vendor/k8s.io/apimachinery/pkg/apis/meta/v1), allow
			// generation to output to the proper relative path (under vendor).
			// Otherwise, the generator will create the file in the wrong location
			// in the output directory.
			// TODO: build a more fundamental concept in gengo for dealing with modifications
			// to vendored packages.
			if strings.HasPrefix(pkg.SourcePath, arguments.OutputBase) {
				expandedPath := strings.TrimPrefix(pkg.SourcePath, arguments.OutputBase)
				if strings.Contains(expandedPath, "/vendor/") {
					path = expandedPath
				}
			}

			// the group is always, at least on K8s api, a constant called GroupName
			apigroupType, ok := pkg.Constants["GroupName"]
			if !ok {
				// We cannot add a deprecated API without knowing its group
				continue
			}
			var apigroup string
			if apigroupType.ConstValue != nil {
				apigroup = *apigroupType.ConstValue
			}

			// Usually the package name should be the version
			apiversion := pkg.Name

			packages = append(packages,
				&generator.DefaultPackage{
					PackageName: strings.Split(filepath.Base(pkg.Path), ".")[0],
					PackagePath: path,
					// HeaderText:  header,
					GeneratorFunc: func(c *generator.Context) (generators []generator.Generator) {
						return []generator.Generator{
							r.NewDeprecatedDefinitionsGen(arguments.OutputFileBaseName, pkg.Path, apigroup, apiversion),
						}
					},
					FilterFunc: func(c *generator.Context, t *types.Type) bool {
						return t.Name.Package == pkg.Path
					},
				})
		}
	}
	return packages
}

type genDeprecatedDefinitions struct {
	generator.DefaultGen
	targetPackage string
	group         string
	version       string
	imports       namer.ImportTracker
	typesForInit  []*types.Type
	registry      *APIRegistry
}

// NewPrereleaseLifecycleGen creates a generator for the prerelease-lifecycle-generator
func (r *APIRegistry) NewDeprecatedDefinitionsGen(sanitizedName, targetPackage, group, version string) generator.Generator {
	return &genDeprecatedDefinitions{
		DefaultGen: generator.DefaultGen{
			OptionalName: sanitizedName,
		},
		targetPackage: targetPackage,
		imports:       generator.NewImportTracker(),
		typesForInit:  make([]*types.Type, 0),
		group:         group,
		version:       version,
		registry:      r,
	}
}

func (g *genDeprecatedDefinitions) Namers(_ *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"public":       namer.NewPublicNamer(1),
		"intrapackage": namer.NewPublicNamer(0),
		"raw":          namer.NewRawNamer("", nil),
	}
}

func (g *genDeprecatedDefinitions) Filter(_ *generator.Context, t *types.Type) bool {
	// Filter out types not being processed or not copyable within the package.
	if !isAPIType(t) {
		klog.V(2).Infof("Type %v is not a valid target for status", t)
		return false
	}
	g.typesForInit = append(g.typesForInit, t)
	return true
}

// isAPIType indicates whether or not a type could be used to serve an API.  That means, "does it have TypeMeta".
// This doesn't mean the type is served, but we will handle all TypeMeta types.
func isAPIType(t *types.Type) bool {
	// Filter out private types.
	if namer.IsPrivateGoName(t.Name.Name) {
		return false
	}

	if t.Kind != types.Struct {
		return false
	}

	for _, currMember := range t.Members {
		if currMember.Embedded && currMember.Name == "TypeMeta" {
			return true
		}
	}

	if t.Kind == types.Alias {
		return isAPIType(t.Underlying)
	}

	return false
}

func (g *genDeprecatedDefinitions) isOtherPackage(pkg string) bool {
	if pkg == g.targetPackage {
		return false
	}
	if strings.HasSuffix(pkg, "\""+g.targetPackage+"\"") {
		return false
	}
	return true
}

func (g *genDeprecatedDefinitions) Imports(_ *generator.Context) (imports []string) {
	importLines := []string{}
	for _, singleImport := range g.imports.ImportLines() {
		if g.isOtherPackage(singleImport) {
			importLines = append(importLines, singleImport)
		}
	}
	return importLines
}

func (g *genDeprecatedDefinitions) argsFromType(_ *generator.Context, t *types.Type) (*APIDeprecation, error) {
	_, introducedMajor, introducedMinor, err := extractIntroducedTag(t)
	if err != nil {
		return nil, err
	}

	// compute based on our policy
	deprecatedMajor := introducedMajor
	deprecatedMinor := introducedMinor + 3
	// if someone intentionally override the deprecation release
	exists, err := tagExists(deprecatedTagName, t)
	if err != nil {
		return nil, err
	}
	if exists {
		_, deprecatedMajor, deprecatedMinor, err = extractDeprecatedTag(t)
		if err != nil {
			return nil, err
		}
	}

	// compute based on our policy
	removedMajor := deprecatedMajor
	removedMinor := deprecatedMinor + 3
	// if someone intentionally override the removed release
	exists, err = tagExists(removedTagName, t)
	if err != nil {
		return nil, err
	}
	if exists {
		_, removedMajor, removedMinor, err = extractRemovedTag(t)
		if err != nil {
			return nil, err
		}
	}

	replacementGroup, replacementVersion, replacementKind, hasReplacement, err := extractReplacementTag(t)
	if err != nil {
		return nil, err
	}

	reg := APIDeprecation{
		IntroducedVersion: Version{
			VersionMajor: introducedMajor,
			VersionMinor: introducedMinor,
		},
		DeprecatedVersion: Version{
			VersionMajor: deprecatedMajor,
			VersionMinor: deprecatedMinor,
		},
		RemovedVersion: Version{
			VersionMajor: removedMajor,
			VersionMinor: removedMinor,
		},
	}

	if hasReplacement {
		reg.Replacement = GroupVersionKind{
			Group:   replacementGroup,
			Version: replacementVersion,
			Kind:    replacementKind,
		}
	}

	return &reg, nil
}

func (g *genDeprecatedDefinitions) Init(_ *generator.Context, _ io.Writer) error {
	return nil
}

func (g *genDeprecatedDefinitions) GenerateType(c *generator.Context, t *types.Type, _ io.Writer) error {
	klog.V(3).Infof("Generating deprecation definitions for type %v", t)

	reg, err := g.argsFromType(c, t)
	if err != nil {
		return err
	}

	reg.Description = strings.Join(t.CommentLines, "\n")

	reg.Group = g.group
	reg.Version = g.version
	reg.Kind = t.Name.Name

	g.registry.mu.Lock()
	defer g.registry.mu.Unlock()

	g.registry.registry = append(g.registry.registry, *reg)
	return nil
}
