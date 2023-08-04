package main

import (
	"fmt"
	"strconv"
	"strings"
	"tg-bot-balance/internal"

	tele "gopkg.in/telebot.v3"
)

func handleTopup(c tele.Context, bClient *internal.BalanceClient) error {
	words := strings.Fields(c.Message().Text)
	if len(words) < 2 {
		return c.Send("Please provide an real amount. Example: /topup 100")
	}
	amountStr := words[1]
	amount, err := strconv.Atoi(amountStr)
	if err != nil {
		return c.Send("Invalid amount. Please provide a number. Example: /topup 100")
	}
	err = bClient.Topup(strconv.FormatInt(c.Sender().ID, 10), amount)
	if err != nil {
		return c.Send(fmt.Sprintf("Can't process topup: %v", err))
	}
	return c.Send("Success")
}

func handleBalance(c tele.Context, bClient *internal.BalanceClient) error {
	balance, err := bClient.Balance(strconv.FormatInt(c.Sender().ID, 10))
	if err != nil {
		return c.Send(fmt.Sprintf("Can't fetch balance: %v", err))
	}
	return c.Send(fmt.Sprintf("Your balance: %d", balance))
}

func handleText(c tele.Context, sqs *internal.SqsClient) error {
	return sqs.Charge(strconv.FormatInt(c.Sender().ID, 10), c.Text())
}
