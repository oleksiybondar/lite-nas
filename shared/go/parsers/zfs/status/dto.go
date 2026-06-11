package status

// ParseMode defines parser strictness behavior.
type ParseMode string

const (
	// ParseModeStrict stops with an error when parser diagnostics contain errors.
	ParseModeStrict ParseMode = "strict"
	// ParseModeTolerant returns best-effort output with diagnostics when possible.
	ParseModeTolerant ParseMode = "tolerant"
)

// DiagnosticSeverity defines the severity level of a parse diagnostic.
type DiagnosticSeverity string

const (
	// DiagnosticSeverityError marks a non-recoverable grammar/structure issue.
	DiagnosticSeverityError DiagnosticSeverity = "error"
	// DiagnosticSeverityWarning marks a recoverable or soft parser issue.
	DiagnosticSeverityWarning DiagnosticSeverity = "warning"
)

// Diagnostic describes one parser issue with source location context.
type Diagnostic struct {
	Severity DiagnosticSeverity
	Message  string
	Line     int
	Column   int
}

// StatusDocument is the root parsed model for one or more zpool status blocks.
type StatusDocument struct {
	Pools []PoolBlock
}

// PoolBlock represents one parsed pool block from zpool status output.
type PoolBlock struct {
	PoolName      string
	Metadata      PoolMetadata
	Config        ConfigTree
	ErrorsSummary string
}

// PoolMetadata contains top-level metadata lines for a pool.
type PoolMetadata struct {
	State  string
	Scan   string
	Status string
	Action string
	See    string
}

// ConfigTree contains parsed header and hierarchical rows from config section.
type ConfigTree struct {
	Header []string
	Roots  []ConfigNode
}

// ConfigNode represents one row in the config section hierarchy.
type ConfigNode struct {
	Name     string
	Columns  map[string]string
	Indent   int
	Children []ConfigNode
}
