//
// Copyright (C) 2025 Kevin Diu <kevindiujp@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// You may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package middleware

import (
	"context"
	"time"

	"github.com/kevindiu/monorepo-go-example/internal/log"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs gRPC calls
func LoggingInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		logger.Info("gRPC call started",
			zap.String("method", info.FullMethod),
			zap.Time("start_time", start),
		)

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("duration", duration),
		}

		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields,
				zap.String("error", err.Error()),
				zap.String("code", st.Code().String()),
			)
			logger.Error("gRPC call failed", fields...)
		} else {
			logger.Info("gRPC call completed", fields...)
		}

		return resp, err
	}
}

// RecoveryInterceptor recovers from panics in gRPC handlers
func RecoveryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("gRPC handler panicked",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
				)
				err = status.Errorf(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

// ValidationInterceptor validates incoming requests
func ValidationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// Add validation logic here if needed
		// For now, just pass through
		return handler(ctx, req)
	}
}

// UnaryLoggingInterceptor is a wrapper around LoggingInterceptor that accepts log.Logger
func UnaryLoggingInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
	return LoggingInterceptor(logger.Logger)
}

// UnaryRecoveryInterceptor is a wrapper around RecoveryInterceptor that accepts log.Logger
func UnaryRecoveryInterceptor(logger *log.Logger) grpc.UnaryServerInterceptor {
	return RecoveryInterceptor(logger.Logger)
}
