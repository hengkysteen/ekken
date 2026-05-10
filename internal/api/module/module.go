package module

import (
	"ekken/internal/config"
	"ekken/internal/db"

	"github.com/gin-gonic/gin"
)

// Module is the contract that must be fulfilled by every modular feature
// to be registered into the API server without tight coupling.
type Module interface {
	// Name returns a unique name for this module.
	Name() string

	// Init initializes the module with access to the database and system configuration.
	Init(db *db.DB, cfg config.Config) error

	// RegisterRoutes registers the module's HTTP endpoints into the router group.
	RegisterRoutes(group *gin.RouterGroup)
}

// ModuleRegistry holds all modules that have registered themselves via init().
var ModuleRegistry = make([]Module, 0)

// RegisterModule is called by modules in their init() function to add themselves to the server.
func RegisterModule(m Module) {
	ModuleRegistry = append(ModuleRegistry, m)
}
