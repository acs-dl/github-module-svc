package data

import (
	"io"
	"net/http"
	"time"
)

const (
	ModuleName             = "github"
	Organization           = "org"
	Repository             = "repo"
	UserOwned              = "User"
	OrganizationOwned      = "Organization"
	Push                   = "push"
	AcceptHeader           = "application/vnd.Github+json"
	GithubApiVersionHeader = "2022-11-28"
)

type ModuleRequest struct {
	ID            string    `db:"id" structs:"id"`
	UserID        int64     `db:"user_id" structs:"user_id"`
	Module        string    `db:"module" structs:"module"`
	Payload       string    `db:"payload" structs:"payload"`
	CreatedAt     time.Time `db:"created_at" structs:"created_at"`
	RequestStatus string    `db:"request_status" structs:"request_status"`
	Error         string    `db:"error" structs:"error"`
}

type ModulePayload struct {
	RequestId   string   `json:"request_id"`
	UserId      string   `json:"user_id"`
	Action      string   `json:"action"`
	Link        string   `json:"link"`
	Links       []string `json:"links"`
	Username    string   `json:"username"`
	AccessLevel string   `json:"access_level"`
	Type        string   `json:"type"`
}

type UnverifiedPayload struct {
	Action string           `json:"action"`
	Users  []UnverifiedUser `json:"users"`
}

var Roles = map[string]string{
	"":         "No access",
	"read":     "Read",
	"triage":   "Triage",
	"write":    "Write",
	"maintain": "Maintain",
	"admin":    "Admin",
	"member":   "Member",
}

type RequestParams struct {
	Method  string
	Link    string
	Body    []byte
	Query   map[string]string
	Header  map[string]string
	Timeout time.Duration
}

type ResponseParams struct {
	Body       io.ReadCloser
	Header     http.Header
	StatusCode int
}
