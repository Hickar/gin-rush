package mailer

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var GmailService *gmail.Service

func Setup() error {
	config := oauth2.Config{
		ClientID:     os.Getenv("GMAIL_CLIENT_ID"),
		ClientSecret: os.Getenv("GMAIL_CLIENT_SECRET"),
		Endpoint:     google.Endpoint,
		RedirectURL:  os.Getenv("API_HOST"),
	}

	token := oauth2.Token{
		AccessToken:  os.Getenv("GMAIL_ACCESS_TOKEN"),
		RefreshToken: os.Getenv("GMAIL_REFRESH_TOKEN"),
		TokenType:    "BEARER",
		Expiry:       time.Now(),
	}

	tokenSource := config.TokenSource(context.Background(), &token)

	srv, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return err
	}

	GmailService = srv
	if GmailService == nil {
		return errors.New("GmailService is nil")
	}

	return nil
}

func SendConfirmationCode(username, email string, code int) error {
	challengeLink := os.Getenv("API_HOST") + "/authorize/email/challenge/" + strconv.Itoa(code)
	body := fmt.Sprintf("Hello <b>%s</b>!<br/>In order to verify your account, please proceed to following link: %s", username, challengeLink)

	return SendMail(email, "Account verification", body)
}

func SendMail(to, subject, body string) error {
	var message gmail.Message

	mime := "MIME-version: 1.0;\nContent-Type: text/plain; charset=\"UTF-8\";\n\n"
	msg := []byte(to + "\n" + subject + "\n" + mime + body)
	message.Raw = base64.URLEncoding.EncodeToString(msg)

	_, err := GmailService.Users.Messages.Send("me", &message).Do()
	return err
}
