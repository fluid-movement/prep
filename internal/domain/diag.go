package domain

import "sort"

// Severity of a diagnostic.
type Severity string

const (
	SevError   Severity = "error"
	SevWarning Severity = "warning"
)

// Class says how a diagnostic gets healed.
type Class string

const (
	ClassFixable Class = "fixable" // repaired by prep fix
	ClassGuided  Class = "guided"  // options explained, user decides
	ClassManual  Class = "manual"  // needs a hand edit
)

// Diagnostic is one finding from validation, with a stable code.
type Diagnostic struct {
	Code     string   `json:"code"`
	Severity Severity `json:"severity"`
	Class    Class    `json:"class"`
	Issue    string   `json:"issue,omitempty"`
	File     string   `json:"file,omitempty"`
	Message  string   `json:"message"`
	Fix      string   `json:"fix"`
}

// Diagnostic codes. They are stable: never reuse or renumber.
const (
	// Project level.
	CodeProjectMissing = "P001" // project.md missing or unreadable
	CodeSchemaMismatch = "P002" // schema version differs from the binary
	CodeConfigInvalid  = "P003" // config.yaml or config.yaml.dist invalid
	CodeProjectUnknown = "P004" // unknown file in .prep
	CodeLocalTracked   = "P007" // .prep/.gitignore does not ignore local/; prep fix adds it
	CodeProjectNoID    = "P008" // project.md has no id; prep fix writes one
	CodeProjectInvalid = "P009" // project.md has an invalid id or name

	// Issue structure.
	CodeBadID              = "I001" // directory name is not a valid ID
	CodeIssueMissing       = "I002" // issue.md missing
	CodeFrontmatter        = "I003" // frontmatter cannot be parsed
	CodeTitleMissing       = "I004"
	CodeKindInvalid        = "I005"
	CodeParentMissing      = "I006"
	CodeParentCycle        = "I007"
	CodeDepMissing         = "I008"
	CodeDepCycle           = "I009"
	CodeDepOnAncestor      = "I010" // deadlock: depends on own parent or ancestor
	CodeUnknownFile        = "I011"
	CodeReadyOutdated      = "I013" // ready.md references an older baseline
	CodeOrphanClaim        = "I014" // claim.md without valid ready sign-off
	CodeRecordInvalid      = "I015" // baseline/ready/claim/resolution malformed
	CodeDepDropped         = "I016"
	CodeNotCanonical       = "I017"
	CodeSelfDependency     = "I021"
	CodeDuplicateDep       = "I022"
	CodeResolvedChildOpen  = "I023" // resolved parent with unresolved children
	CodeStaleInProgress    = "I024"
	CodeContextLinkMissing = "I025"
	CodeTagInvalid         = "I026"
	CodePriorityInvalid    = "I027"
	CodeDuplicateHeading   = "I028" // a ## section heading repeats in a record file

	// Knowledge base.
	CodeKnowledgeFrontmatter = "K001"
	CodeKnowledgeField       = "K002" // required field missing or invalid
	CodeKnowledgeLink        = "K003" // broken link
	CodeKnowledgeSize        = "K004" // entry above size threshold
	CodeLegacyKnowledge      = "K008" // legacy issue knowledge pending conversion
	CodeKnowledgeIndex       = "K007" // OKF index.md missing, outdated or unneeded

)

// SortDiagnostics orders diagnostics deterministically.
func SortDiagnostics(ds []Diagnostic) {
	sort.SliceStable(ds, func(a, b int) bool {
		x, y := ds[a], ds[b]
		if x.Issue != y.Issue {
			return x.Issue < y.Issue
		}
		if x.Code != y.Code {
			return x.Code < y.Code
		}
		if x.File != y.File {
			return x.File < y.File
		}
		return x.Message < y.Message
	})
}
