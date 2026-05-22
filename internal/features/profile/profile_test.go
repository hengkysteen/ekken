package profile

import "testing"

type fakeProfileDB struct {
	item ProfileItem
}

func (db *fakeProfileDB) GetProfile() (ProfileItem, error) {
	return db.item, nil
}

func (db *fakeProfileDB) SaveProfile(item ProfileItem) error {
	db.item = item
	return nil
}

func TestProfileDefaultIsUnlocked(t *testing.T) {
	service := New(&fakeProfileDB{})

	prof, err := service.Get()
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if prof.PINEnabled {
		t.Fatal("PINEnabled = true, want false")
	}

	valid, err := service.VerifyPIN("anything")
	if err != nil {
		t.Fatalf("VerifyPIN() error = %v", err)
	}
	if !valid {
		t.Fatal("VerifyPIN() = false for disabled PIN, want true")
	}
}

func TestEnablePINRequiresPIN(t *testing.T) {
	service := New(&fakeProfileDB{})

	_, err := service.Update(UpdateRequest{
		Name:             "Steen",
		PINEnabled:       true,
		SecurityQuestion: "Dog name?",
		SecurityAnswer:   "Rex",
	})
	if err == nil {
		t.Fatal("Update() error = nil, want validation error for missing PIN")
	}
}

func TestEnablePINRequiresSecurityQA(t *testing.T) {
	service := New(&fakeProfileDB{})

	_, err := service.Update(UpdateRequest{Name: "Steen", PINEnabled: true, PIN: "1234"})
	if err == nil {
		t.Fatal("Update() error = nil, want validation error for missing security question/answer")
	}
}

func TestUpdateRequiresName(t *testing.T) {
	service := New(&fakeProfileDB{})

	_, err := service.Update(UpdateRequest{Name: "   "})
	if err == nil {
		t.Fatal("Update() error = nil, want validation error")
	}
}

func TestVerifyPIN(t *testing.T) {
	service := New(&fakeProfileDB{})

	_, err := service.Update(UpdateRequest{
		Name:             "Steen",
		PINEnabled:       true,
		PIN:              "1234",
		SecurityQuestion: "Dog name?",
		SecurityAnswer:   "Rex",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	valid, err := service.VerifyPIN("1234")
	if err != nil {
		t.Fatalf("VerifyPIN(correct) error = %v", err)
	}
	if !valid {
		t.Fatal("VerifyPIN(correct) = false, want true")
	}

	valid, err = service.VerifyPIN("0000")
	if err != nil {
		t.Fatalf("VerifyPIN(wrong) error = %v", err)
	}
	if valid {
		t.Fatal("VerifyPIN(wrong) = true, want false")
	}
}

func TestDisableKeepsPINAndSecurityData(t *testing.T) {
	db := &fakeProfileDB{}
	service := New(db)

	if _, err := service.Update(UpdateRequest{
		Name:             "Steen",
		PINEnabled:       true,
		PIN:              "1234",
		SecurityQuestion: "Dog name?",
		SecurityAnswer:   "Rex",
	}); err != nil {
		t.Fatalf("enable pin error = %v", err)
	}
	storedHash := db.item.PINHash
	storedUpdatedAt := db.item.PINUpdatedAt
	storedQuestion := db.item.SecurityQuestion
	storedAnswerHash := db.item.SecurityAnswerHash

	if _, err := service.Update(UpdateRequest{Name: "Steen", PINEnabled: false}); err != nil {
		t.Fatalf("disable pin error = %v", err)
	}

	if db.item.PINEnabled {
		t.Fatal("PINEnabled = true, want false")
	}
	if db.item.PINHash != storedHash {
		t.Fatal("PINHash changed after disable")
	}
	if db.item.PINUpdatedAt != storedUpdatedAt {
		t.Fatal("PINUpdatedAt changed after disable")
	}
	if db.item.SecurityQuestion != storedQuestion {
		t.Fatal("SecurityQuestion changed after disable")
	}
	if db.item.SecurityAnswerHash != storedAnswerHash {
		t.Fatal("SecurityAnswerHash changed after disable")
	}
}

func TestResetPIN(t *testing.T) {
	service := New(&fakeProfileDB{})

	// Enable PIN with security Q&A
	_, err := service.Update(UpdateRequest{
		Name:             "Steen",
		PINEnabled:       true,
		PIN:              "1234",
		SecurityQuestion: "Dog name?",
		SecurityAnswer:   "Rex",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	// Try resetting with wrong answer
	success, err := service.ResetPIN(ResetPINRequest{
		Answer: "Wrong",
		NewPIN: "4321",
	})
	if err != nil {
		t.Fatalf("ResetPIN() error = %v", err)
	}
	if success {
		t.Fatal("ResetPIN() success = true, want false for incorrect answer")
	}

	// Reset with correct answer
	success, err = service.ResetPIN(ResetPINRequest{
		Answer: "Rex",
		NewPIN: "4321",
	})
	if err != nil {
		t.Fatalf("ResetPIN() error = %v", err)
	}
	if !success {
		t.Fatal("ResetPIN() success = false, want true")
	}

	// Verify new PIN works
	valid, err := service.VerifyPIN("4321")
	if err != nil {
		t.Fatalf("VerifyPIN() error = %v", err)
	}
	if !valid {
		t.Fatal("VerifyPIN() = false for new PIN, want true")
	}
}
