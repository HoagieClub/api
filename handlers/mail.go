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
	"github.com/microcosm-cc/bluemonday"
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

const (
	mailRoute     = "/mail"
	mailSendRoute = "/mail/send"
)

// BlueMonday sanitizes HTML, preventing unsafe user input
var p = bluemonday.UGCPolicy()

func setupMailHandlers(r *mux.Router, m *jwtmiddleware.JWTMiddleware) {
	// Handle mail send request
	r.Handle(mailSendRoute, m.Handler(sendHandler)).Methods("POST")
}

type MailRequest struct {
	Header string
	Sender string
	Body   string
	Email  string
}

func sendMail(req MailRequest) {
	from := mail.NewEmail(req.Sender, "hoagie@princeton.edu")
	replyTo := mail.NewEmail(req.Sender, req.Email)
	subject := req.Header

	content := mail.NewContent("text/html", req.Body)

	to := mail.NewEmail("Hoagie", "hoagie@princeton.edu")

	message := mail.NewV3MailInit(from, subject, to, content)
	message.SetReplyTo(replyTo)

	tos := []*mail.Email{
		mail.NewEmail("Hoagie", os.Getenv("HOAGIE_TEST_EMAIL")),
	}

	p := mail.NewPersonalization()
	p.AddTos(tos...)

	message.AddPersonalizations(p)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	response, err := client.Send(message)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(response)
	}
}

var sendHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	accessToken := strings.TrimPrefix(r.Header.Get("authorization"), "Bearer ")

	user, err := auth.GetUser(accessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var mailReq MailRequest
	err = json.NewDecoder(r.Body).Decode(&mailReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	const headerLimit int = 256
	if len(mailReq.Header) > headerLimit {
		responseString := fmt.Sprintf("Request cannot be completed because email header of length %d is over the %d-character limit.", len(mailReq.Header), headerLimit)
		http.Error(w, responseString, http.StatusForbidden)
		return
	}

	mailReq.Body = p.Sanitize(mailReq.Body)
	mailReq.Body += `
	<hr />
	<div style="font-size:8pt;">This email was sent instantly to all 
	college listservs with <a href="https://mail.hoagie.io/">mail.hoagie.io</a>. 
	You can use it to automatically send emails to all students without 
	the need to forward it to friends.</div>
	`
	mailReq.Email = user

	sendMail(mailReq)
})
