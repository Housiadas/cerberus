package middleware

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/Housiadas/cerberus/internal/utils/errs"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// BeginTransaction starts a transaction for the core call.
func (m *Middleware) BeginTransaction() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			hasCommitted := false

			m.log.Info(ctx, "BEGIN TRANSACTION")

			tx, err := m.tx.Begin()
			if err != nil {
				err := errs.Errorf(errs.Internal, "BEGIN TRANSACTION: %s", err)
				m.log.Error(ctx, "transaction middleware", err)
				m.Error(w, err, http.StatusInternalServerError)

				return
			}

			defer func() {
				if !hasCommitted {
					m.log.Info(ctx, "ROLLBACK TRANSACTION")
				}

				err := tx.Rollback()
				if err != nil {
					if errors.Is(err, sql.ErrTxDone) {
						return
					}

					m.log.Error(ctx, "ROLLBACK TRANSACTION", "ERROR", err)
				}
			}()

			ctx = pgsql.SetTran(ctx, tx)

			// Create a response recorder to capture the response
			rec := &responseRecorder{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(rec, r.WithContext(ctx))

			// Access the recorded response
			// Check if we can commit transaction
			if rec.statusCode >= http.StatusBadRequest {
				m.log.Info(ctx, "TRANSACTION FAILED, WILL ROLLBACK")

				return
			}

			m.log.Info(ctx, "COMMIT TRANSACTION")

			err = tx.Commit()
			if err != nil {
				err := errs.Errorf(errs.Internal, "COMMIT TRANSACTION: %s", err)
				m.log.Error(ctx, "transaction middleware", err)
				m.Error(w, err, http.StatusInternalServerError)

				return
			}

			hasCommitted = true
		})
	}
}
