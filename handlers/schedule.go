package handlers

import (
	"fmt"
	"net/http"
	"time"

	"eventforge/models"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
)

type ScheduleHandler struct{}

func NewScheduleHandler() *ScheduleHandler {
	return &ScheduleHandler{}
}

func (h *ScheduleHandler) CreateSchedule(c *gin.Context) {
	eventID := c.Param("event_id")

	var req struct {
		Name      string    `json:"name" binding:"required"`
		Date      time.Time `json:"date" binding:"required"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	schedule := models.Schedule{
		EventID:   parseUint(eventID),
		Name:      req.Name,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
	}

	if err := models.DB.Create(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create schedule"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"schedule": schedule})
}

func (h *ScheduleHandler) ListSchedules(c *gin.Context) {
	eventID := c.Param("event_id")

	var schedules []models.Schedule
	if err := models.DB.Preload("AgendaItems.Speaker").Preload("AgendaItems.VenueRoom").
		Where("event_id = ?", eventID).Order("date ASC").Find(&schedules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch schedules"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schedules": schedules})
}

func (h *ScheduleHandler) UpdateSchedule(c *gin.Context) {
	id := c.Param("id")

	var schedule models.Schedule
	if err := models.DB.First(&schedule, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	var req struct {
		Name      string    `json:"name"`
		Date      time.Time `json:"date"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if !req.Date.IsZero() {
		updates["date"] = req.Date
	}
	if !req.StartTime.IsZero() {
		updates["start_time"] = req.StartTime
	}
	if !req.EndTime.IsZero() {
		updates["end_time"] = req.EndTime
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&schedule).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"schedule": schedule})
}

func (h *ScheduleHandler) DeleteSchedule(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.Schedule{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete schedule"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Schedule deleted successfully"})
}

func (h *ScheduleHandler) CreateAgendaItem(c *gin.Context) {
	scheduleID := c.Param("schedule_id")

	var req struct {
		Title       string    `json:"title" binding:"required"`
		Description string    `json:"description"`
		SpeakerID   *uint     `json:"speaker_id"`
		VenueRoomID *uint     `json:"venue_room_id"`
		StartTime   time.Time `json:"start_time"`
		EndTime     time.Time `json:"end_time"`
		Location    string    `json:"location"`
		SortOrder   int       `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var schedule models.Schedule
	if err := models.DB.First(&schedule, parseUint(scheduleID)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Schedule not found"})
		return
	}

	if req.VenueRoomID != nil && !req.StartTime.IsZero() && !req.EndTime.IsZero() {
		var conflicts []models.AgendaItem
		models.DB.Joins("JOIN schedules ON schedules.id = agenda_items.schedule_id").
			Where("schedules.event_id = ? AND agenda_items.venue_room_id = ? AND agenda_items.id != ? AND agenda_items.start_time < ? AND agenda_items.end_time > ?",
				schedule.EventID, *req.VenueRoomID, 0, req.EndTime, req.StartTime).
			Find(&conflicts)

		if len(conflicts) > 0 {
			conflictInfo := make([]map[string]interface{}, 0)
			for _, c := range conflicts {
				conflictInfo = append(conflictInfo, map[string]interface{}{
					"id":         c.ID,
					"title":      c.Title,
					"start_time": c.StartTime,
					"end_time":   c.EndTime,
				})
			}
			c.JSON(http.StatusConflict, gin.H{
				"error":      "Venue room conflict detected",
				"conflicts":  conflictInfo,
				"message":    "The selected venue room is already booked during this time period",
			})
			return
		}
	}

	item := models.AgendaItem{
		ScheduleID:  parseUint(scheduleID),
		Title:       req.Title,
		Description: req.Description,
		SpeakerID:   req.SpeakerID,
		VenueRoomID: req.VenueRoomID,
		StartTime:   req.StartTime,
		EndTime:     req.EndTime,
		Location:    req.Location,
		SortOrder:   req.SortOrder,
	}

	if err := models.DB.Create(&item).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create agenda item"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"agenda_item": item})
}

func (h *ScheduleHandler) UpdateAgendaItem(c *gin.Context) {
	id := c.Param("id")

	var item models.AgendaItem
	if err := models.DB.First(&item, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Agenda item not found"})
		return
	}

	var req struct {
		Title       string    `json:"title"`
		Description string    `json:"description"`
		SpeakerID   *uint     `json:"speaker_id"`
		VenueRoomID *uint     `json:"venue_room_id"`
		StartTime   time.Time `json:"start_time"`
		EndTime     time.Time `json:"end_time"`
		Location    string    `json:"location"`
		SortOrder   int       `json:"sort_order"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	venueRoomID := item.VenueRoomID
	startTime := item.StartTime
	endTime := item.EndTime

	if req.VenueRoomID != nil {
		venueRoomID = req.VenueRoomID
	}
	if !req.StartTime.IsZero() {
		startTime = req.StartTime
	}
	if !req.EndTime.IsZero() {
		endTime = req.EndTime
	}

	if venueRoomID != nil {
		var schedule models.Schedule
		models.DB.First(&schedule, item.ScheduleID)

		var conflicts []models.AgendaItem
		models.DB.Joins("JOIN schedules ON schedules.id = agenda_items.schedule_id").
			Where("schedules.event_id = ? AND agenda_items.venue_room_id = ? AND agenda_items.id != ? AND agenda_items.start_time < ? AND agenda_items.end_time > ?",
				schedule.EventID, *venueRoomID, item.ID, endTime, startTime).
			Find(&conflicts)

		if len(conflicts) > 0 {
			conflictInfo := make([]map[string]interface{}, 0)
			for _, c := range conflicts {
				conflictInfo = append(conflictInfo, map[string]interface{}{
					"id":         c.ID,
					"title":      c.Title,
					"start_time": c.StartTime,
					"end_time":   c.EndTime,
				})
			}
			c.JSON(http.StatusConflict, gin.H{
				"error":     "Venue room conflict detected",
				"conflicts": conflictInfo,
				"message":   "The selected venue room is already booked during this time period",
			})
			return
		}
	}

	updates := map[string]interface{}{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.SpeakerID != nil {
		updates["speaker_id"] = req.SpeakerID
	}
	if req.VenueRoomID != nil {
		updates["venue_room_id"] = req.VenueRoomID
	}
	if !req.StartTime.IsZero() {
		updates["start_time"] = req.StartTime
	}
	if !req.EndTime.IsZero() {
		updates["end_time"] = req.EndTime
	}
	if req.Location != "" {
		updates["location"] = req.Location
	}
	if req.SortOrder != 0 {
		updates["sort_order"] = req.SortOrder
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&item).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"agenda_item": item})
}

func (h *ScheduleHandler) DeleteAgendaItem(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.AgendaItem{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete agenda item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Agenda item deleted successfully"})
}

func (h *ScheduleHandler) CreateSpeaker(c *gin.Context) {
	eventID := c.Param("event_id")

	var req struct {
		Name    string `json:"name" binding:"required"`
		Avatar  string `json:"avatar"`
		Title   string `json:"title"`
		Company string `json:"company"`
		Bio     string `json:"bio"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	speaker := models.Speaker{
		EventID: parseUint(eventID),
		Name:    req.Name,
		Avatar:  req.Avatar,
		Title:   req.Title,
		Company: req.Company,
		Bio:     req.Bio,
	}

	if err := models.DB.Create(&speaker).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create speaker"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"speaker": speaker})
}

func (h *ScheduleHandler) ListSpeakers(c *gin.Context) {
	eventID := c.Param("event_id")

	var speakers []models.Speaker
	if err := models.DB.Where("event_id = ?", eventID).Find(&speakers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch speakers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"speakers": speakers})
}

func (h *ScheduleHandler) UpdateSpeaker(c *gin.Context) {
	id := c.Param("id")

	var speaker models.Speaker
	if err := models.DB.First(&speaker, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Speaker not found"})
		return
	}

	var req struct {
		Name    string `json:"name"`
		Avatar  string `json:"avatar"`
		Title   string `json:"title"`
		Company string `json:"company"`
		Bio     string `json:"bio"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Avatar != "" {
		updates["avatar"] = req.Avatar
	}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Company != "" {
		updates["company"] = req.Company
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&speaker).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"speaker": speaker})
}

func (h *ScheduleHandler) DeleteSpeaker(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.Speaker{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete speaker"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Speaker deleted successfully"})
}

func (h *ScheduleHandler) GetTimelineView(c *gin.Context) {
	eventID := c.Param("event_id")

	var schedules []models.Schedule
	if err := models.DB.Preload("AgendaItems.Speaker").Preload("AgendaItems.VenueRoom").
		Where("event_id = ?", eventID).Order("date ASC").Find(&schedules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch timeline"})
		return
	}

	type TimelineItem struct {
		ID       uint   `json:"id"`
		Time     string `json:"time"`
		Title    string `json:"title"`
		Speaker  string `json:"speaker"`
		Location string `json:"location"`
		Duration string `json:"duration"`
	}

	type TimelineDay struct {
		Date  string         `json:"date"`
		Name  string         `json:"name"`
		Items []TimelineItem `json:"items"`
	}

	var timeline []TimelineDay

	for _, s := range schedules {
		day := TimelineDay{
			Date: s.Date.Format("2006-01-02"),
			Name: s.Name,
		}

		for _, item := range s.AgendaItems {
			speakerName := ""
			if item.Speaker != nil {
				speakerName = item.Speaker.Name
			}

			location := item.Location
			if item.VenueRoom != nil {
				location = item.VenueRoom.Name
			}

			duration := ""
			if !item.StartTime.IsZero() && !item.EndTime.IsZero() {
				duration = item.EndTime.Sub(item.StartTime).String()
			}

			day.Items = append(day.Items, TimelineItem{
				ID:       item.ID,
				Time:     item.StartTime.Format("15:04"),
				Title:    item.Title,
				Speaker:  speakerName,
				Location: location,
				Duration: duration,
			})
		}

		timeline = append(timeline, day)
	}

	c.JSON(http.StatusOK, gin.H{"timeline": timeline})
}

func (h *ScheduleHandler) ExportPDF(c *gin.Context) {
	eventID := c.Param("event_id")

	var event models.Event
	if err := models.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var schedules []models.Schedule
	models.DB.Preload("AgendaItems.Speaker").Preload("AgendaItems.VenueRoom").
		Where("event_id = ?", eventID).Order("date ASC").Find(&schedules)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 18)
	pdf.Cell(0, 15, event.Title)
	pdf.Ln(20)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 8, fmt.Sprintf("时间: %s - %s", event.StartTime.Format("2006-01-02 15:04"), event.EndTime.Format("2006-01-02 15:04")))
	pdf.Ln(6)
	pdf.Cell(0, 8, fmt.Sprintf("地点: %s", event.Location))
	pdf.Ln(15)

	for _, s := range schedules {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, fmt.Sprintf("%s - %s", s.Date.Format("2006-01-02"), s.Name))
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 10)
		for _, item := range s.AgendaItems {
			timeRange := fmt.Sprintf("%s - %s", item.StartTime.Format("15:04"), item.EndTime.Format("15:04"))
			pdf.Cell(30, 6, timeRange)

			title := item.Title
			if item.Speaker != nil {
				title += fmt.Sprintf(" (%s)", item.Speaker.Name)
			}
			if item.VenueRoom != nil {
				title += fmt.Sprintf(" [%s]", item.VenueRoom.Name)
			}
			pdf.Cell(0, 6, title)
			pdf.Ln(6)

			if item.Description != "" {
				pdf.SetFont("Arial", "I", 9)
				pdf.Cell(30, 5, "")
				pdf.MultiCell(0, 5, item.Description, "", "L", false)
				pdf.SetFont("Arial", "", 10)
			}
		}
		pdf.Ln(5)
	}

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=schedule_%s.pdf", eventID))

	var buf []byte
	pdf.Output(pdfWriter{buf: &buf})
	c.Data(http.StatusOK, "application/pdf", buf)
}

type pdfWriter struct {
	buf *[]byte
}

func (w pdfWriter) Write(p []byte) (n int, err error) {
	*w.buf = append(*w.buf, p...)
	return len(p), nil
}
