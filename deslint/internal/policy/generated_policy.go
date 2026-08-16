// Code generated from https://ansldes.dev/schema/design-system-policy.v1.json; DO NOT EDIT.
// policy schema SHA-256: 2b962a65f488095a2b784713d4b7ba64c29ee52b4f098c2824dcaffe4527a36b

package policy

const SchemaVersion = 1
const SchemaSHA256 = "2b962a65f488095a2b784713d4b7ba64c29ee52b4f098c2824dcaffe4527a36b"

type Policy struct {
	Schema        string            `json:"$schema,omitempty"`
	SchemaVersion int               `json:"schemaVersion"`
	DefinitionID  string            `json:"definitionId"`
	Severities    map[string]string `json:"severities"`
	Source        SourcePolicy      `json:"source"`
	Evidence      EvidencePolicy    `json:"evidence"`
	Budgets       Budgets           `json:"budgets"`
	Exceptions    []Exception       `json:"exceptions"`
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
	LayoutDocumentSHA256 string   `json:"layoutDocumentSha256,omitempty"`
}

type Budgets struct {
	Error    int `json:"error"`
	Warning  int `json:"warning"`
	Raw      int `json:"raw"`
	Overflow int `json:"overflow"`
	Overlap  int `json:"overlap"`
}

type Exception struct {
	RuleID    string `json:"ruleId"`
	Path      string `json:"path"`
	Owner     string `json:"owner"`
	Rationale string `json:"rationale"`
	ExpiresAt string `json:"expiresAt"`
}
