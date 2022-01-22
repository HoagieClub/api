package handlers

import (
	"encoding/json"
	"fmt"
	"hoagie-profile/auth"
	"net/http"
	"os"
	"strings"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/gorilla/mux"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"github.com/microcosm-cc/bluemonday"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	mailRoute       = "/mail"
	mailSendRoute   = "/mail/send"
	mailDigestRoute = "/mail/digest"
)

type MailRequest struct {
	Header string
	Sender string
	Body   string
	Email  string
}

// BlueMonday sanitizes HTML, preventing unsafe user input
var p = bluemonday.UGCPolicy()
var client *mongo.Client

func setupMailHandlers(r *mux.Router, m *jwtmiddleware.JWTMiddleware, cl *mongo.Client) {
	// Handle mail send request
	client = cl
	r.Handle(mailSendRoute, m.Handler(sendHandler)).Methods("POST")
	r.Handle(mailDigestRoute, m.Handler(digestSendHandler)).Methods("POST")
	r.Handle(mailDigestRoute, m.Handler(digestStatusHandler)).Methods("GET")
}

func makeRequest(req MailRequest) error {
	mailjetClient := mailjet.NewMailjetClient(os.Getenv("MAILJET_PUBLIC_KEY"), os.Getenv("MAILJET_PRIVATE_KEY"))
	messagesInfo := []mailjet.InfoMessagesV31{
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
					Email: "FIRSTCOMEFIRSTSERV@PRINCETON.EDU",
					Name:  "First",
				},
				mailjet.RecipientV31{
					Email: "matheymail@PRINCETON.EDU",
					Name:  "Mathey",
				},
			},
			Subject:  req.Header,
			TextPart: req.Body,
			HTMLPart: req.Body,
			CustomID: "HoagieMail",
		},
	}
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

func sendMail(req MailRequest) error {
	err := makeRequest(req)
	if err != nil {
		return err
		// TODO: be better with status code handling. Most likely just == 400.
	}
	return nil
}

func userReachedLimit(user string) bool {
	userLimit := getVisitor(user)
	return !userLimit.Allow()
}

var sendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	accessToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	user, err := auth.GetUser(accessToken)
	if err != nil {
		http.Error(w, "You do not have access to send mail.", http.StatusBadRequest)
		return
	}

	if userReachedLimit(user.Email) {
		http.Error(w, `
			You have reached your send limit. 
			You can only send one email every 6 hours. 
			If you need to send an email urgently, 
			please contact hoagie@princeton.edu`,
			http.StatusTooManyRequests)
		return
	}

	var mailReq MailRequest
	err = json.NewDecoder(r.Body).Decode(&mailReq)
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

	// mailReq.Body = p.Sanitize(mailReq.Body)
	mailReq.Email = user.Email
	mailReq.Body += fmt.Sprintf(`
	<hr />
	<div style="font-size:8pt;">This email was instantly sent to all
	college listservs with <a href="https://mail.hoagie.io/">Hoagie Mail</a>. 
	Email composed by %s — if you believe this email
	is offensive, intentionally misleading or harmful, please report it to
	<a href="mailto:hoagie@princeton.edu">hoagie@princeton.edu</a>.</div>
	`, mailReq.Email)

	if os.Getenv("HOAGIE_MODE") == "debug" {
		println(mailReq.Email)
		println(mailReq.Header)
		println(mailReq.Body)
		return
	}
	err = sendMail(mailReq)

	fmt.Printf("MAIL: %s sent an email with title '%s'.", mailReq.Email, mailReq.Header)

	if err != nil {
		http.Error(w, fmt.Sprintf("Hoagie Mail service had an error: %s.", err.Error()), http.StatusNotFound)
		deleteVisitor(user.Email)
		return
	}
})
