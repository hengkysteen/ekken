package skills

type Parameter struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Items       *Schema `json:"items,omitempty"`
}

type SkillCall struct {
	Skill string                 `json:"skill"`
	Args  map[string]interface{} `json:"args"`
}

type Schema struct {
	Type       string               `json:"type"`
	Properties map[string]Parameter `json:"properties"`
	Required   []string             `json:"required"`
}

type SkillInterface interface {
	GetID() string
	GetName() string
	GetDescription() string
	Execute(args map[string]interface{}) (string, error)
}

var Registry = make(map[string]SkillInterface)

func Register(s SkillInterface) {
	Registry[s.GetID()] = s
}
