package main

import (
	"fmt"
	"os"

	"eventforge/config"
	"eventforge/handlers"
	"eventforge/middleware"
	"eventforge/models"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	cfg := config.Load()

	os.MkdirAll(cfg.UploadPath+"/qrcodes", 0755)
	os.MkdirAll(cfg.UploadPath+"/ics", 0755)
	os.MkdirAll(cfg.UploadPath+"/avatars", 0755)

	models.InitDB(cfg.DBPath)

	middleware.StartReminderScheduler()

	authHandler := handlers.NewAuthHandler()
	eventHandler := handlers.NewEventHandler()
	registrationHandler := handlers.NewRegistrationHandler()
	ticketHandler := handlers.NewTicketHandler()
	scheduleHandler := handlers.NewScheduleHandler()
	venueHandler := handlers.NewVenueHandler()
	notificationHandler := handlers.NewNotificationHandler()
	dashboardHandler := handlers.NewDashboardHandler()

	r := gin.Default()

	r.Use(middleware.CORSMiddleware())
	r.Static("/uploads", "./uploads")

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.GET("/profile", middleware.AuthMiddleware(), authHandler.GetProfile)
			auth.PUT("/profile", middleware.AuthMiddleware(), authHandler.UpdateProfile)
		}

		public := api.Group("/public")
		{
			public.GET("/events", eventHandler.ListPublic)
			public.POST("/register", registrationHandler.Register)
			public.GET("/events/:id", func(c *gin.Context) {
				id := c.Param("id")
				var event models.Event
				if err := models.DB.Preload("Venue").Preload("Tickets").First(&event, id).Error; err != nil {
					c.JSON(404, gin.H{"error": "Event not found"})
					return
				}
				var regCount int64
				models.DB.Model(&models.Registration{}).Where("event_id = ? AND status IN ?", event.ID,
					[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).Count(&regCount)
				c.JSON(200, gin.H{"event": event, "registered_count": regCount})
			})
		}

		events := api.Group("/events")
		events.Use(middleware.AuthMiddleware())
		{
			events.GET("", eventHandler.List)
			events.POST("", eventHandler.Create)
			events.GET("/templates", eventHandler.ListTemplates)
			events.POST("/templates", eventHandler.SaveAsTemplate)
			events.GET("/:id", eventHandler.Get)
			events.PUT("/:id", eventHandler.Update)
			events.DELETE("/:id", eventHandler.Delete)
			events.POST("/:id/clone", eventHandler.Clone)
			events.PUT("/:id/status", eventHandler.UpdateStatus)
			events.POST("/:id/save-template", eventHandler.SaveAsTemplate)

			events.GET("/:id/custom-fields", eventHandler.GetCustomFields)
			events.POST("/:id/custom-fields", eventHandler.AddCustomField)
			events.DELETE("/:id/custom-fields/:field_id", eventHandler.DeleteCustomField)

			events.GET("/:id/tickets", ticketHandler.ListTickets)
			events.POST("/:id/tickets", ticketHandler.CreateTicket)
			events.GET("/:id/tickets/stats", ticketHandler.GetTicketStats)
			events.PUT("/tickets/:id", ticketHandler.UpdateTicket)
			events.DELETE("/tickets/:id", ticketHandler.DeleteTicket)

			events.GET("/:id/promo-codes", ticketHandler.ListPromoCodes)
			events.POST("/:id/promo-codes", ticketHandler.CreatePromoCode)
			events.PUT("/promo-codes/:id", ticketHandler.UpdatePromoCode)
			events.DELETE("/promo-codes/:id", ticketHandler.DeletePromoCode)

			events.GET("/:id/schedules", scheduleHandler.ListSchedules)
			events.POST("/:id/schedules", scheduleHandler.CreateSchedule)
			events.GET("/:id/schedules/timeline", scheduleHandler.GetTimelineView)
			events.GET("/:id/schedules/export", scheduleHandler.ExportPDF)
			events.PUT("/schedules/:id", scheduleHandler.UpdateSchedule)
			events.DELETE("/schedules/:id", scheduleHandler.DeleteSchedule)

			events.POST("/schedules/:schedule_id/agenda-items", scheduleHandler.CreateAgendaItem)
			events.PUT("/agenda-items/:id", scheduleHandler.UpdateAgendaItem)
			events.DELETE("/agenda-items/:id", scheduleHandler.DeleteAgendaItem)

			events.GET("/:id/speakers", scheduleHandler.ListSpeakers)
			events.POST("/:id/speakers", scheduleHandler.CreateSpeaker)
			events.PUT("/speakers/:id", scheduleHandler.UpdateSpeaker)
			events.DELETE("/speakers/:id", scheduleHandler.DeleteSpeaker)

			events.GET("/:id/notifications/trigger-update", notificationHandler.TriggerEventUpdateNotification)
		}

		registrations := api.Group("/registrations")
		registrations.Use(middleware.AuthMiddleware())
		{
			registrations.GET("", registrationHandler.List)
			registrations.GET("/mine", registrationHandler.GetMyRegistrations)
			registrations.GET("/:id", registrationHandler.Get)
			registrations.PUT("/:id/cancel", registrationHandler.Cancel)
			registrations.PUT("/:id/check-in", registrationHandler.CheckIn)
			registrations.POST("/check-in-qr", registrationHandler.CheckInByQR)
			registrations.GET("/export/csv", registrationHandler.ExportCSV)
		}

		orders := api.Group("/orders")
		orders.Use(middleware.AuthMiddleware())
		{
			orders.GET("", ticketHandler.ListOrders)
			orders.GET("/:id", ticketHandler.GetOrder)
			orders.PUT("/:id/mark-paid", ticketHandler.MarkAsPaid)
			orders.PUT("/:id/refund", ticketHandler.RefundOrder)
			orders.PUT("/:id/invoice", ticketHandler.UpdateInvoiceInfo)
		}

		payments := api.Group("/payments")
		payments.Use(middleware.AuthMiddleware())
		{
			payments.GET("", ticketHandler.ListPayments)
		}

		venues := api.Group("/venues")
		venues.Use(middleware.AuthMiddleware())
		{
			venues.GET("", venueHandler.ListVenues)
			venues.POST("", venueHandler.CreateVenue)

			venues.GET("/all/rooms", func(c *gin.Context) {
				var rooms []models.VenueRoom
				models.DB.Find(&rooms)
				c.JSON(200, gin.H{"rooms": rooms})
			})

			venues.GET("/:id", venueHandler.GetVenue)
			venues.PUT("/:id", venueHandler.UpdateVenue)
			venues.DELETE("/:id", venueHandler.DeleteVenue)
			venues.GET("/:id/calendar", venueHandler.GetVenueCalendar)
			venues.GET("/:id/rooms", venueHandler.ListRooms)
			venues.POST("/:id/rooms", venueHandler.CreateRoom)

			venues.PUT("/rooms/:id", venueHandler.UpdateRoom)
			venues.DELETE("/rooms/:id", venueHandler.DeleteRoom)

			venues.GET("/rooms/:room_id/seats", venueHandler.ListSeats)
			venues.POST("/rooms/:room_id/seats", venueHandler.CreateSeat)
			venues.PUT("/seats/:id", venueHandler.UpdateSeat)
			venues.DELETE("/seats/:id", venueHandler.DeleteSeat)
		}

		notifications := api.Group("/notifications")
		notifications.Use(middleware.AuthMiddleware())
		{
			notifications.GET("", notificationHandler.ListNotifications)
			notifications.POST("", notificationHandler.SendNotification)
			notifications.GET("/:id", notificationHandler.GetNotification)
			notifications.PUT("/:id", notificationHandler.UpdateNotification)
			notifications.DELETE("/:id", notificationHandler.DeleteNotification)
			notifications.POST("/:id/send", notificationHandler.SendNow)
			notifications.POST("/bulk-send", notificationHandler.BulkSend)
			notifications.GET("/history/sent", notificationHandler.GetSendHistory)

			notifications.GET("/templates/list", notificationHandler.ListTemplates)
			notifications.POST("/templates", notificationHandler.CreateTemplate)
			notifications.PUT("/templates/:id", notificationHandler.UpdateTemplate)
			notifications.DELETE("/templates/:id", notificationHandler.DeleteTemplate)
		}

		dashboard := api.Group("/dashboard")
		dashboard.Use(middleware.AuthMiddleware())
		{
			dashboard.GET("/overview", dashboardHandler.GetOverview)
			dashboard.GET("/registration-trend", dashboardHandler.GetRegistrationTrend)
			dashboard.GET("/events/:id/ticket-sales", dashboardHandler.GetTicketSales)
			dashboard.GET("/events/:id/check-in-rate", dashboardHandler.GetCheckInRate)
			dashboard.GET("/events/:id/feedback", dashboardHandler.GetEventFeedback)
			dashboard.POST("/events/:id/feedback", dashboardHandler.SubmitFeedback)
			dashboard.GET("/events/:id/participant-profile", dashboardHandler.GetParticipantProfile)
			dashboard.GET("/events/:id/stats", dashboardHandler.GetEventStats)
			dashboard.GET("/historical-comparison", dashboardHandler.GetHistoricalComparison)
		}

		admin := api.Group("/admin")
		admin.Use(middleware.AuthMiddleware(), middleware.AdminMiddleware())
		{
			admin.GET("/users", func(c *gin.Context) {
				var users []models.User
				models.DB.Find(&users)
				c.JSON(200, gin.H{"users": users})
			})
		}
	}

	addr := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Server starting on port %d...\n", cfg.Port)
	r.Run(addr)
}
