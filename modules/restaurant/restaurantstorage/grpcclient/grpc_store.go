package grpcrestaurant

import (
	"context"
	"demo/common"
	"demo/proto"
)
type grpcClient struct{
	client proto.RestaurantLikeServiceClient
}

func NewGrpcClient(cw proto.RestaurantLikeServiceClient) *grpcClient {
	return &grpcClient{
		client: cw,
	}
}

func (c *grpcClient) GetRestaurantLikes(ctx context.Context, ids []int) (map[int]int, error) {
	id := make([]int32,len(ids))

	for k,v := range ids{
		id[k] = int32(v)
	}

	res, err := c.client.GetRestaurantLikeStat(ctx, &proto.RestaurantLikeStatRequest{ResIds: id})
	if err != nil{
		return nil, common.ErrDB(err)
	}
	result := make(map[int]int)
	for k,v := range res.Result{
		result[int(k)] = int(v)
	}
	return result, nil
}

