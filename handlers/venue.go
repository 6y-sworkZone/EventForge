package handlers

import (
	"net/http"
	"time"

	"eventforge/models"

	"github.com/gin-gonic/gin"
)

type VenueHandler struct{}

func NewVenueHandler() *VenueHandler {
	return &VenueHandler{}
}

func (h *VenueHandler) CreateVenue(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		Address     string `json:"address"`
		Capacity    int    `json:"capacity"`
		Facilities  string `json:"facilities"`
		FloorPlan   string `json:"floor_plan"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	venue := models.Venue{
		Name:        req.Name,
		Address:     req.Address,
		Capacity:    req.Capacity,
		Facilities:  req.Facilities,
		FloorPlan:   req.FloorPlan,
		Description: req.Description,
	}

	if err := models.DB.Create(&venue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create venue"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"venue": venue})
}

func (h *VenueHandler) ListVenues(c *gin.Context) {
	var venues []models.Venue
	if err := models.DB.Preload("Rooms").Find(&venues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch venues"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"venues": venues})
}

func (h *VenueHandler) GetVenue(c *gin.Context) {
	id := c.Param("id")

	var venue models.Venue
	if err := models.DB.Preload("Rooms").First(&venue, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"venue": venue})
}

func (h *VenueHandler) UpdateVenue(c *gin.Context) {
	id := c.Param("id")

	var venue models.Venue
	if err := models.DB.First(&venue, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Venue not found"})
		return
	}

	var req struct {
		Name        string `json:"name"`
		Address     string `json:"address"`
		Capacity    int    `json:"capacity"`
		Facilities  string `json:"facilities"`
		FloorPlan   string `json:"floor_plan"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Address != "" {
		updates["address"] = req.Address
	}
	if req.Capacity != 0 {
		updates["capacity"] = req.Capacity
	}
	if req.Facilities != "" {
		updates["facilities"] = req.Facilities
	}
	if req.FloorPlan != "" {
		updates["floor_plan"] = req.FloorPlan
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&venue).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"venue": venue})
}

func (h *VenueHandler) DeleteVenue(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.Venue{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete venue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Venue deleted successfully"})
}

func (h *VenueHandler) CreateRoom(c *gin.Context) {
	venueID := c.Param("id")

	var req struct {
		Name     string `json:"name" binding:"required"`
		Capacity int    `json:"capacity"`
		Floor    string `json:"floor"`
		SeatMap  string `json:"seat_map"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room := models.VenueRoom{
		VenueID:  parseUint(venueID),
		Name:     req.Name,
		Capacity: req.Capacity,
		Floor:    req.Floor,
		SeatMap:  req.SeatMap,
	}

	if err := models.DB.Create(&room).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"room": room})
}

func (h *VenueHandler) ListRooms(c *gin.Context) {
	venueID := c.Param("id")

	var rooms []models.VenueRoom
	if err := models.DB.Where("venue_id = ?", venueID).Find(&rooms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": rooms})
}

func (h *VenueHandler) UpdateRoom(c *gin.Context) {
	id := c.Param("id")

	var room models.VenueRoom
	if err := models.DB.First(&room, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	var req struct {
		Name     string `json:"name"`
		Capacity int    `json:"capacity"`
		Floor    string `json:"floor"`
		SeatMap  string `json:"seat_map"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Capacity != 0 {
		updates["capacity"] = req.Capacity
	}
	if req.Floor != "" {
		updates["floor"] = req.Floor
	}
	if req.SeatMap != "" {
		updates["seat_map"] = req.SeatMap
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&room).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"room": room})
}

func (h *VenueHandler) DeleteRoom(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.VenueRoom{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete room"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room deleted successfully"})
}

func (h *VenueHandler) CreateSeat(c *gin.Context) {
	roomID := c.Param("room_id")

	var req struct {
		SeatNumber string  `json:"seat_number" binding:"required"`
		Row        string  `json:"row"`
		Section    string  `json:"section"`
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Status     string  `json:"status"`
		EventID    *uint   `json:"event_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := req.Status
	if status == "" {
		status = "available"
	}

	seat := models.Seat{
		VenueRoomID: parseUint(roomID),
		SeatNumber:  req.SeatNumber,
		Row:         req.Row,
		Section:     req.Section,
		X:           req.X,
		Y:           req.Y,
		Status:      status,
		EventID:     req.EventID,
	}

	if err := models.DB.Create(&seat).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create seat"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"seat": seat})
}

func (h *VenueHandler) ListSeats(c *gin.Context) {
	roomID := c.Param("room_id")
	eventID := c.Query("event_id")

	var seats []models.Seat
	query := models.DB.Where("venue_room_id = ?", roomID)

	if eventID != "" {
		query = query.Where("event_id = ? OR event_id IS NULL", parseUint(eventID))
	}

	if err := query.Find(&seats).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch seats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"seats": seats})
}

func (h *VenueHandler) UpdateSeat(c *gin.Context) {
	id := c.Param("id")

	var seat models.Seat
	if err := models.DB.First(&seat, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Seat not found"})
		return
	}

	var req struct {
		SeatNumber string  `json:"seat_number"`
		Row        string  `json:"row"`
		Section    string  `json:"section"`
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Status     string  `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.SeatNumber != "" {
		updates["seat_number"] = req.SeatNumber
	}
	if req.Row != "" {
		updates["row"] = req.Row
	}
	if req.Section != "" {
		updates["section"] = req.Section
	}
	if req.X != 0 {
		updates["x"] = req.X
	}
	if req.Y != 0 {
		updates["y"] = req.Y
	}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&seat).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"seat": seat})
}

func (h *VenueHandler) DeleteSeat(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.Seat{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete seat"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Seat deleted successfully"})
}

func (h *VenueHandler) GetVenueCalendar(c *gin.Context) {
	venueID := c.Param("venue_id")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var events []models.Event
	query := models.DB.Where("venue_id = ?", venueID)

	if startDate != "" {
		query = query.Where("start_time >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("end_time <= ?", endDate)
	}

	query.Where("status != ?", models.EventStatusCancelled).Find(&events)

	type CalendarEvent struct {
		ID        uint   `json:"id"`
		Title     string `json:"title"`
		StartTime string `json:"start_time"`
		EndTime   string `json:"end_time"`
		Status    string `json:"status"`
	}

	var calendar []CalendarEvent
	for _, e := range events {
		calendar = append(calendar, CalendarEvent{
			ID:        e.ID,
			Title:     e.Title,
			StartTime: e.StartTime.Format("2006-01-02 15:04"),
			EndTime:   e.EndTime.Format("2006-01-02 15:04"),
			Status:    string(e.Status),
		})
	}

	c.JSON(http.StatusOK, gin.H{"calendar": calendar})
}
