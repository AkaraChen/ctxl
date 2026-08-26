package schema

type Kind string

const (
	KindSingular Kind = "singular"
	KindPlural   Kind = "plural"
)

type Format string

const (
	FormatMarkdown Format = "markdown"
	FormatNDJSON   Format = "ndjson"
	FormatSymlink  Format = "symlink"
)

type WriteMode string

const (
	WriteReplace WriteMode = "replace"
	WriteSection WriteMode = "section"
)

type Scope string

const (
	ScopeProject Scope = "project"
	ScopeGlobal  Scope = "global"
	ScopeBoth    Scope = "both"
)

type Location string

const (
	LocationRoot  Location = "root"
	LocationStore Location = "store"
)

type FieldType string

const (
	TypeString FieldType = "string"
	TypeInt    FieldType = "int"
	TypeObject FieldType = "object"
)

type GenerationMode string

const (
	GenerationStandalone     GenerationMode = "standalone"
	GenerationExistingModule GenerationMode = "existing-module"
)

type SkillType string

const (
	SkillBuiltin SkillType = "builtin"
	SkillCustom  SkillType = "custom"
)

type InjectPosition string

const (
	InjectBefore InjectPosition = "before"
	InjectAfter  InjectPosition = "after"
)

type NameOverride struct {
	Name string `json:"name,omitempty"`
}

type Generation struct {
	Mode        GenerationMode `json:"mode,omitempty"`
	Output      string         `json:"output,omitempty"`
	Module      string         `json:"module,omitempty"`
	CtxlVersion string         `json:"ctxl_version,omitempty"`
}

type Skill struct {
	Type          SkillType         `json:"type"`
	Directory     string            `json:"directory,omitempty"`
	Inject        InjectPosition    `json:"inject,omitempty"`
	Name          string            `json:"name,omitempty"`
	Description   string            `json:"description,omitempty"`
	License       string            `json:"license,omitempty"`
	Compatibility string            `json:"compatibility,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	AllowedTools  string            `json:"allowed-tools,omitempty"`
}

type Field struct {
	Name        string    `json:"name"`
	Type        FieldType `json:"type"`
	Required    bool      `json:"required,omitempty"`
	Description string    `json:"description,omitempty"`
}

type Entity struct {
	Name        string       `json:"name"`
	Kind        Kind         `json:"kind"`
	Format      Format       `json:"format"`
	Path        string       `json:"path"`
	Location    Location     `json:"location,omitempty"`
	Scope       Scope        `json:"scope,omitempty"`
	ID          string       `json:"id,omitempty"`
	Write       WriteMode    `json:"write,omitempty"`
	Section     string       `json:"section,omitempty"`
	Target      string       `json:"target,omitempty"`
	Body        string       `json:"body,omitempty"`
	Description string       `json:"description,omitempty"`
	Command     NameOverride `json:"command,omitempty"`
	Fields      []Field      `json:"fields,omitempty"`
}

type Schema struct {
	SchemaURL   string       `json:"$schema,omitempty"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Generation  Generation   `json:"generation,omitempty"`
	CLI         NameOverride `json:"cli,omitempty"`
	Store       NameOverride `json:"store,omitempty"`
	Skills      []Skill      `json:"skills,omitempty"`
	Entities    []Entity     `json:"entities"`
}
