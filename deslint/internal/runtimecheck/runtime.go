// Package runtimecheck evaluates captured runtime failures, resting visibility, and text alignment.
package runtimecheck

import (
	"bytes"
	"encoding/json"
	"fmt"
	"slices"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// Evidence is one independently captured Web or native runtime inventory.
type Evidence struct {
	Schema               string                  `json:"$schema,omitempty"`
	SchemaVersion        int                     `json:"schemaVersion"`
	EvidenceKind         diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform             string                  `json:"platform"`
	SurfaceID            string                  `json:"surfaceId"`
	RuntimePolicyVersion string                  `json:"runtimePolicyVersion"`
	CaptureStatus        string                  `json:"captureStatus"`
	DetectorFailure      *DetectorFailure        `json:"detectorFailure"`
	Routes               []Route                 `json:"routes"`
}

// DetectorFailure describes provider-process failure without turning it into an application finding.
type DetectorFailure struct {
	Stage string `json:"stage"`
	Code  string `json:"code"`
}

// DetectorProcessError distinguishes evidence acquisition failure from captured application failure.
type DetectorProcessError struct {
	SurfaceID string
	Stage     string
	Code      string
}

func (e *DetectorProcessError) Error() string {
	return fmt.Sprintf("runtime detector process failed for surface %q at %s: %s", e.SurfaceID, e.Stage, e.Code)
}

// Route contains runtime observations owned by one consumer route or screen.
type Route struct {
	ID       string               `json:"id"`
	Owner    string               `json:"owner"`
	Failures []RuntimeFailure     `json:"failures"`
	Content  []ContentObservation `json:"content"`
	Text     []TextObservation    `json:"text"`
}

// RuntimeFailure is one browser or native failure signal.
type RuntimeFailure struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
}

// ContentObservation records reveal completion and no-script fallback visibility.
type ContentObservation struct {
	ID                          string `json:"id"`
	Owner                       string `json:"owner"`
	Importance                  string `json:"importance"`
	RevealStatus                string `json:"revealStatus"`
	HiddenAtRest                bool   `json:"hiddenAtRest"`
	DefaultVisibleWithoutScript bool   `json:"defaultVisibleWithoutScript"`
}

// TextObservation records resolved body alignment in a concrete output context.
type TextObservation struct {
	ID        string `json:"id"`
	Owner     string `json:"owner"`
	Role      string `json:"role"`
	Alignment string `json:"alignment"`
	Context   string `json:"context"`
}

// JustifiedTextException is one exact consumer-owned print or export permission.
type JustifiedTextException struct {
	Platform  string
	SurfaceID string
	RouteID   string
	NodeID    string
	Owner     string
	Context   string
}

// Config injects the versioned runtime permission registry and common rule controls.
type Config struct {
	RegistryVersion         string
	JustifiedTextExceptions []JustifiedTextException
	Severity                func(string) diagnostic.Severity
	Active                  func(string) bool
}

var ruleIDs = []string{rules.RuleRuntimeScriptError, rules.RuleRuntimeContentHiddenAtRest, rules.RuleRuntimeJustifiedText}

// Analyze strictly parses and evaluates one runtime evidence payload.
func Analyze(path string, contents []byte, config Config) (Evidence, []diagnostic.Diagnostic, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Evidence{}, nil, err
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Evidence{}, nil, fmt.Errorf("runtime evidence has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var evidence Evidence
	if err := decoder.Decode(&evidence); err != nil {
		return Evidence{}, nil, err
	}
	if evidence.SchemaVersion != 1 || evidence.SurfaceID == "" || evidence.RuntimePolicyVersion == "" {
		return Evidence{}, nil, fmt.Errorf("runtime evidence identity and runtime policy version are required")
	}
	if !platformMatches(evidence.EvidenceKind, evidence.Platform) {
		return Evidence{}, nil, fmt.Errorf("runtime evidence kind %q is incompatible with platform %q", evidence.EvidenceKind, evidence.Platform)
	}
	if config.RegistryVersion == "" || evidence.RuntimePolicyVersion != config.RegistryVersion {
		return Evidence{}, nil, fmt.Errorf("runtime evidence policy version %q does not match consumer policy %q", evidence.RuntimePolicyVersion, config.RegistryVersion)
	}
	if evidence.CaptureStatus == "detector-process-failed" {
		if evidence.DetectorFailure == nil || evidence.DetectorFailure.Stage == "" || evidence.DetectorFailure.Code == "" || len(evidence.Routes) != 0 {
			return Evidence{}, nil, fmt.Errorf("failed detector evidence requires one failure identity and no route observations")
		}
		return evidence, nil, &DetectorProcessError{SurfaceID: evidence.SurfaceID, Stage: evidence.DetectorFailure.Stage, Code: evidence.DetectorFailure.Code}
	}
	if evidence.CaptureStatus != "completed" || evidence.DetectorFailure != nil || len(evidence.Routes) == 0 {
		return Evidence{}, nil, fmt.Errorf("completed runtime evidence requires routes and no detector failure")
	}
	if config.Severity == nil {
		config.Severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	if config.Active == nil {
		config.Active = func(string) bool { return true }
	}

	routes := append([]Route(nil), evidence.Routes...)
	sort.SliceStable(routes, func(i, j int) bool { return routes[i].ID < routes[j].ID })
	seenRoutes := map[string]bool{}
	findings := []diagnostic.Diagnostic{}
	for _, route := range routes {
		if route.ID == "" || route.Owner == "" {
			return Evidence{}, nil, fmt.Errorf("runtime route id and owner are required")
		}
		if seenRoutes[route.ID] {
			return Evidence{}, nil, fmt.Errorf("runtime evidence has duplicate route identity %q", route.ID)
		}
		seenRoutes[route.ID] = true
		routePath := path + "#/routes/" + route.ID

		failures := append([]RuntimeFailure(nil), route.Failures...)
		sort.SliceStable(failures, func(i, j int) bool { return failures[i].ID < failures[j].ID })
		if err := uniqueIDs("failure", failures, func(value RuntimeFailure) string { return value.ID }); err != nil {
			return Evidence{}, nil, fmt.Errorf("runtime route %q: %w", route.ID, err)
		}
		for _, failure := range failures {
			if !failureKindMatches(failure.Kind, evidence.Platform) {
				return Evidence{}, nil, fmt.Errorf("runtime route %q failure %q is incompatible with platform %q", route.ID, failure.Kind, evidence.Platform)
			}
			findings = add(findings, true, rules.RuleRuntimeScriptError, failureMessage(failure.Kind), routePath+"/failures/"+failure.ID, route.Owner, evidence, config)
		}

		content := append([]ContentObservation(nil), route.Content...)
		sort.SliceStable(content, func(i, j int) bool { return content[i].ID < content[j].ID })
		if err := uniqueIDs("content", content, func(value ContentObservation) string { return value.ID }); err != nil {
			return Evidence{}, nil, fmt.Errorf("runtime route %q: %w", route.ID, err)
		}
		for _, observation := range content {
			if err := validateContent(observation); err != nil {
				return Evidence{}, nil, fmt.Errorf("runtime route %q content %q: %w", route.ID, observation.ID, err)
			}
			completedHidden := observation.Importance == "primary" && observation.RevealStatus == "completed" && observation.HiddenAtRest
			failedHidden := observation.Importance == "primary" && observation.RevealStatus == "failed" && !observation.DefaultVisibleWithoutScript
			message := "primary content remains hidden after reveal completion"
			if failedHidden {
				message = "primary content is not default-visible when reveal execution fails"
			}
			findings = add(findings, completedHidden || failedHidden, rules.RuleRuntimeContentHiddenAtRest, message, routePath+"/content/"+observation.ID, observation.Owner, evidence, config)
		}

		textNodes := append([]TextObservation(nil), route.Text...)
		sort.SliceStable(textNodes, func(i, j int) bool { return textNodes[i].ID < textNodes[j].ID })
		if err := uniqueIDs("text", textNodes, func(value TextObservation) string { return value.ID }); err != nil {
			return Evidence{}, nil, fmt.Errorf("runtime route %q: %w", route.ID, err)
		}
		for _, observation := range textNodes {
			if err := validateText(observation); err != nil {
				return Evidence{}, nil, fmt.Errorf("runtime route %q text %q: %w", route.ID, observation.ID, err)
			}
			permission := exactJustifyPermission(config.JustifiedTextExceptions, evidence, route, observation)
			findings = add(findings, observation.Role == "body" && observation.Alignment == "justify" && !permission, rules.RuleRuntimeJustifiedText, "body copy uses justified alignment without an exact print or export permission", routePath+"/text/"+observation.ID, observation.Owner, evidence, config)
		}
	}
	diagnostic.Sort(findings)
	return evidence, diagnostic.MergeCanonical(findings), nil
}

// RuleIDs returns the exact three-rule runtime membership.
func RuleIDs() []string { return append([]string(nil), ruleIDs...) }

func platformMatches(kind diagnostic.EvidenceKind, platform string) bool {
	switch kind {
	case diagnostic.EvidenceWebRendered:
		return platform == "web"
	case diagnostic.EvidenceSimulator:
		return platform == "ios"
	case diagnostic.EvidenceEmulator:
		return platform == "android"
	case diagnostic.EvidencePhysicalDevice:
		return platform == "ios" || platform == "android"
	case diagnostic.EvidenceDefinition, diagnostic.EvidenceWebSource, diagnostic.EvidenceNativeSource,
		diagnostic.EvidenceDesignDocumentSource, diagnostic.EvidenceDesignDocumentComputed,
		diagnostic.EvidenceConsumerConformance, diagnostic.EvidenceConsumerContentRegistry, diagnostic.EvidenceExecution:
		return false
	}
	return false
}

func failureKindMatches(kind, platform string) bool {
	web := []string{"console-error", "unhandled-rejection", "route-render-failure", "uncaught-error", "parse-failure"}
	native := []string{"app-crash", "redbox", "render-error-boundary"}
	if platform == "web" {
		return slices.Contains(web, kind)
	}
	return slices.Contains(native, kind)
}

func failureMessage(kind string) string {
	return map[string]string{
		"console-error":         "console error was captured during route execution",
		"unhandled-rejection":   "unhandled promise rejection was captured during route execution",
		"route-render-failure":  "route render failed",
		"uncaught-error":        "uncaught runtime error was captured",
		"parse-failure":         "loaded script failed to parse",
		"app-crash":             "native application crashed",
		"redbox":                "native development error overlay was captured",
		"render-error-boundary": "native render error boundary was activated",
	}[kind]
}

func validateContent(value ContentObservation) error {
	if value.ID == "" || value.Owner == "" || !slices.Contains([]string{"primary", "supporting", "decorative"}, value.Importance) ||
		!slices.Contains([]string{"none", "completed", "failed"}, value.RevealStatus) {
		return fmt.Errorf("contains a missing identity or unknown enum value")
	}
	return nil
}

func validateText(value TextObservation) error {
	if value.ID == "" || value.Owner == "" || !slices.Contains([]string{"body", "caption", "ui", "data", "code"}, value.Role) ||
		!slices.Contains([]string{"start", "center", "end", "justify"}, value.Alignment) ||
		!slices.Contains([]string{"screen", "print", "export"}, value.Context) {
		return fmt.Errorf("contains a missing identity or unknown enum value")
	}
	return nil
}

func exactJustifyPermission(exceptions []JustifiedTextException, evidence Evidence, route Route, text TextObservation) bool {
	for _, exception := range exceptions {
		if exception.Platform == evidence.Platform && exception.SurfaceID == evidence.SurfaceID && exception.RouteID == route.ID &&
			exception.NodeID == text.ID && exception.Owner == text.Owner && exception.Context == text.Context &&
			(text.Context == "print" || text.Context == "export") {
			return true
		}
	}
	return false
}

func uniqueIDs[T any](kind string, values []T, id func(T) string) error {
	seen := map[string]bool{}
	for _, value := range values {
		identity := id(value)
		if identity == "" {
			return fmt.Errorf("%s identity is required", kind)
		}
		if seen[identity] {
			return fmt.Errorf("duplicate %s identity %q", kind, identity)
		}
		seen[identity] = true
	}
	return nil
}

func add(findings []diagnostic.Diagnostic, condition bool, ruleID, message, path, owner string, evidence Evidence, config Config) []diagnostic.Diagnostic {
	if !condition || !config.Active(ruleID) {
		return findings
	}
	return append(findings, diagnostic.NewWithSources(ruleID, []string{strings.TrimPrefix(ruleID, "runtime/")}, config.Severity(ruleID), message, path, nil, evidence.EvidenceKind, evidence.Platform, owner, "runtime"))
}
