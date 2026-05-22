package profile

import (
	"fmt"
	"strings"
	"time"
)

type Servicer interface {
	Get() (Profile, error)
	Update(req UpdateRequest) (Profile, error)
	VerifyPIN(pin string) (bool, error)
	ResetPIN(req ResetPINRequest) (bool, error)
}

type Database interface {
	GetProfile() (ProfileItem, error)
	SaveProfile(item ProfileItem) error
}

type Service struct {
	db Database
}

func New(database Database) *Service {
	return &Service{db: database}
}

func (s *Service) Get() (Profile, error) {
	item, err := s.db.GetProfile()
	if err != nil {
		return Profile{}, err
	}
	return toProfile(item), nil
}

func (s *Service) Update(req UpdateRequest) (Profile, error) {
	item, err := s.db.GetProfile()
	if err != nil {
		return Profile{}, err
	}

	now := time.Now().Format(time.RFC3339)
	item.Name = strings.TrimSpace(req.Name)
	if item.Name == "" {
		return Profile{}, fmt.Errorf("name is required")
	}
	item.UpdatedAt = now

	if !req.PINEnabled {
		item.PINEnabled = false
		if err := s.db.SaveProfile(item); err != nil {
			return Profile{}, err
		}
		return toProfile(item), nil
	}

	item.PINEnabled = true
	pin := strings.TrimSpace(req.PIN)
	if pin != "" {
		hashed, err := hashPIN(pin)
		if err != nil {
			return Profile{}, err
		}
		item.PINHash = hashed
		item.PINUpdatedAt = now
	} else if item.PINHash == "" {
		return Profile{}, fmt.Errorf("pin is required to enable app lock")
	}

	question := strings.TrimSpace(req.SecurityQuestion)
	answer := strings.TrimSpace(req.SecurityAnswer)
	if question != "" && answer != "" {
		hashedAnswer, err := hashPIN(answer)
		if err != nil {
			return Profile{}, err
		}
		item.SecurityQuestion = question
		item.SecurityAnswerHash = hashedAnswer
	} else if item.SecurityQuestion == "" {
		return Profile{}, fmt.Errorf("security question and answer are required to enable app lock")
	}

	if err := s.db.SaveProfile(item); err != nil {
		return Profile{}, err
	}
	return toProfile(item), nil
}

func (s *Service) VerifyPIN(pin string) (bool, error) {
	item, err := s.db.GetProfile()
	if err != nil {
		return false, err
	}
	if !item.PINEnabled || item.PINHash == "" {
		return true, nil
	}
	return verifyPINHash(item.PINHash, strings.TrimSpace(pin)), nil
}

func (s *Service) ResetPIN(req ResetPINRequest) (bool, error) {
	item, err := s.db.GetProfile()
	if err != nil {
		return false, err
	}
	if !item.PINEnabled || item.SecurityAnswerHash == "" {
		return false, fmt.Errorf("app lock is not active or security question is not set")
	}

	if !verifyPINHash(item.SecurityAnswerHash, strings.TrimSpace(req.Answer)) {
		return false, nil
	}

	newPin := strings.TrimSpace(req.NewPIN)
	if newPin == "" {
		return false, fmt.Errorf("new pin is required")
	}

	hashed, err := hashPIN(newPin)
	if err != nil {
		return false, err
	}
	item.PINHash = hashed
	item.PINUpdatedAt = time.Now().Format(time.RFC3339)

	if err := s.db.SaveProfile(item); err != nil {
		return false, err
	}
	return true, nil
}

func toProfile(item ProfileItem) Profile {
	return Profile{
		Name:             item.Name,
		PINEnabled:       item.PINEnabled,
		UpdatedAt:        item.UpdatedAt,
		PINUpdatedAt:     item.PINUpdatedAt,
		SecurityQuestion: item.SecurityQuestion,
	}
}
