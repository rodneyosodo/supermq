package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgtype" // required for SQL access
	"github.com/mainflux/mainflux/auth/groups"
	"github.com/mainflux/mainflux/internal/postgres"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mainflux/users"
)

var _ users.Repository = (*usersRepo)(nil)

type usersRepo struct {
	db postgres.Database
}

// NewUsersRepo instantiates a PostgreSQL
// implementation of Users repository.
func NewUsersRepo(db postgres.Database) users.Repository {
	return &usersRepo{
		db: db,
	}
}

func (repo usersRepo) Save(ctx context.Context, user users.User) (users.User, error) {
	q := `INSERT INTO users (id, name, tags, owner, identity, secret, metadata, created_at, updated_at, status)
        VALUES (:id, :name, :tags, :owner, :identity, :secret, :metadata, :created_at, :updated_at, :status)
        RETURNING id, name, tags, identity, metadata, COALESCE(owner, '') AS owner, status, created_at, updated_at`
	if user.Owner == "" {
		q = `INSERT INTO users (id, name, tags, identity, secret, metadata, created_at, updated_at, status)
        VALUES (:id, :name, :tags, :identity, :secret, :metadata, :created_at, :updated_at, :status)
        RETURNING id, name, tags, identity, metadata, COALESCE(owner, '') AS owner, status, created_at, updated_at`
	}
	if user.ID == "" || user.Credentials.Identity == "" {
		return users.User{}, errors.ErrMalformedEntity
	}
	dbc, err := toDBUser(user)
	if err != nil {
		return users.User{}, errors.Wrap(errors.ErrCreateEntity, err)
	}

	row, err := repo.db.NamedQueryContext(ctx, q, dbc)
	if err != nil {
		return users.User{}, postgres.HandleError(err, errors.ErrCreateEntity)
	}

	defer row.Close()
	row.Next()
	var rUser dbUser
	if err := row.StructScan(&rUser); err != nil {
		return users.User{}, err
	}

	return toUser(rUser)
}

func (repo usersRepo) RetrieveByID(ctx context.Context, id string) (users.User, error) {
	q := `SELECT id, name, tags, COALESCE(owner, '') AS owner, identity, secret, metadata, created_at, updated_at, status 
        FROM users WHERE id = $1`

	dbc := dbUser{
		ID: id,
	}

	if err := repo.db.QueryRowxContext(ctx, q, id).StructScan(&dbc); err != nil {
		if err == sql.ErrNoRows {
			return users.User{}, errors.Wrap(errors.ErrNotFound, err)

		}
		return users.User{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	return toUser(dbc)
}

func (repo usersRepo) RetrieveByIdentity(ctx context.Context, identity string) (users.User, error) {
	q := fmt.Sprintf(`SELECT id, name, tags, COALESCE(owner, '') AS owner, identity, secret, metadata, created_at, updated_at, status
        FROM users WHERE identity = $1 AND status = %d`, users.EnabledStatus)

	dbc := dbUser{
		Identity: identity,
	}

	if err := repo.db.QueryRowxContext(ctx, q, identity).StructScan(&dbc); err != nil {
		if err == sql.ErrNoRows {
			return users.User{}, errors.Wrap(errors.ErrNotFound, err)

		}
		return users.User{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	return toUser(dbc)
}

func (repo usersRepo) RetrieveAll(ctx context.Context, pm users.Page) (users.UsersPage, error) {
	query, err := pageQuery(pm)
	if err != nil {
		return users.UsersPage{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	q := fmt.Sprintf(`SELECT u.id, u.name, u.tags, u.identity, u.metadata, COALESCE(u.owner, '') AS owner, u.status, u.created_at
						FROM users u %s ORDER BY u.created_at LIMIT :limit OFFSET :offset;`, query)

	dbPage, err := toDBUsersPage(pm)
	if err != nil {
		return users.UsersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	rows, err := repo.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return users.UsersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	defer rows.Close()

	var items []users.User
	for rows.Next() {
		dbu := dbUser{}
		if err := rows.StructScan(&dbu); err != nil {
			return users.UsersPage{}, errors.Wrap(errors.ErrViewEntity, err)
		}

		c, err := toUser(dbu)
		if err != nil {
			return users.UsersPage{}, err
		}

		items = append(items, c)
	}
	uq := fmt.Sprintf(`SELECT COUNT(*) FROM users u %s;`, query)

	total, err := postgres.Total(ctx, repo.db, uq, dbPage)
	if err != nil {
		return users.UsersPage{}, errors.Wrap(errors.ErrViewEntity, err)
	}

	page := users.UsersPage{
		Users: items,
		Page: users.Page{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}

	return page, nil
}

func (repo usersRepo) Members(ctx context.Context, groupID string, pm users.Page) (users.MembersPage, error) {
	emq, err := pageQuery(pm)
	if err != nil {
		return users.MembersPage{}, err
	}

	q := fmt.Sprintf(`SELECT u.id, u.name, u.tags, u.metadata, u.identity, u.status, u.created_at FROM users u
		INNER JOIN policies ON u.id=policies.subject %s AND policies.object = :group_id
		AND EXISTS (SELECT 1 FROM policies WHERE policies.subject = '%s' AND '%s'=ANY(actions))
	  	ORDER BY u.created_at LIMIT :limit OFFSET :offset;`, emq, pm.Subject, pm.Action)
	dbPage, err := toDBUsersPage(pm)
	if err != nil {
		return users.MembersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveAll, err)
	}
	dbPage.GroupID = groupID
	rows, err := repo.db.NamedQueryContext(ctx, q, dbPage)
	if err != nil {
		return users.MembersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembers, err)
	}
	defer rows.Close()

	var items []users.User
	for rows.Next() {
		dbu := dbUser{}
		if err := rows.StructScan(&dbu); err != nil {
			return users.MembersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembers, err)
		}

		c, err := toUser(dbu)
		if err != nil {
			return users.MembersPage{}, err
		}

		items = append(items, c)
	}
	cq := fmt.Sprintf(`SELECT COUNT(*) FROM users u INNER JOIN policies ON u.id=policies.subject %s AND policies.object = :group_id;`, emq)

	total, err := postgres.Total(ctx, repo.db, cq, dbPage)
	if err != nil {
		return users.MembersPage{}, errors.Wrap(postgres.ErrFailedToRetrieveMembers, err)
	}

	page := users.MembersPage{
		Members: items,
		Page: users.Page{
			Total:  total,
			Offset: pm.Offset,
			Limit:  pm.Limit,
		},
	}
	return page, nil
}

func (repo usersRepo) Update(ctx context.Context, user users.User) (users.User, error) {
	var query []string
	var upq string
	if user.Name != "" {
		query = append(query, "name = :name,")
	}
	if user.Metadata != nil {
		query = append(query, "metadata = :metadata,")
	}
	if len(query) > 0 {
		upq = strings.Join(query, " ")
	}
	q := fmt.Sprintf(`UPDATE users SET %s updated_at = :updated_at
        WHERE id = :id AND status = %d
        RETURNING id, name, tags, identity, metadata, COALESCE(owner, '') AS owner, status, created_at, updated_at`,
		upq, users.EnabledStatus)

	dbu, err := toDBUser(user)
	if err != nil {
		return users.User{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}

	row, err := repo.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return users.User{}, postgres.HandleError(err, errors.ErrCreateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return users.User{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	var rUser dbUser
	if err := row.StructScan(&rUser); err != nil {
		return users.User{}, err
	}

	return toUser(rUser)
}

func (repo usersRepo) UpdateTags(ctx context.Context, user users.User) (users.User, error) {
	q := fmt.Sprintf(`UPDATE users SET tags = :tags, updated_at = :updated_at
        WHERE id = :id AND status = %d
        RETURNING id, name, tags, identity, metadata, COALESCE(owner, '') AS owner, status, created_at, updated_at`,
		users.EnabledStatus)

	dbu, err := toDBUser(user)
	if err != nil {
		return users.User{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}
	row, err := repo.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return users.User{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return users.User{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	var rUser dbUser
	if err := row.StructScan(&rUser); err != nil {
		return users.User{}, err
	}

	return toUser(rUser)
}

func (repo usersRepo) UpdateIdentity(ctx context.Context, user users.User) (users.User, error) {
	q := fmt.Sprintf(`UPDATE users SET identity = :identity, updated_at = :updated_at
        WHERE id = :id AND status = %d
        RETURNING id, name, tags, identity, metadata, COALESCE(owner, '') AS owner, status, created_at, updated_at`,
		users.EnabledStatus)

	dbu, err := toDBUser(user)
	if err != nil {
		return users.User{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}
	row, err := repo.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return users.User{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return users.User{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	var rUser dbUser
	if err := row.StructScan(&rUser); err != nil {
		return users.User{}, err
	}

	return toUser(rUser)
}

func (repo usersRepo) UpdateSecret(ctx context.Context, user users.User) (users.User, error) {
	q := fmt.Sprintf(`UPDATE users SET secret = :secret, updated_at = :updated_at
        WHERE identity = :identity AND status = %d
        RETURNING id, name, tags, identity, metadata, COALESCE(owner, '') AS owner, status, created_at, updated_at`,
		users.EnabledStatus)

	dbu, err := toDBUser(user)
	if err != nil {
		return users.User{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}
	row, err := repo.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return users.User{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return users.User{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	var rUser dbUser
	if err := row.StructScan(&rUser); err != nil {
		return users.User{}, err
	}

	return toUser(rUser)
}

func (repo usersRepo) UpdateOwner(ctx context.Context, user users.User) (users.User, error) {
	q := fmt.Sprintf(`UPDATE users SET owner = :owner, updated_at = :updated_at
        WHERE id = :id AND status = %d
        RETURNING id, name, tags, identity, metadata, COALESCE(owner, '') AS owner, status, created_at, updated_at`,
		users.EnabledStatus)

	dbu, err := toDBUser(user)
	if err != nil {
		return users.User{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}
	row, err := repo.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return users.User{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return users.User{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	var rUser dbUser
	if err := row.StructScan(&rUser); err != nil {
		return users.User{}, err
	}

	return toUser(rUser)
}

func (repo usersRepo) ChangeStatus(ctx context.Context, id string, status users.Status) (users.User, error) {
	q := fmt.Sprintf(`UPDATE users SET status = %d WHERE id = :id
        RETURNING id, name, tags, identity, metadata, COALESCE(owner, '') AS owner, status, created_at, updated_at`, status)

	dbu := dbUser{
		ID: id,
	}
	row, err := repo.db.NamedQueryContext(ctx, q, dbu)
	if err != nil {
		return users.User{}, postgres.HandleError(err, errors.ErrUpdateEntity)
	}

	defer row.Close()
	if ok := row.Next(); !ok {
		return users.User{}, errors.Wrap(errors.ErrNotFound, row.Err())
	}
	var rUser dbUser
	if err := row.StructScan(&rUser); err != nil {
		return users.User{}, errors.Wrap(errors.ErrUpdateEntity, err)
	}

	return toUser(rUser)
}

type dbUser struct {
	ID        string           `db:"id"`
	Name      string           `db:"name,omitempty"`
	Tags      pgtype.TextArray `db:"tags,omitempty"`
	Identity  string           `db:"identity"`
	Owner     string           `db:"owner,omitempty"` // nullable
	Secret    string           `db:"secret"`
	Metadata  []byte           `db:"metadata,omitempty"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
	Groups    []groups.Group   `db:"groups"`
	Status    users.Status     `db:"status"`
}

func toDBUser(user users.User) (dbUser, error) {
	data := []byte("{}")
	if len(user.Metadata) > 0 {
		b, err := json.Marshal(user.Metadata)
		if err != nil {
			return dbUser{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
		data = b
	}
	var tags pgtype.TextArray
	if err := tags.Set(user.Tags); err != nil {
		return dbUser{}, err
	}

	return dbUser{
		ID:        user.ID,
		Name:      user.Name,
		Tags:      tags,
		Owner:     user.Owner,
		Identity:  user.Credentials.Identity,
		Secret:    user.Credentials.Secret,
		Metadata:  data,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Status:    user.Status,
	}, nil
}

func toUser(user dbUser) (users.User, error) {
	var metadata users.Metadata
	if user.Metadata != nil {
		if err := json.Unmarshal([]byte(user.Metadata), &metadata); err != nil {
			return users.User{}, errors.Wrap(errors.ErrMalformedEntity, err)
		}
	}
	var tags []string
	for _, e := range user.Tags.Elements {
		tags = append(tags, e.String)
	}

	return users.User{
		ID:    user.ID,
		Name:  user.Name,
		Tags:  tags,
		Owner: user.Owner,
		Credentials: users.Credentials{
			Identity: user.Identity,
			Secret:   user.Secret,
		},
		Metadata:  metadata,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Status:    user.Status,
	}, nil
}

func pageQuery(pm users.Page) (string, error) {
	mq, _, err := postgres.CreateMetadataQuery("", pm.Metadata)
	if err != nil {
		return "", errors.Wrap(errors.ErrViewEntity, err)
	}
	var query []string
	var emq string
	if mq != "" {
		query = append(query, mq)
	}
	if pm.Name != "" {
		query = append(query, fmt.Sprintf("u.name = '%s'", pm.Name))
	}
	if pm.Tag != "" {
		query = append(query, fmt.Sprintf("'%s' = ANY(u.tags)", pm.Tag))
	}
	if pm.Status != users.AllStatus {
		query = append(query, fmt.Sprintf("u.status = %d", pm.Status))
	}
	if len(pm.UserIDs) > 0 {
		query = append(query, fmt.Sprintf("id IN ('%s')", strings.Join(pm.UserIDs, "','")))
	}
	// For listing users that the specified user owns but not sharedby
	if pm.OwnerID != "" && pm.SharedBy == "" {
		query = append(query, fmt.Sprintf("u.owner = '%s'", pm.OwnerID))
	}

	// For listing users that the specified user owns and that are shared with the specified user
	if pm.OwnerID != "" && pm.SharedBy != "" {
		query = append(query, fmt.Sprintf("(u.owner = '%s' OR policies.object IN (SELECT object FROM policies WHERE subject = '%s' AND '%s'=ANY(actions)))", pm.OwnerID, pm.SharedBy, pm.Action))
	}
	// For listing users that the specified user is shared with
	if pm.SharedBy != "" && pm.OwnerID == "" {
		query = append(query, fmt.Sprintf("u.owner != '%s' AND (policies.object IN (SELECT object FROM policies WHERE subject = '%s' AND '%s'=ANY(actions)))", pm.SharedBy, pm.SharedBy, pm.Action))
	}
	if len(query) > 0 {
		emq = fmt.Sprintf("WHERE %s", strings.Join(query, " AND "))
		if strings.Contains(emq, "policies") {
			emq = fmt.Sprintf("JOIN policies ON policies.subject = u.id %s", emq)
		}
	}
	return emq, nil

}

func toDBUsersPage(pm users.Page) (dbUsersPage, error) {
	_, data, err := postgres.CreateMetadataQuery("", pm.Metadata)
	if err != nil {
		return dbUsersPage{}, errors.Wrap(errors.ErrViewEntity, err)
	}
	return dbUsersPage{
		Name:     pm.Name,
		Metadata: data,
		Owner:    pm.OwnerID,
		Total:    pm.Total,
		Offset:   pm.Offset,
		Limit:    pm.Limit,
		Status:   pm.Status,
		Tag:      pm.Tag,
	}, nil
}

type dbUsersPage struct {
	GroupID  string       `db:"group_id"`
	Name     string       `db:"name"`
	Owner    string       `db:"owner"`
	Identity string       `db:"identity"`
	Metadata []byte       `db:"metadata"`
	Tag      string       `db:"tag"`
	Status   users.Status `db:"status"`
	Total    uint64       `db:"total"`
	Limit    uint64       `db:"limit"`
	Offset   uint64       `db:"offset"`
}
