package models

//go:generate rm -f cache.db
//go:generate sql-migrate up -env=cache-development
//go:generate sqlboiler -o . sqlite3
//go:generate rm cache.db
