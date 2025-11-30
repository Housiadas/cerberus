package user_usecase

import (
	"github.com/Housiadas/cerberus/internal/core/domain/user"
	"github.com/Housiadas/cerberus/pkg/order"
)

var defaultOrderBy = order.NewBy("user_id", order.ASC)

var orderByFields = map[string]string{
	"user_id": user.OrderByID,
	"name":    user.OrderByName,
	"email":   user.OrderByEmail,
	"enabled": user.OrderByEnabled,
}
