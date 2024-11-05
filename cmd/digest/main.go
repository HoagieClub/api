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
	"go.mongodb.org/mongo-driver/mongo/options"
)

var REQUEST_TIMEOUT = 10 * time.Second
var SUMMER = false
var sandwich = `<img height="22" src='https://i.imgur.com/gkEZQ4x.png' title='Hoagie' />`
var logo = `<img height="180px" src='https://i.imgur.com/kidY9cT.png' alt='Hoagie Digest' />`

type UserInfo struct {
	Name  string
	Email string
}

type DigestObject struct {
	Title       string
	Category    string
	Contact     string
	Description string
	Link        string
	Thumbnail   string
	Name        string
	Email       string
	Tags        []string
	User        UserInfo
}

func link(text string, link string) string {
	return fmt.Sprintf("<a target='_blank' href=\"%s\">%s</a>", link, text)
}

func link_mail(text string) string {
	return link(text, "mailto:"+text)
}

func formatTag(text string) string {
	return fmt.Sprintf(`<span style="color: #474d66; background-color:#edeff5; padding: 0px 6px; border-radius:4px; margin-right: 1px;">%s</span>`, strings.Title(text))
}

func addTags(email *strings.Builder, tags []string) {
	email.WriteString("<div style='margin-top: 6px;'>")
	for _, tag := range tags {
		email.WriteString(formatTag(tag) + " ")
	}
	email.WriteString("</div>")
}

func formatMessage(message DigestObject) string {
	var email strings.Builder
	name := message.User.Name
	if name == "" {
		// TODO: Old version, remove
		name = message.Name
	}
	switch message.Category {
	case "sale":
		tags := message.Tags
		if tags == nil || len(tags) == 0 {
			tags = strings.Split(message.Title, ", ")
		}
		// TODO: Old version, remove
		if len(message.Link) > 0 {
			email.WriteString("<span><a target='_blank' href=\"" + message.Link + "\">Open Sale Slides</a></span><br />")
		}
		email.WriteString(fmt.Sprintf("<div style='margin:10px 0px;'>%s</div>", message.Description))
		email.WriteString(fmt.Sprintf("<span><b>Contact: </b>%s (%s)</span><br />", name, link_mail(message.Email)))
		addTags(&email, tags)
	case "lost":
		if len(message.Thumbnail) > 0 {
			email.WriteString("<span><a target='_blank' href=\"" + message.Thumbnail + "\">See Picture</a></span><br />")
		}
		email.WriteString("<span><b>" + strings.ToUpper(message.Tags[0]) + ": </b>" + message.Title + "</span><br />")
		email.WriteString("<div style='margin:5px 0px;'>" + message.Description + "</div>")
		email.WriteString(fmt.Sprintf("<span><b>Contact: </b>%s (%s)</span><br />", name, link_mail(message.Email)))
	default:
		email.WriteString("<span><b>" + message.Title + "</b></span><br />")
		email.WriteString("<div style='margin:5px 0px;'>" + message.Description + "</div>")
		email.WriteString(fmt.Sprintf("<span><b>From: </b>%s (%s)</span><br />", name, link_mail(message.Email)))
		addTags(&email, message.Tags)
	}

	return email.String()
}

func main() {
	runDigestScript()
}

func runDigestScript() {
	godotenv.Load(".env.local")

	client, err := db.MongoClient()
	if err != nil {
		panic("Database connection error " + err.Error())
	}
	ctx := context.Background()
	defer client.Disconnect(ctx)

	cursor, err := db.FindMany(client, "apps", "stuff", bson.D{{"sent", false}}, options.Find())
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
	} else if total < 5 {
		// If there are less than 5 posts,
		// only send the Digest if it is a digest day and not summer!
		weekday := time.Now().Weekday()
		allowedDates := []time.Weekday{
			time.Tuesday, time.Thursday, time.Saturday,
		}
		isDigestDay := false
		if !SUMMER {
			for _, allowedDate := range allowedDates {
				if weekday == allowedDate {
					fmt.Println("Today is a Digest Day... Running...")
					isDigestDay = true
					break
				}
			}
		}
		if !isDigestDay {
			fmt.Println("Not a Digest Day... Exiting...")
			return
		}
	} else {
		fmt.Println("5 or more Digest posts... Running...")
	}

	var email strings.Builder
	email.WriteString(`<div style="
	font-family: sans-serif;
	">`)
	email.WriteString(fmt.Sprintf("<center>%s</center>", logo))
	if !SUMMER {
		email.WriteString(`<p><br />Here is a weekly digest of posts made to <a href="https://stuff.hoagie.io/">Hoagie Stuff</a>, 
	from Sales to Lost & Found and more, sent every Tuesday, Thursday, and Saturday.</p>`)
	} else {
		email.WriteString(`<p><br />Here is a digest of posts made to <a href="https://stuff.hoagie.io/">Hoagie Stuff</a></p> over past few days. It's Summer, so Hoagie is taking things slow.</p>`)
	}
	email.WriteString(`<p>
	<a target="_blank" href="https://stuff.hoagie.io/">Open Hoagie Stuff</a> |
	<a target="_blank" href="https://stuff.hoagie.io/create">Add your message to next digest</a> | 
	<a target="_blank" href="https://tally.so/r/mYJjN3">Give feedback</a>
	</p>`)

	email.WriteString("<hr />")
	if len(digest["lost"]) > 0 {
		email.WriteString("<h2>🧭 Lost & Found</h2>")
		email.WriteString(`<div style="margin-bottom:20px; margin-top:-10px;">Access anytime through <a href="https://stuff.hoagie.io/lostfound">stuff.hoagie.io/lostfound</a></div>`)
		for i, message := range digest["lost"] {
			email.WriteString(formatMessage(message))
			if i == len(digest["lost"])-1 {
				email.WriteString("<br />")
			}
			email.WriteString("<hr />")
		}
	}
	if len(digest["sale"]) > 0 {
		email.WriteString("<h2>🛍️ Marketplace</h2>")
		email.WriteString(`<div style="margin-bottom:20px; margin-top:-10px;">Accessible anytime with <a href="https://stuff.hoagie.io/marketplace">stuff.hoagie.io/marketplace</a></div>`)
		for i, message := range digest["sale"] {
			email.WriteString(formatMessage(message))
			if i == len(digest["sale"])-1 {
				email.WriteString("<br />")
			}
			email.WriteString("<hr />")
		}
	}
	if len(digest["bulletin"]) > 0 {
		email.WriteString("<h2>✉️ Bulletins</h2>")
		email.WriteString(`<div style="margin-bottom:20px; margin-top:-10px;">Accessible anytime with <a href="https://stuff.hoagie.io/bulletins">stuff.hoagie.io/bulletins</a></div>`)
		for i, message := range digest["bulletin"] {
			email.WriteString(formatMessage(message))
			if i == len(digest["bulletin"])-1 {
				email.WriteString("<br />")
			}
			email.WriteString("<hr />")
		}
	}
	email.WriteString(fmt.Sprintf(`<p>That's all! This could have been %d emails in your inbox but instead it is just one!<br /><br />
	You don't need to wait for the next digest to see what's new, check out the <a target="_blank" href="https://stuff.hoagie.io/">Hoagie Stuff</a>
	to keep up to date with the latest posts before others.
	</p>
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
				"📬 DIGEST %s: Sales, Lost & Found, and more!",
				time.Now().Format("1/2")),
			Sender: "Hoagie Mail",
			Body:   email.String(),
			Email:  "hoagie@princeton.edu",
		})
		fmt.Println("Successfully sent via Hoagie Mail.")
		db.UpdateMany(client, "apps", "stuff", bson.D{{"sent", false}}, bson.D{{"$set", bson.D{{"sent", true}}}}, options.Update())

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
			CustomID: "HoagieStuffDigest",
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
