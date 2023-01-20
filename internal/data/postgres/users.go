package postgres

import (
	"database/sql"
	"fmt"
	"github.com/fatih/structs"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"

	sq "github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const usersTableName = "users"

type GitlabUsersQ struct {
	db  *pgdb.DB
	sql sq.SelectBuilder
}

var selectedUsersTable = sq.Select("*").From(usersTableName)

func NewUsersQ(db *pgdb.DB) data.Users {
	return &GitlabUsersQ{
		db:  db.Clone(),
		sql: selectedUsersTable,
	}
}

func (q *GitlabUsersQ) New() data.Users {
	return NewUsersQ(q.db)
}

func (q *GitlabUsersQ) Upsert(user data.User) error {
	clauses := structs.Map(user)

	stmt := "ON CONFLICT (github_id) DO NOTHING"
	if user.Id != nil {
		stmt = fmt.Sprintf("ON CONFLICT (github_id) DO UPDATE SET id = %d", *user.Id)
	}
	query := sq.Insert(usersTableName).SetMap(clauses).Suffix(stmt)

	return q.db.Exec(query)
}

func (q *GitlabUsersQ) GetById(id int64) (*data.User, error) {
	query := q.sql.Where(sq.Eq{"id": id})

	var result data.User
	err := q.db.Get(&result, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q *GitlabUsersQ) GetByUsername(username string) (*data.User, error) {
	query := q.sql.Where(sq.Eq{"username": username})

	var result data.User
	err := q.db.Get(&result, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q *GitlabUsersQ) GetByGithubId(githubId int64) (*data.User, error) {
	query := q.sql.Where(sq.Eq{"github_id": githubId})

	var result data.User
	err := q.db.Get(&result, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}
