// Code generated from https://ansldes.dev/schema/consumer-conformance.v1.json; DO NOT EDIT.
// consumer conformance schema SHA-256: 28b1b1e232e7ed4f61d66a95839b225c493701f0584fab74baca7e94153c0975

package conformance

const SchemaVersion = 1
const SchemaSHA256 = "28b1b1e232e7ed4f61d66a95839b225c493701f0584fab74baca7e94153c0975"

type Evidence struct {
	Schema        string    `json:"$schema,omitempty"`
	SchemaVersion int       `json:"schemaVersion"`
	ProfileID     string    `json:"profileId"`
	Platform      string    `json:"platform"`
	SurfaceID     string    `json:"surfaceId"`
	Controls      []Control `json:"controls"`
}

type Control struct {
	ID                   string   `json:"id"`
	ActionID             string   `json:"actionId"`
	Role                 string   `json:"role"`
	Component            string   `json:"component"`
	Label                string   `json:"label"`
	ShapeToken           string   `json:"shapeToken"`
	Icon                 string   `json:"icon,omitempty"`
	Feedback             string   `json:"feedback"`
	States               []string `json:"states"`
	ContractStatus       string   `json:"contractStatus"`
	AffordanceSource     string   `json:"affordanceSource"`
	MotionPurpose        string   `json:"motionPurpose"`
	MotionRecipeStatus   string   `json:"motionRecipeStatus"`
	ReduceMotionFallback bool     `json:"reduceMotionFallback"`
	RawDurationMS        float64  `json:"rawDurationMs,omitempty"`
	Prominence           string   `json:"prominence"`
	NativePrimitive      bool     `json:"nativePrimitive"`
	ExceptionID          string   `json:"exceptionId,omitempty"`
}
