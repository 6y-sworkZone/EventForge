package handlers

import (
	"net/http"
	"time"

	"eventforge/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketHandler struct{}

func NewTicketHandler() *TicketHandler {
	return &TicketHandler{}
}

func (h *TicketHandler) CreateTicket(c *gin.Context) {
	eventID := c.Param("event_id")

	var req struct {
		Name         string          `json:"name" binding:"required"`
		Type         models.TicketType `json:"type" binding:"required"`
		Price        float64         `json:"price"`
		Quantity     int             `json:"quantity" binding:"required"`
		Description  string          `json:"description"`
		SaleStartAt  *time.Time      `json:"sale_start_at"`
		SaleEndAt    *time.Time      `json:"sale_end_at"`
		EarlyPrice   float64         `json:"early_price"`
		EarlyStartAt *time.Time      `json:"early_start_at"`
		EarlyEndAt   *time.Time      `json:"early_end_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ticket := models.Ticket{
		EventID:      parseUint(eventID),
		Name:         req.Name,
		Type:         req.Type,
		Price:        req.Price,
		Quantity:     req.Quantity,
		Description:  req.Description,
		SaleStartAt:  req.SaleStartAt,
		SaleEndAt:    req.SaleEndAt,
		EarlyPrice:   req.EarlyPrice,
		EarlyStartAt: req.EarlyStartAt,
		EarlyEndAt:   req.EarlyEndAt,
	}

	if err := models.DB.Create(&ticket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create ticket"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"ticket": ticket})
}

func (h *TicketHandler) ListTickets(c *gin.Context) {
	eventID := c.Param("event_id")

	var tickets []models.Ticket
	if err := models.DB.Where("event_id = ?", eventID).Find(&tickets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tickets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tickets": tickets})
}

func (h *TicketHandler) UpdateTicket(c *gin.Context) {
	id := c.Param("id")

	var ticket models.Ticket
	if err := models.DB.First(&ticket, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ticket not found"})
		return
	}

	var req struct {
		Name         string          `json:"name"`
		Type         models.TicketType `json:"type"`
		Price        float64         `json:"price"`
		Quantity     int             `json:"quantity"`
		Description  string          `json:"description"`
		SaleStartAt  *time.Time      `json:"sale_start_at"`
		SaleEndAt    *time.Time      `json:"sale_end_at"`
		EarlyPrice   float64         `json:"early_price"`
		EarlyStartAt *time.Time      `json:"early_start_at"`
		EarlyEndAt   *time.Time      `json:"early_end_at"`
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
	if req.Price != 0 {
		updates["price"] = req.Price
	}
	if req.Quantity != 0 {
		updates["quantity"] = req.Quantity
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.SaleStartAt != nil {
		updates["sale_start_at"] = req.SaleStartAt
	}
	if req.SaleEndAt != nil {
		updates["sale_end_at"] = req.SaleEndAt
	}
	if req.EarlyPrice != 0 {
		updates["early_price"] = req.EarlyPrice
	}
	if req.EarlyStartAt != nil {
		updates["early_start_at"] = req.EarlyStartAt
	}
	if req.EarlyEndAt != nil {
		updates["early_end_at"] = req.EarlyEndAt
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&ticket).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"ticket": ticket})
}

func (h *TicketHandler) DeleteTicket(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.Ticket{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete ticket"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Ticket deleted successfully"})
}

func (h *TicketHandler) CreatePromoCode(c *gin.Context) {
	eventID := c.Param("event_id")

	var req struct {
		Code      string              `json:"code" binding:"required"`
		Type      models.PromoCodeType `json:"type" binding:"required"`
		Value     float64             `json:"value" binding:"required"`
		MaxUsage  int                 `json:"max_usage"`
		ExpiresAt *time.Time          `json:"expires_at"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	promo := models.PromoCode{
		EventID:   parseUint(eventID),
		Code:      req.Code,
		Type:      req.Type,
		Value:     req.Value,
		MaxUsage:  req.MaxUsage,
		ExpiresAt: req.ExpiresAt,
		IsActive:  true,
	}

	if err := models.DB.Create(&promo).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create promo code"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"promo_code": promo})
}

func (h *TicketHandler) ListPromoCodes(c *gin.Context) {
	eventID := c.Param("event_id")

	var promos []models.PromoCode
	if err := models.DB.Where("event_id = ?", eventID).Find(&promos).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch promo codes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"promo_codes": promos})
}

func (h *TicketHandler) UpdatePromoCode(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Code      string              `json:"code"`
		Type      models.PromoCodeType `json:"type"`
		Value     float64             `json:"value"`
		MaxUsage  int                 `json:"max_usage"`
		ExpiresAt *time.Time          `json:"expires_at"`
		IsActive  *bool               `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var promo models.PromoCode
	if err := models.DB.First(&promo, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Promo code not found"})
		return
	}

	updates := map[string]interface{}{}
	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Type != "" {
		updates["type"] = req.Type
	}
	if req.Value != 0 {
		updates["value"] = req.Value
	}
	if req.MaxUsage != 0 {
		updates["max_usage"] = req.MaxUsage
	}
	if req.ExpiresAt != nil {
		updates["expires_at"] = req.ExpiresAt
	}
	if req.IsActive != nil {
		updates["is_active"] = req.IsActive
	}
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		models.DB.Model(&promo).Updates(updates)
	}

	c.JSON(http.StatusOK, gin.H{"promo_code": promo})
}

func (h *TicketHandler) DeletePromoCode(c *gin.Context) {
	id := c.Param("id")

	if err := models.DB.Delete(&models.PromoCode{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete promo code"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Promo code deleted successfully"})
}

func (h *TicketHandler) ListOrders(c *gin.Context) {
	eventID := c.Query("event_id")
	status := c.Query("status")

	var orders []models.Order
	query := models.DB.Preload("Event").Preload("Ticket").Preload("PromoCode")

	if eventID != "" {
		query = query.Where("event_id = ?", eventID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"orders": orders})
}

func (h *TicketHandler) GetOrder(c *gin.Context) {
	id := c.Param("id")

	var order models.Order
	if err := models.DB.Preload("Event").Preload("Ticket").Preload("PromoCode").First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

func (h *TicketHandler) MarkAsPaid(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		PaymentMethod string `json:"payment_method" binding:"required"`
		Remark        string `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := models.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":         models.OrderStatusPaid,
		"payment_method": req.PaymentMethod,
		"paid_at":        now,
	}

	models.DB.Model(&order).Updates(updates)

	payment := models.PaymentRecord{
		OrderID:       order.ID,
		Amount:        order.TotalAmount,
		PaymentMethod: req.PaymentMethod,
		Remark:        req.Remark,
	}
	models.DB.Create(&payment)

	if order.TicketID != 0 {
		models.DB.Model(&models.Ticket{}).Where("id = ?", order.TicketID).UpdateColumn("sold_count", gorm.Expr("sold_count + ?", order.Quantity))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order marked as paid"})
}

func (h *TicketHandler) RefundOrder(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order
	if err := models.DB.First(&order, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":       models.OrderStatusRefunded,
		"refund_reason": req.Reason,
		"refunded_at":  now,
	}

	models.DB.Model(&order).Updates(updates)

	if order.TicketID != 0 {
		models.DB.Model(&models.Ticket{}).Where("id = ?", order.TicketID).UpdateColumn("sold_count", gorm.Expr("sold_count - ?", order.Quantity))
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order refunded successfully"})
}

func (h *TicketHandler) UpdateInvoiceInfo(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		InvoiceInfo string `json:"invoice_info" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	models.DB.Model(&models.Order{}).Where("id = ?", id).Update("invoice_info", req.InvoiceInfo)

	c.JSON(http.StatusOK, gin.H{"message": "Invoice info updated"})
}

func (h *TicketHandler) GetTicketStats(c *gin.Context) {
	eventID := c.Param("event_id")

	var tickets []models.Ticket
	models.DB.Where("event_id = ?", eventID).Find(&tickets)

	type TicketStat struct {
		ID          uint    `json:"id"`
		Name        string  `json:"name"`
		Quantity    int     `json:"quantity"`
		SoldCount   int     `json:"sold_count"`
		Remaining   int     `json:"remaining"`
		Price       float64 `json:"price"`
		Revenue     float64 `json:"revenue"`
	}

	var stats []TicketStat
	totalRevenue := 0.0
	totalSold := 0

	for _, t := range tickets {
		revenue := float64(t.SoldCount) * t.Price
		stat := TicketStat{
			ID:        t.ID,
			Name:      t.Name,
			Quantity:  t.Quantity,
			SoldCount: t.SoldCount,
			Remaining: t.Quantity - t.SoldCount,
			Price:     t.Price,
			Revenue:   revenue,
		}
		stats = append(stats, stat)
		totalRevenue += revenue
		totalSold += t.SoldCount
	}

	c.JSON(http.StatusOK, gin.H{
		"tickets":      stats,
		"total_revenue": totalRevenue,
		"total_sold":   totalSold,
	})
}

func (h *TicketHandler) ListPayments(c *gin.Context) {
	orderID := c.Query("order_id")

	var payments []models.PaymentRecord
	query := models.DB.Preload("Order")

	if orderID != "" {
		query = query.Where("order_id = ?", orderID)
	}

	if err := query.Order("created_at DESC").Find(&payments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch payments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payments": payments})
}
