package store

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// tinyDatesPgStore provide access to the Store methods for a PostgreSQL
// backed database.
type tinydatesPgStore struct {
	Db *pgxpool.Pool
}

func NewTinydatesPgStore(db *pgxpool.Pool) Store {
	return &tinydatesPgStore{Db: db}
}

const (
	storeUser = `
        INSERT INTO users (email, password, name, gender, age, location)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
)

func (store *tinydatesPgStore) StoreNewUser(
	ctx context.Context,
	email, password, name, gender string,
	age, location int,
) (int, error) {
	var id int

	if err := store.Db.QueryRow(
		ctx,
		storeUser,
		email,
		password,
		name,
		gender,
		age,
		location,
	).Scan(
		&id,
	); err != nil {
		return 0, err
	}

	return id, nil
}

const (
	getPassword = `
        SELECT password
		FROM users
		WHERE email = $1
	`
)

func (store *tinydatesPgStore) GetPassword(
	ctx context.Context,
	email string,
) (string, error) {
	var password string

	if err := store.Db.QueryRow(
		ctx,
		getPassword,
		email,
	).Scan(
		&password,
	); err != nil {
		return "", err
	}

	return password, nil
}

const (
	discover = `
        SELECT id, name, gender, age, location
		FROM users
		WHERE id != $1
		AND id NOT IN (
		    SELECT swipee
			FROM swipes
			WHERE swiper = $1
		)
	`
)

func (store *tinydatesPgStore) Discover(
	ctx context.Context,
	id int,
) ([]PotentialMatch, error) {
	potentials := make([]PotentialMatch, 0)
	rows, err := store.Db.Query(ctx, discover, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var user PotentialMatch

		if err := rows.Scan(
			&user.Id,
			&user.Name,
			&user.Gender,
			&user.Age,
			&user.Location,
		); err != nil {
			return nil, err
		}
		potentials = append(potentials, user)
	}


	return potentials, nil
}

const (
	discoverWithAge = `
        SELECT id, name, gender, age, location
		FROM users
		WHERE id != $1
		AND id NOT IN (
		    SELECT swipee
			FROM swipes
			WHERE swiper = $1
		)
		AND age BETWEEN $2 AND $3
	`
)

func (store *tinydatesPgStore) DiscoverWithAge(
	ctx context.Context,
	id int,
	minAge int,
	maxAge int,
) ([]PotentialMatch, error) {
	potentials := make([]PotentialMatch, 0)

	rows, err := store.Db.Query(ctx, discoverWithAge, id, minAge, maxAge)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	for rows.Next() {
		var user PotentialMatch

		if err := rows.Scan(
			&user.Id,
			&user.Name,
			&user.Gender,
			&user.Age,
			&user.Location,
		); err != nil {
			return nil, err
		}
		potentials = append(potentials, user)
	}

	return potentials, nil
}

const (
	swipe = `
        INSERT INTO swipes (swiper, swipee, decision)
		VALUES ($1, $2, $3)
		RETURNING id;
	`

	isMatch = `
        SELECT EXISTS(
			SELECT 1 FROM swipes
			WHERE swiper = $1
			AND swipee = $2
		)
	`
)

func (store *tinydatesPgStore) Swipe(
	ctx context.Context,
	swiperId int,
	swipeeId int,
	decision bool,
) (int, bool, error) {
	var matchId int
	var match bool

	if err := store.Db.QueryRow(
		ctx,
		swipe,
		swiperId,
		swipeeId,
		decision,
	).Scan(
		&matchId,
	); err != nil {
		return 0, false, err
	}

	if err := store.Db.QueryRow(
		ctx,
		isMatch,
		swipeeId,
		swiperId,
	).Scan(
		&match,
	); err != nil {
		return 0, false, err
	}

	return matchId, match, nil
}

const (
	location = `
        SELECT location
		FROM users
		WHERE id = $1
	`
)

func (store *tinydatesPgStore) GetLocation(
	ctx context.Context,
	id int,
) (int, error) {
	var userLocation int

	if err := store.Db.QueryRow(
		ctx,
		location,
		id,
	).Scan(&userLocation); err != nil {
		return 0, err
	}

	return userLocation, nil
}
