package adigger

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Playbook represents the entire playbook file.
type Playbook struct {
	Plays        []*Play      `yaml:",inline"`
	Dependencies []Dependency `yaml:"-"` // Populated during analysis
}

// Play represents a single play within a playbook.
type Play struct {
	ID                int       `yaml:"-"` // Assigned during parsing
	Name              string    `yaml:"name"`
	Hosts             string    `yaml:"hosts"`
	Become            bool      `yaml:"become"`
	GatherFacts       bool      `yaml:"gather_facts"`
	Strategy          string    `yaml:"strategy"`
	Serial            any       `yaml:"serial"`
	MaxFailPercentage int       `yaml:"max_fail_percentage"`
	AnyErrorsFatal    bool      `yaml:"any_errors_fatal"`
	ForceHandlers     bool      `yaml:"force_handlers"`
	Order             string    `yaml:"order"`
	Connection        string    `yaml:"connection"`
	PreTasks          []*Task   `yaml:"pre_tasks"`
	PostTasks         []*Task   `yaml:"post_tasks"`
	Tasks             []*Task   `yaml:"tasks"`
	Roles             []*Role   `yaml:"roles"`
	Handlers          []*Task   `yaml:"handlers"`
	Vars              yaml.Node `yaml:"vars"`
	VarsFiles         []string  `yaml:"vars_files"`
}

// Task represents a single Ansible task.
type Task struct {
	ID           string     `yaml:"-"` // Assigned during parsing
	PlayID       int        `yaml:"-"` // Assigned during parsing
	Name         string     `yaml:"name"`
	Include      string     `yaml:"include"`
	IncludeTasks string     `yaml:"include_tasks"`
	ImportTasks  string		`yaml:"import_tasks"`
	When         string     `yaml:"when"`
	Register     string     `yaml:"register"`
	Become       bool       `yaml:"become"`
	Notify       Notify     `yaml:"notify"`
	Tags         Tags       `yaml:"tags"`
	DelegateTo   string     `yaml:"delegate_to"`
	IgnoreErrors bool       `yaml:"ignore_errors"`
	RunOnce      bool       `yaml:"run_once"`
	Async        int        `yaml:"async"`
	Poll         int        `yaml:"poll"`
	NoLog        bool       `yaml:"no_log"`
	Loop         any        `yaml:"loop"`
	WithItems    any        `yaml:"with_items"`
	WithDict     any        `yaml:"with_dict"`
	FailedWhen   string     `yaml:"failed_when"`
	ChangedWhen  string     `yaml:"changed_when"`
	CheckMode    bool       `yaml:"check_mode"`
	Environment  yaml.Node  `yaml:"environment"`
	Retries      int        `yaml:"retries"`
	Delay        int        `yaml:"delay"`
	Until        string     `yaml:"until"`
	Throttle     int        `yaml:"throttle"`
	Diff         bool       `yaml:"diff"`
	Block        []*Task    `yaml:"block"`
	Rescue       []*Task    `yaml:"rescue"`
	Always       []*Task    `yaml:"always"`
	IsCritical   bool       `yaml:"-"` // Set during analysis
	IsHandler    bool       `yaml:"-"` // Set during parsing
	RawNode      *yaml.Node `yaml:"-"` // The original node for tooltips/analysis

	// The yaml.Node below will capture all other keys, which will contain
	// the module name (e.g., "apt") and its arguments.
	Action yaml.Node `yaml:",inline"`
}

// GetModule extracts the module name from the Action node.
// The Action node contains all keys that are not standard task keywords,
// so the first key is the module name.
func (t *Task) GetModule() string {
	if t.Action.Kind == yaml.MappingNode && len(t.Action.Content) > 0 {
		// The module is the first key in the inline map.
		return t.Action.Content[0].Value
	}
	return ""
}

// Role represents a role inclusion.
type Role struct {
	ID              int    `yaml:"-"` // Assigned during parsing
	PlayID          int    `yaml:"-"` // Assigned during parsing
	Name            string `yaml:"role"`
	When            string `yaml:"when"`
	Tags            Tags   `yaml:"tags"`
	Defaults        any    `yaml:"defaults"`
	Meta            any    `yaml:"meta"`
	AllowDuplicates bool   `yaml:"allow_duplicates"`
	Public          bool   `yaml:"public"`

	// Enriched data
	Tasks     []*Task  `yaml:"-"`
	Handlers  []*Task  `yaml:"-"`
	Vars      []string `yaml:"-"`
	Files     []string `yaml:"-"`
	Templates []string `yaml:"-"`
}

// DependencyType defines the type of data flow dependency.
type DependencyType int

const (
	DepTypeRegister DependencyType = iota // From a 'register' variable
	DepTypeFact                           // From 'ansible_facts'
)

// Dependency represents a data flow link between two tasks.
type Dependency struct {
	From  *Task          `yaml:"-"`
	To    *Task          `yaml:"-"`
	Type  DependencyType `yaml:"-"`
	Label string         `yaml:"-"`
}

// UnmarshalYAML allows the Role type to be unmarshalled from either a simple
// string (the role name) or a complex map object.
func (r *Role) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		r.Name = node.Value
		return nil
	}

	// If it's a map, we use an alias to avoid recursion.
	type RoleAlias Role
	var alias RoleAlias
	if err := node.Decode(&alias); err != nil {
		return err
	}
	*r = Role(alias)
	return nil
}

// Tags can be a single string or a list of strings.
type Tags []string

func (t *Tags) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		*t = []string{node.Value}
		return nil
	}
	if node.Kind == yaml.SequenceNode {
		var tags []string
		if err := node.Decode(&tags); err != nil {
			return fmt.Errorf("failed to decode tags sequence: %w", err)
		}
		*t = tags
		return nil
	}
	return fmt.Errorf("tags field must be a string or a sequence of strings")
}

// Notify can be a single string or a list of strings.
type Notify []string

func (n *Notify) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind == yaml.ScalarNode {
		*n = []string{node.Value}
		return nil
	}
	if node.Kind == yaml.SequenceNode {
		var notifies []string
		if err := node.Decode(&notifies); err != nil {
			return fmt.Errorf("failed to decode notify sequence: %w", err)
		}
		*n = notifies
		return nil
	}
	return fmt.Errorf("notify field must be a string or a sequence of strings")
}
