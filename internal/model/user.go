package model

import (
	"encoding/json"
	"fmt"
)

type UserRequest struct {
	FirstName string          `json:"firstName"`
	LastName  string          `json:"lastName"`
	Email     string          `json:"email"`
	Password  string          `json:"password"`
	Anonymous bool            `json:"anonymous"`
	Settings  *SettingRequest `json:"settings"`
}

func (ur *UserRequest) Validate() error {
	if ur.FirstName == "" {
		return fmt.Errorf("first name is required")
	}
	if ur.LastName == "" {
		return fmt.Errorf("last name is required")
	}
	if ur.Email == "" {
		return fmt.Errorf("email is required")
	}
	if ur.Password == "" {
		return fmt.Errorf("password is required")
	}
	return nil
}

func (ur *UserRequest) UnmarshalJSON(b []byte) error {
	err := json.Unmarshal(b, &ur)
	if err != nil {
		return err
	}
	err = ur.Validate()
	if err != nil {
		return err
	}
	return nil
}

func (ur *UserRequest) MarshalJSON() ([]byte, error) {
	if ur == nil {
		return nil, fmt.Errorf("UserRequest is nil")
	}
	err := ur.Validate()
	if err != nil {
		return nil, err
	}
	userObj := *ur
	if userObj.Email == "" {
		return nil, fmt.Errorf("invalid user email")
	}
	if userObj.Password == "" {
		return nil, fmt.Errorf("invalid user password")
	}

	type Alias UserRequest

	auxUser := (Alias)(userObj)
	return json.Marshal(auxUser)
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
	UserId   string         `json:"userId"`
	Settings SettingRequest `json:"settings"`
}
