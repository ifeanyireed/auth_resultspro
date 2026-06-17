package config

import (
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/microsoft"
)

var GoogleOAuthConfig *oauth2.Config
var MicrosoftOAuthConfig *oauth2.Config

func InitConfig() {
	GoogleOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.profile",
			"https://www.googleapis.com/auth/userinfo.email",
		},
		Endpoint: google.Endpoint,
	}

	MicrosoftOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("MICROSOFT_CLIENT_ID"),
		ClientSecret: os.Getenv("MICROSOFT_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("MICROSOFT_REDIRECT_URL"),
		Scopes:       []string{"User.Read"},
		Endpoint:     microsoft.AzureADEndpoint("common"),
	}
}
