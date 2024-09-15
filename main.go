package main

import (
	"superserver/modules/tgbot"
)

func main() {
	bot := tgbot.CreateTgBot()
	bot.Run()
}
