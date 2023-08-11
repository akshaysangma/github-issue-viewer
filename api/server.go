package api

import (
	"errors"

	"github.com/akshaysangma/github-issue-viewer/graphql"
	"github.com/akshaysangma/github-issue-viewer/oauth"
	"github.com/sirupsen/logrus"
)

var (
	ErrOAuthConfigMissing   = errors.New("missing OAuth config")
	ErrGraphQLConfigMissing = errors.New("missing GraphQL config")
)

type server struct {
	listenAddr string
	oAuth      *oauth.OAuth
	gqlCfg     *graphql.GraphQLConfig
	logger     *logrus.Logger
}

type ServerOpt func(*server)

func WithListenAddr(addr string) ServerOpt {
	return func(s *server) {
		s.listenAddr = addr
	}
}

func GetListenAddr(s *server) string {
	return s.listenAddr
}

func WithOAuth(o *oauth.OAuth) ServerOpt {
	return func(s *server) {
		s.oAuth = o
	}
}

func WithLogger(l *logrus.Logger) ServerOpt {
	return func(s *server) {
		s.logger = l
	}
}

func WithGraphqlConfig(gcfg *graphql.GraphQLConfig) ServerOpt {
	return func(s *server) {
		s.gqlCfg = gcfg
	}
}

func New(opts ...ServerOpt) (*server, error) {
	s := &server{
		listenAddr: ":8080",
		logger:     logrus.New(),
	}

	for _, opt := range opts {
		opt(s)
	}

	if s.oAuth == nil {
		return nil, ErrOAuthConfigMissing
	}

	if s.gqlCfg == nil {
		return nil, ErrGraphQLConfigMissing
	}

	return s, nil
}
