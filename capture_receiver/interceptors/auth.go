package interceptors

import (
	"context"
	"fmt"
	"strings"

	"github.com/nrf24l01/sniffly/capture_receiver/postgres"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm"
)

func NewAuthInterceptors(db *gorm.DB) (
    grpc.UnaryServerInterceptor,
    grpc.StreamServerInterceptor,
) {
    unary := func(
        ctx context.Context,
        req interface{},
        info *grpc.UnaryServerInfo,
        handler grpc.UnaryHandler,
    ) (interface{}, error) {
        if err := authorize(ctx, db); err != nil {
            return nil, err
        }
        return handler(ctx, req)
    }

    stream := func(
        srv interface{},
        ss grpc.ServerStream,
        info *grpc.StreamServerInfo,
        handler grpc.StreamHandler,
    ) error {
        if err := authorize(ss.Context(), db); err != nil {
            return err
        }
        return handler(srv, ss)
    }

    return unary, stream
}

func authorize(ctx context.Context, db *gorm.DB) error {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return fmt.Errorf("missing metadata")
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return fmt.Errorf("missing authorization header")
	}

	token := strings.TrimPrefix(values[0], "Bearer ")
	if err := validateToken(token, db); err != nil {
		return err
	}
	return nil
}

func validateToken(token string, db *gorm.DB) error {
	if err := db.Where("api_key = ?", token).First(&postgres.Capturer{}).Error; err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}
	return nil
}