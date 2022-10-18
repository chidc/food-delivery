package subscriber

import (
	"context"
	"demo/component"
)

func Setup(ctx component.AppContext) {
	IncreaseLikeCountAfterUserLikeRestaurant(ctx, context.Background())
}
