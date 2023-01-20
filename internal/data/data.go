package data

import "time"

const (
	ModuleName        = "github"
	Organization      = "org"
	Repository        = "repo"
	UserOwned         = "User"
	OrganizationOwned = "Organization"
	Push              = "push"
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
	RequestId  string `json:"request_id"`
	UserId     int64  `json:"user_id"`
	Action     string `json:"action"`
	Link       string `json:"link"`
	Username   string `json:"username"`
	Permission string `json:"permission"`
	Type       string `json:"type_to"`
}
