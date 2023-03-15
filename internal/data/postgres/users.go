package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/fatih/structs"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/logan/v3/errors"

	sq "github.com/Masterminds/squirrel"
	"gitlab.com/distributed_lab/kit/pgdb"
)

const usersTableName = "users"

type UsersQ struct {
	db  *pgdb.DB
	sql sq.SelectBuilder
}

var selectedUsersTable = sq.Select("*").From(usersTableName)

func NewUsersQ(db *pgdb.DB) data.Users {
	return &UsersQ{
		db:  db.Clone(),
		sql: selectedUsersTable,
	}
}

func (q *UsersQ) New() data.Users {
	return NewUsersQ(q.db)
}

func (q *UsersQ) Upsert(user data.User) error {
	clauses := structs.Map(user)

	updateStmt := "NOTHING"
	var args []interface{}

	if user.Id != nil {
		updateQuery := sq.Update(" ").Set("id", *user.Id)
		updateStmt, args = updateQuery.MustSql()
	}

	query := sq.Insert(usersTableName).SetMap(clauses).Suffix("ON CONFLICT (github_id) DO "+updateStmt, args...)

	return q.db.Exec(query)
}

func (q *UsersQ) GetById(id int64) (*data.User, error) {
	query := q.sql.Where(sq.Eq{"id": id})

	var result data.User
	err := q.db.Get(&result, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q *UsersQ) GetByUsername(username string) (*data.User, error) {
	query := q.sql.Where(sq.Eq{"username": username})

	var result data.User
	err := q.db.Get(&result, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q *UsersQ) GetByGithubId(githubId int64) (*data.User, error) {
	query := q.sql.Where(sq.Eq{"github_id": githubId})

	var result data.User
	err := q.db.Get(&result, query)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q *UsersQ) Delete(githubId int64) error {
	var deleted []data.Response

	query := sq.Delete(subsTableName).
		Where(sq.Eq{
			"github_id": githubId,
		}).
		Suffix("RETURNING *")

	err := q.db.Select(&deleted, query)
	if err != nil {
		return err
	}
	if len(deleted) == 0 {
		return errors.Errorf("no rows with `%d` github id", githubId)
	}

	return nil
}

func (q *UsersQ) FilterById(id *int64) data.Users {
	stmt := sq.Eq{usersTableName + ".id": id}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *UsersQ) Get() (*data.User, error) {
	var result data.User

	err := q.db.Get(&result, q.sql)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q *UsersQ) Select() ([]data.User, error) {
	var result []data.User

	err := q.db.Select(&result, q.sql)

	return result, err
}

func (q *UsersQ) Page(pageParams pgdb.OffsetPageParams) data.Users {
	q.sql = pageParams.ApplyTo(q.sql, "username")

	return q
}

func (q *UsersQ) SearchBy(search string) data.Users {
	search = strings.Replace(search, " ", "%", -1)
	search = fmt.Sprint("%", search, "%")

	q.sql = q.sql.Where(sq.ILike{"username": search})
	return q
}

func (q *UsersQ) Count() data.Users {
	q.sql = sq.Select("COUNT (*)").From(usersTableName)

	return q
}

func (q *UsersQ) GetTotalCount() (int64, error) {
	var count int64
	err := q.db.Get(&count, q.sql)

	return count, err
}
