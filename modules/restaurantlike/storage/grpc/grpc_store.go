package grpcrestaurantlike

import (
	"context"
	restaurantlikestorage "demo/modules/restaurantlike/storage"
	"demo/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type gRPCServer struct {
	db *gorm.DB
	proto.UnimplementedRestaurantLikeServiceServer
}

func NewGRPCServer(db *gorm.DB) *gRPCServer{
	return &gRPCServer{db:db}
}
func (s *gRPCServer) GetRestaurantLikeStat(ctx context.Context, req *proto.RestaurantLikeStatRequest) (*proto.RestaurantLikeStatResponse, error){
	storage := restaurantlikestorage.NewSQLStore(s.db)
	ids := make([]int,len(req.ResIds))
	for i:= range ids{
		 ids[i] = int(req.ResIds[i])
	}
	result, err := storage.GetRestaurantLikes(ctx, ids)
	if err != nil{
		return nil, status.Errorf(codes.Internal, "method GetRestaurantLikeStat has something error %s", err.Error())
	}
	rep := make(map[int32]int32,len(result))
	for k,v := range  result{
		rep[int32(k)]= int32(v)
	}
	return &proto.RestaurantLikeStatResponse{
		Result:rep,
	},nil
}
