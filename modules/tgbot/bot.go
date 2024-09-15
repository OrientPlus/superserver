package tgbot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math/rand"
	"regexp"
	"superserver/loggers"
	"superserver/modules/tgbot/inst"
	"time"
)

const token string = "6739454793:AAFTDRXnqDTNGvN7IWQBom6a5YkHeO6YpzQ"

type Command string

const (
	funnyCat   Command = "funny_cat"
	instaReels Command = "insta_reels"
)

type TgBot interface {
	Run()
}

type tgBot struct {
	lg            loggers.Logger
	botApi        *tgbotapi.BotAPI
	reelRegex     *regexp.Regexp
	funnyCat      *regexp.Regexp
	unluckyCat    *regexp.Regexp
	updates       tgbotapi.UpdateConfig
	instModule    inst.ReelsDownloader
	chats         map[string][]tgbotapi.User
	lastCat       tgbotapi.User
	lastCatChoise time.Time
	lastPes       tgbotapi.User
	lastPesChoise time.Time
}

func CreateTgBot() TgBot {
	bot := tgBot{}

	logger, err := loggers.CreateLogger(loggers.LoggerConfig{
		Name:           "Default",
		Path:           "./DefLogs.txt",
		Level:          loggers.DebugLevel,
		WriteToConsole: true,
		UseColor:       true,
	})
	if err != nil {
		panic(err)
	}

	bot.lg = logger
	bot.botApi, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		bot.lg.Error(err.Error())
	}

	bot.updates = tgbotapi.NewUpdate(0)
	bot.updates.Timeout = 60

	bot.reelRegex = regexp.MustCompile(`https?://(www\.)?instagram\.com/reel/[^\s]+`)
	bot.funnyCat = regexp.MustCompile(`^\/lucky_cat$`)
	bot.unluckyCat = regexp.MustCompile(`^\/unlucky_cat$`)

	bot.instModule = inst.ReelsDownloader{}

	bot.chats = make(map[string][]tgbotapi.User)

	return &bot
}

func (bot *tgBot) Run() {

	updates := bot.botApi.GetUpdatesChan(bot.updates)
	for update := range updates {

		if update.Message != nil {
			bot.checkUser(update)

			text := update.Message.Text
			if bot.reelRegex.MatchString(text) {
				bot.lg.Info("распознана ссылка: " + text)

				link := bot.reelRegex.FindString(text)
				if link == "" {
					bot.lg.Error("ссылка не отработана" + text)
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Я не смог распознать ссыль((( У меня лапки...")
					bot.botApi.Send(msg)
					continue
				}

				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "разраб выпил недостаточно кофе для реализации этой фичи, ожидайте, скоро заработает")
				bot.botApi.Send(msg)

				// Скачиваем видео
				videoURL, err := bot.instModule.DownloadInstagramReel(link)
				if err != nil {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при скачивании видео: "+err.Error())
					bot.botApi.Send(msg)
					continue
				}

				// Отправляем видео в чат
				videoMsg := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FileURL(videoURL))
				bot.botApi.Send(videoMsg)
			} else if update.Message.Text == "/start" {
				bot.lg.Info("распознана команда: " + text)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нажми кнопку, чтобы выбрать котика или пса дня!")

				// Создаем inline-кнопку
				buttonCat := tgbotapi.NewInlineKeyboardButtonData("Выбрать Котеночка дня", "choose_kitten")
				buttonPes := tgbotapi.NewInlineKeyboardButtonData("Выбрать Псину дня", "choose_pes")
				keyboardCat := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttonCat), tgbotapi.NewInlineKeyboardRow(buttonPes))

				msg.ReplyMarkup = keyboardCat
				bot.botApi.Send(msg)
			}

		}
		// если это нажатие на кнопку
		if update.CallbackQuery != nil {
			bot.checkUser(update)

			if update.CallbackQuery.Data == "choose_kitten" {
				bot.lg.Info("нажата кнопка 'котеночек дня' пользователем: " + update.CallbackQuery.Message.From.UserName)

				randomUser := bot.getRandomUser(update.CallbackQuery.Message, []tgbotapi.User{bot.lastCat, bot.lastPes})
				if randomUser.ID == -1 {
					bot.lg.Warn(fmt.Sprintf("в чате %s обнаружено слишком мало участников для выполнения команды", update.CallbackQuery.Message.Chat.Title))
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пока что я знаю мало людей в чате, чтобы выбрать котеночка( Попробуй позже")
					bot.botApi.Send(msg)
					continue
				} else if randomUser.ID == -2 {
					bot.lg.Warn(fmt.Sprintf("в чате %s не удалось выбить рандомного юзера", update.CallbackQuery.Message.Chat.Title))
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вас слишком мало чате, не получилось выбрать котика( Попробуй позже")
					bot.botApi.Send(msg)
					continue
				}

				if isNextDay(bot.lastCatChoise) == false {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "рановато нового выбирать, еще старый хорош")
					bot.botApi.Send(msg)
					continue
				}

				// Формируем сообщение с упоминанием пользователя
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("@%s, поздравляю, ты котеночек дня! Чмок в пупок!", randomUser.UserName))
				bot.botApi.Send(msg)
				bot.lastCatChoise = time.Now()
			} else if update.CallbackQuery.Data == "choose_pes" {
				bot.lg.Info("нажата кнопка 'псина дня' пользователем: " + update.CallbackQuery.Message.From.UserName)

				randomUser := bot.getRandomUser(update.CallbackQuery.Message, []tgbotapi.User{bot.lastPes, bot.lastCat})
				if randomUser.ID == -1 {
					bot.lg.Warn(fmt.Sprintf("в чате %s обнаружено слишком мало участников для выполнения команды", update.CallbackQuery.Message.Chat.Title))
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пока что я знаю мало людей в чате, чтобы выбрать псину( Попробуй позже")
					bot.botApi.Send(msg)
					continue
				} else if randomUser.ID == -2 {
					bot.lg.Warn(fmt.Sprintf("в чате %s не удалось выбить рандомного юзера", update.CallbackQuery.Message.Chat.Title))
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вас слишком мало чате, не получилось выбрать псину( Попробуй позже")
					bot.botApi.Send(msg)
					continue
				}

				if isNextDay(bot.lastPesChoise) == false {
					msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "рановато нового выбирать, еще старый пЭс годен")
					bot.botApi.Send(msg)
					continue
				}

				// Формируем сообщение с упоминанием пользователя
				msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("@%s, поздравляю, ты пЭс этого дня! ", randomUser.UserName))
				bot.botApi.Send(msg)
				bot.lastPesChoise = time.Now()
			}
		}

	}

}

// Функция для выбора случайного пользователя
func (bot *tgBot) getRandomUser(message *tgbotapi.Message, bannedUser []tgbotapi.User) tgbotapi.User {
	rand.Seed(time.Now().UnixNano())
	usersCount := len(bot.chats[message.Chat.Title])
	if usersCount < 1 {
		return tgbotapi.User{ID: -1}
	}

	var luckyUser tgbotapi.User
	for range 10 {
		luckyUser = bot.chats[message.Chat.Title][rand.Intn(usersCount)]
		badUser := false
		for _, user := range bannedUser {
			if luckyUser.ID == user.ID {
				badUser = true
				break
			}
		}
		if badUser == false {
			break
		}
	}

	return luckyUser
}

func (bot *tgBot) checkUser(update tgbotapi.Update) {
	var message *tgbotapi.Message
	if update.Message == nil {
		message = update.CallbackQuery.Message
	} else {
		message = update.Message
	}

	chatName := message.Chat.Title
	if chatName == "" {
		bot.lg.Warn("не удалось определить имя группы")
	}

	for _, user := range bot.chats[chatName] {
		if user.ID == message.From.ID {
			return
		}
	}
	bot.lg.Info("добавлен новый пользователь: " + message.From.UserName)
	bot.chats[chatName] = append(bot.chats[chatName], *message.From)
}

func handleCommand(update tgbotapi.Update) {

}

func isNextDay(prev time.Time) bool {
	currentTime := time.Now()

	if prev.IsZero() {
		return true
	}

	return currentTime.Year() != prev.Year() || currentTime.YearDay() != prev.YearDay()
}
