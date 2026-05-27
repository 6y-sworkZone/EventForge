package utils

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"time"

	"eventforge/config"

	"github.com/jung-kurt/gofpdf"
	"github.com/skip2/go-qrcode"
	"gopkg.in/gomail.v2"
)

func GenerateQRCode(content string) ([]byte, error) {
	qr, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return nil, err
	}
	return qr.PNG(256)
}

func SaveQRCode(content, filename string) (string, error) {
	cfg := config.Load()
	dir := cfg.UploadPath + "/qrcodes"
	os.MkdirAll(dir, 0755)

	filepath := dir + "/" + filename
	err := qrcode.WriteFile(content, qrcode.Medium, 256, filepath)
	if err != nil {
		return "", err
	}
	return "/uploads/qrcodes/" + filename, nil
}

func GenerateICS(eventTitle string, startTime, endTime time.Time, location, description string) string {
	return fmt.Sprintf(`BEGIN:VCALENDAR
VERSION:2.0
PRODID:-//EventForge//Event//EN
BEGIN:VEVENT
UID:%s@eventforge
DTSTAMP:%s
DTSTART:%s
DTEND:%s
SUMMARY:%s
DESCRIPTION:%s
LOCATION:%s
END:VEVENT
END:VCALENDAR`,
		fmt.Sprintf("%d", time.Now().Unix()),
		time.Now().UTC().Format("20060102T150405Z"),
		startTime.UTC().Format("20060102T150405Z"),
		endTime.UTC().Format("20060102T150405Z"),
		eventTitle,
		description,
		location,
	)
}

func SendEmail(to, subject, body string, attachments ...string) error {
	cfg := config.Load()

	if cfg.SMTPHost == "" {
		fmt.Printf("[EMAIL] To: %s\nSubject: %s\nBody:\n%s\n", to, subject, body)
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.SMTPFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/plain", body)

	for _, attachment := range attachments {
		m.Attach(attachment)
	}

	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	return d.DialAndSend(m)
}

func SendHTMLEmail(to, subject, htmlBody string, attachments ...string) error {
	cfg := config.Load()

	if cfg.SMTPHost == "" {
		fmt.Printf("[EMAIL] To: %s\nSubject: %s\nBody:\n%s\n", to, subject, htmlBody)
		return nil
	}

	m := gomail.NewMessage()
	m.SetHeader("From", cfg.SMTPFrom)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", htmlBody)

	for _, attachment := range attachments {
		m.Attach(attachment)
	}

	d := gomail.NewDialer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass)
	return d.DialAndSend(m)
}

func GenerateCSV(headers []string, rows [][]string) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)

	if err := w.Write(headers); err != nil {
		return nil, err
	}

	for _, row := range rows {
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}

	w.Flush()
	return buf.Bytes(), nil
}

func GenerateSchedulePDF(eventTitle string, schedules []map[string]interface{}) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, eventTitle)
	pdf.Ln(15)

	for _, s := range schedules {
		pdf.SetFont("Arial", "B", 12)
		pdf.Cell(0, 8, fmt.Sprintf("%s - %s", s["date"], s["name"]))
		pdf.Ln(10)

		pdf.SetFont("Arial", "", 10)
		if items, ok := s["items"].([]map[string]interface{}); ok {
			for _, item := range items {
				timeRange := fmt.Sprintf("%s - %s", item["start_time"], item["end_time"])
				pdf.Cell(30, 6, timeRange)
				pdf.Cell(0, 6, fmt.Sprintf("%s", item["title"]))
				if speaker, ok := item["speaker"]; ok && speaker != nil {
					pdf.Cell(0, 6, fmt.Sprintf(" (%s)", speaker))
				}
				pdf.Ln(6)
			}
		}
		pdf.Ln(5)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
