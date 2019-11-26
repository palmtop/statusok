package notify

//Inspired from https://github.com/zbindenren/logrus_mail
import (
	"fmt"
	"bytes"
	"net"
	"net/mail"
	"net/smtp"
	"strconv"
	"time"
)

type MailNotify struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Host     string `json:"smtpHost"`
	Port     int    `json:"port"`
	Subject  string `json:"subject"`
	SenderName string `json:"senderName"`
	From     string `json:"from"`
	ReceiverName string `json:"receiverName"`
	To       string `json:"to"`
}

var (
	isAuthorized bool
	client       *smtp.Client
)

func (mailNotify MailNotify) GetClientName() string {
	return "Smtp Mail"
}

func (mailNotify MailNotify) Initialize() error {

	// Check if server listens on that port.
	if len(mailNotify.Username) == 0 && len(mailNotify.Password) == 0 {
		isAuthorized = false

		conn, err := smtp.Dial(mailNotify.Host + ":" + strconv.Itoa(mailNotify.Port))

		if err != nil {
			return err
		}

		client = conn

	} else {
		isAuthorized = true
		conn, err := net.DialTimeout("tcp", mailNotify.Host+":"+strconv.Itoa(mailNotify.Port), 3*time.Second)
		if err != nil {
			return err
		}
		if conn != nil {
			defer conn.Close()
		}
	}
	// Validate sender and recipient
	_, err := mail.ParseAddress(mailNotify.From)
	if err != nil {
		return err
	}
	_, err = mail.ParseAddress(mailNotify.To)
	//TODO: validate port and email host
	if err != nil {
		return err
	}

	return nil
}

func (mailNotify MailNotify) SendResponseTimeNotification(responseTimeNotification ResponseTimeNotification) error {
	if isAuthorized {

		auth := smtp.PlainAuth("", mailNotify.Username, mailNotify.Password, mailNotify.Host)

		message := getMessageFromResponseTimeNotification(responseTimeNotification)
		mail_message := getMailMessageFromMessage(message)

		// Connect to the server, authenticate, set the sender and recipient,
		// and send the email all in one step.
		err := smtp.SendMail(
			mailNotify.Host+":"+strconv.Itoa(mailNotify.Port),
			auth,
			mailNotify.From,
			[]string{mailNotify.To},
			bytes.NewBufferString(mail_message).Bytes(),
		)

		if err != nil {
			return err
		}
		return nil
	} else {
		wc, err := client.Data()

		if err != nil {
			return err
		}

		defer wc.Close()

		message := bytes.NewBufferString(getMessageFromResponseTimeNotification(responseTimeNotification))

		if _, err = message.WriteTo(wc); err != nil {
			return err
		}

		return nil
	}
}

func (mailNotify MailNotify) SendErrorNotification(errorNotification ErrorNotification) error {
	if isAuthorized {

		auth := smtp.PlainAuth("", mailNotify.Username, mailNotify.Password, mailNotify.Host)

		message := getMessageFromErrorNotification(errorNotification)
		mail_message := getMailMessageFromMessage(message)

		// Connect to the server, authenticate, set the sender and recipient,
		// and send the email all in one step.
		err := smtp.SendMail(
			mailNotify.Host+":"+strconv.Itoa(mailNotify.Port),
			auth,
			mailNotify.From,
			[]string{mailNotify.To},
			bytes.NewBufferString(mail_message).Bytes(),
		)
		if err != nil {
			return err
		}
		return nil
	} else {
		wc, err := client.Data()

		if err != nil {
			return err
		}

		defer wc.Close()

		message := bytes.NewBufferString(getMessageFromErrorNotification(errorNotification))

		if _, err = message.WriteTo(wc); err != nil {
			return err
		}

		return nil
	}
}

func (mailNotify MailNotify) getMailMessageFromMessage(message string) string {
	mail_from_header := ""
	mail_to_header := ""
	mail_subject_header := ""

	if len(mailNotify.SenderName) > 0 {
		mail_from_header := fmt.Sprintf("From: %v <%v>\n", mailNotify.SenderName, mailNotify.From)
	} else {
		mail_from_header := fmt.Sprintf("From: %v\n", mailNotify.From)
	}

	if len(mailNotify.ReceiverName) > 0 {
		mail_to_header := fmt.Sprintf("To: %v <%v>\n", mailNotify.ReceiverName, mailNotify.To)	
	} else {
		mail_to_header := fmt.Sprintf("To: %v\n", mailNotify.To)
	}
	
	if len(mailNotify.Subject) > 0 {
		mail_subject_header := fmt.Sprintf("Subject: %v\n", mailNotify.Subject)
	}

	mail_content := fmt.Sprintf("%v%v%v\n%v", mail_from_header, mail_to_header, mail_subject_header, message)

	return mail_content
}