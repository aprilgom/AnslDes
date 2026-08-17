// Code generated from https://ansldes.dev/schema/design-system-policy.v1.json; DO NOT EDIT.
// policy schema SHA-256: b0af57833aa55b97ccd9f7f628a4734fc05ecaa61c12c09492b2a41531ef2295

package policy

const SchemaVersion = 1
const SchemaSHA256 = "b0af57833aa55b97ccd9f7f628a4734fc05ecaa61c12c09492b2a41531ef2295"

type Policy struct {
	Schema        string                `json:"$schema,omitempty"`
	SchemaVersion int                   `json:"schemaVersion"`
	DefinitionID  string                `json:"definitionId"`
	Profile       *ConsumerProfile      `json:"profile,omitempty"`
	Severities    map[string]string     `json:"severities"`
	Source        SourcePolicy          `json:"source"`
	Content       *ContentPolicy        `json:"content,omitempty"`
	Assets        *AssetPolicy          `json:"assets,omitempty"`
	Runtime       *RuntimePolicy        `json:"runtime,omitempty"`
	Native        *NativePolicy         `json:"native,omitempty"`
	Web           *WebPolicy            `json:"web,omitempty"`
	Evidence      EvidencePolicy        `json:"evidence"`
	Budgets       Budgets               `json:"budgets"`
	Exceptions    []Exception           `json:"exceptions"`
	RulePacks     []RulePackRequirement `json:"rulePacks"`
	RuleOverrides []RuleOverride        `json:"ruleOverrides"`
	Governance    GovernancePolicy      `json:"governance"`
}

type ContentPolicy struct {
	RegistryVersion  string                         `json:"registryVersion"`
	Locales          map[string]LocaleContentPolicy `json:"locales"`
	SourceReferences []string                       `json:"sourceReferences"`
}

type LocaleContentPolicy struct {
	PhraseRegistryVersion string   `json:"phraseRegistryVersion"`
	MarketingBuzzwords    []string `json:"marketingBuzzwords"`
	TheaterPhrases        []string `json:"theaterPhrases"`
	ProtectedTerms        []string `json:"protectedTerms"`
	RecoveryCopyIDs       []string `json:"recoveryCopyIds"`
}

type AssetPolicy struct {
	RegistryVersion string                `json:"registryVersion"`
	Entries         map[string]AssetEntry `json:"entries"`
}

type AssetEntry struct {
	Owner                string   `json:"owner"`
	Role                 string   `json:"role"`
	ImplementationSource string   `json:"implementationSource"`
	Consumers            []string `json:"consumers"`
	FingerprintSHA256    string   `json:"fingerprintSha256"`
	IntentionallyOmitted bool     `json:"intentionallyOmitted"`
	Decorative           bool     `json:"decorative"`
}

type RuntimePolicy struct {
	RegistryVersion         string                   `json:"registryVersion"`
	JustifiedTextExceptions []JustifiedTextException `json:"justifiedTextExceptions"`
}

type JustifiedTextException struct {
	Platform  string `json:"platform"`
	SurfaceID string `json:"surfaceId"`
	RouteID   string `json:"routeId"`
	NodeID    string `json:"nodeId"`
	Owner     string `json:"owner"`
	Context   string `json:"context"`
}

type NativePolicy struct {
	RegistryVersion          string                     `json:"registryVersion"`
	IOSAdjacentTargetSpacing float64                    `json:"iosAdjacentTargetSpacing"`
	Thresholds               NativeThresholds           `json:"thresholds"`
	RequiredRuntimeCaptures  []NativeRuntimeRequirement `json:"requiredRuntimeCaptures"`
}

type NativeThresholds struct {
	MaxSynchronousStartupMS     float64 `json:"maxSynchronousStartupMs"`
	MaxInitializationMS         float64 `json:"maxInitializationMs"`
	MaxMainThreadWorkMS         float64 `json:"maxMainThreadWorkMs"`
	MaxFrameDropRatio           float64 `json:"maxFrameDropRatio"`
	MaxThumbnailDecodeRatio     float64 `json:"maxThumbnailDecodeRatio"`
	MaxJSBundleRegressionBytes  int     `json:"maxJsBundleRegressionBytes"`
	MaxAppBinaryRegressionBytes int     `json:"maxAppBinaryRegressionBytes"`
}

type NativeRuntimeRequirement struct {
	ID               string  `json:"id"`
	Platform         string  `json:"platform"`
	EvidenceKind     string  `json:"evidenceKind"`
	FormFactor       string  `json:"formFactor"`
	Orientation      string  `json:"orientation"`
	WindowMode       string  `json:"windowMode"`
	FoldPosture      string  `json:"foldPosture"`
	Theme            string  `json:"theme"`
	MinimumFontScale float64 `json:"minimumFontScale"`
	ReduceMotion     bool    `json:"reduceMotion"`
}

type WebPolicy struct {
	RegistryVersion    string                  `json:"registryVersion"`
	BuildCommand       string                  `json:"buildCommand"`
	Routes             []WebRoute              `json:"routes"`
	Viewports          []WebViewport           `json:"viewports"`
	Themes             []string                `json:"themes"`
	FontScales         []float64               `json:"fontScales"`
	ReduceMotion       []bool                  `json:"reduceMotion"`
	RequiredCaptures   []WebCaptureRequirement `json:"requiredCaptures"`
	ArtifactExclusions []WebArtifactExclusion  `json:"artifactExclusions"`
}

type WebRoute struct {
	ID     string `json:"id"`
	Owner  string `json:"owner"`
	Target string `json:"target"`
}

type WebViewport struct {
	ID     string `json:"id"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

type WebCaptureRequirement struct {
	ID           string  `json:"id"`
	Provider     string  `json:"provider"`
	RouteID      string  `json:"routeId"`
	ViewportID   string  `json:"viewportId"`
	Theme        string  `json:"theme"`
	FontScale    float64 `json:"fontScale"`
	ReduceMotion bool    `json:"reduceMotion"`
}

type WebArtifactExclusion struct {
	Path                string `json:"path"`
	FingerprintSHA256   string `json:"fingerprintSha256"`
	Owner               string `json:"owner"`
	Rationale           string `json:"rationale"`
	ReproductionCommand string `json:"reproductionCommand"`
}

type ConsumerProfile struct {
	ID                       string            `json:"id"`
	PrimaryUserGoal          string            `json:"primaryUserGoal"`
	Density                  string            `json:"density"`
	NoveltyTolerance         string            `json:"noveltyTolerance"`
	NativeAffordancePriority string            `json:"nativeAffordancePriority"`
	SeverityOverrides        map[string]string `json:"severityOverrides"`
	Thresholds               ProfileThresholds `json:"thresholds"`
	RequiredEvidence         []string          `json:"requiredEvidence"`
	Rationale                string            `json:"rationale,omitempty"`
	Reviewer                 string            `json:"reviewer,omitempty"`
	EvidenceOwner            string            `json:"evidenceOwner,omitempty"`
}

type ProfileThresholds struct {
	MaxOversizedActions    int `json:"maxOversizedActions"`
	MaxInconsistentActions int `json:"maxInconsistentActions"`
}

type RulePackRequirement struct {
	ID                string `json:"id"`
	Version           string `json:"version"`
	FingerprintSHA256 string `json:"fingerprintSha256"`
}

type RuleOverride struct {
	RuleID        string `json:"ruleId"`
	PackID        string `json:"packId"`
	PackVersion   string `json:"packVersion"`
	Status        string `json:"status"`
	Owner         string `json:"owner"`
	Rationale     string `json:"rationale"`
	Reviewer      string `json:"reviewer"`
	ExpiresAt     string `json:"expiresAt"`
	ReviewTrigger string `json:"reviewTrigger"`
}

type GovernancePolicy struct {
	ReviewedAt              string            `json:"reviewedAt"`
	ReviewIntervalDays      int               `json:"reviewIntervalDays"`
	Reviewer                string            `json:"reviewer"`
	ReviewSubjects          []string          `json:"reviewSubjects"`
	AdvisoryMode            string            `json:"advisoryMode"`
	ForbiddenFlags          []string          `json:"forbiddenFlags"`
	RequireExitCode2        bool              `json:"requireExitCode2"`
	RequireUnmodifiedReport bool              `json:"requireUnmodifiedReport"`
	PassingReportsOnly      bool              `json:"passingReportsOnly"`
	Ignores                 []IgnoreAllowance `json:"ignores"`
}

type IgnoreAllowance struct {
	Kind          string `json:"kind"`
	RuleID        string `json:"ruleId,omitempty"`
	Engine        string `json:"engine"`
	Platform      string `json:"platform"`
	Path          string `json:"path"`
	Property      string `json:"property,omitempty"`
	Value         string `json:"value,omitempty"`
	Owner         string `json:"owner"`
	Rationale     string `json:"rationale"`
	Reviewer      string `json:"reviewer"`
	ExpiresAt     string `json:"expiresAt"`
	ReviewTrigger string `json:"reviewTrigger"`
}

type SourcePolicy struct {
	RawProperties RawProperties `json:"rawProperties"`
	ExactExcludes []string      `json:"exactExcludes"`
}

type RawProperties struct {
	Color  []string `json:"color"`
	Number []string `json:"number"`
	Motion []string `json:"motion"`
}

type EvidencePolicy struct {
	RequiredKinds        []string `json:"requiredKinds"`
	DeferredKinds        []string `json:"deferredKinds,omitempty"`
	LayoutDocumentSHA256 string   `json:"layoutDocumentSha256,omitempty"`
}

type Budgets struct {
	Error     int `json:"error"`
	Warning   int `json:"warning"`
	Raw       int `json:"raw"`
	Overflow  int `json:"overflow"`
	Overlap   int `json:"overlap"`
	Blocking  int `json:"blocking"`
	Advisory  int `json:"advisory"`
	Exception int `json:"exception"`
	NotRun    int `json:"notRun"`
	Deferred  int `json:"deferred"`
}

type Exception struct {
	RuleID        string `json:"ruleId"`
	Engine        string `json:"engine"`
	Platform      string `json:"platform"`
	Path          string `json:"path"`
	Owner         string `json:"owner"`
	Rationale     string `json:"rationale"`
	Reviewer      string `json:"reviewer"`
	ExpiresAt     string `json:"expiresAt"`
	ReviewTrigger string `json:"reviewTrigger"`
}
