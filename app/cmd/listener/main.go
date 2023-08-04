package main

import (
	"encoding/json"
	"log"
	"os"
	"tg-bot-balance/internal"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	tele "gopkg.in/telebot.v3"
)

func main() {
	publicUrl := os.Getenv("PUBLIC_URL")
	table := os.Getenv("TABLE_NAME")
	region := os.Getenv("REGION")
	queue := os.Getenv("QUEUE")
	balanceClient := internal.NewBalanceClient(table, region)
	sqsClient := internal.NewSqsClient(region, queue)
	pref := tele.Settings{
		Token:       os.Getenv("TG_BOT_TOKEN"),
		Synchronous: true,
		Verbose:     true,
	}
	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}
	setHandlers(bot, balanceClient, sqsClient)
	webhookEndpoint := &tele.WebhookEndpoint{
		PublicURL: publicUrl,
	}
	err = bot.SetWebhook(&tele.Webhook{
		Endpoint: webhookEndpoint,
	})
	if err != nil {
		log.Fatal(err)
	}

	lambda.Start(func(req events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		var u tele.Update
		if err = json.Unmarshal([]byte(req.Body), &u); err == nil {
			bot.ProcessUpdate(u)
		}
		return events.APIGatewayProxyResponse{Body: "ok", StatusCode: 200}, nil
	})
}

func setHandlers(bot *tele.Bot, bClient *internal.BalanceClient, sqsClient *internal.SqsClient) {
	bot.Handle("/topup", func(c tele.Context) error {
		return handleTopup(c, bClient)
	})
	bot.Handle("/balance", func(c tele.Context) error {
		return handleBalance(c, bClient)
	})
	bot.Handle(tele.OnText, func(c tele.Context) error {
		return handleText(c, sqsClient)
	})
}
