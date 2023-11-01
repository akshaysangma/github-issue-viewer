package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"text/template"
	"time"
)

// query {
//   repository(owner: "golang", name: "go") {
//     nameWithOwner
//     issues(first: 1, orderBy: {field: UPDATED_AT, direction: DESC}, states: [OPEN]) {
//       totalCount
//       pageInfo {
//         startCursor
//         hasNextPage
//         hasPreviousPage
//       }
//       nodes {
//         title
//         updatedAt
//         assignees(first: 1) {
//           nodes {
//             name
//           }
//         }
//         milestone {
//           title
//         }
//         labels(first: 10) {
//           nodes {
//             name
//             color
//           }
//         }
//         author {
//           login
//         }
//       }
//     }
//   }
// }

// TODO: Just testing revisit later
func (s *server) GetIssueForRepositoryHandler(
	w http.ResponseWriter,
	r *http.Request,
	accessToken string,
) {
	client := s.gqlCfg.NewClient(accessToken)
	// List of repositories
	repositories := []string{"railwayapp/cli", "golang/go"}
	var collection []any
	// Define the GraphQL query using Hasura-generated types
	for _, repo := range repositories {

		var query struct {
			Repository struct {
				NameWithOwner string
				Issues        struct {
					TotalCount int
					PageInfo   struct {
						StartCursor     string
						HasNextPage     bool
						HasPreviousPage bool
					}
					Nodes []struct {
						Title     string
						UpdatedAt time.Time
						Assignees struct {
							Nodes []struct {
								Name string
							}
						} `graphql:"assignees(first: 1)"`
						Milestone struct {
							Title string
						}
						Labels struct {
							Nodes []struct {
								Name  string
								Color string
							}
						} `graphql:"labels(first: 10)"`
						Author struct {
							Login string
						}
					}
				} `graphql:"issues(first: 1, orderBy: {field: UPDATED_AT, direction: DESC}, states: [OPEN])"`
			} `graphql:"repository(owner: $owner, name: $name)"`
		}

		variables := map[string]interface{}{
			"owner": strings.Split(repo, "/")[0],
			"name":  strings.Split(repo, "/")[1],
		}

		if err := client.Query(context.Background(), &query, variables); err != nil {
			log.Printf("Error querying issues for repository %s: %v", repo, err)
			continue
		}

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

func (s *server) SearchForRepositoryHandler(
	w http.ResponseWriter,
	r *http.Request,
	accessToken string,
) {
	searchInput := r.URL.Query().Get("input")

	client := s.gqlCfg.NewClient(accessToken)

	variables := map[string]interface{}{
		"query": searchInput,
		"first": 5,
	}

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
