package middleware

import (
	"fmt"
	"log"
	"time"

	"eventforge/models"
	"eventforge/utils"
)

type ReminderLog struct {
	EventID uint
	ReminderType string
	LastSent time.Time
}

var reminderSent = make(map[string]ReminderLog)

func StartReminderScheduler() {
	log.Println("[Scheduler] Starting event reminder scheduler...")

	ticker := time.NewTicker(5 * time.Minute)
	go func() {
		for range ticker.C {
			checkAndSendReminders()
		}
	}()
}

func checkAndSendReminders() {
	now := time.Now()

	var events []models.Event
	models.DB.Where("is_template = ? AND status = ?", false, models.EventStatusOpen).Find(&events)

	for _, event := range events {
		timeUntilStart := event.StartTime.Sub(now)

		if timeUntilStart <= 24*time.Hour && timeUntilStart > 23*time.Hour {
			key := fmt.Sprintf("%d-1day", event.ID)
			if !shouldSendReminder(key, event.StartTime) {
				continue
			}
			sendReminder(event, "1天")
			markReminderSent(key, event.ID, "1day")
		}

		if timeUntilStart <= 1*time.Hour && timeUntilStart > 0 {
			key := fmt.Sprintf("%d-1hour", event.ID)
			if !shouldSendReminder(key, event.StartTime) {
				continue
			}
			sendReminder(event, "1小时")
			markReminderSent(key, event.ID, "1hour")
		}
	}
}

func shouldSendReminder(key string, eventStartTime time.Time) bool {
	log, exists := reminderSent[key]
	if !exists {
		return true
	}

	if log.LastSent.Year() != eventStartTime.Year() ||
		log.LastSent.YearDay() != eventStartTime.YearDay() {
		return true
	}

	return false
}

func markReminderSent(key string, eventID uint, reminderType string) {
	reminderSent[key] = ReminderLog{
		EventID:      eventID,
		ReminderType: reminderType,
		LastSent:     time.Now(),
	}
}

func sendReminder(event models.Event, reminderTime string) {
	var registrations []models.Registration
	models.DB.Where("event_id = ? AND status IN ?", event.ID,
		[]models.RegistrationStatus{models.RegistrationStatusConfirmed, models.RegistrationStatusCheckedIn}).
		Distinct("email").Find(&registrations)

	if len(registrations) == 0 {
		return
	}

	subject := fmt.Sprintf("活动提醒 - %s 还有%s开始", event.Title, reminderTime)
	content := fmt.Sprintf("您好，\n\n您报名的活动「%s」还有%s就要开始了！\n\n活动时间：%s\n活动地点：%s\n\n请准时参加，期待您的到来！\n\n此致\nEventForge 团队",
		event.Title, reminderTime, event.StartTime.Format("2006-01-02 15:04"), event.Location)

	now := time.Now()
	sentCount := 0

	for _, r := range registrations {
		err := utils.SendEmail(r.Email, subject, content)
		if err == nil {
			sentCount++
		}
	}

	models.DB.Create(&models.Notification{
		EventID: &event.ID,
		Type:    models.NotificationTypeEventReminder,
		Subject: subject,
		Content: content,
		Status:  models.NotificationStatusSent,
		SentAt:  &now,
	})

	log.Printf("[Scheduler] Sent %s reminder for event %d to %d recipients", reminderTime, event.ID, sentCount)
}
