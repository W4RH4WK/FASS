package fass

import (
	"net/smtp"
)

// Mail addresses are represented as strings.
type Mail = string

func DistributeToken(token Token, to Mail, course Course, config Config) error {
	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + course.Identifier + " - FASS Token\r\n" +
		"\r\n" +
		"Here is your token for the " + course.Name + " course:\r\n" +
		"\r\n" +
		token + "\r\n")

	var auth smtp.Auth = nil
	if config.MailUseAuth {
		auth = smtp.PlainAuth(config.MailAuthIdent, config.MailAuthUser, config.MailAuthPass, config.MailHost)
	}

	return smtp.SendMail(config.MailHost, auth, config.MailFrom, []string{to}, msg)
}
