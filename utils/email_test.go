package utils

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sesv2"
)

type mockSESClient struct {
	SESClientAPI
	err    error
	input  *sesv2.SendEmailInput
}

func (m *mockSESClient) SendEmail(ctx context.Context, params *sesv2.SendEmailInput, optFns ...func(*sesv2.Options)) (*sesv2.SendEmailOutput, error) {
	m.input = params
	return &sesv2.SendEmailOutput{}, m.err
}

func TestSendEmail(t *testing.T) {
	os.Setenv("SMTP_FROM", "test@example.com")
	defer os.Unsetenv("SMTP_FROM")

	tests := []struct {
		name    string
		to      string
		subject string
		body    string
		mockErr error
		wantErr bool
	}{
		{
			name:    "successful email sending",
			to:      "recipient@example.com",
			subject: "Subject",
			body:    "<h1>Body</h1>",
			mockErr: nil,
			wantErr: false,
		},
		{
			name:    "failed email sending",
			to:      "recipient@example.com",
			subject: "Subject",
			body:    "<h1>Body</h1>",
			mockErr: errors.New("SES error"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSESClient{err: tt.mockErr}
			sesClient = mock
			err := SendEmail(tt.to, tt.subject, tt.body)
			if (err != nil) != tt.wantErr {
				t.Errorf("SendEmail() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if mock.input.Destination.ToAddresses[0] != tt.to {
					t.Errorf("expected to %s, got %s", tt.to, mock.input.Destination.ToAddresses[0])
				}
				if *mock.input.Content.Simple.Subject.Data != tt.subject {
					t.Errorf("expected subject %s, got %s", tt.subject, *mock.input.Content.Simple.Subject.Data)
				}
				if *mock.input.Content.Simple.Body.Html.Data != tt.body {
					t.Errorf("expected body %s, got %s", tt.body, *mock.input.Content.Simple.Body.Html.Data)
				}
			}
		})
	}
}

func TestSendVerificationEmail(t *testing.T) {
	os.Setenv("SMTP_FROM", "test@example.com")
	defer os.Unsetenv("SMTP_FROM")

	mock := &mockSESClient{}
	sesClient = mock
	token := "test-token"
	to := "user@example.com"
	err := SendVerificationEmail(to, token)
	if err != nil {
		t.Errorf("SendVerificationEmail() error = %v", err)
	}

	if mock.input.Destination.ToAddresses[0] != to {
		t.Errorf("expected to %s, got %s", to, mock.input.Destination.ToAddresses[0])
	}
	expectedBodySub := "token=test-token"
	if !strings.Contains(*mock.input.Content.Simple.Body.Html.Data, expectedBodySub) {
		t.Errorf("expected body to contain %s", expectedBodySub)
	}
}

func TestSendPasswordResetEmail(t *testing.T) {
	os.Setenv("SMTP_FROM", "test@example.com")
	defer os.Unsetenv("SMTP_FROM")

	mock := &mockSESClient{}
	sesClient = mock
	token := "test-token"
	to := "user@example.com"
	err := SendPasswordResetEmail(to, token)
	if err != nil {
		t.Errorf("SendPasswordResetEmail() error = %v", err)
	}

	if mock.input.Destination.ToAddresses[0] != to {
		t.Errorf("expected to %s, got %s", to, mock.input.Destination.ToAddresses[0])
	}
	expectedBodySub := "token=test-token"
	if !strings.Contains(*mock.input.Content.Simple.Body.Html.Data, expectedBodySub) {
		t.Errorf("expected body to contain %s", expectedBodySub)
	}
}
