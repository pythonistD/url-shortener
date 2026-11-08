package postgres

var URLQueries = struct {
	Save          string
	GetByOriginal string
	GetByShort    string
}{
	Save: `
		INSERT INTO urls (id, original, short, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
	`,
	GetByOriginal: `
		SELECT id, original, short, created_at, expires_at
		FROM urls
		WHERE original = $1
		ORDER BY created_at DESC
		LIMIT 1
	`,
	GetByShort: `
		SELECT id, original, short, created_at, expires_at
		FROM urls
		WHERE short = $1
	`,
}
