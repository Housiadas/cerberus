package pgsql

import "errors"

var (
	ErrInvalidTransactorType = errors.New("transactor not of type *sql.Tx")
	ErrTransactionNotFound   = errors.New("transaction not found in context")
)
