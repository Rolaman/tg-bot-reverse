package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"tg-bot-balance/internal"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	tele "gopkg.in/telebot.v3"
)

var (
	ErrInvalidUserId = errors.New("invalid user id")
)

func main() {
	table := os.Getenv("TABLE_NAME")
	region := os.Getenv("REGION")
	balanceClient := internal.NewBalanceClient(table, region)
	pref := tele.Settings{
		Token:       os.Getenv("TG_BOT_TOKEN"),
		Synchronous: true,
		Verbose:     true,
	}
	bot, err := tele.NewBot(pref)
	if err != nil {
		log.Fatal(err)
	}
	lambda.Start(func(sqsEvent events.SQSEvent) {
		for _, message := range sqsEvent.Records {
			err = sqsHandler(bot, balanceClient, message)
			if err != nil {
				log.Printf("Error while processing message: %v", err)
			}
		}
	})
}

func sqsHandler(bot *tele.Bot, balanceClient *internal.BalanceClient, msg events.SQSMessage) error {
	price := calculatePrice()
	splits := strings.SplitN(msg.Body, ":", 2)
	userId, err := strconv.ParseInt(splits[0], 10, 64)
	text := reverseString(splits[1])
	if err != nil {
		return ErrInvalidUserId
	}
	user := &tele.User{
		ID: userId,
	}
	err = balanceClient.Charge(strconv.FormatInt(userId, 10), price)
	if err != nil {
		if errors.Is(err, internal.ErrInsufficient) {
			_, err = bot.Send(user, "Not enough balance")
			if err != nil {
				return err
			}
			return nil
		}
		return err
	}
	_, err = bot.Send(user, fmt.Sprintf("Result: %v\nMessage price: %d", text, price))
	return err
}

func calculatePrice() int {
	return rand.Intn(101)
}

func reverseString(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
