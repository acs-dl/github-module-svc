package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	sq "github.com/Masterminds/squirrel"
	"github.com/fatih/structs"
	"gitlab.com/distributed_lab/acs/github-module/internal/data"
	"gitlab.com/distributed_lab/kit/pgdb"
	"gitlab.com/distributed_lab/logan/v3/errors"
)

const subsTableName = "subs"

type SubsQ struct {
	db  *pgdb.DB
	sql sq.SelectBuilder
}

var (
	subsColumns = []string{
		subsTableName + ".id",
		subsTableName + ".link as subs_link",
		subsTableName + ".path",
		subsTableName + ".type as subs_type",
		subsTableName + ".parent_id",
	}
	selectedSubsTable = sq.Select(subsColumns...).From(subsTableName)
)

func NewSubsQ(db *pgdb.DB) data.Subs {
	return &SubsQ{
		db:  db.Clone(),
		sql: selectedSubsTable,
	}
}

func (q *SubsQ) New() data.Subs {
	return NewSubsQ(q.db)
}

func (q *SubsQ) Insert(sub data.Sub) error {
	clauses := structs.Map(sub)

	query := sq.Insert(subsTableName).SetMap(clauses)

	return q.db.Exec(query)
}

func (q *SubsQ) Upsert(sub data.Sub) error {
	query := sq.Insert(subsTableName).SetMap(structs.Map(sub)).
		Suffix(fmt.Sprintf("ON CONFLICT DO NOTHING"))

	return q.db.Exec(query)
}

func (q *SubsQ) Select() ([]data.Sub, error) {
	var result []data.Sub

	err := q.db.Select(&result, q.sql)

	return result, err
}

func (q *SubsQ) Get() (*data.Sub, error) {
	var result data.Sub

	err := q.db.Get(&result, q.sql)
	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &result, err
}

func (q *SubsQ) Delete(subId int64, typeTo, link string) error {
	var deleted []data.Sub

	query := sq.Delete(subsTableName).
		Where(sq.Eq{
			"id":   subId,
			"type": typeTo,
			"link": link,
		}).
		Suffix("RETURNING *")

	err := q.db.Select(&deleted, query)
	if err != nil {
		return err
	}
	if len(deleted) == 0 {
		return errors.Errorf("no rows with `%s` link", link)
	}

	return nil
}

func (q *SubsQ) FilterByParentLinks(parentLinks ...string) data.Subs {
	stmt := sq.Eq{permissionsTableName + ".parent_link": parentLinks}
	if len(parentLinks) == 0 {
		stmt = sq.Eq{permissionsTableName + ".parent_link": nil}
	}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *SubsQ) FilterByLinks(links ...string) data.Subs {
	stmt := sq.Eq{subsTableName + ".link": links}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *SubsQ) FilterByIds(ids ...int64) data.Subs {
	stmt := sq.Eq{subsTableName + ".id": ids}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *SubsQ) FilterByUserIds(userIds ...int64) data.Subs {
	stmt := sq.Eq{permissionsTableName + ".user_id": userIds}

	if len(userIds) == 0 {
		stmt = sq.Eq{permissionsTableName + ".user_id": nil}
	}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *SubsQ) FilterByGithubIds(githubIds ...int64) data.Subs {
	stmt := sq.Eq{permissionsTableName + ".github_id": githubIds}

	q.sql = q.sql.Where(stmt)

	return q
}

func (q *SubsQ) OrderBy(columns ...string) data.Subs {
	q.sql = q.sql.OrderBy(columns...)

	return q
}

func (q *SubsQ) WithPermissions() data.Subs {
	q.sql = sq.Select().Columns(subsColumns...).
		Columns(
			permissionsTableName+".request_id", permissionsTableName+".user_id",
			permissionsTableName+".username", permissionsTableName+".github_id",
			permissionsTableName+".access_level", permissionsTableName+".has_child",
			permissionsTableName+".expires_at", permissionsTableName+".parent_link").
		From(subsTableName).
		LeftJoin(fmt.Sprint(permissionsTableName, " ON ", permissionsTableName, ".link = ", subsTableName, ".link")).
		Where(sq.NotEq{permissionsTableName + ".request_id": nil})

	return q
}

func (q *SubsQ) CountWithPermissions() data.Subs {
	q.sql = sq.Select("COUNT(*)").From(subsTableName).
		LeftJoin(fmt.Sprint(permissionsTableName, " ON ", permissionsTableName, ".link = ", subsTableName, ".link")).
		Where(sq.NotEq{permissionsTableName + ".request_id": nil})

	return q
}

func (q *SubsQ) SearchBy(search string) data.Subs {
	search = strings.Replace(search, " ", "%", -1)
	search = fmt.Sprint("%", search, "%")

	q.sql = q.sql.Where(sq.ILike{"subs.path": search})

	return q
}

func (q *SubsQ) FilterByHasParent(HasParent bool) data.Subs {
	q.sql = q.sql.Where(sq.Eq{permissionsTableName + ".has_parent": HasParent})

	return q
}

func (q *SubsQ) Count() data.Subs {
	q.sql = sq.Select("COUNT (*)").From(subsTableName)

	return q
}

func (q *SubsQ) GetTotalCount() (int64, error) {
	var count int64

	err := q.db.Get(&count, q.sql)

	return count, err
}

func (q *SubsQ) Page(pageParams pgdb.OffsetPageParams) data.Subs {
	q.sql = pageParams.ApplyTo(q.sql, subsTableName+".link")

	return q
}
