package utils

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

// SESClientAPI defines the interface for the SES client
type SESClientAPI interface {
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

var (
	sesClient SESClientAPI
	once      sync.Once
)

func getSESClient() (SESClientAPI, error) {
	var err error
	once.Do(func() {
		if sesClient != nil {
			return
		}

		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1" // Default region
		}

		// Load AWS configuration (uses default credential chain)
		// Requires AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY in env
		cfg, err2 := config.LoadDefaultConfig(context.TODO(), config.WithRegion(region))
		if err2 != nil {
			err = fmt.Errorf("unable to load SDK config: %v", err2)
			return
		}

		sesClient = sesv2.NewFromConfig(cfg)
	})

	if err != nil {
		return nil, err
	}
	if sesClient == nil {
		return nil, fmt.Errorf("ses client not initialized")
	}

	return sesClient, nil
}

// SetSESClient allows overriding the SES client (mainly for tests)
func SetSESClient(client SESClientAPI) {
	sesClient = client
}

// ResetSESClient clears the initialized client and allow re-initialization (mainly for live testing different configs)
func ResetSESClient() {
	sesClient = nil
	once = sync.Once{}
}

func SendEmail(to string, subject string, htmlBody string) error {
	client, err := getSESClient()
	if err != nil {
		return err
	}

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
