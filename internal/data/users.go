package data

type Users interface {
	New() Users

	Upsert(user User) error

	GetById(id int64) (*User, error)
	GetByUsername(username string) (*User, error)
	GetByGithubId(githubId int64) (*User, error)
}

type User struct {
	Id       *int64 `db:"id" structs:"id,omitempty"`
	Username string `db:"username" structs:"username"`
	GithubId int64  `db:"github_id" structs:"github_id"`
}
