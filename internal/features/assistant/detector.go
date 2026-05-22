package assistant

import (
	"fmt"
	"regexp"
	"strings"

	"ekken/internal/features/assistant/agents"
	"ekken/internal/features/assistant/skills"

	"github.com/goccy/go-yaml"
)

// SkillCallDetector scans the assistant's clean content for a valid ~ekken skill ... ekken~ block.
func (o *Orchestrator) SkillCallDetector(content string) (bool, skills.SkillCall, int, string) {
	re := regexp.MustCompile(`(?is)~\s*ekken\s+skill\s+([a-zA-Z0-9_]+)\s*(.*?)\s*ekken\s*~`)
	match := re.FindStringSubmatch(content)
	if len(match) == 3 {
		skillName := match[1]
		argsStr := match[2]

		var args map[string]interface{}
		argsStr = strings.TrimSpace(argsStr)

		if argsStr == "" || argsStr == "()" || argsStr == "{}" {
			args = make(map[string]interface{})
		} else {
			if err := yaml.Unmarshal([]byte(argsStr), &args); err != nil {
				return false, skills.SkillCall{}, -1, ""
			}
		}

		loc := re.FindStringIndex(content)
		return true, skills.SkillCall{
			Skill: skillName,
			Args:  args,
		}, loc[0], match[0]
	}
	return false, skills.SkillCall{}, -1, ""
}

func (o *Orchestrator) SkillCallFallback(content string, isSkillCall bool, parsedCall skills.SkillCall, activeAgent *agents.Agent) string {
	hasNewFormatTag := regexp.MustCompile(`(?is)~\s*ekken`).MatchString(content)
	hasLegacyFormat := regexp.MustCompile(`(?i)<function=([a-zA-Z0-9_]+)>`).MatchString(content)

	if (hasNewFormatTag || hasLegacyFormat) && !isSkillCall {
		if skillName, parseErr := skillCallArgsError(content); parseErr != nil {
			return fmt.Sprintf("%sError: Invalid skill arguments for '%s': %v. Output exactly one skill call and fix only the arguments inside it.", TagSkillResult, skillName, parseErr)
		}
		return TagSkillResult + "Error: Invalid format! You must use the following format: ~ekken skill SKILL_NAME args ekken~"
	}
	if isSkillCall {
		isAuthorized := true
		if activeAgent != nil {
			found := false
			for _, s := range activeAgent.Skills {
				if s.GetID() == parsedCall.Skill {
					found = true
					break
				}
			}
			if !found {
				isAuthorized = false
			}
		}
		if _, ok := skills.Registry[parsedCall.Skill]; !ok {
			isAuthorized = false
		}
		if !isAuthorized {
			return fmt.Sprintf("%sError: Skill '%s' not found or not allowed!", TagSkillResult, parsedCall.Skill)
		}
	}
	return ""
}

func skillCallArgsError(content string) (string, error) {
	re := regexp.MustCompile(`(?is)~\s*ekken\s+skill\s+([a-zA-Z0-9_]+)\s*(.*?)\s*ekken\s*~`)
	match := re.FindStringSubmatch(content)
	if len(match) != 3 {
		return "", nil
	}

	argsStr := strings.TrimSpace(match[2])
	if argsStr == "" || argsStr == "()" || argsStr == "{}" {
		return "", nil
	}

	var args map[string]interface{}
	if err := yaml.Unmarshal([]byte(argsStr), &args); err != nil {
		return match[1], err
	}
	return "", nil
}
