package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	netsmtp "net/smtp"
	"strings"
	"time"
)

type Provider interface {
	Send(ctx context.Context, msg Message) error
	Name() string
}

type Message struct {
	From    string
	To      string
	Subject string
	HTML    string
}

type Config struct {
	Host           string
	Port           int
	Username       string
	Password       string
	From           string
	UseTLS         bool
	UseStartTLS    bool
	SkipTLSVerify  bool
	ConnectTimeout time.Duration
	SendTimeout    time.Duration
}

type SMTPProvider struct {
	cfg  Config
	name string
}

func NewSMTPProvider(name string, cfg Config) *SMTPProvider {
	return &SMTPProvider{name: name, cfg: cfg}
}

func (p *SMTPProvider) Name() string { return p.name }

func (p *SMTPProvider) Send(ctx context.Context, msg Message) error {
	if p.cfg.Host == "" || p.cfg.Port == 0 {
		return fmt.Errorf("%s: smtp config is empty", p.name)
	}
	if msg.To == "" {
		return fmt.Errorf("%s: empty recipient", p.name)
	}

	from := msg.From
	if from == "" {
		from = p.cfg.From
	}
	if from == "" {
		return fmt.Errorf("%s: empty from", p.name)
	}

	addr := fmt.Sprintf("%s:%d", p.cfg.Host, p.cfg.Port)

	dialer := &net.Dialer{}
	if p.cfg.ConnectTimeout > 0 {
		dialer.Timeout = p.cfg.ConnectTimeout
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return fmt.Errorf("%s: dial %s: %w", p.name, addr, err)
	}
	defer conn.Close()

	if p.cfg.SendTimeout > 0 {
		_ = conn.SetDeadline(time.Now().Add(p.cfg.SendTimeout))
	}

	c, err := netsmtp.NewClient(conn, p.cfg.Host)
	if err != nil {
		return fmt.Errorf("%s: new client: %w", p.name, err)
	}
	defer c.Quit()

	if err := c.Hello("localhost"); err != nil {
		return fmt.Errorf("%s: hello: %w", p.name, err)
	}

	if p.cfg.UseStartTLS {
		if ok, _ := c.Extension("STARTTLS"); ok {
			tlsCfg := &tls.Config{
				ServerName:         p.cfg.Host,
				InsecureSkipVerify: p.cfg.SkipTLSVerify,
			}
			if err := c.StartTLS(tlsCfg); err != nil {
				return fmt.Errorf("%s: starttls: %w", p.name, err)
			}

			if err := c.Hello("localhost"); err != nil {
				return fmt.Errorf("%s: hello after starttls: %w", p.name, err)
			}
		}
	}

	if p.cfg.Username != "" {
		auth := netsmtp.PlainAuth("", p.cfg.Username, p.cfg.Password, p.cfg.Host)
		if err := c.Auth(auth); err != nil {
			return fmt.Errorf("%s: auth: %w", p.name, err)
		}
	}

	if err := c.Mail(from); err != nil {
		return fmt.Errorf("%s: MAIL FROM: %w", p.name, err)
	}
	if err := c.Rcpt(msg.To); err != nil {
		return fmt.Errorf("%s: RCPT TO: %w", p.name, err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("%s: DATA: %w", p.name, err)
	}
	defer w.Close()

	raw := buildMIME(from, msg.To, msg.Subject, msg.HTML)

	if _, err := w.Write([]byte(raw)); err != nil {
		return fmt.Errorf("%s: write: %w", p.name, err)
	}

	return nil
}

func buildMIME(from, to, subject, html string) string {
	var b strings.Builder
	b.WriteString("From: " + from + "\r\n")
	b.WriteString("To: " + to + "\r\n")
	b.WriteString("Subject: " + encodeHeader(subject) + "\r\n")
	b.WriteString("MIME-Version: 1.0\r\n")
	b.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	b.WriteString("Content-Transfer-Encoding: 8bit\r\n")
	b.WriteString("\r\n")
	b.WriteString(html)
	return b.String()
}

func encodeHeader(s string) string {
	return s
}
