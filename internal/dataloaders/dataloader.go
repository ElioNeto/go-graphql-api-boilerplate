package dataloaders

import (
	"context"
	"net/http"
	"time"

	"github.com/ElioNeto/go-graphql-api-boilerplate/internal/models"
	"github.com/ElioNeto/go-graphql-api-boilerplate/internal/repositories"
	"github.com/vikstrous/dataloadgen"
)

type key string

const LoadersKey key = "dataloaders"

type Loaders struct {
	UserLoader *dataloadgen.Loader[string, *models.User]
}

// Middleware injects dataloaders into the context of each request.
// This is critical for solving the N+1 problem in GraphQL.
func Middleware(userRepo repositories.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			// Batch function for UserLoader
			fetchUsers := func(ctx context.Context, userIDs []string) ([]*models.User, []error) {
				users, err := userRepo.GetByIDs(ctx, userIDs)
				if err != nil {
					// Return same error for all keys
					errs := make([]error, len(userIDs))
					for i := range errs {
						errs[i] = err
					}
					return nil, errs
				}

				// Map users by ID to ensure correct ordering
				userMap := make(map[string]*models.User, len(users))
				for _, u := range users {
					userMap[u.ID] = u
				}

				result := make([]*models.User, len(userIDs))
				errs := make([]error, len(userIDs))

				for i, id := range userIDs {
					if user, ok := userMap[id]; ok {
						result[i] = user
					} else {
						errs[i] = repositories.ErrNotFound
					}
				}

				return result, errs
			}

			loaders := &Loaders{
				UserLoader: dataloadgen.NewLoader(fetchUsers, dataloadgen.WithWait(10*time.Millisecond)),
			}

			ctx = context.WithValue(ctx, LoadersKey, loaders)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// For retrieves the dataloaders from context
func For(ctx context.Context) *Loaders {
	return ctx.Value(LoadersKey).(*Loaders)
}
