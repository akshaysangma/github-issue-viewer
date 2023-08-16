package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"text/template"
)

// TODO: Just testing revisit later
func (s *server) GetIssueForRepositoryHandler(w http.ResponseWriter, r *http.Request, accessToken string) {
	client := s.gqlCfg.NewClient(accessToken)
	// List of repositories
	repositories := []string{"railwayapp/cli", "cockroachdb/cockroach"}
	var collection []any
	// Define the GraphQL query using Hasura-generated types
	for _, repo := range repositories {
		var query struct {
			Repository struct {
				Issues struct {
					Nodes []struct {
						Number int
						Title  string
					}
				} `graphql:"issues(first: 10)"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		// Define variables for the query
		variables := map[string]interface{}{
			"owner": strings.Split(repo, "/")[0],
			"name":  strings.Split(repo, "/")[1],
		}

		// Construct and execute the GraphQL query
		if err := client.Query(context.Background(), &query, variables); err != nil {
			log.Printf("Error querying issues for repository %s: %v", repo, err)
			continue
		}

		// Access the issues for the repository using the response struct
		collection = append(collection, query.Repository)
	}

	data, err := json.Marshal(collection)
	if err != nil {
		s.logger.Error(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func (s *server) SearchForRepositoryHandler(w http.ResponseWriter, r *http.Request, accessToken string) {
	searchInput := r.URL.Query().Get("input")

	client := s.gqlCfg.NewClient(accessToken)
	variables := map[string]interface{}{
		"query": searchInput,
		"first": 5,
	}

	// Define the GraphQL query
	var query struct {
		Search struct {
			RepositoryCount int
			PageInfo        struct {
				StartCursor string
				HasNextPage bool
				EndCursor   string
			}
			Edges []struct {
				Node struct {
					Repository struct {
						NameWithOwner  string
						Description    string
						StargazerCount int
						URL            string
					} `graphql:"... on Repository"`
				}
			}
		} `graphql:"search(query: $query, type: REPOSITORY, first: $first)"`
	}

	// Construct and execute the GraphQL query
	if err := client.Query(context.Background(), &query, variables); err != nil {
		s.logger.Fatal(err)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	tmpl := template.Must(template.ParseFiles("templates/component/search_result.html"))
	err := tmpl.Execute(w, query)
	if err != nil {
		s.logger.Error(err)
	}
}
