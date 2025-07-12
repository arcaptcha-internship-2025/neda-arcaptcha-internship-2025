package service

import (
	"fmt"
	"net/http"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/config"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/app"
	httpserver "github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/handlers"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/utils"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/repositories"
)

type UserService struct {
	cfg            *config.Config
	userRepository repositories.UserRepository
	userHandler    *handlers.UserHandler
	httpServer     *httpserver.HTTPServer
}

func NewUserService(cfg *config.Config) *UserService {
	return &UserService{
		cfg:        cfg,
		httpServer: httpserver.NewHTTPServer(cfg),
	}
}

func (s *UserService) Start() error {
	db, err := app.ConnectToDatabase(s.cfg.Postgres)
	if err != nil {
		return fmt.Errorf("failed to connect to Postgres: %v", err)
	}

	repo, err := repositories.NewUserRepository(s.cfg.Postgres.AutoCreate, db)
	if err != nil {
		return fmt.Errorf("failed to create user repository: %v", err)
	}
	s.userRepository = repo

	s.userHandler = handlers.NewUserHandler(s.userRepository)

	//with our route setup
	return s.httpServer.Start(s, "user-service")
}

// implements the RouteSetup interface
func (s *UserService) SetupRoutes(mux *http.ServeMux) {
	api := utils.APIPrefix(mux)

	// public routes
	api.HandleFunc("/user/signup", utils.MethodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.SignUp,
	}))

	api.HandleFunc("/user/login", utils.MethodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.Login,
	}))

	// manager routes
	managerRoutes := http.NewServeMux()
	api.Handle("/manager/", http.StripPrefix("/manager", middleware.JWTAuthMiddleware(models.Manager)(managerRoutes)))

	managerRoutes.HandleFunc("/user/get-all", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetAllUsers,
	}))

	managerRoutes.HandleFunc("/user/get", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetUser,
	}))

	managerRoutes.HandleFunc("/user/delete", utils.MethodHandler(map[string]http.HandlerFunc{
		"DELETE": s.userHandler.DeleteUser,
	}))

	// resident routes
	residentRoutes := http.NewServeMux()
	api.Handle("/resident/", http.StripPrefix("/resident", middleware.JWTAuthMiddleware(models.Resident)(residentRoutes)))

	residentRoutes.HandleFunc("/profile", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetProfile,
		"PUT": s.userHandler.UpdateProfile,
	}))

	residentRoutes.HandleFunc("/profile/picture", utils.MethodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.UploadProfilePicture,
	}))
}

func (s *UserService) Stop() error {
	return s.httpServer.Stop()
}
