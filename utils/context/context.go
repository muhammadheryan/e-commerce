package context

import (
	"context"

	"github.com/muhammadheryan/e-commerce/constant"
)

func GetUserID(ctx context.Context) (uint64, bool) {
	v := ctx.Value(constant.UserIDKey)
	if v == nil {
		return 0, false
	}
	id, ok := v.(uint64)
	return id, ok
}
