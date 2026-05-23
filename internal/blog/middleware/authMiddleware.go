package middleware

import (
	"context"
	grpcclient "gobr/internal/blog/grpc_client"
	"net/http"
	"strings"
)

type contextKey string
const UserIDKey contextKey = "userID"

func AuthMiddleWare(authClient *grpcclient.AuthClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func (w http.ResponseWriter, r *http.Request){
			authHeader:=r.Header.Get("Authorization")
			if authHeader=="" {
				http.Error(w,"missing authorization header", http.StatusUnauthorized)
				return
			}
			parts:=strings.SplitN(authHeader," ", 2)
			if len(parts)!=2 || strings.ToLower(parts[0])!="bearer" {
				http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
				return
			}
			token:=parts[1]
			resp,err:=authClient.ValidateToken(r.Context(),token)
			if err!=nil || !resp.Valid {
				http.Error(w, "invalid or expired token",http.StatusUnauthorized)
				return
			}
			ctx:=context.WithValue(r.Context(),UserIDKey,resp.Id)
			next.ServeHTTP(w,r.WithContext(ctx))
		})
	}
}

func GetUser(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(UserIDKey).(string)
	return userID, ok
}