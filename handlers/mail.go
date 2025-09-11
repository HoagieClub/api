package handlers

import (
	"encoding/json"
	"fmt"
	"hoagie-profile/auth"
	"hoagie-profile/db"
	"net/http"
	"os"
	"time"

	mailjet "github.com/mailjet/mailjet-apiv3-go"
	bluemonday "github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
)

// BlueMonday sanitizes HTML, preventing unsafe user input
var p = bluemonday.UGCPolicy()

const NORMAL_EMAIL_FOOTER = `<hr />` +
	`<div style="font-size:8pt;">This email was instantly sent to all ` +
	`college listservs with <a href="https://mail.hoagie.io/">Hoagie Mail</a>. ` +
	`Email composed by %s (%s) — if you believe this email is offensive, ` +
	`intentionally misleading or harmful, please report it to ` +
	`<a href="mailto:hoagie@princeton.edu">hoagie@princeton.edu</a>.</div>`

const TEST_EMAIL_FOOTER = `<hr />` +
	`<div style="font-size:8pt;">This test email was instantly sent only ` +
	`to you with <a href="https://mail.hoagie.io/">Hoagie Mail</a>. ` +
	`Email composed by %s (%s).</div>`

type MailRequest struct {
	Header   string
	Sender   string
	Body     string
	Email    string
	Schedule string
}

func getListServs() *mailjet.RecipientsV31 {
	return &mailjet.RecipientsV31{
		mailjet.RecipientV31{
			Email: "BUTLERBUZZ@PRINCETON.EDU",
			Name:  "Butler",
		},
		mailjet.RecipientV31{
			Email: "WHITMANWIRE@PRINCETON.EDU",
			Name:  "Whitman",
		},
		mailjet.RecipientV31{
			Email: "RockyWire@PRINCETON.EDU",
			Name:  "Rocky",
		},
		mailjet.RecipientV31{
			Email: "Re-INNformer@PRINCETON.EDU",
			Name:  "Forbes",
		},
		mailjet.RecipientV31{
			Email: "westwire@princeton.edu",
			Name:  "NCW",
		},
		mailjet.RecipientV31{
			Email: "matheymail@PRINCETON.EDU",
			Name:  "Mathey",
		},
		mailjet.RecipientV31{
			Email: "yehyellowpages@princeton.edu",
			Name:  "Yeh",
		},
		mailjet.RecipientV31{
			Email: "hoagiemailgradstudents@princeton.edu",
			Name:  "hoagiemailgradstudents",
		},
	}
}

func createInfoMessage(req MailRequest, toEmail string) []mailjet.InfoMessagesV31 {
	return []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: "hoagie@princeton.edu",
				Name:  req.Sender,
			},
			ReplyTo: &mailjet.RecipientV31{
				Email: req.Email,
				Name:  req.Sender,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: toEmail,
					Name:  req.Sender,
				},
			},
			Subject:  req.Header,
			TextPart: req.Body,
			HTMLPart: req.Body,
			CustomID: "HoagieMail",
		},
	}
}

func sendEmail(req MailRequest) error {

	var messagesInfo []mailjet.InfoMessagesV31

	if req.Schedule != "test" {
		messagesInfo = createInfoMessage(req, "hoagie@princeton.edu")
		messagesInfo[0].Cc = getListServs()
	} else {
		messagesInfo = createInfoMessage(req, req.Email)
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}

	if os.Getenv("HOAGIE_MODE") == "debug" {
		printDebug(messages)
		return nil
	}

	mailjetClient := mailjet.NewMailjetClient(os.Getenv("MAILJET_PUBLIC_KEY"), os.Getenv("MAILJET_PRIVATE_KEY"))
	res, err := mailjetClient.SendMailV31(&messages)

	if err != nil {
		return err
	}
	if len(res.ResultsV31) > 0 && res.ResultsV31[0].Status == "success" {
		return nil
	}

	return fmt.Errorf("mail service received an error, possibly because of limits")
}

func printDebug(messages mailjet.MessagesV31) {
	fmt.Printf("Sender Name: %s\n", (*messages.Info[0].From).Name)
	fmt.Printf("From: %s\n", (*messages.Info[0].From).Email)
	fmt.Printf("ReplyTo: %s\n", (*messages.Info[0].ReplyTo).Email)
	fmt.Printf("To: %s\n", (*messages.Info[0].To)[0].Email)
	if messages.Info[0].Cc == nil {
		fmt.Printf("CC: 0 recipients\n")
	} else {
		ccRecipients := (*messages.Info[0].Cc)
		fmt.Printf("CC: %d recipients\n", len(ccRecipients))
		for _, recipient := range ccRecipients {
			fmt.Printf("\t%s\n", recipient.Email)
		}
	}
	fmt.Printf("Subject: %s\n", messages.Info[0].Subject)
	fmt.Printf("Body: %s\n", messages.Info[0].TextPart)
}

func handleScheduledEmail(w http.ResponseWriter, mailReq MailRequest, user auth.User) bool {
	// Validate that schedule is valid
	if !scheduleValid(mailReq.Schedule) {
		http.Error(
			w,
			"Your email could not be scheduled at the specified time. Please refresh the page and select a later time.",
			http.StatusBadRequest,
		)
		return false
	}

	// Convert time to EST and check for errors
	est, err := time.LoadLocation("America/New_York")
	if err != nil {
		http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
		return false
	}
	scheduleEST, err := time.ParseInLocation(time.RFC3339, mailReq.Schedule, est)
	if err != nil {
		http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
		return false
	}

	// Check that user doesn't have an already-existing entry
	currentScheduledMail, err := getScheduled(user, scheduleEST)
	if err != nil {
		http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
		return false
	}
	if currentScheduledMail != (ScheduledMail{}) {
		errString := "You already have an email scheduled for this time. If you would like to change"
		errString += " your message, please delete your mail in the Scheduled Emails page and try again."
		http.Error(w, errString, http.StatusBadRequest)
		return false
	}

	// Add to MongoDB
	db.InsertOne(client, "apps", "mail", bson.D{
		{Key: "Email", Value: mailReq.Email},
		{Key: "Sender", Value: mailReq.Sender},
		{Key: "Header", Value: mailReq.Header},
		{Key: "Body", Value: mailReq.Body},
		{Key: "Schedule", Value: scheduleEST},
		{Key: "UserName", Value: user.Name},
		{Key: "CreatedAt", Value: time.Now()},
	})
	return true
}

func handleEmailNow(w http.ResponseWriter, mailReq MailRequest, user auth.User) bool {

	// Ignore user limits when debugging
	if os.Getenv("HOAGIE_MODE") != "debug" {
		visitor := getVisitor(user.Email)

		if mailReq.Schedule == "test" {
			if !visitor.testEmailLimiter.Allow() {
				http.Error(w, "You have reached your send limit. "+
					"You can only send one test email every 1 minute.",
					http.StatusTooManyRequests)
				return false
			}
		} else if !visitor.emailLimiter.Allow() {
			http.Error(w, "You have reached your send limit. "+
				"You can only send one email every 6 hours. "+
				"If you need to send an email urgently, "+
				"please contact hoagie@princeton.edu",
				http.StatusTooManyRequests)
			return false
		}
	}

	if mailReq.Schedule != "test" {
		mailReq.Body += fmt.Sprintf(NORMAL_EMAIL_FOOTER, user.Name, mailReq.Email)
	} else {
		mailReq.Body += fmt.Sprintf(TEST_EMAIL_FOOTER, user.Name, mailReq.Email)
	}

	err := sendEmail(mailReq)

	if err != nil {
		http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
		deleteVisitor(user.Email)
		return false
	}

	if mailReq.Schedule != "test" {
		fmt.Printf("MAIL: %s sent an email with title '%s'.\n", mailReq.Email, mailReq.Header)
	}

	return true
}

// POST /mail/send
var sendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	user, success := getUser(r.Header.Get("authorization"))
	if !success {
		http.Error(w, "You do not have access to send mail.", http.StatusBadRequest)
		return
	}

	if len(user.Name) == 0 {
		http.Error(w, `Hoagie Mail has been updated. Please log-out and log-in again.`, http.StatusBadRequest)
		return
	}

	var mailReq MailRequest
	err := json.NewDecoder(r.Body).Decode(&mailReq)
	if err != nil {
		http.Error(w, "Message did not contain correct fields.", http.StatusBadRequest)
		return
	}
	if notBetween(w, mailReq.Sender, "sender name", 3, 30) {
		return
	}
	if notBetween(w, mailReq.Header, "email subject", 3, 150) {
		return
	}

	mailReq.Body = p.Sanitize(mailReq.Body)
	mailReq.Email = user.Email

	if mailReq.Schedule != "now" && mailReq.Schedule != "test" {
		success := handleScheduledEmail(w, mailReq, user)
		if !success {
			return
		}
	} else {
		success := handleEmailNow(w, mailReq, user)
		if !success {
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Status\": \"OK\"}"))
})
