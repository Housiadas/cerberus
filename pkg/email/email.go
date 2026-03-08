// Package email provides an SMTP email sender.
package email

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
)

// ErrTLSConnType is returned when the dialer returns an unexpected connection type.
var ErrTLSConnType = errors.New("tls dial: unexpected connection type")

// Config holds the SMTP configuration.
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	TLS      bool
}

// Client sends emails via SMTP.
type Client struct {
	cfg Config
}

// New constructs an email Client.
func New(cfg Config) *Client {
	return &Client{cfg: cfg}
}

// Send delivers an email to the given recipient.
func (c *Client) Send(ctx context.Context, to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", c.cfg.Host, c.cfg.Port)

	msg := buildMessage(c.cfg.From, to, subject, body)

	if c.cfg.TLS {
		return c.sendTLS(ctx, addr, to, msg)
	}

	return c.sendPlain(addr, to, msg)
}

func (c *Client) sendPlain(addr, to string, msg []byte) error {
	var auth smtp.Auth
	if c.cfg.Username != "" {
		auth = smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)
	}

	err := smtp.SendMail(addr, auth, c.cfg.From, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("smtp send: %w", err)
	}

	return nil
}

func (c *Client) sendTLS(ctx context.Context, addr, to string, msg []byte) error {
	tlsCfg := &tls.Config{
		ServerName: c.cfg.Host,
		MinVersion: tls.VersionTLS12,
	}

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{},
		Config:    tlsCfg,
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("tls dial: %w", err)
	}
	defer conn.Close()

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return ErrTLSConnType
	}

	client, err := smtp.NewClient(tlsConn, c.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp client: %w", err)
	}
	defer client.Close()

	if c.cfg.Username != "" {
		auth := smtp.PlainAuth("", c.cfg.Username, c.cfg.Password, c.cfg.Host)

		err = client.Auth(auth)
		if err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	err = client.Mail(c.cfg.From)
	if err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}

	err = client.Rcpt(to)
	if err != nil {
		return fmt.Errorf("smtp rcpt to: %w", err)
	}

	wc, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	defer wc.Close()

	_, err = wc.Write(msg)
	if err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}

	return nil
}

func buildMessage(from, to, subject, body string) []byte {
	headers := fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n",
		from,
		to,
		subject,
	)

	return []byte(headers + body)
}
