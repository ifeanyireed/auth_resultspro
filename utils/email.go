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
        subject := otp + " is your ResultsPro verification code"
        textBody := fmt.Sprintf("Your ResultsPro verification code is: %s. This code will expire in 24 hours.", otp)
        htmlBody := fmt.Sprintf(`
<div style="font-family: 'Google Sans',Roboto,RobotoDraft,Helvetica,Arial,sans-serif; background-color: #ffffff; padding: 40px; margin: 0; color: #3c4043; border: 1px solid #dadce0; border-radius: 8px; max-width: 600px;">
  <img src="https://resultspro.ng/logo.png" alt="ResultsPro" style="height: 32px; margin-bottom: 24px;">
  <h1 style="font-size: 24px; font-weight: 400; color: #202124; margin-bottom: 24px; margin-top: 0;">Verify your email address</h1>
  <p style="font-size: 14px; line-height: 20px; margin-bottom: 24px;">
    To finish setting up your ResultsPro account, we need to make sure this email address is yours.
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
    You received this email because it was used to register for ResultsPro. 
    If you're not sure why you're receiving this, please contact support.
  </p>
</div>`, otp)
        return SendEmail(to, subject, htmlBody, textBody)
}

func SendPasswordResetEmail(to string, token string) error {
        subject := "Reset your password - ResultsPro"
        textBody := fmt.Sprintf("Reset your password by visiting: https://classroompro.com/reset-password?token=%s", token)
        htmlBody := fmt.Sprintf("<p>Reset your password by clicking the link below:</p><p><a href=\"https://classroompro.com/reset-password?token=%s\">Reset Password</a></p>", token)
        return SendEmail(to, subject, htmlBody, textBody)
}
