package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"hoagie-profile/db"

	godotenv "github.com/joho/godotenv"
	mailjet "github.com/mailjet/mailjet-apiv3-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MailRequest struct {
	Header string
	Sender string
	Body   string
	Email  string
	UserName string
	Schedule time.Time
	CreatedAt time.Time
}

func main() {
	runScheduledSendScript()
}

// This function queries the MongoDB collection for mail and sends emails
// with a schedule field that's less than the current time in EST + 60 minutes
func runScheduledSendScript() {
	godotenv.Load(".env.local")

	client, err := db.MongoClient()
	if err != nil {
		panic("Database connection error " + err.Error())
	}
	ctx := context.Background()
	defer client.Disconnect(ctx)

	estLocation, err := time.LoadLocation("America/New_York")
	if err != nil {
		panic(err)
	}
	currentTimeEST := time.Now().In(estLocation).Add(60 * time.Minute)

	// Grace period of 60 minutes because Heroku Scheduler isn't exact
	filter := bson.D{
		{"Schedule", bson.D{
        	{"$lte", currentTimeEST},
    	}},
	}
	cursor, err := db.FindMany(client, "apps", "mail", filter, options.Find())
	if err != nil {
		panic("Error getting scheduled emails" + err.Error())
	}
	defer cursor.Close(ctx)

	total := 0
	errorTotal := 0
	for cursor.Next(ctx) {
		var mailReq MailRequest
		if err = cursor.Decode(&mailReq); err != nil {
			fmt.Printf("Error decoding scheduled mail: %s", err)
			errorTotal++
			continue
		}
		mailReq.Body += fmt.Sprintf(`
		<hr />
		<div style="font-size:8pt;">This email was instantly sent to all
		college listservs with <a href="https://mail.hoagie.io/">Hoagie Mail</a>. 
		Email composed by %s (%s) — if you believe this email
		is offensive, intentionally misleading or harmful, please report it to
		<a href="mailto:hoagie@princeton.edu">hoagie@princeton.edu</a>.</div>
		`, mailReq.UserName, mailReq.Email)

		if os.Getenv("HOAGIE_MODE") == "debug" {
			println("Email: " + mailReq.Email)
			println("Sender: " + mailReq.Sender)
			println("Header: " + mailReq.Header)
			println("Body: " + mailReq.Body)
			println("Schedule: " + mailReq.Schedule.String())
			println("UserName: " + mailReq.UserName)
			println("CreatedAt: " + mailReq.CreatedAt.String())
		} 

		if os.Getenv("HOAGIE_MODE") == "production" {
			err = makeRequest(mailReq)
			if err != nil {
				fmt.Println(err)
				errorTotal++
				continue // have better error handling?
			}
			currentMailFilter := bson.D{
				{"Email", mailReq.Email},
				{"Sender", mailReq.Sender},
				{"Header", mailReq.Header},
				{"Schedule", mailReq.Schedule},
			}
			db.DeleteOne(client, "apps", "mail", currentMailFilter)
		}
		total++
	}
	if errorTotal != 0 {
		fmt.Printf("%d scheduled emails had errors\n", errorTotal)
	}
	if total == 0 {
		fmt.Println("No emails sent at this time")
	} else {
		fmt.Printf("Successfully sent %d scheduled emails via Hoagie Mail.\n", total)
	}
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
