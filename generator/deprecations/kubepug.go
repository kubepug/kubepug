package deprecations

import (
	"io"
	"path"
	"strings"
	"sync"

	"k8s.io/gengo/v2/generator"
	"k8s.io/gengo/v2/namer"
	"k8s.io/gengo/v2/types"
	"k8s.io/klog/v2"
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

type genPreleaseLifecycle struct {
	generator.GoGenerator
	targetPackage  string
	imports        namer.ImportTracker
	typesForInit   []*types.Type
	registry       *APIRegistry
	group, version string
}

// NewPrereleaseLifecycleGen creates a generator for the prerelease-lifecycle-generator
func (r *APIRegistry) NewDeprecatedDefinitionsGen(targetPackage, group, version string) generator.Generator {
	return &genPreleaseLifecycle{
		GoGenerator: generator.GoGenerator{
			OutputFilename: "/tmp/xxx",
		},
		group:         group,
		version:       version,
		targetPackage: targetPackage,
		imports:       generator.NewImportTracker(),
		typesForInit:  make([]*types.Type, 0),
		registry:      r,
	}
}

func (g *genPreleaseLifecycle) GenerateType(c *generator.Context, t *types.Type, _ io.Writer) error {
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

func (g *genPreleaseLifecycle) argsFromType(_ *generator.Context, t *types.Type) (*APIDeprecation, error) {
	_, introducedMajor, introducedMinor, err := extractIntroducedTag(t)
	if err != nil {
		return nil, err
	}
	reg := APIDeprecation{
		IntroducedVersion: Version{
			VersionMajor: introducedMajor,
			VersionMinor: introducedMinor,
		},
	}
	// Take version from package last segment.
	// Use heuristic to determine whether the package is GA or prerelease.
	// If the package is GA, the version matches the format vN where N is a number.
	version := path.Base(t.Name.Package)
	isGAVersion := isGAVersionRegex.MatchString(version)

	// compute based on our policy
	hasDeprecated := tagExists(deprecatedTagName, t)
	hasRemoved := tagExists(removedTagName, t)

	deprecatedMajor := introducedMajor
	deprecatedMinor := introducedMinor + 3
	// if someone intentionally override the deprecation release
	if hasDeprecated {
		_, deprecatedMajor, deprecatedMinor, err = extractDeprecatedTag(t)
		if err != nil {
			return nil, err
		}
	}

	if !isGAVersion || hasDeprecated {
		reg.DeprecatedVersion = Version{
			VersionMajor: deprecatedMajor,
			VersionMinor: deprecatedMinor,
		}
	}

	// compute based on our policy
	removedMajor := deprecatedMajor
	removedMinor := deprecatedMinor + 3
	// if someone intentionally override the removed release
	if hasRemoved {
		_, removedMajor, removedMinor, err = extractRemovedTag(t)
		if err != nil {
			return nil, err
		}
	}

	if !isGAVersion || hasRemoved {
		reg.RemovedVersion = Version{
			VersionMajor: removedMajor,
			VersionMinor: removedMinor,
		}
	}

	replacementGroup, replacementVersion, replacementKind, hasReplacement, err := extractReplacementTag(t)
	if err != nil {
		return nil, err
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
