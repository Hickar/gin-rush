package mailer

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/Hickar/gin-rush/internal/config"
	"golang.org/x/oauth2"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

var _mailer *Mailer

type ConfirmationMessage struct {
	Username string
	Email    string
	Code     string
}

type Credentials struct {
	ClientID     string   `json:"client_id"`
	ClientSecret string   `json:"client_secret"`
	RefreshToken string   `json:"refresh_token"`
	RedirectURIs []string `json:"redirect_uris"`
	AuthURI      string   `json:"auth_uri"`
	TokenURI     string   `json:"token_uri"`
}

type Mailer struct {
	GmailService *gmail.Service
}

func NewMailer(conf *config.GmailConfig) (*Mailer, error) {
	if conf == nil {
		return nil, errors.New("no mailer configuration was provided")
	}

	ctx := context.Background()

	oauthConfig := oauth2.Config{
		ClientID:     conf.ClientID,
		ClientSecret: conf.ClientSecret,
		Endpoint: oauth2.Endpoint{
			AuthURL:  conf.AuthURI,
			TokenURL: conf.TokenURI,
		},
		RedirectURL: conf.RedirectURIs[0],
		Scopes:      []string{gmail.GmailSendScope},
	}

	tokenSource := oauthConfig.TokenSource(ctx, &oauth2.Token{
		RefreshToken: conf.RefreshToken,
	})

	client := oauth2.NewClient(ctx, tokenSource)

	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
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
	challengeLink := config.GetConfig().Server.HostUrl + "/authorize/email/challenge/" + code
	body := fmt.Sprintf("Hello <b>%s</b>!<br/>In order to verify your account, please proceed to following link: <a href=\"%s\">%s</a>", username, challengeLink, challengeLink)

	return m.SendMail(email, "Account verification", body)
}

func (m *Mailer) SendMail(to, subject, body string) error {
	var message gmail.Message

	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := []byte("To: " + to + "\n" + "Subject: " + subject + "\n" + mime + body)
	message.Raw = base64.URLEncoding.EncodeToString(msg)

	_, err := m.GmailService.Users.Messages.Send("me", &message).Do()
	if err != nil {
		errMsg := errors.Unwrap(err).Error()
		return fmt.Errorf("error during sending mail: %s", errMsg)
	}

	return nil
}
