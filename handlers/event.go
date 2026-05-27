package handlers

import (
	"fmt"
	"net/http"
	"time"

	"eventforge/models"
	"eventforge/utils"

	"github.com/gin-gonic/gin"
)

type EventHandler struct{}

func NewEventHandler() *EventHandler {
	return &EventHandler{}
}

func (h *EventHandler) List(c *gin.Context) {
	var events []models.Event
	query := models.DB.Where("is_template = ?", false)

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	eventType := c.Query("type")
	if eventType != "" {
		query = query.Where("type = ?", eventType)
	}

	if err := query.Order("created_at DESC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (h *EventHandler) ListPublic(c *gin.Context) {
	var events []models.Event
	query := models.DB.Where("is_template = ? AND is_public = ? AND status = ?", false, true, models.EventStatusOpen)

	eventType := c.Query("type")
	if eventType != "" {
		query = query.Where("type = ?", eventType)
	}

	if err := query.Order("start_time ASC").Find(&events).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch events"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"events": events})
}

func (h *EventHandler) Get(c *gin.Context) {
	id := c.Param("id")

	var event models.Event
	if err := models.DB.Preload("Venue").Preload("Tickets").First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var regCount int64
	models.DB.Model(&models.Registration{}).Where("event_id = ? AND status IN ?", event.ID,
		[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).Count(&regCount)

	c.JSON(http.StatusOK, gin.H{
		"event":           event,
		"registered_count": regCount,
	})
}

func (h *EventHandler) Create(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Title        string          `json:"title" binding:"required"`
		Description  string          `json:"description"`
		CoverImage   string          `json:"cover_image"`
		Type         models.EventType `json:"type" binding:"required"`
		StartTime    time.Time       `json:"start_time" binding:"required"`
		EndTime      time.Time       `json:"end_time" binding:"required"`
		Location     string          `json:"location"`
		VenueID      *uint           `json:"venue_id"`
		MaxCapacity  int             `json:"max_capacity"`
		IsPublic     bool            `json:"is_public"`
		IsPaid       bool            `json:"is_paid"`
		TemplateID   *uint           `json:"template_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	event := models.Event{
		Title:       req.Title,
		Description: req.Description,
		CoverImage:  req.CoverImage,
		Type:        req.Type,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
		VenueID:     req.VenueID,
		MaxCapacity: req.MaxCapacity,
		IsPublic:    req.IsPublic,
		IsPaid:      req.IsPaid,
		Status:      models.EventStatusDraft,
		CreatedBy:   userID,
		TemplateID:  req.TemplateID,
	}

	if err := models.DB.Create(&event).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"event": event})
}

func (h *EventHandler) Update(c *gin.Context) {
	id := c.Param("id")

	var event models.Event
	if err := models.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var req struct {
		Title       string          `json:"title"`
		Description string          `json:"description"`
		CoverImage  string          `json:"cover_image"`
		Type        models.EventType `json:"type"`
		StartTime   *time.Time      `json:"start_time"`
		EndTime     *time.Time      `json:"end_time"`
		Location    string          `json:"location"`
		VenueID     *uint           `json:"venue_id"`
		MaxCapacity int             `json:"max_capacity"`
		IsPublic    *bool           `json:"is_public"`
		IsPaid      *bool           `json:"is_paid"`
		Status      models.EventStatus `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.CoverImage != "" {
		updates["cover_image"] = req.CoverImage
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.StartTime != nil {
		updates["start_time"] = req.StartTime
	}
	if req.EndTime != nil {
		updates["end_time"] = req.EndTime
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.VenueID != nil {
		updates["venue_id"] = req.VenueID
	}
	if req.MaxCapacity != 0 {
		updates["max_capacity"] = req.MaxCapacity
	}
	if req.IsPublic != nil {
		updates["is_public"] = req.IsPublic
	}
	if req.IsPaid != nil {
		updates["is_paid"] = req.IsPaid
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&event).Updates(updates)

		models.DB.First(&event, id)

		timeChanged := req.StartTime != nil || req.EndTime != nil
		locationChanged := req.Location != "" || (req.VenueID != nil)
		if timeChanged || locationChanged {
			var registrations []models.Registration
			models.DB.Where("event_id = ? AND status IN ?", event.ID,
				[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).
				Distinct("email").Find(&registrations)

			changeDetails := ""
			if timeChanged {
				changeDetails += fmt.Sprintf("活动时间：%s 至 %s\n", event.StartTime.Format("2006-01-02 15:04"), event.EndTime.Format("2006-01-02 15:04"))
			}
			if locationChanged {
				changeDetails += fmt.Sprintf("活动地点：%s\n", event.Location)
			}

			now := time.Now()
			subject := fmt.Sprintf("活动信息更新 - %s", event.Title)
			content := fmt.Sprintf("您好，\n\n您报名的活动「%s」信息有更新：\n\n%s\n\n请查看最新活动信息，如有疑问请联系我们。\n\n此致\nEventForge 团队", event.Title, changeDetails)

			for _, r := range registrations {
				utils.SendEmail(r.Email, subject, content)
			}

			models.DB.Create(&models.Notification{
				EventID: &event.ID,
				Type:    models.NotificationTypeEventUpdate,
				Subject: subject,
				Content: content,
				Status:  models.NotificationStatusSent,
				SentAt:  &now,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"event": event})
}

func (h *EventHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.Event{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete event"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Event deleted successfully"})
}

func (h *EventHandler) Clone(c *gin.Context) {
	id := c.Param("id")
	userID := c.GetUint("user_id")

	var event models.Event
	if err := models.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	cloned := models.Event{
		Title:        event.Title + " (副本)",
		Description:  event.Description,
		CoverImage:   event.CoverImage,
		Type:         event.Type,
		StartTime:    event.StartTime,
		EndTime:      event.EndTime,
		Location:     event.Location,
		VenueID:      event.VenueID,
		MaxCapacity:  event.MaxCapacity,
		IsPublic:     event.IsPublic,
		IsPaid:       event.IsPaid,
		Status:       models.EventStatusDraft,
		CreatedBy:    userID,
		ClonedFromID: &event.ID,
	}

	if err := models.DB.Create(&cloned).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clone event"})
		return
	}

	var tickets []models.Ticket
	models.DB.Where("event_id = ?", event.ID).Find(&tickets)
	for _, t := range tickets {
		newTicket := models.Ticket{
			EventID:      cloned.ID,
			Name:         t.Name,
			Type:         t.Type,
			Price:        t.Price,
			Quantity:     t.Quantity,
			Description:  t.Description,
			SaleStartAt:  t.SaleStartAt,
			SaleEndAt:    t.SaleEndAt,
			EarlyPrice:   t.EarlyPrice,
			EarlyStartAt: t.EarlyStartAt,
			EarlyEndAt:   t.EarlyEndAt,
		}
		models.DB.Create(&newTicket)
	}

	var customFields []models.CustomField
	models.DB.Where("event_id = ?", event.ID).Find(&customFields)
	for _, cf := range customFields {
		newCF := models.CustomField{
			EventID:   cloned.ID,
			Label:     cf.Label,
			FieldType: cf.FieldType,
			Required:  cf.Required,
			Options:   cf.Options,
		}
		models.DB.Create(&newCF)
	}

	c.JSON(http.StatusCreated, gin.H{"event": cloned})
}

func (h *EventHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Status models.EventStatus `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := models.DB.Model(&models.Event{}).Where("id = ?", id).Update("status", req.Status)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully"})
}

func (h *EventHandler) SaveAsTemplate(c *gin.Context) {
	id := c.Param("id")

	var event models.Event
	if err := models.DB.First(&event, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var req struct {
		TemplateName string `json:"template_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	template := models.Event{
		Title:        req.TemplateName,
		Description:  event.Description,
		CoverImage:   event.CoverImage,
		Type:         event.Type,
		StartTime:    event.StartTime,
		EndTime:      event.EndTime,
		Location:     event.Location,
		VenueID:      event.VenueID,
		MaxCapacity:  event.MaxCapacity,
		IsPublic:     event.IsPublic,
		IsPaid:       event.IsPaid,
		Status:       models.EventStatusDraft,
		CreatedBy:    event.CreatedBy,
		IsTemplate:   true,
		TemplateName: req.TemplateName,
	}

	if err := models.DB.Create(&template).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save template"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"template": template})
}

func (h *EventHandler) ListTemplates(c *gin.Context) {
	var templates []models.Event
	if err := models.DB.Where("is_template = ?", true).Order("created_at DESC").Find(&templates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"templates": templates})
}

func (h *EventHandler) GetCustomFields(c *gin.Context) {
	eventID := c.Param("id")

	var fields []models.CustomField
	if err := models.DB.Where("event_id = ?", eventID).Find(&fields).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch custom fields"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"custom_fields": fields})
}

func (h *EventHandler) AddCustomField(c *gin.Context) {
	eventID := c.Param("id")

	var req struct {
		Label     string `json:"label" binding:"required"`
		FieldType string `json:"field_type" binding:"required"`
		Required  bool   `json:"required"`
		Options   string `json:"options"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	field := models.CustomField{
		EventID:   parseUint(eventID),
		Label:     req.Label,
		FieldType: req.FieldType,
		Required:  req.Required,
		Options:   req.Options,
	}

	if err := models.DB.Create(&field).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create custom field"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"custom_field": field})
}

func (h *EventHandler) DeleteCustomField(c *gin.Context) {
	eventID := c.Param("id")
	fieldID := c.Param("field_id")

	result := models.DB.Where("id = ? AND event_id = ?", fieldID, eventID).Delete(&models.CustomField{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete custom field"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Custom field deleted successfully"})
}

func parseUint(s string) uint {
	var n uint
	for _, c := range s {
		if c >= '0' && c <= '9' {
			n = n*10 + uint(c-'0')
		}
	}
	return n
}
