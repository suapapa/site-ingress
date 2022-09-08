package main

import (
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
)

var (
	tgBot       *tgbotapi.BotAPI
	tgAPIToke   = os.Getenv("TELEGRAM_APITOKEN")
	tgRoomIDStr = os.Getenv("TELEGRAM_ROOM_ID")
)

func sendMsgToTelegram(msg string) error {
	var err error
	if tgBot == nil {
		tgBot, err = tgbotapi.NewBotAPI(tgAPIToke)
		if err != nil {
			return errors.Wrap(err, "fail to send msg to telegram")
		}
		// tgBot.Debug = true
	}

	tgRoomID, err := strconv.Atoi(tgRoomIDStr)
	if err != nil {
		return errors.Wrap(err, "fail to send msg to telegram")
	}

	// TODO: parse pretty msg from incomming json msg bytes
	c := tgbotapi.NewMessage(int64(tgRoomID), msg)
	// c.ParseMode = tgbotapi.ModeMarkdown // NOT WORKING :(
	if _, err := tgBot.Send(c); err != nil {
		return errors.Wrap(err, "fail to send msg to telegram")
	}
	return nil
}
