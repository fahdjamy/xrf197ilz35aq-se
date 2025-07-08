package model

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"time"
	"xrf197ilz35aq/internal"
)

type UserRequest struct {
	LastName  string          `json:"lastName"`
	Settings  *SettingRequest `json:"settings"`
	FirstName string          `json:"firstName"`
	Anonymous bool            `json:"anonymous"`
	Email     string          `json:"email" validate:"required"`
	Password  string          `json:"password" validate:"required,min=8"`
}

func (ur *UserRequest) Validate() error {
	_, err := mail.ParseAddress(ur.Email)
	if err != nil {
		return fmt.Errorf("invalid email address, Err :: '%s'", err.Error())
	}
	lastNameLen := len(ur.LastName)
	if lastNameLen != 0 && lastNameLen < 3 {
		return fmt.Errorf("if last name is specified, it should be at least 3 characters long")
	}
	firstNameLen := len(ur.FirstName)
	if firstNameLen != 0 && firstNameLen < 3 {
		return fmt.Errorf("if first name is specified, it should be at least 3 characters long")
	}
	if (ur.FirstName != "" && ur.LastName == "") || (ur.FirstName == "" && ur.LastName != "") {
		return fmt.Errorf("if FirstName or LastName is specified, then both fields should not be empty")
	}
	if err := ValidatePassword(ur.Password); err != nil {
		return err
	}
	return nil
}

func ValidatePassword(password string) error {
	// Minimum length of 8 characters
	// At least 1 uppercase letter and 1 lowercase letter
	// At least one digit
	if len(password) < 8 {
		return fmt.Errorf("password must be at least 8 characters long")
	}

	if !regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one special character")
	}

	if !regexp.MustCompile(`[A-Z]`).MatchString(password) || !regexp.MustCompile(`[a-z]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one lowercase and an uppercase letter")
	}

	if !regexp.MustCompile(`[0-9]`).MatchString(password) {
		return fmt.Errorf("password must contain at least one digit")
	}

	return nil
}

func (ur *UserRequest) UnmarshalJSON(bytes []byte) error {
	clientErr := &internal.ExternalError{
		Message: "Invalid JSON request",
		Code:    http.StatusBadRequest,
	}
	type Alias UserRequest
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(ur),
	}
	if err := json.Unmarshal(bytes, &aux); err != nil {
		clientErr.Message = "Failed to unmarshal JSON"
		return clientErr
	}

	if aux.Email == "" {
		clientErr.Message = "Invalid email address"
		return clientErr
	}
	if aux.Password == "" {
		clientErr.Message = "Invalid password"
		return clientErr
	}

	return nil
}

func (ur *UserRequest) String() string {
	return fmt.Sprintf("{firstName:%s, lastName%s, anonymous=%t}", ur.FirstName, ur.LastName, ur.Anonymous)
}

type SettingRequest struct {
	RotateKey     bool   `json:"rotateKey"`
	RotateAfter   int    `json:"rotateAfter"`
	EncryptionKey string `json:"encryptionKey"`
}

type UserResponse struct {
	UserId    string          `json:"userId"`
	FirstName string          `json:"firstName,omitempty"`
	LastName  string          `json:"lastName,omitempty"`
	Anonymous bool            `json:"anonymous"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Settings  SettingResponse `json:"settings,omitempty"`
}

type SettingResponse struct {
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
	EncryptionKey string    `json:"encryptionKey"`
	RotateKey     bool      `json:"rotateKey"`
}

func (s *SettingResponse) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		CreatedAt time.Time `json:"createdAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		RotateKey bool      `json:"rotateKey"`
	}{
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
		RotateKey: s.RotateKey,
	})
}
