package pgsql

import "errors"

var ErrInvalidTransactorType = errors.New("transactor not of type *sql.Tx")
