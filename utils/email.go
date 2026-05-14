package utils

import (
	"fmt"
	"net/smtp"
	"os"
)

func SendEmail(to string, subject string, body string) error {
	from := os.Getenv("SMTP_FROM")
	password := os.Getenv("SMTP_PASS")
	user := os.Getenv("SMTP_USER")
	host := os.Getenv("SMTP_HOST")
	port := os.Getenv("SMTP_PORT")

	auth := smtp.PlainAuth("", user, password, host)

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + subject + "\r\n" +
		"\r\n" +
		body + "\r\n")

	addr := fmt.Sprintf("%s:%s", host, port)
	return smtp.SendMail(addr, auth, from, []string{to}, msg)
}

func SendVerificationEmail(to string, token string) error {
	subject := "Verify your email - ResultsPro"
	body := fmt.Sprintf("Please verify your email by clicking the link below:\r\nhttps://auth.resultspro.ng/verify-email?token=%s", token)
	return SendEmail(to, subject, body)
}

func SendPasswordResetEmail(to string, token string) error {
	subject := "Reset your password - ResultsPro"
	body := fmt.Sprintf("Reset your password by clicking the link below:\r\nhttps://auth.resultspro.ng/reset-password?token=%s", token)
	return SendEmail(to, subject, body)
}
