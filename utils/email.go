package utils

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sesv2"
	"github.com/aws/aws-sdk-go-v2/service/sesv2/types"
)

type SESClientAPI interface {
	SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error)
}

var (
	sesClient SESClientAPI
	once      sync.Once
)

func SetSESClient(client SESClientAPI) {
	sesClient = client
}

func getSESClient() (SESClientAPI, error) {
	var err error
	once.Do(func() {
		if sesClient != nil {
			return
		}
		region := os.Getenv("AWS_REGION")
		if region == "" {
			region = "us-east-1"
		}
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
	return sesClient, nil
}

func SendEmail(to string, subject string, htmlBody string, textBody string) error {
	client, err := getSESClient()
	if err != nil {
		return err
	}
	from := os.Getenv("SMTP_FROM")
	if from == "" {
		from = "noreply@resultspro.ng" // Fallback
	}
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
					Text: &types.Content{
						Data: aws.String(textBody),
					},
				},
			},
		},
	}
	_, err = client.SendEmail(context.TODO(), input)
	return err
}

func SendVerificationEmail(to string, otp string) error {
	subject := otp + " is your ResultsPRO verification code"
	textBody := fmt.Sprintf("Your ResultsPRO verification code is: %s. This code will expire in 24 hours.", otp)
	htmlBody := fmt.Sprintf(`
<div style="font-family: 'Google Sans', Roboto, RobotoDraft, Helvetica, Arial, sans-serif; background-color: #ffffff; padding: 40px; margin: 0; color: #3c4043; border: 1px solid #dadce0; border-radius: 8px; max-width: 600px;">
  <div style="margin-bottom: 24px; display: flex; align-items: center;">
    <img src="https://resultspro.ng/logo.png" alt="ResultsPRO" style="height: 32px; margin-right: 4px; border-radius: 6px;">
	<div style="line-height: 1.1;">
	  <div style="font-size: 18px; font-weight: 700; color: #202124; letter-spacing: -0.01em; font-family: sans-serif;">ResultsPRO</div>
	  <div style="font-size: 9px; font-weight: 700; color: #1a73e8; text-transform: uppercase; letter-spacing: 0.15em; margin-top: 2px; font-family: sans-serif;">SUITE</div>
	</div>
  </div>
  <h1 style="font-size: 24px; font-weight: 400; color: #202124; margin-bottom: 24px; margin-top: 0;">Verify your email address</h1>
  <p style="font-size: 14px; line-height: 20px; margin-bottom: 24px;">
    To finish setting up your ResultsPRO account, we need to make sure this email address is yours.
  </p>
  <p style="font-size: 14px; line-height: 20px; margin-bottom: 8px;">
    Use this verification code to complete your signup:
  </p>
  <div style="background-color: #f8f9fa; border-radius: 4px; padding: 16px; text-align: center; margin-bottom: 24px;">
    <span style="font-size: 32px; font-weight: 500; letter-spacing: 8px; color: #1a73e8;">%s</span>
  </div>
  <p style="font-size: 14px; line-height: 20px; margin-bottom: 24px;">
    This code will expire in 24 hours. If you didn't request this code, you can safely ignore this email.
  </p>
  <hr style="border: none; border-top: 1px solid #dadce0; margin-bottom: 24px;">
  <p style="font-size: 12px; line-height: 16px; color: #70757a;">
    You received this email because it was used to register for ResultsPRO Suite. 
    If you're not sure why you're receiving this, please contact support.
  </p>
</div>`, otp)
	return SendEmail(to, subject, htmlBody, textBody)
}

func SendPasswordResetEmail(to string, token string, resetURL string) error {
	if resetURL == "" {
		resetURL = "https://auth.resultspro.ng/reset-password"
	}

	link := fmt.Sprintf("%s?token=%s", resetURL, token)
	if strings.Contains(resetURL, "?") {
		link = fmt.Sprintf("%s&token=%s", resetURL, token)
	}

	subject := "Reset your password - ResultsPRO Suite"
	textBody := fmt.Sprintf("Reset your password by visiting: %s", link)
	htmlBody := fmt.Sprintf(`
<div style="font-family: 'Google Sans', Roboto, RobotoDraft, Helvetica, Arial, sans-serif; background-color: #ffffff; padding: 40px; margin: 0; color: #3c4043; border: 1px solid #dadce0; border-radius: 8px; max-width: 600px;">
  <div style="margin-bottom: 24px; display: flex; align-items: center;">
    <img src="https://resultspro.ng/logo.png" alt="ResultsPRO" style="height: 32px; margin-right: 4px; border-radius: 6px;">
    <div style="display: flex; flex-direction: column; line-height: 1;">
      <span style="font-size: 18px; font-weight: 700; color: #202124; letter-spacing: -0.01em;">ResultsPRO</span>
      <span style="font-size: 9px; font-weight: 700; color: #1a73e8; text-transform: uppercase; letter-spacing: 0.15em; margin-top: 2px;">SUITE</span>
    </div>
  </div>
  <h1 style="font-size: 24px; font-weight: 400; color: #202124; margin-bottom: 24px; margin-top: 0;">Reset your password</h1>
  <p style="font-size: 14px; line-height: 20px; margin-bottom: 24px;">
    We received a request to reset your ResultsPRO Suite password. Click the link below to choose a new one:
  </p>
  <div style="text-align: center; margin-bottom: 24px;">
    <a href="%s" style="background-color: #1a73e8; color: #ffffff; padding: 12px 24px; border-radius: 4px; text-decoration: none; font-weight: 500; display: inline-block;">Reset Password</a>
  </div>
  <p style="font-size: 14px; line-height: 20px; margin-bottom: 24px;">
    If you didn't request a password reset, you can safely ignore this email. This link will expire in 1 hour.
  </p>
  <hr style="border: none; border-top: 1px solid #dadce0; margin-bottom: 24px;">
  <p style="font-size: 12px; line-height: 16px; color: #70757a;">
    You received this email because a password reset was requested for your ResultsPRO Suite account.
  </p>
</div>`, link)
	return SendEmail(to, subject, htmlBody, textBody)
}
