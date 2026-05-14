package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type ResendPayload struct {
	From    string `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

func SendEmail(to string, subject string, htmlBody string) error {
	apiKey := os.Getenv("RESEND_API_KEY")
	from := os.Getenv("SMTP_FROM") // We'll keep this variable name for consistency or rename it

	if apiKey == "" {
		return fmt.Errorf("RESEND_API_KEY is not set")
	}

	payload := ResendPayload{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    htmlBody,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("resend api returned non-ok status: %d", resp.StatusCode)
	}

	return nil
}

func SendVerificationEmail(to string, token string) error {
	subject := "Verify your email - ResultsPro"
	body := fmt.Sprintf("<p>Please verify your email by clicking the link below:</p><p><a href=\"https://auth.resultspro.ng/verify-email?token=%s\">Verify Email</a></p>", token)
	return SendEmail(to, subject, body)
}

func SendPasswordResetEmail(to string, token string) error {
	subject := "Reset your password - ResultsPro"
	body := fmt.Sprintf("<p>Reset your password by clicking the link below:</p><p><a href=\"https://auth.resultspro.ng/reset-password?token=%s\">Reset Password</a></p>", token)
	return SendEmail(to, subject, body)
}
