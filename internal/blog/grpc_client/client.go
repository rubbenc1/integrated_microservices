package grpcclient

import (
	"context"
	pb "gobr/internal/blog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)
type AuthClient struct {
	client pb.AuthServiceClient
	conn *grpc.ClientConn
}

func NewAuthClient(addr string) (*AuthClient, error) {
	conn, err:=grpc.NewClient(addr,grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err!= nil {
		return nil, err
	}
	return &AuthClient{
		client: pb.NewAuthServiceClient(conn),
		conn: conn,
	}, nil
}

func (a *AuthClient) Close() error {
	return a.conn.Close()
}

func (a *AuthClient) ValidateToken(ctx context.Context, token string) (*pb.ValidateTokenRes, error) {
	ctx,cancel:=context.WithTimeout(ctx,2*time.Second)
	defer cancel()
	return a.client.ValidateToken(ctx,&pb.ValidateTokenReq{Token: token})
}