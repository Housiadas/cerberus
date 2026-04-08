package pgsql

import sq "github.com/Masterminds/squirrel"

// Builder is the shared squirrel statement builder configured for Postgres
// ($N placeholders).
var Builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)
