package mailer

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Hickar/gin-rush/internal/config"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var _mailer *Mailer

type Mailer struct {
	GmailService *gmail.Service
}

func NewMailer(conf *config.GmailConfig) (*Mailer, error) {
	oauthConf := oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		Endpoint:     google.Endpoint,
		RedirectURL:  conf.RedirectUrl,
	}

	token := oauth2.Token{
		AccessToken:  conf.AccessToken,
		RefreshToken: conf.RefreshToken,
		TokenType:    "BEARER",
		Expiry:       time.Now(),
	}

	tokenSource := oauthConf.TokenSource(context.Background(), &token)

	srv, err := gmail.NewService(context.Background(), option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, err
	}

	if srv == nil {
		return nil, errors.New("GmailService is nil")
	}

	_mailer = &Mailer{GmailService: srv}
	return _mailer, nil
}

func GetMailer() *Mailer {
	return _mailer
}

func (m *Mailer) SendConfirmationCode(username, email, code string) error {
	challengeLink := os.Getenv("API_HOST")+"/authorize/email/challenge/"+code
	body := fmt.Sprintf("Hello <b>%s</b>!<br/>In order to verify your account, please proceed to following link: <a href=\"%s\">%s</a>", username, challengeLink, challengeLink)

	return m.SendMail(email, "Account verification", body)
}

func (m *Mailer) SendMail(to, subject, body string) error {
	var message gmail.Message

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte("To: "+to+"\n"+"Subject: "+subject + "\n" + mime + body)
	message.Raw = base64.URLEncoding.EncodeToString(msg)

	_, err := m.GmailService.Users.Messages.Send("me", &message).Do()
	if err != nil {
		errMsg := errors.Unwrap(err).Error()
		return fmt.Errorf("error during sending mail: %s", errMsg)
	}

	return nil
}