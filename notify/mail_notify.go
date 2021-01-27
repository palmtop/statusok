package notify

//Inspired from https://github.com/zbindenren/logrus_mail
import (
	"bytes"
	"fmt"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"
)

//Structure containing the needed fields for mail notification
type MailNotify struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	Host         string `json:"smtpHost"`
	Port         int    `json:"port"`
	Subject      string `json:"subject"`
	SenderName   string `json:"senderName"`
	From         string `json:"from"`
	ReceiverName string `json:"receiverName"`
	To           string `json:"to"`
}

var (
	isAuthorized bool
)

//GetClientName returns the name of the notify client
func (mailNotify MailNotify) GetClientName() string {
	return "Smtp Mail"
}

//Initialize checks if the parameters for the SMTP server are correct
func (mailNotify MailNotify) Initialize() error {
	// If both username and password is given, then use them to authorize mail sending
	isAuthorized = len(mailNotify.Username) > 0 && len(mailNotify.Password) > 0

	// Check if server listens on that port. (Try to connect to the TCP port, then release it)
	conn, err := net.DialTimeout("tcp", mailNotify.Host+":"+strconv.Itoa(mailNotify.Port), 3*time.Second)
	if err != nil {
		return err
	}
	if conn != nil {
		defer conn.Close()
	}

	// Validate sender and recipient
	_, err = mail.ParseAddress(mailNotify.From)
	if err != nil {
		return err
	}
	_, err = mail.ParseAddress(mailNotify.To)

	if err != nil {
		return err
	}

	return nil
}

//SendResponseTimeNotification send a Response time notification over SMTP
func (mailNotify MailNotify) SendResponseTimeNotification(responseTimeNotification ResponseTimeNotification) error {

	message := getMessageFromResponseTimeNotification(responseTimeNotification)
	return SendMailMessage(message, mailNotify)

}

//SendErrorNotification send an Error notification over SMTP
func (mailNotify MailNotify) SendErrorNotification(errorNotification ErrorNotification) error {
	message := getMessageFromErrorNotification(errorNotification)
	return SendMailMessage(message, mailNotify)

}

//SendMailMessage - sends the mail message, using the method dependent of authorization
func SendMailMessage(message string, mailNotify MailNotify) error {
	mailMessage := getMailMessageFromMessage(message, mailNotify)
	if isAuthorized {
		auth := smtp.PlainAuth("", mailNotify.Username, mailNotify.Password, mailNotify.Host)
		return smtp.SendMail(
			mailNotify.Host+":"+strconv.Itoa(mailNotify.Port),
			auth,
			mailNotify.From,
			[]string{mailNotify.To},
			bytes.NewBufferString(mailMessage).Bytes(),
		)
	} else {
		return SendMail(mailNotify.Host+":"+strconv.Itoa(mailNotify.Port),
			mailNotify.From,
			mailMessage,
			[]string{mailNotify.To},
		)
	}

}

//getMailMessageFromMessage creates the SMTP message body, by creating the needed
//headers (To, From, Subject) and adding the message body
func getMailMessageFromMessage(message string, mailNotify MailNotify) string {
	mailFromHeader := ""
	mailToHeader := ""
	mailSubjectHeader := ""

	if len(mailNotify.SenderName) > 0 {
		mailFromHeader = fmt.Sprintf("From: %v <%v>\n", mailNotify.SenderName, mailNotify.From)
	} else {
		mailFromHeader = fmt.Sprintf("From: %v\n", mailNotify.From)
	}

	if len(mailNotify.ReceiverName) > 0 {
		mailToHeader = fmt.Sprintf("To: %v <%v>\n", mailNotify.ReceiverName, mailNotify.To)
	} else {
		mailToHeader = fmt.Sprintf("To: %v\n", mailNotify.To)
	}

	if len(mailNotify.Subject) > 0 {
		mailSubjectHeader = fmt.Sprintf("Subject: %v\n", mailNotify.Subject)
	}

	return fmt.Sprintf("%v%v%v\n%v", mailFromHeader, mailToHeader, mailSubjectHeader, message)

}

//SendMail sends an email message to an unauthorized SMTP server
//The body is the SMTP message body, including the headers
func SendMail(addr, from, body string, to []string) error {
	r := strings.NewReplacer("\r\n", "", "\r", "", "\n", "", "%0a", "", "%0d", "")

	c, err := smtp.Dial(addr)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Mail(r.Replace(from)); err != nil {
		return err
	}
	for i := range to {
		to[i] = r.Replace(to[i])
		if err = c.Rcpt(to[i]); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(body))
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}
	return c.Quit()
}
