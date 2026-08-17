// Package nativecheck evaluates React Native source contracts and native runtime captures.
package nativecheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
)

// Thresholds are consumer-owned deterministic performance boundaries.
type Thresholds struct {
	MaxSynchronousStartupMS     float64
	MaxInitializationMS         float64
	MaxMainThreadWorkMS         float64
	MaxFrameDropRatio           float64
	MaxThumbnailDecodeRatio     float64
	MaxJSBundleRegressionBytes  int
	MaxAppBinaryRegressionBytes int
}

// RuntimeRequirement is one exact platform capture required by consumer policy.
type RuntimeRequirement struct {
	ID               string
	Platform         string
	EvidenceKind     diagnostic.EvidenceKind
	FormFactor       string
	Orientation      string
	WindowMode       string
	FoldPosture      string
	Theme            string
	MinimumFontScale float64
	ReduceMotion     bool
}

// Config injects versioned policy thresholds, capture requirements, and common rule controls.
type Config struct {
	RegistryVersion          string
	IOSAdjacentTargetSpacing float64
	Thresholds               Thresholds
	RequiredRuntimeCaptures  []RuntimeRequirement
	Severity                 func(string) diagnostic.Severity
	Active                   func(string) bool
}

func normalizeConfig(config Config) Config {
	if config.Severity == nil {
		config.Severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	if config.Active == nil {
		config.Active = func(string) bool { return true }
	}
	return config
}

func strictDecode(contents []byte, label string, target any) error {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return err
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return fmt.Errorf("%s has duplicate keys: %s", label, strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	return nil
}

func add(findings []diagnostic.Diagnostic, condition bool, ruleID, sourceID, message, path, owner string, kind diagnostic.EvidenceKind, platform string, config Config) []diagnostic.Diagnostic {
	if !condition || !config.Active(ruleID) {
		return findings
	}
	return append(findings, diagnostic.NewWithSources(ruleID, []string{sourceID}, config.Severity(ruleID), message, path, nil, kind, platform, owner, "native-conformance"))
}

func uniqueOwnedIDs[T any](label string, values []T, identity func(T) (string, string)) error {
	seen := map[string]bool{}
	for _, value := range values {
		id, owner := identity(value)
		if id == "" || owner == "" {
			return fmt.Errorf("%s id and owner are required", label)
		}
		if seen[id] {
			return fmt.Errorf("duplicate %s identity %q", label, id)
		}
		seen[id] = true
	}
	return nil
}
