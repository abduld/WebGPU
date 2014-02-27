package controllers

import (
	. "wb/app/config"

	"github.com/abduld/go-mailgun"
)

var MailClient *mailgun.Client = nil

type mail struct {
	from      string
	to        []string
	cc        []string
	bcc       []string
	subject   string
	html      string
	text      string
	headers   map[string]string
	options   map[string]string
	variables map[string]string
}

func (m *mail) From() string                 { return m.from }
func (m *mail) To() []string                 { return m.to }
func (m *mail) Cc() []string                 { return m.cc }
func (m *mail) Bcc() []string                { return m.bcc }
func (m *mail) Subject() string              { return m.subject }
func (m *mail) Html() string                 { return m.html }
func (m *mail) Text() string                 { return m.text }
func (m *mail) Headers() map[string]string   { return m.headers }
func (m *mail) Options() map[string]string   { return m.options }
func (m *mail) Variables() map[string]string { return m.variables }

func SendEmail(from string, to string, subject string, text string) error {
	if MailClient == nil {
		MailClient = mailgun.New(MailgunKey)
	}
	m := &mail{
		from:    from,
		to:      []string{to},
		subject: subject,
		text:    text,
	}
	_, err := MailClient.Send(m)
	return err
}
