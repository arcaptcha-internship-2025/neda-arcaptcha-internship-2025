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
	managerRoutes.HandleFunc("/user/{user_id}", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET":    s.userHandler.GetUser,
		"DELETE": s.userHandler.DeleteUser,
	}))

	managerRoutes.HandleFunc("/apartment", s.methodHandler(map[string]http.HandlerFunc{
		"POST":   s.apartmentHandler.CreateApartment,
		"GET":    s.apartmentHandler.GetApartmentByID,
		"PUT":    s.apartmentHandler.UpdateApartment,
		"DELETE": s.apartmentHandler.DeleteApartment,
	}))
	managerRoutes.HandleFunc("/apartments/get-all/resident/{user_id}", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetAllApartmentsForResident,
	}))
	managerRoutes.HandleFunc("/apartment/{apartment_id}/residents", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.GetResidentsInApartment,
	}))
	managerRoutes.HandleFunc("/apartment/{apartment_id}/invite/resident/{telegram_username}", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.InviteUserToApartment,
	}))
	managerRoutes.HandleFunc("/bill/{apartment_id}/create", utils.MethodHandler(map[string]http.HandlerFunc{
		"POST": s.billHandler.CreateBill,
	}))

	managerRoutes.HandleFunc("/bills/{apartment_id}/divide/{bill_type}", utils.MethodHandler(map[string]http.HandlerFunc{
		"POST": s.billHandler.DivideBillByType,
	}))

	managerRoutes.HandleFunc("/bills/{apartment_id}/divide-all", utils.MethodHandler(map[string]http.HandlerFunc{
		"POST": s.billHandler.DivideAllBills,
	}))

	managerRoutes.HandleFunc("/bill", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET":    s.billHandler.GetBillByID,
		"PUT":    s.billHandler.UpdateBill,
		"DELETE": s.billHandler.DeleteBill,
	}))
	managerRoutes.HandleFunc("/bills/get-all", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.billHandler.GetBillsByApartment,
	}))
	// resident routes
	residentRoutes := http.NewServeMux()
	v1.Handle("/resident/", http.StripPrefix("/resident", middleware.JWTAuthMiddleware(models.Resident, models.Manager)(residentRoutes)))

	residentRoutes.HandleFunc("/profile", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.userHandler.GetProfile,
		"PUT": s.userHandler.UpdateProfile,
	}))
	residentRoutes.HandleFunc("/apartment/invite/{invitation_code}", s.methodHandler(map[string]http.HandlerFunc{
		"GET": s.apartmentHandler.JoinApartment,
	}))
	residentRoutes.HandleFunc("/apartment/leave", s.methodHandler(map[string]http.HandlerFunc{
		"POST": s.apartmentHandler.LeaveApartment,
	}))

	residentRoutes.HandleFunc("/bills/pay",
		middleware.IdempotentKeyMiddleware(
			utils.MethodHandler(map[string]http.HandlerFunc{
				"POST": s.billHandler.PayBills,
			}),
		).ServeHTTP,
	)

	residentRoutes.HandleFunc("/bills/pay-batch",
		middleware.IdempotentKeyMiddleware(
			utils.MethodHandler(map[string]http.HandlerFunc{
				"POST": s.billHandler.PayBatchBills,
			}),
		).ServeHTTP,
	)

	residentRoutes.HandleFunc("/bills/get-unpaid", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.billHandler.GetUnpaidBills,
	}))
	residentRoutes.HandleFunc("/bill/status", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.billHandler.GetBillWithPaymentStatus,
	}))
	residentRoutes.HandleFunc("/bills/payment-history", utils.MethodHandler(map[string]http.HandlerFunc{
		"GET": s.billHandler.GetUserPaymentHistory,
	}))

}
