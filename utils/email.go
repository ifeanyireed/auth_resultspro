package utils

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

func SendEmail(to string, subject string, htmlBody string) error {
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1" // Default region
	}

	// Load AWS configuration (uses default credential chain)
	// Requires AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY in env
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %v", err)
	}

	client := sesv2.NewFromConfig(cfg)

	from := os.Getenv("SMTP_FROM")

	input := &sesv2.SendEmailInput{
		FromEmailAddress: aws.String(from),
		Destination: &types.Destination{
			ToAddresses: []string{to},
		},
		Content: &types.EmailContent{
			Simple: &types.Message{
				Subject: &types.Content{
					Data: aws.String(subject),
				},
				Body: &types.Body{
					Html: &types.Content{
						Data: aws.String(htmlBody),
					},
				},
			},
		},
	}

	_, err = client.SendEmail(context.TODO(), input)
	if err != nil {
		return fmt.Errorf("failed to send email via SES: %v", err)
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
