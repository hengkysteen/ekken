package agents

import (
	"ekken/internal/features/assistant/skills"
	_ "embed"
)

//go:embed prompts/workflow.md
var workflowPrompt string

func init() {
	Register("workflow", func() Agent {
		return Agent{
			Name:         "Workflow Engineer",
			Description:  "Specialized agent for designing and managing automated workflows.",
			SystemPrompt: workflowPrompt,
			Skills: []skills.SkillInterface{
				skills.Registry["nodes"],
				skills.Registry["nodes_actions"],
				skills.Registry["create_workflow"],
				skills.Registry["save_workflow"],
			},
		}
	})
}
