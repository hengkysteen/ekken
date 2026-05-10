package module

import (
	"ekken/internal/api/module"
	"ekken/internal/config"
	"ekken/internal/db"
	"ekken/internal/features/credential"
	"ekken/internal/features/workflow/node"

	"github.com/gin-gonic/gin"
)

type CredentialModule struct {
	service credential.Servicer
}

func NewModule() *CredentialModule {
	return &CredentialModule{}
}

func init() {
	module.RegisterModule(NewModule())
}

func (m *CredentialModule) Name() string {
	return "credential"
}

func (m *CredentialModule) Init(database *db.DB, cfg config.Config) error {
	repo, err := credential.NewRepository(database)
	if err != nil {
		return err
	}
	m.service = credential.New(repo)

	// Register global resolver for node templates
	node.CredentialResolver = m.service.GetValueByKey

	return nil
}

func (m *CredentialModule) RegisterRoutes(api *gin.RouterGroup) {
	h := &CredentialHandler{service: m.service}

	creds := api.Group("/credentials")
	{
		creds.GET("", h.ListCredentials)
		creds.POST("", h.CreateCredential)
		creds.GET("/:id", h.GetCredential)
		creds.PUT("/:id", h.UpdateCredential)
		creds.DELETE("/:id", h.DeleteCredential)
	}
}
