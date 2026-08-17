// Package colorcheck evaluates theme-scoped computed color evidence.
package colorcheck

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/aprilgom/AnslDes/deslint/internal/diagnostic"
	"github.com/aprilgom/AnslDes/deslint/internal/jsoncheck"
	"github.com/aprilgom/AnslDes/deslint/internal/rules"
)

// Evidence contains screenshot and computed-color records for one exact theme.
type Evidence struct {
	Schema            string                  `json:"$schema,omitempty"`
	SchemaVersion     int                     `json:"schemaVersion"`
	EvidenceKind      diagnostic.EvidenceKind `json:"evidenceKind"`
	Platform          string                  `json:"platform"`
	Theme             string                  `json:"theme"`
	SurfaceID         string                  `json:"surfaceId"`
	ScreenshotPath    string                  `json:"screenshotPath"`
	ComputedColorPath string                  `json:"computedColorPath"`
	Nodes             []Node                  `json:"nodes"`
}

// Node is one text, surface, or asset color relationship.
type Node struct {
	ID                   string  `json:"id"`
	Owner                string  `json:"owner"`
	MediaRole            string  `json:"mediaRole"`
	Context              string  `json:"context"`
	ColorSource          string  `json:"colorSource"`
	Foreground           string  `json:"foreground"`
	Background           string  `json:"background"`
	FontSize             float64 `json:"fontSize"`
	FontWeight           int     `json:"fontWeight"`
	FillKind             string  `json:"fillKind"`
	PalettePattern       string  `json:"palettePattern"`
	PaletteApproved      bool    `json:"paletteApproved"`
	PaletteID            string  `json:"paletteId,omitempty"`
	SurfaceTemperature   string  `json:"surfaceTemperature"`
	GlowPresent          bool    `json:"glowPresent"`
	GlowColor            string  `json:"glowColor"`
	GlowBlur             float64 `json:"glowBlur"`
	NeutralElevation     bool    `json:"neutralElevation"`
	RadialSaturated      bool    `json:"radialSaturated"`
	RadialOpacity        float64 `json:"radialOpacity"`
	RadialDecorative     bool    `json:"radialDecorative"`
	MonochromeAsset      bool    `json:"monochromeAsset"`
	MediumExceptionOwner string  `json:"mediumExceptionOwner,omitempty"`
}

// PalettePermission is an exact consumer-definition approval scope.
type PalettePermission struct {
	Contexts []string
	Themes   []string
}

// Config supplies consumer-owned contrast thresholds and palette scopes.
type Config struct {
	BodyContrast     float64
	LargeContrast    float64
	ApprovedPalettes map[string]PalettePermission
	Severity         func(string) diagnostic.Severity
	Active           func(string) bool
}

var ruleIDs = []string{rules.RuleColorGradientText, rules.RuleColorAiColorPalette, rules.RuleColorCreamPalette, rules.RuleColorDarkGlow, rules.RuleColorRadialHalo, rules.RuleColorRadialSpotlightGlow, rules.RuleColorGrayOnColor, rules.RuleColorLowContrast, rules.RuleColorPureExtremeSurface}

// Analyze strictly parses and evaluates color evidence.
func Analyze(path string, contents []byte, config Config) (Evidence, []diagnostic.Diagnostic, error) {
	duplicates, err := jsoncheck.DuplicateKeys(contents)
	if err != nil {
		return Evidence{}, nil, err
	}
	if len(duplicates) > 0 {
		sort.Strings(duplicates)
		return Evidence{}, nil, fmt.Errorf("color evidence has duplicate keys: %s", strings.Join(duplicates, ", "))
	}
	decoder := json.NewDecoder(bytes.NewReader(contents))
	decoder.DisallowUnknownFields()
	var e Evidence
	if err := decoder.Decode(&e); err != nil {
		return Evidence{}, nil, err
	}
	if e.SchemaVersion != 1 || e.Theme == "" || e.SurfaceID == "" || e.ScreenshotPath == "" || e.ComputedColorPath == "" || len(e.Nodes) == 0 {
		return Evidence{}, nil, fmt.Errorf("color evidence identity, screenshot, computed colors, and nodes are required")
	}
	validEvidence := []diagnostic.EvidenceKind{diagnostic.EvidenceWebRendered, diagnostic.EvidenceDesignDocumentComputed, diagnostic.EvidenceSimulator, diagnostic.EvidenceEmulator, diagnostic.EvidencePhysicalDevice}
	if !slices.Contains(validEvidence, e.EvidenceKind) {
		return Evidence{}, nil, fmt.Errorf("color evidence kind %q is invalid", e.EvidenceKind)
	}
	if !platformMatches(e.EvidenceKind, e.Platform) {
		return Evidence{}, nil, fmt.Errorf("color evidence kind %q is incompatible with platform %q", e.EvidenceKind, e.Platform)
	}
	if config.BodyContrast == 0 {
		config.BodyContrast = 4.5
	}
	if config.LargeContrast == 0 {
		config.LargeContrast = 3
	}
	if config.BodyContrast < 4.5 || config.BodyContrast > 21 || config.LargeContrast < 3 || config.LargeContrast > 21 {
		return Evidence{}, nil, fmt.Errorf("color contrast registry thresholds are invalid")
	}
	if config.Severity == nil {
		config.Severity = func(string) diagnostic.Severity { return diagnostic.SeverityError }
	}
	if config.Active == nil {
		config.Active = func(string) bool { return true }
	}
	nodes := append([]Node(nil), e.Nodes...)
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	out := []diagnostic.Diagnostic{}
	seen := map[string]bool{}
	for _, n := range nodes {
		if err := validateNode(n); err != nil {
			return Evidence{}, nil, fmt.Errorf("color node %q: %w", n.ID, err)
		}
		if seen[n.ID] {
			return Evidence{}, nil, fmt.Errorf("color evidence has duplicate node identity %q", n.ID)
		}
		seen[n.ID] = true
		p := path + "#/nodes/" + n.ID
		text := n.MediaRole == "text"
		surface := n.MediaRole == "surface"
		exactMedium := n.MediumExceptionOwner != "" && n.MediumExceptionOwner == n.Owner
		paletteApproved := paletteAllowed(n, e.Theme, config.ApprovedPalettes)
		out = add(out, text && n.FillKind != "solid", rules.RuleColorGradientText, "text uses a gradient fill", p, n, e, config.Severity, config.Active, nil)
		out = add(out, !paletteApproved && n.PalettePattern != "none", rules.RuleColorAiColorPalette, "palette pattern is not approved for this theme and context", p, n, e, config.Severity, config.Active, []string{"hallmark-eight-01"})
		out = add(out, surface && n.SurfaceTemperature == "warm-cream" && !paletteApproved, rules.RuleColorCreamPalette, "warm cream surface is outside the approved palette", p, n, e, config.Severity, config.Active, nil)
		out = add(out, n.GlowPresent && n.GlowBlur >= 12 && !n.NeutralElevation && chromatic(n.GlowColor), rules.RuleColorDarkGlow, "chromatic blurred glow is decorative rather than neutral elevation", p, n, e, config.Severity, config.Active, nil)
		out = add(out, n.FillKind == "radial-gradient" && n.RadialSaturated && n.RadialDecorative, rules.RuleColorRadialHalo, "saturated decorative radial halo is used", p, n, e, config.Severity, config.Active, nil)
		out = add(out, n.FillKind == "radial-gradient" && n.RadialDecorative && n.RadialOpacity > 0 && n.RadialOpacity <= 0.3, rules.RuleColorRadialSpotlightGlow, "low-opacity decorative radial spotlight is used", p, n, e, config.Severity, config.Active, nil)
		out = add(out, text && neutral(n.Foreground) && chromatic(n.Background), rules.RuleColorGrayOnColor, "neutral gray text is placed on a chromatic surface", p, n, e, config.Severity, config.Active, nil)
		threshold := config.BodyContrast
		if n.FontSize >= 24 || (n.FontSize >= 18.66 && n.FontWeight >= 700) {
			threshold = config.LargeContrast
		}
		ratio, _ := contrast(n.Foreground, n.Background)
		out = add(out, text && ratio < threshold, rules.RuleColorLowContrast, fmt.Sprintf("computed contrast %.2f is below %.1f:1", ratio, threshold), p, n, e, config.Severity, config.Active, nil)
		extreme := strings.EqualFold(n.Background, "#000000") || strings.EqualFold(n.Background, "#FFFFFF")
		out = add(out, surface && extreme && !n.MonochromeAsset && !exactMedium, rules.RuleColorPureExtremeSurface, "pure black or white is used as a surface", p, n, e, config.Severity, config.Active, nil)
	}
	diagnostic.Sort(out)
	return e, diagnostic.MergeCanonical(out), nil
}

// RuleIDs returns the exact color rule membership for adapters and tests.
func RuleIDs() []string { return append([]string(nil), ruleIDs...) }

func platformMatches(kind diagnostic.EvidenceKind, platform string) bool {
	switch kind {
	case diagnostic.EvidenceWebRendered:
		return platform == "web"
	case diagnostic.EvidenceDesignDocumentComputed:
		return platform == "design-document"
	case diagnostic.EvidenceSimulator:
		return platform == "ios"
	case diagnostic.EvidenceEmulator:
		return platform == "android"
	case diagnostic.EvidencePhysicalDevice:
		return platform == "ios" || platform == "android"
	case diagnostic.EvidenceDefinition, diagnostic.EvidenceWebSource, diagnostic.EvidenceNativeSource,
		diagnostic.EvidenceDesignDocumentSource, diagnostic.EvidenceConsumerConformance,
		diagnostic.EvidenceConsumerContentRegistry, diagnostic.EvidenceExecution:
		return false
	}
	return false
}

func validateNode(n Node) error {
	if n.ID == "" || n.Owner == "" {
		return fmt.Errorf("id and owner are required")
	}
	if !slices.Contains([]string{"text", "surface", "monochrome-asset", "color-asset"}, n.MediaRole) ||
		!slices.Contains([]string{"page", "component", "hero", "brand", "status"}, n.Context) ||
		!slices.Contains([]string{"literal", "token", "computed"}, n.ColorSource) ||
		!slices.Contains([]string{"solid", "linear-gradient", "radial-gradient"}, n.FillKind) ||
		!slices.Contains([]string{"none", "purple-violet-gradient", "cyan-on-dark"}, n.PalettePattern) ||
		!slices.Contains([]string{"neutral", "warm-cream"}, n.SurfaceTemperature) {
		return fmt.Errorf("contains an unknown enum value")
	}
	if _, ok := rgb(n.Foreground); !ok {
		return fmt.Errorf("foreground must be #RRGGBB")
	}
	if _, ok := rgb(n.Background); !ok {
		return fmt.Errorf("background must be #RRGGBB")
	}
	if _, ok := rgb(n.GlowColor); !ok {
		return fmt.Errorf("glowColor must be #RRGGBB")
	}
	if n.FontSize < 0 || n.FontWeight < 1 || n.FontWeight > 1000 || n.GlowBlur < 0 || n.RadialOpacity < 0 || n.RadialOpacity > 1 {
		return fmt.Errorf("contains an out-of-range numeric value")
	}
	return nil
}

func paletteAllowed(n Node, theme string, registry map[string]PalettePermission) bool {
	if !n.PaletteApproved || n.PaletteID == "" {
		return false
	}
	permission, ok := registry[n.PaletteID]
	return ok && slices.Contains(permission.Contexts, n.Context) && slices.Contains(permission.Themes, theme)
}

func add(v []diagnostic.Diagnostic, c bool, id, msg, path string, n Node, e Evidence, s func(string) diagnostic.Severity, a func(string) bool, extra []string) []diagnostic.Diagnostic {
	if !c || !a(id) {
		return v
	}
	sources := append([]string{strings.TrimPrefix(id, "color/")}, extra...)
	return append(v, diagnostic.NewWithSources(id, sources, s(id), msg, path, nil, e.EvidenceKind, e.Platform, n.Owner, "color"))
}
func rgb(value string) ([3]float64, bool) {
	var out [3]float64
	b, err := hex.DecodeString(strings.TrimPrefix(value, "#"))
	if err != nil || len(b) != 3 {
		return out, false
	}
	for i := range b {
		out[i] = float64(b[i]) / 255
	}
	return out, true
}
func chromatic(value string) bool {
	v, ok := rgb(value)
	if !ok {
		return false
	}
	maxV := math.Max(v[0], math.Max(v[1], v[2]))
	minV := math.Min(v[0], math.Min(v[1], v[2]))
	return maxV-minV >= 0.12
}
func neutral(value string) bool {
	v, ok := rgb(value)
	if !ok {
		return false
	}
	maxV := math.Max(v[0], math.Max(v[1], v[2]))
	minV := math.Min(v[0], math.Min(v[1], v[2]))
	return maxV-minV <= 0.05
}
func luminance(value string) (float64, bool) {
	v, ok := rgb(value)
	if !ok {
		return 0, false
	}
	for i, c := range v {
		if c <= 0.04045 {
			v[i] = c / 12.92
		} else {
			v[i] = math.Pow((c+0.055)/1.055, 2.4)
		}
	}
	return 0.2126*v[0] + 0.7152*v[1] + 0.0722*v[2], true
}
func contrast(foreground, background string) (float64, bool) {
	a, okA := luminance(foreground)
	b, okB := luminance(background)
	if !okA || !okB {
		return 0, false
	}
	if a < b {
		a, b = b, a
	}
	return (a + 0.05) / (b + 0.05), true
}
