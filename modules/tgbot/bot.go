package tgbot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"math/rand"
	"os"
	"regexp"
	"superserver/db"
	"superserver/entity"
	"superserver/loggers"
	"superserver/modules/tgbot/inst"
	"time"
)

const token string = "6739454793:AAFTDRXnqDTNGvN7IWQBom6a5YkHeO6YpzQ"

type Command string

const (
	helpGroupChatOutput string = "/start - для запуска меню\n" +
		"\nКнпчк 'Котик дня' - выбирает случайным образом 'котика дня' среди участников чата" +
		"\nКнпчк 'Псина дня' - выбирает случайным образом 'пса дня' среди участников чата" +
		"\nБот знает только тех участников чата, которые хоть раз писали в чат с момента добавления бота." +
		"\nКотик и пес дня сбрасываются после 24:00." +
		"\nБот распознает ссылки на instagram reels, скачивает рилс и отправляет в чат вместо распознанной ссылки, с указанием того, кто скинул рилс."
	helpPrivateChatOutput string = "Бот распознает ссылки на instagram reels, скачивает рилс и отправляет в чат вместо распознанной ссылки"
)

const (
	steelCat1  string = "Не спеши, старый котеночек ещё в силе!"
	steelCat2  string = "Ещё рано, сегодняшний котеночек не сдал свои позиции!"
	steelCat3  string = "Погоди, давай дадим старому котеночку насладиться моментом."
	steelCat4  string = "Спокойно, нынешний котеночек ещё не успел насладиться своим триумфом."
	steelCat5  string = "Текущий котеночек ещё не наигрался, подождем!"
	steelCat6  string = "Котеночек дня всё ещё в строю, давай не торопить события."
	steelCat7  string = "Постой, ещё не вечер для сегодняшнего котеночка!"
	steelCat8  string = "Терпение, старый котеночек всё ещё царствует!"
	steelCat9  string = "Рано, котеночек дня всё ещё на своём заслуженном посту."
	steelCat10 string = "Давай дадим котеночку дня насладиться своим званием чуть дольше!"

	steelPes1  string = "Не спеши, этот пёс ещё не выбегал своё счастье!"
	steelPes2  string = "Погоди, старый пёс ещё в строю, не время менять его!"
	steelPes3  string = "Текущий пёс дня ещё не налаялся вдоволь!"
	steelPes4  string = "Терпение, пёс дня ещё не показал все свои трюки!"
	steelPes5  string = "Этот пёс ещё не всех порадовал, рановато для нового!"
	steelPes6  string = "Подожди, пёс дня ещё патрулирует свои владения!"
	steelPes7  string = "Старый пёс ещё лает, не торопись с новым!"
	steelPes8  string = "Не время для нового пса, этот ещё хвостом не намахался!"
	steelPes9  string = "Ещё рановато, пёс дня всё ещё в форме!"
	steelPes10 string = "Этот пёс ещё не исчерпал свою энергию, давай дадим ему доиграться!"
)

var steelCatPhrases = []string{steelCat1, steelCat2, steelCat3, steelCat4, steelCat5, steelCat6, steelCat7, steelCat8, steelCat9, steelCat10}
var steelPesPhrases = []string{steelPes1, steelPes2, steelPes3, steelPes4, steelPes5, steelPes6, steelPes7, steelPes8, steelPes9, steelPes10}

type TgBot interface {
	Run()
}

type tgBot struct {
	logger       loggers.Logger
	botApi       *tgbotapi.BotAPI
	reelRegex    *regexp.Regexp
	funnyCat     *regexp.Regexp
	unluckyCat   *regexp.Regexp
	updateConfig tgbotapi.UpdateConfig
	instModule   inst.ReelModule
	repo         db.Repo
}

func CreateTgBot() TgBot {
	bot := tgBot{}

	logger := loggers.CreateLogger(loggers.LoggerConfig{
		Name:           "MainLog",
		Path:           "./MainLogs.txt",
		Level:          loggers.InfoLevel,
		WriteToConsole: false,
		UseColor:       true,
	})

	var err error
	bot.logger = logger
	bot.botApi, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		bot.logger.Error(err.Error())
	}

	bot.updateConfig = tgbotapi.NewUpdate(0)
	bot.updateConfig.Timeout = 60

	bot.reelRegex = regexp.MustCompile(`^https?://(www\.)?instagram\.com/(reel|reels)/[A-Za-z0-9_-]+/?`)
	bot.funnyCat = regexp.MustCompile(`^\/lucky_cat$`)
	bot.unluckyCat = regexp.MustCompile(`^\/unlucky_cat$`)

	bot.instModule, err = inst.NewReelsDownloader()
	if err != nil {
		bot.logger.Error(err.Error())
	}

	return &bot
}

func (bot *tgBot) Run() {

	updates := bot.botApi.GetUpdatesChan(bot.updateConfig)
	for update := range updates {
		go func(upd tgbotapi.Update) {
			bot.checkUser(update)
			bot.handleCommand(upd)
		}(update)

	}

}

func (bot *tgBot) handleCommand(update tgbotapi.Update) {
	message := entity.GetMessage(update)
	if message != nil {
		text := message.Text
		if bot.reelRegex.MatchString(text) {
			bot.handleCommandInstReel(update)
		}
		if message.Text == "/start" {
			bot.handleCommandLuckyPet(update)
		}
		if message.Text == "/help" {
			bot.handleCommandHelp(update)
		}
		if message.Text == "/random" {
			bot.handleCommandRandom(update)
		}

		return
	}
	if update.CallbackQuery != nil {
		bot.handleCommandButtonLuckyPet(update)
		return
	}
}

func (bot *tgBot) getRandomUser(parameters entity.Chat) tgbotapi.User {
	rand.Seed(time.Now().UnixNano())
	usersCount := len(parameters.Members)
	if usersCount < 2 {
		return tgbotapi.User{ID: -1}
	}

	var luckyUser tgbotapi.User
	for range 10 {
		luckyUser = parameters.Members[rand.Intn(usersCount)]
		badUser := false
		if luckyUser.ID == parameters.LastCat.ID || luckyUser.ID == parameters.LastPes.ID {
			badUser = true
		}
		if badUser == false {
			break
		}
	}

	return luckyUser
}

func (bot *tgBot) checkUser(update tgbotapi.Update) {
	message := entity.GetMessage(update)

	if message.From.UserName == "ninjaConnectionBot" {
		return
	}
	if message.Chat.Type == "private" || message.Chat.Type == "channel" {
		return
	}

	chatName := message.Chat.Title
	if chatName == "" {
		bot.logger.Warn("не удалось определить имя группы")
		return
	}

	if bot.repo.CheckUserAndGroup(message.Chat, message.From) {
		return
	}
}

func isNextDay(prev time.Time) bool {
	currentTime := time.Now()

	if prev.IsZero() {
		return true
	}

	return currentTime.Year() != prev.Year() || currentTime.YearDay() != prev.YearDay()
}

func (bot *tgBot) handleCommandLuckyPet(update tgbotapi.Update) {
	message := entity.GetMessage(update)
	text := message.Text
	bot.logger.Info(fmt.Sprintf("распознана команда: %s; User: %s", text, update.Message.From.UserName))
	if message.Chat.Type == "private" || message.Chat.Type == "channel" {
		bot.logger.Info(fmt.Sprintf("команда '/start' проигнорирована. User: %s; Name: %s", message.From.UserName, message.From.FirstName))
		return
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Нажми кнопку, чтобы выбрать котика или пса дня!")

	// Создаем inline-кнопку
	buttonCat := tgbotapi.NewInlineKeyboardButtonData("Выбрать Котеночка дня", "choose_kitten")
	buttonPes := tgbotapi.NewInlineKeyboardButtonData("Выбрать Псину дня", "choose_pes")
	keyboardCat := tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(buttonCat), tgbotapi.NewInlineKeyboardRow(buttonPes))

	msg.ReplyMarkup = keyboardCat
	_, err := bot.botApi.Send(msg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось отправить сообщение %s: %v", msg.Text, err))
	}
}

func (bot *tgBot) handleCommandButtonLuckyPet(update tgbotapi.Update) {
	defer func() {
		callback := tgbotapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.botApi.Request(callback); err != nil {
			bot.logger.Error(fmt.Sprintf("Ошибка при отправке CallbackQuery ответа: %s", err))
		}
	}()
	message := entity.GetMessage(update)
	if message.Chat.Type == "private" || message.Chat.Type == "channel" {
		bot.logger.Info(fmt.Sprintf("нажатие кнопки проигнорировано для типа чата %s. User tag: %s; Name: %s", message.Chat.Type, message.From.UserName, message.From.FirstName))
		return
	}

	if update.CallbackQuery.Data == "choose_kitten" {
		parameters := bot.repo.GetChatParameters(message.Chat.Title)
		if !parameters.LastPressButtonLuckyCat.IsZero() && time.Now().Sub(parameters.LastPressButtonLuckyCat) < 30*time.Second {
			return
		}
		parameters.LastPressButtonLuckyCat = time.Now()
		bot.logger.Info(fmt.Sprintf("нажата кнопка 'choose_kitten'; User tag: %s", message.From.UserName))

		randomUser := bot.getRandomUser(parameters)
		if randomUser.ID == -1 {
			bot.logger.Warn(fmt.Sprintf("в чате %s обнаружено слишком мало участников для выполнения команды", update.CallbackQuery.Message.Chat.Title))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пока что я знаю мало людей в чате, чтобы выбрать котеночка( Попробуй позже")
			bot.botApi.Send(msg)
			return
		} else if randomUser.ID == -2 {
			bot.logger.Warn(fmt.Sprintf("в чате %s не удалось выбить рандомного юзера", update.CallbackQuery.Message.Chat.Title))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вас слишком мало чате, не получилось выбрать котика( Попробуй позже")
			bot.botApi.Send(msg)
			return
		}

		if isNextDay(bot.lastCatChoice) == false {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, getRandomCatAnswerPhrase())
			bot.botApi.Send(msg)
			return
		}

		// Формируем сообщение с упоминанием пользователя
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("@%s, поздравляю, ты котеночек дня! Чмок в пупок!", randomUser.UserName))
		bot.botApi.Send(msg)
		bot.lastCatChoice = time.Now()
		bot.lastCat = randomUser
	} else if update.CallbackQuery.Data == "choose_pes" {
		if !bot.lastPressButtonLuckyPes.IsZero() && time.Now().Sub(bot.lastPressButtonLuckyPes) < 30*time.Second {
			return
		}
		bot.lastPressButtonLuckyPes = time.Now()
		bot.logger.Info(fmt.Sprintf("нажата кнопка 'choose_pes'; User: %s", update.CallbackQuery.Message.From.UserName))

		bot.logger.Info("нажата кнопка 'псина дня' пользователем: " + update.CallbackQuery.Message.From.UserName)

		randomUser := bot.getRandomUser(update.CallbackQuery.Message, []tgbotapi.User{bot.lastPes, bot.lastCat})
		if randomUser.ID == -1 {
			bot.logger.Warn(fmt.Sprintf("в чате %s обнаружено слишком мало участников для выполнения команды", update.CallbackQuery.Message.Chat.Title))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пока что я знаю мало людей в чате, чтобы выбрать псину( Попробуй позже")
			bot.botApi.Send(msg)
			return
		} else if randomUser.ID == -2 {
			bot.logger.Warn(fmt.Sprintf("в чате %s не удалось выбить рандомного юзера", update.CallbackQuery.Message.Chat.Title))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вас слишком мало чате, не получилось выбрать псину( Попробуй позже")
			bot.botApi.Send(msg)
			return
		}

		if isNextDay(bot.lastPesChoice) == false {
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, getRandomPesAnswerPhrase())
			bot.botApi.Send(msg)
			return
		}

		// Формируем сообщение с упоминанием пользователя
		msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("@%s, поздравляю, ты пЭс этого дня! ", randomUser.UserName))
		bot.botApi.Send(msg)
		bot.lastPesChoice = time.Now()
		bot.lastPes = randomUser
	}
}

func (bot *tgBot) handleCommandInstReel(update tgbotapi.Update) {
	message := entity.GetMessage(update)
	text := message.Text
	bot.logger.Info("распознана ссылка: " + text)

	if bot.instModule == nil {
		bot.sendMessage(message, "Модуль инстаграма временно недоступен((")
		return
	}

	answerMsg, err := bot.sendMessage(message, "Вижу ссыль, обрабатываю!")
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось отправить сообщение: %v", err))
		return
	}
	defer func() {
		deleteMsg := tgbotapi.DeleteMessageConfig{
			ChatID:    answerMsg.Chat.ID,
			MessageID: answerMsg.MessageID,
		}
		_, err = bot.botApi.Request(deleteMsg)
		if err != nil {
			bot.logger.Warn(fmt.Sprintf("не удалось удалить сообщение: %v", err))
		}
	}()

	link := bot.reelRegex.FindString(text)
	if link == "" {
		bot.logger.Error("ссылка не отработана" + text)
		bot.sendMessage(message, "Я не смог распознать ссыль((( У меня лапки...")
		return
	}

	// Скачиваем видео
	videoPath, err := bot.instModule.DownloadReel(link)
	bot.logger.Info(fmt.Sprintf("скачано видео: %s", videoPath))
	if err != nil {
		bot.sendMessage(message, "Я не смог скачать видосик((( У меня лапки...")
		return
	}

	// Отправляем видео в чат
	videoFile, _ := os.OpenFile(videoPath, os.O_RDONLY, os.ModePerm)
	videoMsg := tgbotapi.NewVideo(message.Chat.ID, tgbotapi.FileReader{
		Name:   "video.mp4",
		Reader: videoFile,
	})
	if message.Chat.Type == "group" || message.Chat.Type == "supergroup" {
		videoMsg.Caption = update.Message.From.UserName + " скинул видос"
	}

	_, err = bot.botApi.Send(videoMsg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось отправить видео в чат %s; error: %s", message.Chat.Title, err))
		return
	}

	deleteMsg := tgbotapi.DeleteMessageConfig{
		ChatID:    message.Chat.ID,
		MessageID: message.MessageID,
	}
	_, err = bot.botApi.Request(deleteMsg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось удалить сообщение: %s", err))
	}
}

func (bot *tgBot) handleCommandHelp(update tgbotapi.Update) {
	var message *tgbotapi.Message
	if update.Message == nil {
		message = update.CallbackQuery.Message
	} else {
		message = update.Message
	}

	var msg tgbotapi.MessageConfig
	if message.Chat.Type == "private" {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, helpPrivateChatOutput)
	} else {
		msg = tgbotapi.NewMessage(update.Message.Chat.ID, helpGroupChatOutput)
	}
	_, err := bot.botApi.Send(msg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось отправить сообщение %s; Chat ID: %s; From: %s", err.Error(), message.Chat.ID, message.Chat.UserName))
	}

}

func (bot *tgBot) goEatEvent() {

}

func getRandomCatAnswerPhrase() string {
	rand.Seed(time.Now().UnixNano())
	return steelCatPhrases[rand.Intn(len(steelCatPhrases))]
}

func getRandomPesAnswerPhrase() string {
	rand.Seed(time.Now().UnixNano())
	return steelPesPhrases[rand.Intn(len(steelPesPhrases))]
}

func (bot *tgBot) sendMessage(message *tgbotapi.Message, text string) (tgbotapi.Message, error) {
	msg := tgbotapi.NewMessage(message.Chat.ID, text)
	sentMsg, err := bot.botApi.Send(msg)
	if err != nil {
		errorMsg := fmt.Sprintf("не удалось отправить сообщение %s в чат %s; error: %s", text, message.Chat.Title, err)
		bot.logger.Error(errorMsg)
		return tgbotapi.Message{}, err
	}

	return sentMsg, nil
}
