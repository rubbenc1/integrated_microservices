package grpcserver

import (
	"context"
	pb "gobr/internal/auth"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type AuthServiceManager struct {
	pb.UnimplementedAuthServiceServer
	jwtSecret []byte
}

func NewAuthServiceManager(jwtSecret string) *AuthServiceManager {
	return &AuthServiceManager{
		jwtSecret: []byte(jwtSecret),
	}
}

func (a *AuthServiceManager) ValidateToken(ctx context.Context, req *pb.ValidateTokenReq) (*pb.ValidateTokenRes, error) {
	tokenStr := strings.TrimPrefix(req.Token, "Bearer ")
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return a.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return &pb.ValidateTokenRes{Valid: false}, nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &pb.ValidateTokenRes{Valid: false}, nil
	}
	id, _ := claims["id"].(string)
	return &pb.ValidateTokenRes{
		Valid: true,
		Id:    id,
	}, nil
}
