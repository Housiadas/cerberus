package auth

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/user"
	"github.com/Housiadas/cerberus/internal/sdk/errs"
	"github.com/Housiadas/cerberus/internal/types/password"
)

// ResetPassword validates the reset token and updates the user's password.
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordReq) error {
	tkn, err := s.resetTokenSvc.QueryByToken(ctx, req.Token)
	if err != nil {
		return errs.Errorf(
			errs.NotFound,
			errs.CodeInvalidToken,
			"invalid or expired reset token",
		)
	}

	if tkn.IsExpired() {
		return errs.Errorf(
			errs.InvalidArgument,
			errs.CodeExpiredToken,
			"reset token has expired",
		)
	}

	if tkn.Used() {
		return errs.Errorf(
			errs.InvalidArgument,
			errs.CodeInvalidToken,
			"reset token has already been used",
		)
	}

	newPassword, parseErr := password.ParseConfirm(req.Password, req.PasswordConfirm)
	if parseErr != nil {
		return errs.NewFieldErrors("password", parseErr)
	}

	currentUsr, queryErr := s.userService.QueryByID(ctx, tkn.UserID())
	if queryErr != nil {
		return errs.Errorf(errs.Internal, errs.CodeInternal, "query user: %s", queryErr)
	}

	verifyErr := s.userService.VerifyPassword(currentUsr, req.OldPassword)
	if verifyErr != nil {
		return errs.Errorf(
			errs.InvalidArgument,
			errs.CodeAuthFailed,
			"old password is incorrect",
		)
	}

	uu := user.UpdateUser{
		Password: &newPassword,
	}

	_, updateErr := s.userService.Update(ctx, currentUsr, uu)
	if updateErr != nil {
		return fmt.Errorf("update password: %w", updateErr)
	}

	deleteErr := s.resetTokenSvc.Delete(ctx, tkn)
	if deleteErr != nil {
		return fmt.Errorf("delete reset token: %w", deleteErr)
	}

	return nil
}
