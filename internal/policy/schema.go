package policy

// Action is the decision a policy resource applies to matching commands.
//
// Only ALLOW/DENY are valid in the YAML schema (spec.action, when.action) --
// kept binary and symmetric with Istio's own ALLOW/DENY core. CONFIRM is
// derived at evaluation time, never written directly: a rule reaching
// ALLOW with its `options.confirm.enabled: true` (or a matching
// `when.confirm: true`) gets gated behind a human confirmation instead of
// proceeding silently. See Options.Confirm / When.Confirm.
type Action string

const (
	ActionAllow Action = "ALLOW"
	ActionDeny  Action = "DENY"

	// ActionConfirm never appears in a policy file. It is produced by the
	// engine from ALLOW + a confirm modifier, or as the built-in fallback
	// when no policy matches at all (see DefaultAction).
	ActionConfirm Action = "CONFIRM"
)

// Mode is the fixed vocabulary used to categorize an operation for reporting
// and matching purposes.
type Mode string

const (
	ModeRead        Mode = "read"
	ModeWrite       Mode = "write"
	ModeDestructive Mode = "destructive"
	ModeNetwork     Mode = "network"
)

// ValidModes lists every Mode value guard lint accepts. Keep in sync with
// the `oneof=...` validate tag on To.Mode below.
var ValidModes = []Mode{ModeRead, ModeWrite, ModeDestructive, ModeNetwork}

// ValidActions lists every Action value that may appear directly in a
// policy file (spec.action / when.action). Keep in sync with the
// `oneof=...` validate tags on Spec.Action and When.Action below.
// ActionConfirm is deliberately excluded -- see the Action doc comment.
var ValidActions = []Action{ActionAllow, ActionDeny}

const (
	APIVersion = "ai-guard/v1"
	Kind       = "CommandGatePolicy"

	// ModulePermission is the only module value supported in v1. It is kept
	// as an explicit field so future modules (e.g. network, notification)
	// can be introduced without a breaking schema change.
	ModulePermission = "permission"
)

// Policy is one CommandGatePolicy manifest, one file under .guard/policies/.
//
// `validate` tags declare required-ness/enum constraints and are enforced
// by validateStruct (see validate.go); lint.go only hand-codes checks that
// can't be expressed as a single field constraint (e.g. detecting two
// catch-all rules across a slice).
type Policy struct {
	APIVersion string   `yaml:"apiVersion" validate:"eq=ai-guard/v1"`
	Kind       string   `yaml:"kind" validate:"eq=CommandGatePolicy"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`

	// SourcePath is set by the loader to the file this policy came from, for
	// error messages and lint output. Not part of the YAML itself.
	SourcePath string `yaml:"-" validate:"-"`
}

type Metadata struct {
	Name string `yaml:"name" validate:"required"`
}

type Spec struct {
	Target Target `yaml:"target"`
	Action Action `yaml:"action" validate:"oneof=ALLOW DENY"`
	Rules  []Rule `yaml:"rules" validate:"min=1,dive"`
}

type Target struct {
	Module string `yaml:"module" validate:"eq=permission"`
	Tool   string `yaml:"tool" validate:"required"`
}

type Rule struct {
	To      To       `yaml:"to"`
	Options *Options `yaml:"options,omitempty"`

	// Message is this rule's general-purpose reason/next (e.g. why a DENY
	// rule denies). Options.Confirm.Message, when set, takes precedence
	// over this field for a confirmation gate's wording; this field is
	// the fallback otherwise.
	Message *Message `yaml:"message,omitempty"`
}

// To is the match condition for a rule. Functions is optional: when absent,
// the rule is a catch-all for every invocation of the target tool (used for
// flat commands such as rm that have no subcommand concept). Args is also
// optional: when absent, the rule applies regardless of flags.
type To struct {
	Mode      Mode       `yaml:"mode" validate:"oneof=read write destructive network"`
	Functions []string   `yaml:"functions,omitempty"`
	Args      *ArgsMatch `yaml:"args,omitempty"`
}

// ArgsMatch restricts a rule to commands whose argv contains at least one
// of AnyOf as a literal token (e.g. distinguishing "git push" from
// "git push -f" without pattern matching). Deliberately no regex/globs --
// keep matching predictable and auditable; this is usually paired with a
// separate, stricter resource (see git/push-force-deny.yaml) rather than
// baked into one resource with per-rule actions.
type ArgsMatch struct {
	AnyOf []string `yaml:"anyOf" validate:"min=1,dive,required"`
}

// Message is attached to a rule's decision and surfaced to the AI (via
// Claude Code's permissionDecisionReason), not shown as end-user UI text --
// write Reason/Next in English. `{cmd}` in either field is replaced with
// the exact command that matched the rule (e.g. "git add -A"); it's the
// only placeholder supported in v1.
type Message struct {
	Reason string `yaml:"reason"`
	Next   string `yaml:"next"`
}

type Options struct {
	// Confirm gates this rule's ALLOW behind a human confirmation instead
	// of proceeding silently. Only meaningful when the resource's
	// spec.action is ALLOW -- lint rejects Confirm.Enabled on a DENY rule,
	// since DENY is meant to be unconditional, not negotiable via
	// confirmation.
	Confirm     *Confirm     `yaml:"confirm,omitempty"`
	CustomCheck *CustomCheck `yaml:"customCheck,omitempty"`
}

// Confirm gates a rule's ALLOW behind a human confirmation step.
type Confirm struct {
	Enabled bool `yaml:"enabled"`

	// Message overrides Rule.Message with wording specific to this
	// confirmation gate. When unset, the rule's own Message is used
	// instead.
	Message *Message `yaml:"message,omitempty"`
}

// CustomCheck delegates part of the decision to an external script or
// inline shell command, mirroring Istio's CUSTOM authorization action.
// Exactly one of Path or Command must be set -- required_without/
// excluded_with together express that XOR.
type CustomCheck struct {
	Path           string   `yaml:"path,omitempty" validate:"required_without=Command,excluded_with=Command"`
	Command        string   `yaml:"command,omitempty" validate:"required_without=Path,excluded_with=Path"`
	Args           []string `yaml:"args,omitempty"`
	TimeoutSeconds int      `yaml:"timeoutSeconds,omitempty" validate:"gte=0"`
	When           []When   `yaml:"when,omitempty" validate:"dive"`
}

const DefaultCustomCheckTimeoutSeconds = 10

type When struct {
	Key    string `yaml:"key" validate:"eq=exit"`
	Value  int    `yaml:"value"`
	Action Action `yaml:"action" validate:"oneof=ALLOW DENY"`
	// Confirm mirrors Options.Confirm: when this entry matches and Action
	// is ALLOW, the result is gated behind human confirmation. Invalid
	// (lint error) when Action is DENY, for the same reason as above.
	Confirm bool `yaml:"confirm,omitempty"`
}

// DefaultAction is applied when no policy rule matches a given command.
const DefaultAction = ActionConfirm
