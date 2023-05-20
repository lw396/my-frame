package service

import (
	"my-frame/internal/repository"
	"my-frame/pkg/log"
	"my-frame/pkg/redis"

	"go.opentelemetry.io/otel/trace"
)

type Service struct {
	*options
}

type JWTConfig struct {
	Secret          string
	ExpireSecs      int
	DefaultPassword string
}

type GoogleOAuth struct {
	ClientId string
}

type options struct {
	rep    repository.Repository
	logger log.Logger
	tracer trace.Tracer
	redis  redis.RedisClient
}

type Option func(*options)

func WithRepository(rep repository.Repository) Option {
	return func(o *options) {
		o.rep = rep
	}
}

func WithLogger(logger log.Logger) Option {
	return func(o *options) {
		o.logger = logger
	}
}

func WithTracer(tracer trace.Tracer) Option {
	return func(o *options) {
		o.tracer = tracer
	}
}

func WithRedis(rc redis.RedisClient) Option {
	return func(o *options) {
		o.redis = rc
	}
}

func New(opts ...Option) *Service {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}
	return &Service{
		options: o,
	}
}
