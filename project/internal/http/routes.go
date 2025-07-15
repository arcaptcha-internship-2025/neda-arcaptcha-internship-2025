package http

import (
	"net/http"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/utils"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
)

func (s *ApartmantService) SetupRoutes(mux *http.ServeMux) {
	v1 := utils.APIPrefix(mux)

	// public routes
	v1.HandleFunc("/user/signup", utils.MethodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.SignUp,
	}))

	v1.HandleFunc("/user/login", utils.MethodHandler(map[string]http.HandlerFunc{
		"POST": s.userHandler.Login,
	}))

	// manager routes
	managerRoutes := http.NewServeMux()
	v1.Handle("/manager/", http.StripPrefix("/manager", middleware.JWTAuthMiddleware(models.Manager)(managerRoutes)))

	managerRoutes.HandleFunc("/user/get-all", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetAllUsers,
	}))

	managerRoutes.HandleFunc("/user/get", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetUser,
	}))

	managerRoutes.HandleFunc("/user/delete", utils.MethodHandler(map[string]http.HandlerFunc{
		"DELETE": s.userHandler.DeleteUser,
	}))

	managerRoutes.HandleFunc("/apartment/create", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.CreateApartment,
	}))
	managerRoutes.HandleFunc("/apartment/get", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetApartmentByID,
	}))
	managerRoutes.HandleFunc("/apartments/get-all/resident/{userID}", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetAllApartmentsForResident,
	}))
	managerRoutes.HandleFunc("/apartment/update", s.methodHandler(map[string]http.HandlerFunc{
		"PUT": s.apartmentHandler.UpdateApartment,
	}))
	managerRoutes.HandleFunc("/apartment/delete", s.methodHandler(map[string]http.HandlerFunc{
		"DELETE": s.apartmentHandler.DeleteApartment,
	}))
	managerRoutes.HandleFunc("/apartment/residents", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetResidentsInApartment,
	}))
	managerRoutes.HandleFunc("/apartment/invite", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.InviteUserToApartment,
	}))

	// resident routes
	residentRoutes := http.NewServeMux()
	v1.Handle("/resident/", http.StripPrefix("/resident", middleware.JWTAuthMiddleware(models.Resident, models.Manager)(residentRoutes)))

	residentRoutes.HandleFunc("/profile", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetProfile,
		"PUT": s.userHandler.UpdateProfile,
	}))

	residentRoutes.HandleFunc("/apartment/join", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.JoinApartment,
	}))
	residentRoutes.HandleFunc("/apartment/leave", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.LeaveApartment,
	}))
}
