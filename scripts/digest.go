package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"hoagie-profile/db"

	"github.com/joho/godotenv"
	"github.com/mailjet/mailjet-apiv3-go"
	"go.mongodb.org/mongo-driver/bson"
)

var REQUEST_TIMEOUT = 10 * time.Second
var sandwich = `<img height="22" src='https://i.imgur.com/gkEZQ4x.png' title='Hoagie' />`
var logo = `<img height="68" src='https://i.imgur.com/bu1KWlQ.png' alt='Hoagie Digest' />`

type DigestObject struct {
	Title       string
	Category    string
	Contact     string
	Description string
	Link        string
	Name        string
	Email       string
}

func link(text string, link string) string {
	return fmt.Sprintf("<a target='_blank' href=\"%s\">%s</a>", link, text)
}

func link_mail(text string) string {
	return link(text, "mailto:"+text)
}

func formatMessage(message DigestObject) string {
	var email strings.Builder
	switch message.Category {
	case "sale":
		if len(message.Link) > 0 {
			email.WriteString("<span><a target='_blank' href=\"" + message.Link + "\">Open Sale Slides</a></span><br />")
		}
		email.WriteString(fmt.Sprintf("<span><b>Contact: </b>%s (%s)</span><br />", message.Name, link_mail(message.Email)))
		email.WriteString(fmt.Sprintf("<span><b>Categories: </b>%s</span><br />", message.Title))
		email.WriteString(fmt.Sprintf("<span><b>Description: </b>%s</span><br />", message.Description))
	case "lost":
		if len(message.Link) > 0 {
			email.WriteString("<span><a target='_blank' href=\"" + message.Link + "\">See Picture</a></span><br />")
		}
		email.WriteString(fmt.Sprintf("<span><b>Contact: </b>%s (%s)</span><br />", message.Name, link_mail(message.Email)))
		email.WriteString("<span><b>Title: </b>" + message.Title + "</span><br />")
		email.WriteString("<span><b>Description: </b>" + message.Description + "</span><br />")
	default:
		email.WriteString(fmt.Sprintf("<span><b>From: </b>%s (%s)</span><br />", message.Name, message.Email))
		email.WriteString("<span><b>Title: </b>" + message.Title + "</span><br />")
		email.WriteString("<span><b>Message: </b>" + message.Description + "</span><br />")
	}

	return email.String()
}

func main() {
	weekday := time.Now().Weekday()
	allowedDates := []time.Weekday{
		time.Wednesday, time.Saturday,
	}
	for _, allowedDate := range allowedDates {
		if weekday == allowedDate {
			fmt.Println("Today is a Digest Day... Running...")
			runDigestScript()
			return
		}
	}
	fmt.Println("Not a Digest Day... Quitting...")
}

func runDigestScript() {
	godotenv.Load(".env.local")

	client, err := db.MongoClient()
	if err != nil {
		panic("Database connection error " + err.Error())
	}
	ctx := context.Background()
	defer client.Disconnect(ctx)

	cursor, err := db.FindMany(client, "apps", "mail", bson.D{})
	if err != nil {
		panic("Error getting digest emails" + err.Error())
	}
	defer cursor.Close(ctx)

	digest := make(map[string][]DigestObject)
	total := 0
	for cursor.Next(ctx) {
		var message DigestObject
		if err = cursor.Decode(&message); err != nil {
			fmt.Printf("Error decoding digest email: %s", err)
			return
		}
		cat := message.Category
		digest[cat] = append(digest[cat], message)
		total += 1
	}
	if total < 1 {
		fmt.Println("No messages found...")
		return
	}

	var email strings.Builder
	email.WriteString(`<div style="
	font-family: sans-serif;
	">`)
	email.WriteString(fmt.Sprintf("<center>%s</center>", logo))
	email.WriteString(`<p><br />Here is a weekly digest of student messages, 
	from Sales to Lost & Found and more, sent at noon every Wednesday and Saturday, powered by <b>hoagiemail</b></p>`)
	email.WriteString(`<p>
	<a target="_blank" href="https://mail.hoagie.io">Add your message to next digest</a> | 
	<a target="_blank" href="https://tally.so/r/mYJjN3">Give feedback</a>
	</p>`)

	email.WriteString("<hr />")
	if len(digest["lost"]) > 0 {
		email.WriteString("<h2>🧭 Lost & Found</h2>")
		for i, message := range digest["lost"] {
			email.WriteString(formatMessage(message))
			if i == len(digest["lost"])-1 {
				email.WriteString("<br />")
			}
			email.WriteString("<hr />")
		}
	}
	if len(digest["sale"]) > 0 {
		email.WriteString("<h2>🛍️ Student Sales</h2>")
		for i, message := range digest["sale"] {
			email.WriteString(formatMessage(message))
			if i == len(digest["sale"])-1 {
				email.WriteString("<br />")
			}
			email.WriteString("<hr />")
		}
	}
	if len(digest["misc"]) > 0 {
		email.WriteString("<h2>✉️ Other</h2>")
		for i, message := range digest["misc"] {
			email.WriteString(formatMessage(message))
			if i == len(digest["misc"])-1 {
				email.WriteString("<br />")
			}
			email.WriteString("<hr />")
		}
	}
	email.WriteString(fmt.Sprintf(`<p>That's all! This could have been %d emails in your inbox but instead it is just one!</p>
	`, total))
	email.WriteString(fmt.Sprintf(`
		<center>
		%s <br />
		<div style="font-size:8pt; margin-top:8px;">
		Powered by <a target="_blank" href="https://mail.hoagie.io/">HoagieMail</a><br />
		In the Hoagie world, hoagies digest you!
		</div>
		</center>
	`, sandwich))
	email.WriteString("</div>")
	fmt.Println(email.String())
	if os.Getenv("HOAGIE_MODE") == "production" {
		makeRequest(MailRequest{
			Header: fmt.Sprintf(
				"📬 DIGEST %s: Sales, lost & found, and more!",
				time.Now().Format("2/1")),
			Sender: "Hoagie Mail",
			Body:   email.String(),
			Email:  "hoagie@princeton.edu",
		})
		fmt.Println("Successfully sent via Hoagie Mail.")
		db.Drop(client, "apps", "mail")
	}
}

type MailRequest struct {
	Header string
	Sender string
	Body   string
	Email  string
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
					Email: "hoagie@princeton.edu",
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
