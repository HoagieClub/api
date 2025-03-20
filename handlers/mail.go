package handlers

import (
	"encoding/json"
	"fmt"
	"hoagie-profile/db"
	"net/http"
	"os"
	"time"

	mailjet "github.com/mailjet/mailjet-apiv3-go"
	bluemonday "github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/bson"
)

type MailRequest struct {
	Header   string
	Sender   string
	Body     string
	Email    string
	Schedule string
}

// BlueMonday sanitizes HTML, preventing unsafe user input
var p = bluemonday.UGCPolicy()

// makeRequest makes a request to mailjet to send the contents of req
//
//	to either all the listservs or simply the sender themselves
//	depending on whether email is marked as a testMail or not
//
// params:
//
//	req        - contents of email request
//	isTestMail - if true, req is sent back to the sender
//	             if false, req is sent to all the listservs
func makeRequest(req MailRequest, isTestMail bool) error {
	mailjetClient := mailjet.NewMailjetClient(os.Getenv("MAILJET_PUBLIC_KEY"), os.Getenv("MAILJET_PRIVATE_KEY"))

	var messagesInfo []mailjet.InfoMessagesV31
	if !isTestMail { // This request is normal and should send to all listservs
		messagesInfo = []mailjet.InfoMessagesV31{
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
						Email: req.Email,
						Name:  req.Sender,
					},
				},
				Cc: &mailjet.RecipientsV31{
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
				},
				Subject:  req.Header,
				TextPart: req.Body,
				HTMLPart: req.Body,
				CustomID: "HoagieMail",
			},
		}
	} else { // Case where email is test email
		messagesInfo = []mailjet.InfoMessagesV31{
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
						Email: req.Email,
						Name:  req.Sender,
					},
				},
				Subject:  req.Header,
				TextPart: req.Body,
				HTMLPart: req.Body,
				CustomID: "HoagieMail",
			},
		}
	} // endif !isTestMail

	messages := mailjet.MessagesV31{Info: messagesInfo}
	res, err := mailjetClient.SendMailV31(&messages)
	if err != nil {
		return err
	}
	if len(res.ResultsV31) > 0 && res.ResultsV31[0].Status == "success" {
		return nil
	}
	return fmt.Errorf("mail service received an error, possibly because of limits")
}

func sendMail(req MailRequest, isTestMail bool) error {
	err := makeRequest(req, isTestMail)
	if err != nil {
		return err
		// TODO: be better with status code handling. Most likely just == 400.
	}
	return nil
}

// userReachedLimit finds whether user has exceeded limits for either
// normal emails or test emails
// NOTE: this function has the side effect of using up a rate limiter token
// if the user has yet to exceed limits
func userReachedLimit(user string, isTestMail bool) bool {
	visitor := getVisitor(user)
	if !isTestMail {
		return !visitor.limiter.Allow()
	}
	return !visitor.testLimiter.Allow()
}

func processSendRequest(w http.ResponseWriter, r *http.Request, isTestEmail bool) {
	user, success := getUser(r.Header.Get("authorization"))
	if !success {
		http.Error(w, "You do not have access to send mail.", http.StatusBadRequest)
		return
	}

	// Ignore user limits when debugging
	if os.Getenv("HOAGIE_MODE") != "debug" {
		if userReachedLimit(user.Email, isTestEmail) {
			http.Error(w, `
				You have reached your send limit. 
				You can only send one email every 6 hours. 
				If you need to send an email urgently, 
				please contact hoagie@princeton.edu`,
				http.StatusTooManyRequests)
			return
		}
	}

	if len(user.Name) == 0 {
		http.Error(w, `Hoagie Mail has been updated. Please log-out and log-in again.`, http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}

	var mailReq MailRequest
	err := json.NewDecoder(r.Body).Decode(&mailReq)
	if err != nil {
		http.Error(w, "Message did not contain correct fields.", http.StatusBadRequest)
		deleteVisitor(user.Email)
		return
	}
	if notBetween(w, mailReq.Sender, "sender name", 3, 30) {
		deleteVisitor(user.Email)
		return
	}
	if notBetween(w, mailReq.Header, "email subject", 3, 150) {
		deleteVisitor(user.Email)
		return
	}

	mailReq.Body = p.Sanitize(mailReq.Body)
	mailReq.Email = user.Email

	// Always send test emails immediately; do not attempt to schedule
	if isTestEmail {
		mailReq.Schedule = "now"
	}

	// Scheduled send
	if mailReq.Schedule != "now" {
		// Validate that schedule is valid
		if !scheduleValid(mailReq.Schedule) {
			deleteVisitor(user.Email)
			http.Error(
				w,
				"Your email could not be scheduled at the specified time. Please refresh the page and select a later time.",
				http.StatusBadRequest,
			)
			return
		}
		// Convert time to EST and check for errors
		est, err := time.LoadLocation("America/New_York")
		if err != nil {
			http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
			deleteVisitor(user.Email)
			return
		}
		scheduleEST, err := time.ParseInLocation(time.RFC3339, mailReq.Schedule, est)
		if err != nil {
			http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
			deleteVisitor(user.Email)
			return
		}

		// Check that user doesn't have an already-existing entry
		currentScheduledMail, err := getScheduled(user, scheduleEST)
		if err != nil {
			http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
			deleteVisitor(user.Email)
			return
		}
		if currentScheduledMail != (ScheduledMail{}) {
			errString := "You already have an email scheduled for this time. If you would like to change"
			errString += " your message, please delete your mail in the Scheduled Emails page and try again."
			http.Error(w, errString, http.StatusBadRequest)
			deleteVisitor(user.Email)
			return
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
	}

	mailReq.Body += fmt.Sprintf(`
	<hr />
	<div style="font-size:8pt;">This email was instantly sent to all
	college listservs with <a href="https://mail.hoagie.io/">Hoagie Mail</a>. 
	Email composed by %s (%s) — if you believe this email
	is offensive, intentionally misleading or harmful, please report it to
	<a href="mailto:hoagie@princeton.edu">hoagie@princeton.edu</a>.</div>
	`, user.Name, mailReq.Email)

	if os.Getenv("HOAGIE_MODE") == "debug" {
		println("Email: " + mailReq.Email)
		println("Sender: " + mailReq.Sender)
		println("Header: " + mailReq.Header)
		println("Body: " + mailReq.Body)
		println("Schedule: " + mailReq.Schedule)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("{\"Status\": \"OK\"}"))
		return
	}

	// Normal send
	if mailReq.Schedule == "now" {
		err = sendMail(mailReq, isTestEmail)

		fmt.Printf("MAIL: %s sent an email with title '%s'.", mailReq.Email, mailReq.Header)

		if err != nil {
			http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
			deleteVisitor(user.Email)
			return
		}
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"Status\": \"OK\"}"))
}

// POST /mail/send
var sendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	processSendRequest(w, r, false)
})

// POST /mail/sendTestMail
var sendTestMailHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	processSendRequest(w, r, true)
})
