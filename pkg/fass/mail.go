package fass

import "net/smtp"

// Mail addresses are represented as strings.
type Mail = string

func DistributeToken(token Token, to Mail, course Course) error {
	const host = "localhost"
	const from = "fass@fass.dps.uibk.ac.at"

	msg := []byte("To: " + to + "\r\n" +
		"Subject: " + course.Identifier + " - FASS Token\r\n" +
		"\r\n" +
		"Here is your token for the " + course.Name + " course:\r\n" +
		"\r\n" +
		token + "\r\n")

	return smtp.SendMail("localhost", nil, from, []string{to}, msg)
}
