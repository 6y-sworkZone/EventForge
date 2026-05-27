package handlers

import (
	"fmt"
	"net/http"
	"time"

	"eventforge/models"
	"eventforge/utils"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct{}

func NewNotificationHandler() *NotificationHandler {
	return &NotificationHandler{}
}

func (h *NotificationHandler) SendNotification(c *gin.Context) {
	var req struct {
		EventID      *uint                 `json:"event_id"`
		Type         models.NotificationType `json:"type" binding:"required"`
		Subject      string                `json:"subject" binding:"required"`
		Content      string                `json:"content" binding:"required"`
		UserID       *uint                 `json:"user_id"`
		ScheduledAt  *time.Time            `json:"scheduled_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	notification := models.Notification{
		EventID:     req.EventID,
		UserID:      req.UserID,
		Type:        req.Type,
		Subject:     req.Subject,
		Content:     req.Content,
		Status:      models.NotificationStatusPending,
		ScheduledAt: req.ScheduledAt,
	}

	if req.ScheduledAt == nil {
		notification.Status = models.NotificationStatusSent
		now := time.Now()
		notification.SentAt = &now
	}

	if err := models.DB.Create(&notification).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create notification"})
		return
	}

	if notification.Status == models.NotificationStatusSent && req.UserID != nil {
		var user models.User
		if err := models.DB.First(&user, *req.UserID).Error; err == nil {
			utils.SendEmail(user.Email, req.Subject, req.Content)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"notification": notification})
}

func (h *NotificationHandler) ListNotifications(c *gin.Context) {
	eventID := c.Query("event_id")
	status := c.Query("status")
	notificationType := c.Query("type")

	var notifications []models.Notification
	query := models.DB.Preload("Event")

	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if notificationType != "" {
		query = query.Where("type = ?", notificationType)
	}

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func (h *NotificationHandler) GetNotification(c *gin.Context) {
	id := c.Param("id")

	var notification models.Notification
	if err := models.DB.Preload("Event").First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notification": notification})
}

func (h *NotificationHandler) UpdateNotification(c *gin.Context) {
	id := c.Param("id")

	var notification models.Notification
	if err := models.DB.First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	var req struct {
		Subject     string                `json:"subject"`
		Content     string                `json:"content"`
		ScheduledAt *time.Time            `json:"scheduled_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Subject != "" {
		updates["subject"] = req.Subject
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.ScheduledAt != nil {
		updates["scheduled_at"] = req.ScheduledAt
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&notification).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"notification": notification})
}

func (h *NotificationHandler) DeleteNotification(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.Notification{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete notification"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification deleted successfully"})
}

func (h *NotificationHandler) SendNow(c *gin.Context) {
	id := c.Param("id")

	var notification models.Notification
	if err := models.DB.First(&notification, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	var emails []string

	if notification.UserID != nil {
		var user models.User
		if err := models.DB.First(&user, *notification.UserID).Error; err == nil {
			emails = append(emails, user.Email)
		}
	} else if notification.EventID != nil {
		var registrations []models.Registration
		models.DB.Where("event_id = ? AND status IN ?", *notification.EventID,
			[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).
			Distinct("email").Find(&registrations)

		for _, r := range registrations {
			emails = append(emails, r.Email)
		}
	}

	for _, email := range emails {
		utils.SendEmail(email, notification.Subject, notification.Content)
	}

	now := time.Now()
	models.DB.Model(&notification).Updates(map[string]interface{}{
		"status":  models.NotificationStatusSent,
		"sent_at": now,
	})

	c.JSON(http.StatusOK, gin.H{"message": "Notification sent", "recipient_count": len(emails)})
}

func (h *NotificationHandler) BulkSend(c *gin.Context) {
	var req struct {
		EventID *uint  `json:"event_id"`
		Subject string `json:"subject" binding:"required"`
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var emails []string

	if req.EventID != nil {
		var registrations []models.Registration
		models.DB.Where("event_id = ? AND status IN ?", *req.EventID,
			[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).
			Distinct("email").Find(&registrations)

		for _, r := range registrations {
			emails = append(emails, r.Email)
		}
	} else {
		var users []models.User
		models.DB.Distinct("email").Find(&users)
		for _, u := range users {
			emails = append(emails, u.Email)
		}
	}

	for _, email := range emails {
		utils.SendEmail(email, req.Subject, req.Content)
	}

	notification := models.Notification{
		EventID: req.EventID,
		Type:    models.NotificationTypeCustom,
		Subject: req.Subject,
		Content: req.Content,
		Status:  models.NotificationStatusSent,
		SentAt:  func() *time.Time { t := time.Now(); return &t }(),
	}
	models.DB.Create(&notification)

	c.JSON(http.StatusOK, gin.H{"message": "Bulk email sent", "recipient_count": len(emails)})
}

func (h *NotificationHandler) CreateTemplate(c *gin.Context) {
	var req struct {
		Name      string `json:"name" binding:"required"`
		Type      string `json:"type"`
		Subject   string `json:"subject" binding:"required"`
		Content   string `json:"content" binding:"required"`
		Variables string `json:"variables"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := models.NotificationTemplate{
		Name:      req.Name,
		Type:      req.Type,
		Subject:   req.Subject,
		Content:   req.Content,
		Variables: req.Variables,
	}

	if err := models.DB.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"template": template})
}

func (h *NotificationHandler) ListTemplates(c *gin.Context) {
	var templates []models.NotificationTemplate
	if err := models.DB.Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *NotificationHandler) UpdateTemplate(c *gin.Context) {
	id := c.Param("id")

	var template models.NotificationTemplate
	if err := models.DB.First(&template, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
		return
	}

	var req struct {
		Name      string `json:"name"`
		Type      string `json:"type"`
		Subject   string `json:"subject"`
		Content   string `json:"content"`
		Variables string `json:"variables"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Subject != "" {
		updates["subject"] = req.Subject
	}
	if req.Content != "" {
		updates["content"] = req.Content
	}
	if req.Variables != "" {
		updates["variables"] = req.Variables
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&template).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"template": template})
}

func (h *NotificationHandler) DeleteTemplate(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.NotificationTemplate{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Template deleted successfully"})
}

func (h *NotificationHandler) GetSendHistory(c *gin.Context) {
	eventID := c.Query("event_id")
	limit := 50

	var notifications []models.Notification
	query := models.DB.Preload("Event").Where("status = ?", models.NotificationStatusSent)

	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}

	if err := query.Order("sent_at DESC").Limit(limit).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch history"})
		return
	}

	type HistoryItem struct {
		ID        uint   `json:"id"`
		Type      string `json:"type"`
		Subject   string `json:"subject"`
		EventID   *uint  `json:"event_id"`
		EventName string `json:"event_name"`
		SentAt    string `json:"sent_at"`
	}

	var history []HistoryItem
	for _, n := range notifications {
		eventName := ""
		if n.Event != nil {
			eventName = n.Event.Title
		}
		sentAt := ""
		if n.SentAt != nil {
			sentAt = n.SentAt.Format("2006-01-02 15:04:05")
		}

		history = append(history, HistoryItem{
			ID:        n.ID,
			Type:      string(n.Type),
			Subject:   n.Subject,
			EventID:   n.EventID,
			EventName: eventName,
			SentAt:    sentAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

func (h *NotificationHandler) TriggerEventUpdateNotification(c *gin.Context) {
	eventID := c.Param("event_id")

	var event models.Event
	if err := models.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var registrations []models.Registration
	models.DB.Where("event_id = ? AND status IN ?", event.ID,
		[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).
		Find(&registrations)

	subject := fmt.Sprintf("活动信息更新 - %s", event.Title)
	content := fmt.Sprintf("您好，\n\n您报名的活动「%s」信息有更新。\n\n活动时间：%s 至 %s\n活动地点：%s\n\n请查看最新活动信息。\n\n此致\nEventForge 团队",
		event.Title,
		event.StartTime.Format("2006-01-02 15:04"),
		event.EndTime.Format("2006-01-02 15:04"),
		event.Location,
	)

	for _, r := range registrations {
		utils.SendEmail(r.Email, subject, content)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notifications sent", "count": len(registrations)})
}
