package viewmodels

import "github.com/gin-gonic/gin"

type API struct {
	Path            string
	RegisteredPath  string
	Method          string
	Middlewares     gin.HandlersChain
	HeadMiddlewares gin.HandlersChain
	Handler         gin.HandlerFunc
	// AuthenticatedUser bool
	// Authenticated
	AllowedPolicies [][]string
}



type APIs []API