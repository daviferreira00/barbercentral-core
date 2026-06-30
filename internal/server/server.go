package server

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jmoiron/sqlx"

	"barbercentral-core/internal/admin"
	"barbercentral-core/internal/appointment"
	"barbercentral-core/internal/auth"
	"barbercentral-core/internal/clientconfig"
	"barbercentral-core/internal/config"
	"barbercentral-core/internal/customer"
	"barbercentral-core/internal/email"
	"barbercentral-core/internal/finance"
	"barbercentral-core/internal/middleware"
	"barbercentral-core/internal/professional"
	"barbercentral-core/internal/service"
	"barbercentral-core/internal/stock"
)

func New(db *sqlx.DB, cfg *config.Config) http.Handler {
	router := chi.NewRouter()

	// Middlewares globais
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestLogger)

	// Health check
	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	// Servidor de arquivos estáticos para fotos locais (uploads)
	router.Handle("/uploads/*", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./uploads"))))

	// 1. Instanciar Repositórios
	authRepo := auth.NewAuthRepository(db)
	profRepo := professional.NewProfessionalRepository(db)
	svcRepo := service.NewServiceRepository(db)
	configRepo := clientconfig.NewConfigRepository(db)
	custRepo := customer.NewCustomerRepository(db)
	finRepo := finance.NewRepository(db)
	stockRepo := stock.NewRepository(db)
	appRepo := appointment.NewAppointmentRepository(db)
	admRepo := admin.NewAdminRepository(db)

	// 1.5. Instanciar Cliente de E-mail
	emailClient := email.NewClient(cfg.Resend.APIKey, cfg.Resend.FromEmail)

	// 2. Instanciar Serviços
	authService := auth.NewAuthService(authRepo, cfg.JWT.Secret, emailClient)
	profService := professional.NewService(profRepo)
	svcService := service.NewServiceService(svcRepo)
	configService := clientconfig.NewConfigService(configRepo)
	custService := customer.NewService(custRepo)
	finService := finance.NewService(finRepo)
	stockService := stock.NewService(stockRepo)
	appService := appointment.NewService(appRepo, configRepo, profRepo, svcRepo, emailClient, custService, stockService)
	admService := admin.NewAdminService(admRepo)

	// 3. Instanciar Handlers
	authHandler := auth.NewAuthHandler(authService)
	profHandler := professional.NewProfessionalHandler(profService)
	svcHandler := service.NewServiceHandler(svcService)
	configHandler := clientconfig.NewConfigHandler(configService, profService, svcService)
	custHandler := customer.NewCustomerHandler(custService)
	finHandler := finance.NewHandler(finService)
	stockHandler := stock.NewHandler(stockService)
	appHandler := appointment.NewAppointmentHandler(appService)
	admHandler := admin.NewAdminHandler(admService)

	// 4. Middlewares de Segurança
	authMiddleware := middleware.RequireAuth(authService)

	// Role limits
	receptionistRoleLimit := middleware.RequireRole("receptionist", "professional", "manager", "owner")
	managerRoleLimit := middleware.RequireRole("manager", "owner")
	ownerRoleLimit := middleware.RequireRole("owner")
	adminRoleLimit := middleware.RequireRole("admin")

	// 5. Configurar Versionamento de API (/api/v1)
	router.Route("/api/v1", func(r chi.Router) {
		// Rotas públicas de auth
		authHandler.RegisterRoutes(r)

		// Rotas públicas do portal de auto-agendamento e cancelamentos
		r.Get("/public/{slug}", configHandler.GetPublicData)
		r.Get("/public/{slug}/services", configHandler.GetPublicServices)
		r.Get("/public/{slug}/professionals", configHandler.GetPublicProfessionals)
		r.Get("/public/{slug}/availability", appHandler.GetAvailability)
		r.Post("/public/{slug}/appointments", appHandler.CreatePublic)
		r.Get("/public/appointments/cancel/{token}", appHandler.GetByCancelToken)
		r.Post("/public/appointments/cancel/{token}", appHandler.CancelByToken)

		// Endpoint de serviço de lembretes automáticos
		r.Post("/internal/reminders/send", appHandler.SendReminders)

		// Rotas protegidas
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)

			// Perfil do Usuário
			r.Get("/auth/me", authHandler.Me)

			// Configurações da barbearia
			r.Get("/config", configHandler.GetConfig)
			r.With(managerRoleLimit).Put("/config", configHandler.UpdateConfig)
			r.With(managerRoleLimit).Post("/config/logo", configHandler.UploadLogo)

			// Módulo de Profissionais
			r.Get("/professionals", profHandler.List)
			r.Get("/professionals/{id}", profHandler.GetByID)
			r.Get("/professionals/{id}/schedule", profHandler.GetSchedule)
			r.Get("/professionals/{id}/services", profHandler.GetLinkedServices)

			r.With(managerRoleLimit).Post("/professionals", profHandler.Create)
			r.With(managerRoleLimit).Put("/professionals/{id}", profHandler.Update)
			r.With(managerRoleLimit).Post("/professionals/{id}/photo", profHandler.UploadPhoto)
			r.With(managerRoleLimit).Put("/professionals/{id}/schedule", profHandler.SaveSchedule)
			r.With(managerRoleLimit).Post("/professionals/{id}/services", profHandler.LinkService)
			r.With(managerRoleLimit).Delete("/professionals/{id}/services/{service_id}", profHandler.UnlinkService)

			r.With(ownerRoleLimit).Delete("/professionals/{id}", profHandler.Delete)

			// Módulo de Serviços
			r.Get("/services", svcHandler.List)
			r.Get("/services/{id}", svcHandler.GetByID)
			r.Get("/service-categories", svcHandler.ListCategories)

			r.With(managerRoleLimit).Post("/services", svcHandler.Create)
			r.With(managerRoleLimit).Put("/services/{id}", svcHandler.Update)
			r.With(managerRoleLimit).Post("/service-categories", svcHandler.CreateCategory)

			r.With(ownerRoleLimit).Delete("/services/{id}", svcHandler.Delete)

			// Módulo de Agenda e Bloqueios
			r.Get("/blocked-slots", appHandler.ListBlockedSlots)
			r.Get("/appointments", appHandler.List)
			r.Get("/appointments/{id}", appHandler.GetByID)
			r.Get("/appointments/availability", appHandler.GetAvailabilityInternal)
			r.Post("/appointments", appHandler.Create)
			r.Put("/appointments/{id}", appHandler.Update)
			r.Get("/appointments/{id}/logs", appHandler.GetStatusLogs)
			r.With(managerRoleLimit).Delete("/appointments/{id}", appHandler.Cancel)

			r.With(managerRoleLimit).Post("/blocked-slots", appHandler.CreateBlockedSlot)
			r.With(managerRoleLimit).Delete("/blocked-slots/{id}", appHandler.DeleteBlockedSlot)

			r.With(receptionistRoleLimit).Patch("/appointments/{id}/status", appHandler.UpdateStatus)

			// Módulo CRM de Clientes
			r.Get("/customers", custHandler.List)
			r.Get("/customers/search", custHandler.Search)
			r.Get("/customers/{id}", custHandler.GetByID)
			r.Post("/customers", custHandler.Create)
			r.Put("/customers/{id}", custHandler.Update)
			r.Get("/customers/{id}/appointments", custHandler.GetAppointmentsHistory)
			r.With(managerRoleLimit).Delete("/customers/{id}", custHandler.Delete)
			r.With(managerRoleLimit).Get("/customers/export", custHandler.ExportCSV)

			// Módulo de Caixa e Financeiro
			r.With(managerRoleLimit).Get("/cash-registers", finHandler.List)
			r.Get("/cash-registers/current", finHandler.GetCurrent)
			r.With(managerRoleLimit).Post("/cash-registers/open", finHandler.Open)
			r.With(managerRoleLimit).Post("/cash-registers/{id}/close", finHandler.Close)
			r.With(managerRoleLimit).Get("/cash-registers/{id}/summary", finHandler.GetSummary)
			r.With(managerRoleLimit).Get("/cash-registers/{id}/export", finHandler.ExportCSV)
			r.Get("/cash-registers/{id}/transactions", finHandler.ListTransactions)
			r.Post("/cash-registers/{id}/transactions", finHandler.CreateTransaction)
			r.With(managerRoleLimit).Delete("/cash-registers/{id}/transactions/{tx_id}", finHandler.DeleteTransaction)
			r.Post("/appointments/{id}/payment", finHandler.RegisterPayment)
			r.With(managerRoleLimit).Put("/appointments/{id}/payment", finHandler.UpdatePayment)

			// Módulo de Controle de Estoque
			r.Get("/products", stockHandler.List)
			r.Get("/products/{id}", stockHandler.GetByID)
			r.With(managerRoleLimit).Post("/products", stockHandler.Create)
			r.With(managerRoleLimit).Put("/products/{id}", stockHandler.Update)
			r.With(managerRoleLimit).Delete("/products/{id}", stockHandler.Delete)
			r.With(managerRoleLimit).Get("/products/low-stock", stockHandler.GetLowStock)
			r.With(managerRoleLimit).Get("/products/{id}/movements", stockHandler.ListMovementsByProduct)
			r.With(managerRoleLimit).Post("/products/{id}/movements", stockHandler.CreateMovement)
			r.With(managerRoleLimit).Get("/stock/movements", stockHandler.ListAllMovements)
			r.With(managerRoleLimit).Get("/stock/report", stockHandler.StockReport)
			r.With(managerRoleLimit).Get("/stock/export", stockHandler.ExportCSV)

			// Vínculos de produtos consumidos em serviços
			r.With(managerRoleLimit).Get("/services/{id}/products", stockHandler.ListServiceProducts)
			r.With(managerRoleLimit).Post("/services/{id}/products", stockHandler.LinkServiceProduct)
			r.With(managerRoleLimit).Put("/services/{id}/products/{product_id}", stockHandler.UpdateServiceProductQty)
			r.With(managerRoleLimit).Delete("/services/{id}/products/{product_id}", stockHandler.UnlinkServiceProduct)

			// Módulo Admin Geral (exclusivo para platform_admin)
			r.Group(func(r chi.Router) {
				r.Use(adminRoleLimit)
				r.Get("/admin/clients", admHandler.ListClients)
				r.Post("/admin/clients", admHandler.CreateClient)
				r.Put("/admin/clients/{id}", admHandler.UpdateClient)
				r.Post("/admin/clients/{id}/block", admHandler.BlockClient)
				r.Post("/admin/clients/{id}/unblock", admHandler.UnblockClient)
				r.Get("/admin/clients/{id}/users", admHandler.ListClientUsers)
				r.Post("/admin/clients/{id}/users", admHandler.CreateClientUser)
				r.Post("/admin/impersonate/{client_id}", authHandler.Impersonate)
			})
		})
	})

	return router
}
