package request

type Credentials struct {
	Email    string `json:"email" binding:"required,email,max=254" example:"operator@example.com"`
	Password string `json:"password" binding:"required,min=8,max=128" example:"UptarisDemo!2026"`
}
