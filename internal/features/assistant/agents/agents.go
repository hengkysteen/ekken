package agents

import (
	"ekken/internal/features/assistant/skills"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed prompts/platform.md
var platform string

type Agent struct {
	Name         string                  `json:"name"`
	Description  string                  `json:"description"`
	SystemPrompt string                  `json:"-"`
	Skills       []skills.SkillInterface `json:"skills"`
}

var Registry = make(map[string]func() Agent)

func Register(name string, factory func() Agent) {
	Registry[name] = factory
}

func GetAgent(name string) (Agent, error) {
	if factory, ok := Registry[name]; ok {
		return factory(), nil
	}
	return Agent{}, fmt.Errorf("unknown agent: %s", name)
}

func ListAgents() []map[string]string {
	list := make([]map[string]string, 0, len(Registry))
	for name, factory := range Registry {
		a := factory()
		list = append(list, map[string]string{"name": name, "description": a.Description})
	}
	return list
}

func (a Agent) BuildSystemPrompt() string {
	prompt := platform + "\n\n" + a.SystemPrompt
	if len(a.Skills) == 0 {
		return prompt
	}
	var sb strings.Builder
	sb.WriteString(prompt)
	sb.WriteString("\n\n# AVAILABLE SKILLS\n")
	for _, s := range a.Skills {
		fmt.Fprintf(&sb, "- %s: %s\n", s.GetID(), s.GetDescription())
	}
	return sb.String()
}
