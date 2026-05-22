package module

import (
	"ekken/internal/api/module"
	"ekken/internal/config"
	"ekken/internal/db"
	"ekken/internal/features/profile"

	"github.com/gin-gonic/gin"
)

type ProfileModule struct {
	service profile.Servicer
}

func NewModule() *ProfileModule {
	return &ProfileModule{}
}

func init() {
	module.RegisterModule(NewModule())
}

type profileName string

func (m *ProfileModule) Name() string {
	return "profile"
}

func (m *ProfileModule) Init(database *db.DB, cfg config.Config) error {
	repo := profile.NewRepository(database)
	m.service = profile.New(repo)
	return nil
}

func (m *ProfileModule) RegisterRoutes(api *gin.RouterGroup) {
	h := &ProfileHandler{service: m.service}

	profileRoutes := api.Group("/profile")
	{
		profileRoutes.GET("", h.GetProfile)
		profileRoutes.PUT("", h.UpdateProfile)
		profileRoutes.POST("/pin/verify", h.VerifyPIN)
		profileRoutes.POST("/pin/reset", h.ResetPIN)
	}
}
