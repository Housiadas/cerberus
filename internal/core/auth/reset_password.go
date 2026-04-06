package auth

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/user"
	errs2 "github.com/Housiadas/cerberus/internal/errs"
	"github.com/Housiadas/cerberus/internal/types/password"
)

// ResetPassword validates the reset token and updates the user's password.
func (s *Service) ResetPassword(ctx context.Context, req ResetPasswordReq) error {
	tkn, err := s.resetTokenSvc.QueryByToken(ctx, req.Token)
	if err != nil {
		return errs2.Errorf(
			errs2.NotFound,
			errs2.CodeInvalidToken,
			"invalid or expired reset token",
		)
	}

	if tkn.IsExpired() {
		return errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeExpiredToken,
			"reset token has expired",
		)
	}

	if tkn.Used() {
		return errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeInvalidToken,
			"reset token has already been used",
		)
	}

	newPassword, parseErr := password.ParseConfirm(req.Password, req.PasswordConfirm)
	if parseErr != nil {
		return errs2.NewFieldErrors("password", parseErr)
	}

	currentUsr, queryErr := s.userService.QueryByID(ctx, tkn.UserID())
	if queryErr != nil {
		return errs2.Errorf(errs2.Internal, errs2.CodeInternal, "query user: %s", queryErr)
	}

	verifyErr := s.userService.VerifyPassword(currentUsr, req.OldPassword)
	if verifyErr != nil {
		return errs2.Errorf(
			errs2.InvalidArgument,
			errs2.CodeAuthFailed,
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
