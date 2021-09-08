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

func sendMail(req MailRequest) error {
	from := mail.NewEmail(req.Sender, "hoagie@princeton.edu")
	replyTo := mail.NewEmail(req.Sender, req.Email)
	subject := req.Header

	content := mail.NewContent("text/html", req.Body)

	to := mail.NewEmail(req.Sender, req.Email)

	message := mail.NewV3MailInit(from, subject, to, content)
	message.SetReplyTo(replyTo)

	tos := []*mail.Email{
		mail.NewEmail(req.Sender, req.Email),
	}
	ccs := []*mail.Email{
		mail.NewEmail("Butler", "BUTLERBUZZ@PRINCETON.EDU"),
		mail.NewEmail("Whitman", "WHITMANWIRE@PRINCETON.EDU"),
		mail.NewEmail("Rocky", "RockyWire@PRINCETON.EDU"),
		mail.NewEmail("Forbes", "Re-INNformer@PRINCETON.EDU"),
		mail.NewEmail("First", "FIRSTCOMEFIRSTSERV@PRINCETON.EDU"),
		mail.NewEmail("Mathey", "matheymail@PRINCETON.EDU"),
		mail.NewEmail("Hoagie Mail", "hoagie+mail@princeton.edu"),
	}

	p := mail.NewPersonalization()
	p.AddTos(tos...)
	p.AddCCs(ccs...)

	message.AddPersonalizations(p)

	client := sendgrid.NewSendClient(os.Getenv("SENDGRID_API_KEY"))
	resp, err := client.Send(message)
	if err != nil {
		return err
		// TODO: be better with status code handling. Most likely just == 400.
	} else if resp.StatusCode == 400 {
		return fmt.Errorf("reached the mail send limit for the day, try again tomorrow", resp.Body)
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

	if userReachedLimit(user) {
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
		return
	}
	if notBetween(w, mailReq.Sender, "sender name", 3, 30) {
		return
	}
	if notBetween(w, mailReq.Header, "email header", 3, 150) {
		return
	}

	// mailReq.Body = p.Sanitize(mailReq.Body)
	mailReq.Email = user
	mailReq.Body += fmt.Sprintf(`
	<hr />
	<div style="font-size:8pt;">This email was instantly sent to all
	college listservs with <a href="https://mail.hoagie.io/">mail.hoagie.io</a>. 
	Want to be part of projects like this? <a href="https://club.hoagie.io/">Apply to Hoagie!</a>
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
		return
	}
})
