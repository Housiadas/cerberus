package auth_usecase

import (
	"context"
	"fmt"

	"github.com/Housiadas/cerberus/internal/core/domain/password"
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/internal/utils/errs"
)

// ResetPassword validates the reset token and updates the user's password.
func (u *UseCase) ResetPassword(ctx context.Context, req ResetPasswordReq) error {
	tkn, err := u.resetTokenSvc.QueryByToken(ctx, req.Token)
	if err != nil {
		return errs.Errorf(errs.NotFound, "invalid or expired reset token")
	}

	if tkn.IsExpired() {
		return errs.Errorf(errs.InvalidArgument, "reset token has expired")
	}

	if tkn.Used() {
		return errs.Errorf(errs.InvalidArgument, "reset token has already been used")
	}

	newPassword, parseErr := password.ParseConfirm(req.Password, req.PasswordConfirm)
	if parseErr != nil {
		return errs.NewFieldErrors("password", parseErr)
	}

	currentUsr, queryErr := u.userService.QueryByID(ctx, tkn.UserID())
	if queryErr != nil {
		return errs.Errorf(errs.Internal, "query user: %s", queryErr)
	}

	verifyErr := u.userService.VerifyPassword(currentUsr, req.OldPassword)
	if verifyErr != nil {
		return errs.Errorf(errs.InvalidArgument, "old password is incorrect")
	}

	uu := user.UpdateUser{
		Password: &newPassword,
	}

	_, updateErr := u.userService.Update(ctx, currentUsr, uu)
	if updateErr != nil {
		return fmt.Errorf("update password: %w", updateErr)
	}

	deleteErr := u.resetTokenSvc.Delete(ctx, tkn)
	if deleteErr != nil {
		return fmt.Errorf("delete reset token: %w", deleteErr)
	}

	return nil
}
