package data

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"greenlight.alexedwards.net/internal/validator"
	"time"
)

type Game struct {
	ID          int64     `json:"id"`
	CreatedAt   time.Time `json:"-"`
	Title       string    `json:"name"`
	Year        int32     `json:"year,omitempty"`
	Runtime     Runtime   `json:"dateOfcreate,omitempty"`
	Genres      []string  `json:"genres,omitempty"`
	Description []string  `json:"description"`
	Size        float64   `json:"size"`
	Price       float64   `json:"price"`
	Version     int32     `json:"version"`
}

type GameModel struct {
	DB *sql.DB
}

func ValidateGame(v *validator.Validator, game *Game) {
	v.Check(game.Title != "", "title", "must be provided")
	v.Check(len(game.Title) <= 500, "title", "must not be more than 500 bytes long")
	v.Check(game.Year != 0, "year", "must be provided")
	v.Check(game.Year >= 1888, "year", "must be greater than 1888")
	v.Check(game.Year <= int32(time.Now().Year()), "year", "must not be in the future")
	v.Check(game.Runtime != 0, "runtime", "must be provided")
	v.Check(game.Runtime > 0, "runtime", "must be a positive integer")
	v.Check(game.Genres != nil, "genres", "must be provided")
	v.Check(len(game.Genres) >= 1, "genres", "must contain at least 1 genre")
	v.Check(len(game.Genres) <= 5, "genres", "must not contain more than 5 genres")
	v.Check(validator.Unique(game.Genres), "genres", "must not contain duplicate values")
}

func (m GameModel) Insert(game *Game) error {
	query := `
INSERT INTO game (title, year, runtime, genres,)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, version`
	args := []any{game.Title, game.Year, game.Runtime, pq.Array(game.Genres)}
	// Create a context with a 3-second timeout.
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Use QueryRowContext() and pass the context as the first argument.
	return m.DB.QueryRowContext(ctx, query, args...).Scan(&game.ID, &game.CreatedAt, &game.Version)
}
func (m GameModel) Get(id int64) (*Game, error) {
	if id < 1 {
		return nil, ErrRecordNotFound
	}
	// Remove the pg_sleep(10) clause.
	query := `
SELECT id, runtime, title, description, genres, year, version, size, price,CreatedAt
FROM games
WHERE id = $1`
	var game Game
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Remove &[]byte{} from the first Scan() destination.
	err := m.DB.QueryRowContext(ctx, query, id).Scan(
		&game.ID,
		&game.CreatedAt,
		&game.Title,
		&game.Year,
		&game.Runtime,
		pq.Array(&game.Genres),
		&game.Version,
		&game.Price,
		&game.Size,
		&game.Description,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &game, nil
}
func (m GameModel) Update(game *Game) error {
	query := `
UPDATE games
SET title = $1, year = $2, runtime = $3, genres = $4, version = version + 1
WHERE id = $5 AND version = $6
RETURNING version`
	args := []any{
		game.Title,
		game.Year,
		game.Runtime,
		pq.Array(game.Genres),
		game.ID,
		game.Version,
	}
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Use QueryRowContext() and pass the context as the first argument.
	// Use QueryRowContext() and pass the context as the first argument.
	err := m.DB.QueryRowContext(ctx, query, args...).Scan(&game.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}
func (m GameModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}
	query := `
DELETE FROM games
WHERE id = $1`
	// Create a context with a 3-second timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	// Use ExecContext() and pass the context as the first argument.
	result, err := m.DB.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return ErrRecordNotFound
	}
	return nil
}

// Create a new GetAll() method which returns a slice of games. Although we're not
// using them right now, we've set this up to accept the various filter parameters as
// arguments.
func (m GameModel) GetAll(title string, genres []string, filters Filters) ([]*Game, Metadata, error) {
	// Update the SQL query to include the window function which counts the total
	// (filtered) records.
	query := fmt.Sprintf(`
SELECT count(*) OVER(), id, created_at, title, year, runtime, genres, version
FROM games
WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '')
AND (genres @> $2 OR $2 = '{}')
ORDER BY %s %s, id ASC
LIMIT $3 OFFSET $4`, filters.sortColumn(), filters.sortDirection())
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	args := []any{title, pq.Array(genres), filters.limit(), filters.offset()}
	rows, err := m.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, Metadata{}, err // Update this to return an empty Metadata struct.
	}
	defer rows.Close()
	// Declare a totalRecords variable.
	totalRecords := 0
	games := []*Game{}
	for rows.Next() {
		var game Game
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&game.ID,
			&game.CreatedAt,
			&game.Title,
			&game.Year,
			&game.Runtime,
			pq.Array(&game.Genres),
			&game.Version,
		)
		if err != nil {
			return nil, Metadata{}, err // Update this to return an empty Metadata struct.
		}
		games = append(games, &game)
	}
	if err = rows.Err(); err != nil {
		return nil, Metadata{}, err // Update this to return an empty Metadata struct.
	}
	// Generate a Metadata struct, passing in the total record count and pagination
	// parameters from the client.
	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)
	// Include the metadata struct when returning.
	return games, metadata, nil
}
