package profile

type Profile struct {
	Name             string `json:"name"`
	PINEnabled       bool   `json:"pin_enabled"`
	UpdatedAt        string `json:"updated_at,omitempty"`
	PINUpdatedAt     string `json:"pin_updated_at,omitempty"`
	SecurityQuestion string `json:"security_question,omitempty"`
}

type UpdateRequest struct {
	Name             string `json:"name"`
	PINEnabled       bool   `json:"pin_enabled"`
	PIN              string `json:"pin,omitempty"`
	SecurityQuestion string `json:"security_question,omitempty"`
	SecurityAnswer   string `json:"security_answer,omitempty"`
}

type VerifyPINRequest struct {
	PIN string `json:"pin"`
}

type VerifyPINResponse struct {
	Valid bool `json:"valid"`
}

type ResetPINRequest struct {
	Answer string `json:"answer"`
	NewPIN string `json:"new_pin"`
}

type ProfileItem struct {
	Name               string
	PINEnabled         bool
	PINHash            string
	UpdatedAt          string
	PINUpdatedAt       string
	SecurityQuestion   string
	SecurityAnswerHash string
}
