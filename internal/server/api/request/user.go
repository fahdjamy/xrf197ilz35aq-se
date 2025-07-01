package request

type UserRequest struct {
	FirstName string          `json:"firstName"`
	LastName  string          `json:"lastName"`
	Email     string          `json:"email"`
	Password  string          `json:"password"`
	Anonymous bool            `json:"anonymous"`
	Settings  *SettingRequest `json:"settings"`
}

type SettingRequest struct {
	RotateKey     bool   `json:"rotateKey"`
	RotateAfter   int    `json:"rotateAfter"`
	EncryptionKey string `json:"encryptionKey"`
}
