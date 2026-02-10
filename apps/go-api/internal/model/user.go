package model

// User represents a user in the system
type User struct {
	ID        int64   `json:"id"`
	Email     string  `json:"email"`
	Name      *string `json:"name,omitempty"`
	CreatedAt string  `json:"created_at"`
}

// AuthUser represents authenticated user info stored in context
type AuthUser struct {
	ID    int64
	Email string
}

// Request/Response types

type CreateUserRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OAuthCallbackRequest struct {
	Provider          string `json:"provider"`
	ProviderAccountID string `json:"provider_account_id"`
	Email             string `json:"email"`
	Name              string `json:"name"`
}

type OAuthCallbackResponse struct {
	User      User `json:"user"`
	IsNewUser bool `json:"is_new_user"`
}

type TokenResponse struct {
	Token string `json:"token"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
