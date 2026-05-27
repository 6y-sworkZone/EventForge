package handlers

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"eventforge/models"
	"eventforge/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type RegistrationHandler struct{}

func NewRegistrationHandler() *RegistrationHandler {
	return &RegistrationHandler{}
}

type EventRegistrationRequest struct {
	EventID        uint   `json:"event_id" binding:"required"`
	TicketID       *uint  `json:"ticket_id"`
	Name           string `json:"name" binding:"required"`
	Email          string `json:"email" binding:"required,email"`
	Phone          string `json:"phone"`
	Company        string `json:"company"`
	Position       string `json:"position"`
	DietPreference string `json:"diet_preference"`
	CustomFields   string `json:"custom_fields"`
	PromoCode      string `json:"promo_code"`
}

func (h *RegistrationHandler) Register(c *gin.Context) {
	var req EventRegistrationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var event models.Event
	if err := models.DB.First(&event, req.EventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	if event.Status != models.EventStatusOpen {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Registration is not open for this event"})
		return
	}

	var existing models.Registration
	if models.DB.Where("event_id = ? AND email = ? AND status != ?", req.EventID, req.Email, models.RegistrationStatusCancelled).First(&existing).Error == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Already registered for this event"})
		return
	}

	var regCount int64
	models.DB.Model(&models.Registration{}).Where("event_id = ? AND status IN ?", req.EventID,
		[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).Count(&regCount)

	status := models.RegistrationStatusConfirmed
	waitlistOrder := 0

	if event.MaxCapacity > 0 && int(regCount) >= event.MaxCapacity {
		status = models.RegistrationStatusWaitlist
		var maxWaitlistOrder int
		models.DB.Model(&models.Registration{}).Where("event_id = ? AND status = ?", req.EventID, models.RegistrationStatusWaitlist).Select("COALESCE(MAX(waitlist_order), 0)").Scan(&maxWaitlistOrder)
		waitlistOrder = maxWaitlistOrder + 1
	}

	regID := uuid.New().String()
	qrPath, _ := utils.SaveQRCode(regID, regID+".png")

	registration := models.Registration{
		EventID:        req.EventID,
		TicketID:       req.TicketID,
		Name:           req.Name,
		Email:          req.Email,
		Phone:          req.Phone,
		Company:        req.Company,
		Position:       req.Position,
		DietPreference: req.DietPreference,
		CustomFields:   req.CustomFields,
		Status:         status,
		QRCode:         qrPath,
		WaitlistOrder:  waitlistOrder,
	}

	if req.TicketID != nil && event.IsPaid {
		var ticket models.Ticket
		if err := models.DB.First(&ticket, req.TicketID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
			return
		}

		if ticket.SoldCount >= ticket.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Ticket sold out"})
			return
		}

		orderNumber := fmt.Sprintf("ORD-%d-%s", time.Now().Unix(), uuid.New().String()[:8])

		discount := 0.0
		var promoCodeID *uint

		if req.PromoCode != "" {
			var promo models.PromoCode
			if err := models.DB.Where("code = ? AND event_id = ? AND is_active = ?", req.PromoCode, req.EventID, true).First(&promo).Error; err == nil {
				if promo.ExpiresAt == nil || promo.ExpiresAt.After(time.Now()) {
					if promo.MaxUsage == 0 || promo.UsedCount < promo.MaxUsage {
						if promo.Type == models.PromoCodeTypePercent {
							discount = ticket.Price * promo.Value / 100
						} else {
							discount = promo.Value
						}
						if discount > ticket.Price {
							discount = ticket.Price
						}
						promoCodeID = &promo.ID
						models.DB.Model(&promo).Update("used_count", promo.UsedCount+1)
					}
				}
			}
		}

		totalAmount := ticket.Price - discount

		order := models.Order{
			EventID:       req.EventID,
			TicketID:      *req.TicketID,
			OrderNumber:   orderNumber,
			Quantity:      1,
			UnitPrice:     ticket.Price,
			Discount:      discount,
			TotalAmount:   totalAmount,
			Status:        models.OrderStatusPending,
			PromoCodeID:   promoCodeID,
		}

		if totalAmount == 0 {
			order.Status = models.OrderStatusPaid
			order.PaymentMethod = "free"
			now := time.Now()
			order.PaidAt = &now
			models.DB.Model(&ticket).Update("sold_count", ticket.SoldCount+1)
		}

		models.DB.Create(&order)
		registration.OrderID = &order.ID
	}

	if err := models.DB.Create(&registration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create registration"})
		return
	}

	if status == models.RegistrationStatusConfirmed {
		icsContent := utils.GenerateICS(event.Title, event.StartTime, event.EndTime, event.Location, event.Description)
		icsPath := fmt.Sprintf("./uploads/ics/%s.ics", regID)
		models.DB.Create(&models.Notification{
			EventID: &event.ID,
			Type:    models.NotificationTypeRegistrationSuccess,
			Subject: fmt.Sprintf("活动报名成功 - %s", event.Title),
			Content: fmt.Sprintf("您好 %s，\n\n您已成功报名参加「%s」活动。\n\n活动时间：%s\n活动地点：%s\n\n报名编号：%s\n\n此致\nEventForge 团队",
				req.Name, event.Title, event.StartTime.Format("2006-01-02 15:04"), event.Location, regID),
			Status:  models.NotificationStatusPending,
		})

		_ = icsContent
		_ = icsPath
	}

	c.JSON(http.StatusCreated, gin.H{
		"registration": registration,
		"waitlist":     status == models.RegistrationStatusWaitlist,
	})
}

func (h *RegistrationHandler) List(c *gin.Context) {
	eventID := c.Query("event_id")
	status := c.Query("status")

	var registrations []models.Registration
	query := models.DB.Preload("Event").Preload("Ticket")

	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&registrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch registrations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"registrations": registrations})
}

func (h *RegistrationHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var registration models.Registration
	if err := models.DB.Preload("Event").Preload("Ticket").Preload("Order").First(&registration, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"registration": registration})
}

func (h *RegistrationHandler) Cancel(c *gin.Context) {
	id := c.Param("id")

	var registration models.Registration
	if err := models.DB.First(&registration, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
		return
	}

	wasConfirmed := registration.Status == models.RegistrationStatusConfirmed

	registration.Status = models.RegistrationStatusCancelled
	models.DB.Save(&registration)

	if registration.OrderID != nil {
		models.DB.Model(&models.Order{}).Where("id = ?", registration.OrderID).Update("status", models.OrderStatusCancelled)
	}

	if wasConfirmed && registration.EventID != 0 {
		var waitlistReg models.Registration
		result := models.DB.Where("event_id = ? AND status = ?", registration.EventID, models.RegistrationStatusWaitlist).
			Order("waitlist_order ASC").First(&waitlistReg)

		if result.Error == nil {
			waitlistReg.Status = models.RegistrationStatusConfirmed
			models.DB.Save(&waitlistReg)

			models.DB.Model(&models.Notification{}).Create(&models.Notification{
				EventID: &registration.EventID,
				Type:    models.NotificationTypeRegistrationSuccess,
				Subject: "候补递补成功通知",
				Content: fmt.Sprintf("您好 %s，\n\n由于有其他报名者取消，您已从候补名单递补成功，正式获得活动参加资格！\n\n此致\nEventForge 团队", waitlistReg.Name),
				Status:  models.NotificationStatusPending,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Registration cancelled successfully"})
}

func (h *RegistrationHandler) CheckIn(c *gin.Context) {
	id := c.Param("id")

	var registration models.Registration
	if err := models.DB.First(&registration, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Registration not found"})
		return
	}

	if registration.Status != models.RegistrationStatusConfirmed && registration.Status != models.RegistrationStatusCheckedIn {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot check in: invalid registration status"})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.RegistrationStatusCheckedIn,
		"checked_in_at": now,
	}

	models.DB.Model(&registration).Updates(updates)

	c.JSON(http.StatusOK, gin.H{"message": "Check-in successful"})
}

func (h *RegistrationHandler) CheckInByQR(c *gin.Context) {
	var req struct {
		QRCode string `json:"qr_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var registration models.Registration
	if err := models.DB.Where("qr_code LIKE ?", "%"+req.QRCode+"%").First(&registration).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid QR code"})
		return
	}

	if registration.Status != models.RegistrationStatusConfirmed && registration.Status != models.RegistrationStatusCheckedIn {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot check in: invalid registration status"})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":        models.RegistrationStatusCheckedIn,
		"checked_in_at": now,
	}

	models.DB.Model(&registration).Updates(updates)

	c.JSON(http.StatusOK, gin.H{
		"message":      "Check-in successful",
		"registration": gin.H{"id": registration.ID, "name": registration.Name, "email": registration.Email},
	})
}

func (h *RegistrationHandler) ExportCSV(c *gin.Context) {
	eventID := c.Query("event_id")

	var registrations []models.Registration
	query := models.DB.Where("status IN ?", []models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn})

	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}

	query.Find(&registrations)

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=registrations_%s.csv", time.Now().Format("20060102")))

	w := csv.NewWriter(c.Writer)
	w.Write([]string{"ID", "活动ID", "姓名", "邮箱", "手机", "公司", "职位", "饮食偏好", "状态", "签到时间", "报名时间"})

	for _, reg := range registrations {
		checkedInAt := ""
		if reg.CheckedInAt != nil {
			checkedInAt = reg.CheckedInAt.Format("2006-01-02 15:04:05")
		}

		w.Write([]string{
			strconv.Itoa(int(reg.ID)),
			strconv.Itoa(int(reg.EventID)),
			reg.Name,
			reg.Email,
			reg.Phone,
			reg.Company,
			reg.Position,
			reg.DietPreference,
			string(reg.Status),
			checkedInAt,
			reg.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}

	w.Flush()
}

func (h *RegistrationHandler) GetMyRegistrations(c *gin.Context) {
	userID := c.GetUint("user_id")

	var user models.User
	if err := models.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	var registrations []models.Registration
	if err := models.DB.Preload("Event").Where("email = ?", user.Email).Order("created_at DESC").Find(&registrations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch registrations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"registrations": registrations})
}
