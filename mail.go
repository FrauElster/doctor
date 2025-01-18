package main

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"gitlab.com/tozd/go/errors"
)

// loginAuth is a custom authentication mechanism that implements LOGIN auth
type loginAuth struct {
	username string
	password string
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

func (a *loginAuth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", nil, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if !more {
		return nil, nil
	}

	prompt := string(fromServer)
	switch prompt {
	case "Username:", "User Name:", "Username", "User:", "Identity:":
		return []byte(a.username), nil
	case "Password:", "Password":
		return []byte(a.password), nil
	default:
		if strings.Contains(prompt, "Username") {
			return []byte(a.username), nil
		}
		if strings.Contains(prompt, "Password") {
			return []byte(a.password), nil
		}
		return nil, fmt.Errorf("unknown prompt from server: %q", prompt)
	}
}

// Rest of EmailConfig remains the same
type EmailConfig struct {
	From         string   `json:"from"`
	Password     string   `json:"password"`
	SMTPHost     string   `json:"smtpHost"`
	SMTPPort     string   `json:"smtpPort"`
	ToEmails     []string `json:"toEmails"`
	AuthType     string   `json:"authType,omitempty"`
	UseTLS       bool     `json:"useTLS,omitempty"`
	StartTLSAuth bool     `json:"startTLSAuth,omitempty"`
}

func (c EmailConfig) getAuth() smtp.Auth {
	switch c.AuthType {
	case "login":
		return LoginAuth(c.From, c.Password)
	default:
		// Fallback to LOGIN auth since that's what the server supports
		return LoginAuth(c.From, c.Password)
	}
}

func (c EmailConfig) sendMail(subject, body string) error {
	message := []byte(fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n",
		strings.Join(c.ToEmails, ","),
		subject,
		body))

	auth := c.getAuth()
	addr := c.SMTPHost + ":" + c.SMTPPort

	if c.UseTLS {
		return c.sendMailTLS(auth, message)
	} else if c.StartTLSAuth {
		return c.sendMailStartTLS(auth, message)
	}

	return smtp.SendMail(addr, auth, c.From, c.ToEmails, message)
}

func (c EmailConfig) sendMailTLS(auth smtp.Auth, message []byte) error {
	addr := c.SMTPHost + ":" + c.SMTPPort

	tlsConfig := &tls.Config{
		ServerName: c.SMTPHost,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return errors.Errorf("failed to create TLS connection: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, c.SMTPHost)
	if err != nil {
		return errors.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	return c.sendMailWithClient(client, auth, message)
}

// Rest of the code remains the same...
func (c EmailConfig) sendMailStartTLS(auth smtp.Auth, message []byte) error {
	addr := c.SMTPHost + ":" + c.SMTPPort

	client, err := smtp.Dial(addr)
	if err != nil {
		return errors.Errorf("failed to dial SMTP server: %w", err)
	}
	defer client.Close()

	if err := client.StartTLS(&tls.Config{
		ServerName: c.SMTPHost,
		MinVersion: tls.VersionTLS12,
	}); err != nil {
		return errors.Errorf("failed to start TLS: %w", err)
	}

	return c.sendMailWithClient(client, auth, message)
}

func (c EmailConfig) sendMailWithClient(client *smtp.Client, auth smtp.Auth, message []byte) error {
	if err := client.Auth(auth); err != nil {
		return errors.Errorf("failed to authenticate: %w", err)
	}

	if err := client.Mail(c.From); err != nil {
		return errors.Errorf("failed to set FROM address: %w", err)
	}

	for _, addr := range c.ToEmails {
		if err := client.Rcpt(addr); err != nil {
			return errors.Errorf("failed to set TO address: %w", err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return errors.Errorf("failed to open mail writer: %w", err)
	}
	defer w.Close()

	_, err = w.Write(message)
	if err != nil {
		return errors.Errorf("failed to write mail content: %w", err)
	}

	return nil
}

func NewEmailAlert(config EmailConfig) AlertFunc {
	return func(target HealthTarget, result Result) error {
		if result.Healthy {
			return nil
		}

		subject := fmt.Sprintf("Health Check Alert: %s is DOWN", target.URLString)
		body := fmt.Sprintf(`Health Check Failed for %s

Details:
- Target ID: %s
- URL: %s
- Status Code: %d
- Timestamp: %s
- Duration: %s
`,
			target.URLString,
			target.ID,
			target.URLString,
			result.Status,
			result.Timestamp.Format(time.RFC3339),
			result.Duration.String(),
		)

		if result.Error != nil {
			body += fmt.Sprintf("- Error: %v\n", result.Error)
		}

		return config.sendMail(subject, body)
	}
}

func NewEmailResolve(config EmailConfig) AlertFunc {
	return func(target HealthTarget, result Result) error {
		if !result.Healthy {
			return nil
		}

		subject := fmt.Sprintf("Health Check Resolved: %s is UP", target.URLString)
		body := fmt.Sprintf(`Health Check Recovered for %s

Details:
- Target ID: %s
- URL: %s
- Status Code: %d
- Timestamp: %s
- Duration: %s
- Status: RESOLVED
`,
			target.URLString,
			target.ID,
			target.URLString,
			result.Status,
			result.Timestamp.Format(time.RFC3339),
			result.Duration.String(),
		)

		return config.sendMail(subject, body)
	}
}
