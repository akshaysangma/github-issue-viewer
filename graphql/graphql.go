package graphql

import (
	"context"

	gqlclient "github.com/hasura/go-graphql-client"
	"golang.org/x/oauth2"
)

type GraphQLConfig struct {
	Endpoint string
}

func New(endpoint string) *GraphQLConfig {
	return &GraphQLConfig{
		Endpoint: endpoint,
	}
}

func (gqlCfg *GraphQLConfig) NewClient(accessToken string) *gqlclient.Client {

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: accessToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	return gqlclient.NewClient(gqlCfg.Endpoint, httpClient)
}
