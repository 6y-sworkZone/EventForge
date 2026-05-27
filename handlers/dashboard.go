package handlers

import (
	"net/http"
	"time"

	"eventforge/models"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) GetOverview(c *gin.Context) {
	var totalEvents int64
	var totalRegistrations int64
	var totalRevenue float64
	var upcomingEvents int64

	models.DB.Model(&models.Event{}).Where("is_template = ?", false).Count(&totalEvents)
	models.DB.Model(&models.Registration{}).Where("status IN ?",
		[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).Count(&totalRegistrations)

	now := time.Now()
	models.DB.Model(&models.Event{}).Where("is_template = ? AND start_time > ?", false, now).Count(&upcomingEvents)

	type OrderTotal struct {
		Total float64
	}
	var result OrderTotal
	models.DB.Model(&models.Order{}).Where("status = ?", models.OrderStatusPaid).Select("COALESCE(SUM(total_amount), 0) as total").Scan(&result)
	totalRevenue = result.Total

	c.JSON(http.StatusOK, gin.H{
		"total_events":       totalEvents,
		"total_registrations": totalRegistrations,
		"total_revenue":      totalRevenue,
		"upcoming_events":    upcomingEvents,
	})
}

func (h *DashboardHandler) GetRegistrationTrend(c *gin.Context) {
	eventID := c.Query("event_id")
	days := 7

	type DailyCount struct {
		Date  string `json:"date"`
		Count int64  `json:"count"`
	}

	var trends []DailyCount

	baseQuery := models.DB.Model(&models.Registration{})
	if eventID != "" {
		baseQuery = baseQuery.Where("event_id = ?", eventID)
	}

	for i := days - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		endOfDay := startOfDay.Add(24 * time.Hour)

		var count int64
		baseQuery.Where("created_at >= ? AND created_at < ?", startOfDay, endOfDay).Count(&count)

		trends = append(trends, DailyCount{
			Date:  date.Format("2006-01-02"),
			Count: count,
		})
	}

	c.JSON(http.StatusOK, gin.H{"trends": trends})
}

func (h *DashboardHandler) GetTicketSales(c *gin.Context) {
	eventID := c.Param("event_id")

	var tickets []models.Ticket
	query := models.DB.Where("event_id = ?", eventID)
	query.Find(&tickets)

	type TicketSale struct {
		ID        uint    `json:"id"`
		Name      string  `json:"name"`
		Quantity  int     `json:"quantity"`
		SoldCount int     `json:"sold_count"`
		Remaining int     `json:"remaining"`
		Price     float64 `json:"price"`
		Revenue   float64 `json:"revenue"`
		Progress  float64 `json:"progress"`
	}

	var sales []TicketSale
	totalSold := 0
	totalRevenue := 0.0
	totalCapacity := 0

	for _, t := range tickets {
		revenue := float64(t.SoldCount) * t.Price
		progress := 0.0
		if t.Quantity > 0 {
			progress = float64(t.SoldCount) / float64(t.Quantity) * 100
		}

		sales = append(sales, TicketSale{
			ID:        t.ID,
			Name:      t.Name,
			Quantity:  t.Quantity,
			SoldCount: t.SoldCount,
			Remaining: t.Quantity - t.SoldCount,
			Price:     t.Price,
			Revenue:   revenue,
			Progress:  progress,
		})

		totalSold += t.SoldCount
		totalRevenue += revenue
		totalCapacity += t.Quantity
	}

	c.JSON(http.StatusOK, gin.H{
		"tickets":        sales,
		"total_sold":     totalSold,
		"total_revenue":  totalRevenue,
		"total_capacity": totalCapacity,
	})
}

func (h *DashboardHandler) GetCheckInRate(c *gin.Context) {
	eventID := c.Param("event_id")

	var totalRegistered int64
	var totalCheckedIn int64

	query := models.DB.Model(&models.Registration{}).Where("event_id = ?", eventID)
	query.Where("status IN ?", []models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).Count(&totalRegistered)

	query.Where("status = ?", models.RegistrationStatusCheckedIn).Count(&totalCheckedIn)

	rate := 0.0
	if totalRegistered > 0 {
		rate = float64(totalCheckedIn) / float64(totalRegistered) * 100
	}

	c.JSON(http.StatusOK, gin.H{
		"registered":  totalRegistered,
		"checked_in":  totalCheckedIn,
		"check_in_rate": rate,
	})
}

func (h *DashboardHandler) GetEventFeedback(c *gin.Context) {
	eventID := c.Param("event_id")

	var feedbacks []models.Feedback
	models.DB.Where("event_id = ?", eventID).Find(&feedbacks)

	totalRating := 0
	count := len(feedbacks)
	for _, f := range feedbacks {
		totalRating += f.Rating
	}

	avgRating := 0.0
	if count > 0 {
		avgRating = float64(totalRating) / float64(count)
	}

	type RatingDistribution struct {
		Rating int   `json:"rating"`
		Count  int64 `json:"count"`
	}

	var distribution []RatingDistribution
	for i := 1; i <= 5; i++ {
		var count int64
		models.DB.Model(&models.Feedback{}).Where("event_id = ? AND rating = ?", eventID, i).Count(&count)
		distribution = append(distribution, RatingDistribution{
			Rating: i,
			Count:  count,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"feedback_count":    count,
		"average_rating":    avgRating,
		"rating_distribution": distribution,
		"feedbacks":         feedbacks,
	})
}

func (h *DashboardHandler) SubmitFeedback(c *gin.Context) {
	eventID := c.Param("event_id")
	userID := c.GetUint("user_id")

	var req struct {
		Rating  int    `json:"rating" binding:"required,min=1,max=5"`
		Comment string `json:"comment"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var existing models.Feedback
	if models.DB.Where("event_id = ? AND user_id = ?", eventID, userID).First(&existing).Error == nil {
		existing.Rating = req.Rating
		existing.Comment = req.Comment
		models.DB.Save(&existing)
		c.JSON(http.StatusOK, gin.H{"feedback": existing})
		return
	}

	feedback := models.Feedback{
		EventID: parseUint(eventID),
		UserID:  userID,
		Rating:  req.Rating,
		Comment: req.Comment,
	}

	if err := models.DB.Create(&feedback).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit feedback"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"feedback": feedback})
}

func (h *DashboardHandler) GetParticipantProfile(c *gin.Context) {
	eventID := c.Param("event_id")

	type CompanyDist struct {
		Company string `json:"company"`
		Count   int64  `json:"count"`
	}

	type PositionDist struct {
		Position string `json:"position"`
		Count    int64  `json:"count"`
	}

	type RegionDist struct {
		Region string `json:"region"`
		Count  int64  `json:"count"`
	}

	var companies []CompanyDist
	var positions []PositionDist
	var regions []RegionDist

	query := models.DB.Model(&models.Registration{}).Where("event_id = ? AND status IN ?", eventID,
		[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn})

	query.Select("company, count(*) as count").Group("company").Having("company != ''").Order("count DESC").Limit(10).Scan(&companies)

	query.Select("position, count(*) as count").Group("position").Having("position != ''").Order("count DESC").Limit(10).Scan(&positions)

	query.Select("region, count(*) as count").Group("region").Having("region != ''").Order("count DESC").Limit(10).Scan(&regions)

	c.JSON(http.StatusOK, gin.H{
		"companies": companies,
		"positions": positions,
		"regions":   regions,
	})
}

func (h *DashboardHandler) GetHistoricalComparison(c *gin.Context) {
	var events []models.Event
	models.DB.Where("is_template = ? AND status = ?", false, models.EventStatusEnded).
		Order("end_time DESC").Limit(10).Find(&events)

	type EventComparison struct {
		ID              uint    `json:"id"`
		Title           string  `json:"title"`
		Date            string  `json:"date"`
		Registrations   int64   `json:"registrations"`
		CheckedIn       int64   `json:"checked_in"`
		Revenue         float64 `json:"revenue"`
		AvgRating       float64 `json:"avg_rating"`
	}

	var comparisons []EventComparison

	for _, e := range events {
		var regCount int64
		models.DB.Model(&models.Registration{}).Where("event_id = ? AND status IN ?", e.ID,
			[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).Count(&regCount)

		var checkInCount int64
		models.DB.Model(&models.Registration{}).Where("event_id = ? AND status = ?", e.ID, models.RegistrationStatusCheckedIn).Count(&checkInCount)

		var revenue float64
		type RevTotal struct {
			Total float64
		}
		var rev RevTotal
		models.DB.Model(&models.Order{}).Where("event_id = ? AND status = ?", e.ID, models.OrderStatusPaid).Select("COALESCE(SUM(total_amount), 0) as total").Scan(&rev)
		revenue = rev.Total

		var avgRating float64
		type RatingAvg struct {
			Avg float64
		}
		var rat RatingAvg
		models.DB.Model(&models.Feedback{}).Where("event_id = ?", e.ID).Select("COALESCE(AVG(rating), 0) as avg").Scan(&rat)
		avgRating = rat.Avg

		comparisons = append(comparisons, EventComparison{
			ID:            e.ID,
			Title:         e.Title,
			Date:          e.StartTime.Format("2006-01-02"),
			Registrations: regCount,
			CheckedIn:     checkInCount,
			Revenue:       revenue,
			AvgRating:     avgRating,
		})
	}

	c.JSON(http.StatusOK, gin.H{"comparisons": comparisons})
}

func (h *DashboardHandler) GetEventStats(c *gin.Context) {
	eventID := c.Param("event_id")

	var event models.Event
	if err := models.DB.First(&event, eventID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Event not found"})
		return
	}

	var regCount int64
	models.DB.Model(&models.Registration{}).Where("event_id = ? AND status IN ?", event.ID,
		[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).Count(&regCount)

	var waitlistCount int64
	models.DB.Model(&models.Registration{}).Where("event_id = ? AND status = ?", event.ID, models.RegistrationStatusWaitlist).Count(&waitlistCount)

	var checkInCount int64
	models.DB.Model(&models.Registration{}).Where("event_id = ? AND status = ?", event.ID, models.RegistrationStatusCheckedIn).Count(&checkInCount)

	var revenue float64
	type RevTotal struct {
		Total float64
	}
	var rev RevTotal
	models.DB.Model(&models.Order{}).Where("event_id = ? AND status = ?", event.ID, models.OrderStatusPaid).Select("COALESCE(SUM(total_amount), 0) as total").Scan(&rev)
	revenue = rev.Total

	var ticketCount int64
	models.DB.Model(&models.Ticket{}).Where("event_id = ?", event.ID).Count(&ticketCount)

	var scheduleCount int64
	models.DB.Model(&models.Schedule{}).Where("event_id = ?", event.ID).Count(&scheduleCount)

	var speakerCount int64
	models.DB.Model(&models.Speaker{}).Where("event_id = ?", event.ID).Count(&speakerCount)

	c.JSON(http.StatusOK, gin.H{
		"event":           event,
		"registered":      regCount,
		"waitlist":        waitlistCount,
		"checked_in":      checkInCount,
		"revenue":         revenue,
		"ticket_types":    ticketCount,
		"schedules":       scheduleCount,
		"speakers":        speakerCount,
		"capacity":        event.MaxCapacity,
		"capacity_usage":  func() float64 { if event.MaxCapacity > 0 { return float64(regCount) / float64(event.MaxCapacity) * 100 }; return 0 }(),
	})
}
