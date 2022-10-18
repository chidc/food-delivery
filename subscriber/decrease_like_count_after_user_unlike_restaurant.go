package subscriber

import (
	"context"
	"demo/component"
	"demo/modules/restaurant/restaurantstorage"
	"demo/pubsub"
	"demo/skio"
	"go.opencensus.io/trace"
)

func RunDecreaseLikeCountAfterUserUnlikeRestaurant(appCtx component.AppContext, rtEngine skio.RealtimeEngine) consumerJob {
	return consumerJob{
		Title: "Decrease like count after user unlikes restaurant",
		Hld: func(ctx context.Context, message *pubsub.Message) error {
			store := restaurantstorage.NewSQLStore(appCtx.GetMainDBConnection())
			likeData := message.Data().(HasRestaurantId)

			ctx1, span := trace.StartSpan(ctx, "pubsub.sub.RunDecreaseLikeCountAfterUserUnlikeRestaurant")
			defer span.End()

			return store.DecreaseLikeCount(ctx1, likeData.GetRestaurantId())
		},
	}
}
