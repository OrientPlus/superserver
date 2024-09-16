package tgbot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math/rand"
	"os"
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

const (
	youAlreadyCat string = "Да ты себя видел?? Ты и так котик, тут бот не нужен"
	youAlreadyPes string = "Да ты и так пЭс сутулый. Выпрямился быстро! Спинку ровно"
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
	instModule    inst.ReelModule
	chats         map[string][]tgbotapi.User
	lastCat       tgbotapi.User
	lastCatChoise time.Time
	lastPes       tgbotapi.User
	lastPesChoise time.Time
}

func CreateTgBot() TgBot {
	bot := tgBot{}

	logger := loggers.CreateLogger(loggers.LoggerConfig{
		Name:           "MainLog",
		Path:           "./MainLogs.txt",
		Level:          loggers.DebugLevel,
		WriteToConsole: true,
		UseColor:       true,
	})

	var err error
	bot.lg = logger
	bot.botApi, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		bot.lg.Error(err.Error())
	}

	bot.updates = tgbotapi.NewUpdate(0)
	bot.updates.Timeout = 60

	bot.reelRegex = regexp.MustCompile(`^https?://(www\.)?instagram\.com/(reel|reels)/[A-Za-z0-9_-]+/?`)
	bot.funnyCat = regexp.MustCompile(`^\/lucky_cat$`)
	bot.unluckyCat = regexp.MustCompile(`^\/unlucky_cat$`)

	bot.instModule, err = inst.NewReelsDownloader()
	if err != nil {
		bot.lg.Error(err.Error())
	}

	bot.chats = make(map[string][]tgbotapi.User)

	return &bot
}

func (bot *tgBot) Run() {

	updates := bot.botApi.GetUpdatesChan(bot.updates)
	for update := range updates {
		bot.checkUser(update)

		bot.handleCommand(update)
	}

}

// Функция для выбора случайного пользователя
func (bot *tgBot) getRandomUser(message *tgbotapi.Message, bannedUser []tgbotapi.User) tgbotapi.User {
	rand.Seed(time.Now().UnixNano())
	usersCount := len(bot.chats[message.Chat.Title])
	if usersCount < 2 {
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

	if message.From.UserName == "NamorBot" {
		return
	}
	if message.Chat.Type == "private" {
		return
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

func isNextDay(prev time.Time) bool {
	currentTime := time.Now()

	if prev.IsZero() {
		return true
	}

	return currentTime.Year() != prev.Year() || currentTime.YearDay() != prev.YearDay()
}

func (bot *tgBot) handleCommand(update tgbotapi.Update) {
	if update.Message != nil {
		text := update.Message.Text
		if bot.reelRegex.MatchString(text) {
			bot.handleCommandInstReel(update)
		}
		if update.Message.Text == "/start" {
			bot.handleCommandLuckyPet(update)
		}
		if update.Message.Text == "/help" {
			bot.handleCommandHelp(update)
		}

		return
	}
	if update.CallbackQuery != nil {
		bot.handleCommandButtonLuckyPet(update)
		return
	}
}

func (bot *tgBot) handleCommandLuckyPet(update tgbotapi.Update) {
	text := update.Message.Text
	bot.lg.Info(fmt.Sprintf("распознана команда: %s; User: %s", text, update.Message.From.UserName))
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нажми кнопку, чтобы выбрать котика или пса дня!")

	// Создаем inline-кнопку
	buttonCat := tgbotapi.NewInlineKeyboardButtonData("Выбрать Котеночка дня", "choose_kitten")
	buttonPes := tgbotapi.NewInlineKeyboardButtonData("Выбрать Псину дня", "choose_pes")
	keyboardCat := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttonCat), tgbotapi.NewInlineKeyboardRow(buttonPes))

	msg.ReplyMarkup = keyboardCat
	_, err := bot.botApi.Send(msg)
	if err != nil {
		bot.lg.Error(fmt.Sprintf("не удалось отправить сообщение %s: %v", msg.Text, err))
	}
}

func (bot *tgBot) handleCommandButtonLuckyPet(update tgbotapi.Update) {
	if update.CallbackQuery.Data == "choose_kitten" {
		bot.lg.Info(fmt.Sprintf("нажата кнопка 'choose_kitten'; User: %s", update.CallbackQuery.Message.From.UserName))

		if update.CallbackQuery.Message.Chat.Type == "private" {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, youAlreadyCat)
			bot.botApi.Send(msg)
			return
		}

		bot.lg.Info("нажата кнопка 'котеночек дня' пользователем: " + update.CallbackQuery.Message.From.UserName)
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.botApi.Request(callback); err != nil {
			bot.lg.Error(fmt.Sprintf("Ошибка при отправке CallbackQuery ответа: %s", err))
		}

		randomUser := bot.getRandomUser(update.CallbackQuery.Message, []tgbotapi.User{bot.lastCat, bot.lastPes})
		if randomUser.ID == -1 {
			bot.lg.Warn(fmt.Sprintf("в чате %s обнаружено слишком мало участников для выполнения команды", update.CallbackQuery.Message.Chat.Title))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пока что я знаю мало людей в чате, чтобы выбрать котеночка( Попробуй позже")
			bot.botApi.Send(msg)
			return
		} else if randomUser.ID == -2 {
			bot.lg.Warn(fmt.Sprintf("в чате %s не удалось выбить рандомного юзера", update.CallbackQuery.Message.Chat.Title))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вас слишком мало чате, не получилось выбрать котика( Попробуй позже")
			bot.botApi.Send(msg)
			return
		}

		if isNextDay(bot.lastCatChoise) == false {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "рановато нового выбирать, еще старый хорош")
			bot.botApi.Send(msg)
			return
		}

		// Формируем сообщение с упоминанием пользователя
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("@%s, поздравляю, ты котеночек дня! Чмок в пупок!", randomUser.UserName))
		bot.botApi.Send(msg)
		bot.lastCatChoise = time.Now()
	} else if update.CallbackQuery.Data == "choose_pes" {
		bot.lg.Info(fmt.Sprintf("нажата кнопка 'choose_pes'; User: %s", update.CallbackQuery.Message.From.UserName))
		if update.CallbackQuery.Message.Chat.Type == "private" {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, youAlreadyPes)
			bot.botApi.Send(msg)
			return
		}

		bot.lg.Info("нажата кнопка 'псина дня' пользователем: " + update.CallbackQuery.Message.From.UserName)
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.botApi.Request(callback); err != nil {
			bot.lg.Error(fmt.Sprintf("Ошибка при отправке CallbackQuery ответа: %s", err))
		}

		randomUser := bot.getRandomUser(update.CallbackQuery.Message, []tgbotapi.User{bot.lastPes, bot.lastCat})
		if randomUser.ID == -1 {
			bot.lg.Warn(fmt.Sprintf("в чате %s обнаружено слишком мало участников для выполнения команды", update.CallbackQuery.Message.Chat.Title))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пока что я знаю мало людей в чате, чтобы выбрать псину( Попробуй позже")
			bot.botApi.Send(msg)
			return
		} else if randomUser.ID == -2 {
			bot.lg.Warn(fmt.Sprintf("в чате %s не удалось выбить рандомного юзера", update.CallbackQuery.Message.Chat.Title))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вас слишком мало чате, не получилось выбрать псину( Попробуй позже")
			bot.botApi.Send(msg)
			return
		}

		if isNextDay(bot.lastPesChoise) == false {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "рановато нового выбирать, еще старый пЭс годен")
			bot.botApi.Send(msg)
			return
		}

		// Формируем сообщение с упоминанием пользователя
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("@%s, поздравляю, ты пЭс этого дня! ", randomUser.UserName))
		bot.botApi.Send(msg)
		bot.lastPesChoise = time.Now()
	}
}

func (bot *tgBot) handleCommandInstReel(update tgbotapi.Update) {
	text := update.Message.Text

	bot.lg.Info("распознана ссылка: " + text)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Вижу ссыль, обрабатываю!")
	answerMsg, err := bot.botApi.Send(msg)
	if err != nil {
		bot.lg.Error(fmt.Sprintf("не удалось отправить сообщение: %v", err))
	}
	defer func() {
		deleteMsg := tgbotapi.DeleteMessageConfig{
			ChatID:    answerMsg.Chat.ID,
			MessageID: answerMsg.MessageID,
		}
		_, err = bot.botApi.Request(deleteMsg)
		if err != nil {
			bot.lg.Error(fmt.Sprintf("не удалось удалить сообщение: %v", err))
		}
	}()

	link := bot.reelRegex.FindString(text)
	if link == "" {
		bot.lg.Error("ссылка не отработана" + text)
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Я не смог распознать ссыль((( У меня лапки...")
		bot.botApi.Send(msg)
		return
	}

	// Скачиваем видео
	videoPath, err := bot.instModule.DownloadReel(link)
	bot.lg.Info(fmt.Sprintf("скачано видео: %s", videoPath))
	if err != nil {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Я не смог скачать видосик((( У меня лапки...")
		bot.botApi.Send(msg)
		return
	}
	defer os.Remove(videoPath)

	// Отправляем видео в чат
	videoFile, _ := os.OpenFile(videoPath, os.O_RDONLY, os.ModePerm)
	videoMsg := tgbotapi.NewVideo(update.Message.Chat.ID, tgbotapi.FileReader{
		Name:   "video.mp4", // Название файла
		Reader: videoFile,   // Файл, который нужно отправить
	})
	videoMsg.Caption = update.Message.From.UserName + " скинул видос"
	_, err = bot.botApi.Send(videoMsg)
	if err != nil {
		bot.lg.Error(fmt.Sprintf("не удалось отправить видео в чат: %v", err))
	}

	deleteMsg := tgbotapi.DeleteMessageConfig{
		ChatID:    update.Message.Chat.ID,
		MessageID: update.Message.MessageID,
	}
	_, err = bot.botApi.Request(deleteMsg)
	if err != nil {
		bot.lg.Error(fmt.Sprintf("не удалось удалить сообщение: %v", err))
	}
}

func (bot *tgBot) handleCommandHelp(update tgbotapi.Update) {

}
