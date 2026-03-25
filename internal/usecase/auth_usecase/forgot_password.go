package auth_usecase

import (
	"context"
	"fmt"
	"net/mail"

	"github.com/Housiadas/cerberus/internal/core/domain/email_notification_outbox"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/types/event"
	"github.com/Housiadas/cerberus/pkg/pgsql"
)

// emailPayload is the JSON payload stored in email_notification_outbox.
type emailPayload struct {
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	ResetURL string `json:"resetUrl"`
}

// ForgotPassword initiates the password reset flow.
// It always returns nil to prevent email enumeration attacks.
func (u *UseCase) ForgotPassword(ctx context.Context, req ForgotPasswordReq) error {
	addr, err := mail.ParseAddress(req.Email)
	if err != nil {
		return errs2.NewFieldErrors("email", err)
	}

	usr, lookupErr := u.userService.QueryByEmail(ctx, *addr)
	if lookupErr != nil {
		// Silently succeed - do not reveal whether the email exists.
		u.log.Info(ctx, "forgot_password", "email_not_found", req.Email)

		return nil //nolint:nilerr
	}

	txErr := pgsql.RunInTx(ctx, u.log, u.tx, func(tran pgsql.CommitRollbacker) error {
		emailOutboxTx, initErr := u.emailNotificationOutboxSvc.NewWithTx(tran)
		if initErr != nil {
			return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "email_outbox tx: %s", initErr)
		}

		tkn, createErr := u.resetTokenSvc.Create(ctx, usr.ID())
		if createErr != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"create reset token: %s",
				createErr,
			)
		}

		resetURL := fmt.Sprintf("%s/reset-password?token=%s", u.frontendURL, tkn.Token())

		payload := emailPayload{
			Subject:  "Password Reset Request",
			Body:     "Use the link below to reset your password. It expires in 15 minutes.\n\n" + resetURL,
			ResetURL: resetURL,
		}

		createErr = emailOutboxTx.Create(ctx, email_notification_outbox.NewEmailNotificationOutbox{
			EventType: event.PasswordResetRequested,
			ToEmail:   addr.Address,
			Payload:   payload,
		})
		if createErr != nil {
			return errs2.Errorf(
				errs2.Internal,
				errs2.CodeInternal,
				"email_outbox create: %s",
				createErr,
			)
		}

		return nil
	})
	if txErr != nil {
		return fmt.Errorf("forgot password: %w", txErr)
	}

	return nil
}
