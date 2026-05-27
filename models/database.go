package models

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitDB(dbPath string) {
	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		log.Fatalf("Failed to connect database: %v", err)
	}

	err = DB.AutoMigrate(
		&User{},
		&Event{},
		&CustomField{},
		&Feedback{},
		&Registration{},
		&Ticket{},
		&PromoCode{},
		&Order{},
		&PaymentRecord{},
		&Schedule{},
		&AgendaItem{},
		&Speaker{},
		&Venue{},
		&VenueRoom{},
		&Seat{},
		&Notification{},
		&NotificationTemplate{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	seedData()
}

func seedData() {
	var count int64
	DB.Model(&User{}).Count(&count)
	if count > 0 {
		return
	}

	admin := User{
		Email:    "admin@eventforge.local",
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy",
		Name:     "Administrator",
		Role:     "admin",
	}
	DB.Create(&admin)

	templates := []NotificationTemplate{
		{
			Name:      "registration_success",
			Type:      string(NotificationTypeRegistrationSuccess),
			Subject:   "活动报名成功 - {{.EventTitle}}",
			Content:   "您好 {{.Name}}，\n\n您已成功报名参加「{{.EventTitle}}」活动。\n\n活动时间：{{.StartTime}}\n活动地点：{{.Location}}\n\n请保存此邮件作为凭证，活动开始前我们将发送提醒通知。\n\n报名编号：{{.RegistrationID}}\n签到二维码：{{.QRCode}}\n\n此致\nEventForge 团队",
			Variables: `["Name","EventTitle","StartTime","Location","RegistrationID","QRCode"]`,
		},
		{
			Name:      "event_update",
			Type:      string(NotificationTypeEventUpdate),
			Subject:   "活动信息更新 - {{.EventTitle}}",
			Content:   "您好 {{.Name}}，\n\n您报名的活动「{{.EventTitle}}」信息有更新：\n\n{{.UpdateContent}}\n\n请查看最新活动信息，如有疑问请联系我们。\n\n此致\nEventForge 团队",
			Variables: `["Name","EventTitle","UpdateContent"]`,
		},
		{
			Name:      "event_reminder",
			Type:      string(NotificationTypeEventReminder),
			Subject:   "活动提醒 - {{.EventTitle}} 即将开始",
			Content:   "您好 {{.Name}}，\n\n您报名的活动「{{.EventTitle}}」即将开始！\n\n活动时间：{{.StartTime}}\n活动地点：{{.Location}}\n\n请准时参加，期待您的到来！\n\n此致\nEventForge 团队",
			Variables: `["Name","EventTitle","StartTime","Location"]`,
		},
	}
	DB.Create(&templates)
}
