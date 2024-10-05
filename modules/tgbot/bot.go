package tgbot

import (
	"fmt"
	"math/rand"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	tgapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/robfig/cron/v3"
	"superserver/db"
	"superserver/entity"
	"superserver/loggers"
	vl "superserver/modules/tgbot/inst"
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
	botApi       *tgapi.BotAPI
	reelRegex    *regexp.Regexp
	funnyCat     *regexp.Regexp
	unluckyCat   *regexp.Regexp
	eventRegex   *regexp.Regexp
	updateConfig tgapi.UpdateConfig
	instModule   vl.ReelModule
	repo         db.Repo
	cron         *cron.Cron
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
	bot.botApi, err = tgapi.NewBotAPI(token)
	if err != nil {
		bot.logger.Error(err.Error())
	}

	bot.updateConfig = tgapi.NewUpdate(0)
	bot.updateConfig.Timeout = 15

	bot.cron = cron.New()
	bot.cron.Start()

	bot.reelRegex = regexp.MustCompile(`^https?://(www\.)?instagram\.com/(reel|reels)/[A-Za-z0-9_-]+/?`)
	bot.funnyCat = regexp.MustCompile(`^\/lucky_cat$`)
	bot.unluckyCat = regexp.MustCompile(`^\/unlucky_cat$`)
	bot.eventRegex = regexp.MustCompile(`^(.*?);\s*([\*\d\/,\-]+ [\*\d\/,\-]+ [\*\d\/,\-]+ [\*\d\/,\-]+ [\*\d\/,\-]+(?: [\*\d\/,\-]+)?);\s*(.*)$`)

	bot.instModule, err = vl.NewReelsDownloader()
	if err != nil {
		bot.logger.Error(err.Error())
	}

	return &bot
}

func (bot *tgBot) Run() {

	updates := bot.botApi.GetUpdatesChan(bot.updateConfig)
	for update := range updates {
		go func(upd tgapi.Update) {
			bot.checkUser(update)
			bot.handleCommand(upd)
		}(update)

	}

}

func (bot *tgBot) handleCommand(update tgapi.Update) {
	message := entity.GetMessage(update)
	chat, _ := bot.repo.GetChatParameters(message.Chat.Title)
	if !chat.OpPerTime.Allow() {
		return
	}
	if message != nil {
		text := message.Text
		if bot.reelRegex.MatchString(text) {
			bot.handleCommandInstReel(update)
		}
		if message.Text == "/start" && (chat.LuckyPesLimiter.Allow() || chat.LuckyCatLimiter.Allow()) {
			bot.handleCommandLuckyPet(update)
		}
		if message.Text == "/help" {
			bot.handleCommandHelp(update)
		}
		if strings.Contains(message.Text, "/event") {
			bot.HandleEvent(update)
		}
		if message.Text == "/event_list" {
			bot.HandleEventList(update)
		}
		if strings.Contains(message.Text, "/del_event") {
			bot.HandleDeleteEvent(update)
		}
		if message.Text == "/del_all_events" {
			bot.HandleDelAllEvents(update)
		}
		if message.Text == "/random" {
			bot.handleCommandRandom(update)
		}
		if message.Text == "/admin_message" {
			bot.handleCommandNewVersion(update)
		}

		return
	}
	if update.CallbackQuery != nil {
		bot.handleCommandButtonLuckyPet(update)
		return
	}
}

func (bot *tgBot) getRandomUser(parameters entity.Chat) entity.User {
	rand.Seed(time.Now().UnixNano())
	usersCount := len(parameters.Members)
	if usersCount < 2 {
		return entity.User{TgID: -1}
	}

	var luckyUser entity.User
	for range 10 {
		luckyUser = parameters.Members[rand.Intn(usersCount)]
		badUser := false
		if luckyUser.TgID == parameters.LastCat.TgID || luckyUser.TgID == parameters.LastPes.TgID {
			badUser = true
		}
		if badUser == false {
			break
		}
	}

	return luckyUser
	/*return tgapi.User{
		TgID:                      luckyUser.TgID,
		IsBot:                   luckyUser.IsBot,
		FirstName:               luckyUser.FirstName,
		LastName:                luckyUser.LastName,
		UserName:                luckyUser.UserName,
		LanguageCode:            luckyUser.LanguageCode,
		CanJoinGroups:           luckyUser.CanJoinGroups,
		CanReadAllGroupMessages: luckyUser.CanReadAllGroupMessages,
		SupportsInlineQueries:   luckyUser.SupportsInlineQueries,
	}*/
}

func (bot *tgBot) checkUser(update tgapi.Update) {
	message := entity.GetMessage(update)

	if message.From.UserName == "ninjaConnectionBot" {
		return
	}

	chat := entity.NewChat(message.Chat)
	user := entity.NewUser(message.From)
	_, err := bot.repo.AddUser(&user)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("ошибка добавления юзера %s в БД %s", user.UserName, err.Error()))
	}
	_, err = bot.repo.AddChat(&chat)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("ошибка добавления чата %s в БД %s", chat.Title, err.Error()))
	}
	return
}

func isNextDay(prev time.Time) bool {
	currentTime := time.Now()

	if prev.IsZero() {
		return true
	}

	return currentTime.Year() != prev.Year() || currentTime.YearDay() != prev.YearDay()
}

func (bot *tgBot) handleCommandLuckyPet(update tgapi.Update) {
	message := entity.GetMessage(update)
	text := message.Text
	bot.logger.Info(fmt.Sprintf("распознана команда: %s; User: %s", text, update.Message.From.UserName))
	if message.Chat.Type == "private" || message.Chat.Type == "channel" {
		bot.logger.Info(fmt.Sprintf("команда '/start' проигнорирована. User: %s; Name: %s", message.From.UserName, message.From.FirstName))
		return
	}
	msg := tgapi.NewMessage(update.Message.Chat.ID, "Нажми кнопку, чтобы выбрать котика или пса дня!")

	// Создаем inline-кнопку
	buttonCat := tgapi.NewInlineKeyboardButtonData("Выбрать Котеночка дня", "choose_kitten")
	buttonPes := tgapi.NewInlineKeyboardButtonData("Выбрать Псину дня", "choose_pes")
	keyboardCat := tgapi.NewInlineKeyboardMarkup(tgapi.NewInlineKeyboardRow(buttonCat), tgapi.NewInlineKeyboardRow(buttonPes))

	msg.ReplyMarkup = keyboardCat
	_, err := bot.botApi.Send(msg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось отправить сообщение %s: %v", msg.Text, err))
	}
}

func (bot *tgBot) handleCommandButtonLuckyPet(update tgapi.Update) {
	defer func() {
		callback := tgapi.NewCallback(update.CallbackQuery.ID, "")
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
		parameters, _ := bot.repo.GetChatParameters(message.Chat.Title)
		if !parameters.LuckyCatLimiter.Allow() {
			bot.logger.Warn(fmt.Sprintf("юзер %s из чата %s дудосит бота", message.Chat.UserName, message.Chat.Title))
			return
		}
		bot.logger.Info(fmt.Sprintf("нажата кнопка 'choose_kitten'; User tag: %s", message.Chat.UserName))

		randomUser := bot.getRandomUser(parameters)
		if randomUser.TgID == -1 {
			bot.logger.Warn(fmt.Sprintf("в чате %s обнаружено слишком мало участников для выполнения команды", update.CallbackQuery.Message.Chat.Title))
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пока что я знаю мало людей в чате, чтобы выбрать котеночка( Попробуй позже")
			bot.botApi.Send(msg)
			return
		} else if randomUser.TgID == -2 {
			bot.logger.Warn(fmt.Sprintf("в чате %s не удалось выбить рандомного юзера", update.CallbackQuery.Message.Chat.Title))
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вас слишком мало чате, не получилось выбрать котика( Попробуй позже")
			bot.botApi.Send(msg)
			return
		}

		if isNextDay(parameters.LastCatChoice) == false {
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, getRandomCatAnswerPhrase())
			bot.botApi.Send(msg)
			return
		}

		// Формируем сообщение с упоминанием пользователя
		msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("@%s, поздравляю, ты котеночек дня! Чмок в пупок!", randomUser.UserName))
		bot.botApi.Send(msg)
		parameters.LastCatChoice = time.Now()
		parameters.LastCat = randomUser
	} else if update.CallbackQuery.Data == "choose_pes" {
		parameters, _ := bot.repo.GetChatParameters(message.Chat.Title)
		if !parameters.LuckyPesLimiter.Allow() {
			bot.logger.Warn(fmt.Sprintf("юзер %s из чата %s дудосит бота", message.Chat.UserName, message.Chat.Title))
			return
		}
		bot.logger.Info(fmt.Sprintf("нажата кнопка 'choose_pes'; User tag: %s", message.Chat.UserName))

		randomUser := bot.getRandomUser(parameters)
		if randomUser.TgID == -1 {
			bot.logger.Warn(fmt.Sprintf("в чате %s обнаружено слишком мало участников для выполнения команды", message.Chat.Title))
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пока что я знаю мало людей в чате, чтобы выбрать псину( Попробуй позже")
			bot.botApi.Send(msg)
			return
		} else if randomUser.TgID == -2 {
			bot.logger.Warn(fmt.Sprintf("в чате %s не удалось выбить рандомного юзера", update.CallbackQuery.Message.Chat.Title))
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Вас слишком мало чате, не получилось выбрать псину( Попробуй позже")
			bot.botApi.Send(msg)
			return
		}

		if isNextDay(parameters.LastPesChoice) == false {
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, getRandomPesAnswerPhrase())
			bot.botApi.Send(msg)
			return
		}

		// Формируем сообщение с упоминанием пользователя
		msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, fmt.Sprintf("@%s, поздравляю, ты пЭс этого дня! ", randomUser.UserName))
		bot.botApi.Send(msg)
		parameters.LastPesChoice = time.Now()
		parameters.LastPes = randomUser
	}
}

func (bot *tgBot) handleCommandInstReel(update tgapi.Update) {
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
		deleteMsg := tgapi.DeleteMessageConfig{
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
	videoMsg := tgapi.NewVideo(message.Chat.ID, tgapi.FileReader{
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

	deleteMsg := tgapi.DeleteMessageConfig{
		ChatID:    message.Chat.ID,
		MessageID: message.MessageID,
	}
	_, err = bot.botApi.Request(deleteMsg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось удалить сообщение: %s", err))
	}
}

func (bot *tgBot) handleCommandHelp(update tgapi.Update) {
	var message *tgapi.Message
	if update.Message == nil {
		message = update.CallbackQuery.Message
	} else {
		message = update.Message
	}

	var msg string
	if message.Chat.Type == "private" {
		msg = helpPrivateChatOutput
	} else {
		msg = helpGroupChatOutput
	}
	_, err := bot.sendMessage(message, msg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось отправить сообщение %s; Chat TgID: %s; From: %s", err.Error(), message.Chat.ID, message.Chat.UserName))
	}

}

func (bot *tgBot) handleCommandRandom(update tgapi.Update) {
	// TODO implement this method
	/*var message *tgapi.Message
	if update.Message == nil {
		message = update.CallbackQuery.Message
	} else {
		message = update.Message
	}*/

}

func (bot *tgBot) handleCommandNewVersion(update tgapi.Update) {
	// TODO implement this method
}

func (bot *tgBot) HandleEvent(update tgapi.Update) {
	var message *tgapi.Message
	if update.Message == nil {
		message = update.CallbackQuery.Message
	} else {
		message = update.Message
	}

	chat, chatId := bot.repo.GetChatParameters(message.Chat.Title)

	match := bot.eventRegex.FindStringSubmatch(message.Text)
	if match == nil || len(match) < 4 {
		bot.sendMessage(message, "Кажется, параметры неверные(")
		return
	}
	event := entity.ChatEvent{
		CronID:     0,
		TgID:       chat.TgID,
		Title:      match[1],
		Message:    match[2],
		TimeConfig: match[3],
	}

	cronID, err := bot.cron.AddFunc(event.TimeConfig, func() {
		// пример: "30 12 * * 1-5" каждый будний день в 12:30
		bot.sendMessage(message, event.Message)
	})
	if err != nil {
		bot.sendMessage(message, "Мне не удалось создать ивент для тебя( У меня лапки((")
		return
	}
	bot.sendMessage(message, fmt.Sprintf("Отлично, я создал ивент %s. Его TgID: %d", event.Title, event.CronID))

	event.CronID = int64(cronID)
	chat.Events = append(chat.Events, event)

	eventId, err := bot.repo.AddEvent(event)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("ошибка добавления ивента в бд: %s", err))
		return
	}

	eventId, err = bot.repo.AddEventInChat(eventId, chatId)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("ошибка добавления ивента в чат в бд: %s", err))
		return
	}

	return
}

func (bot *tgBot) HandleEventList(update tgapi.Update) {
	var message *tgapi.Message
	if update.Message == nil {
		message = update.CallbackQuery.Message
	} else {
		message = update.Message
	}

	chat, _ := bot.repo.GetChatParameters(message.Chat.Title)
	if len(chat.Events) == 0 {
		bot.sendMessage(message, "Похоже в чатике нет активных ивентов")
		return
	}

	answer := "Ивенты чата:"
	for _, event := range chat.Events {
		answer = answer + "\nID: " + strconv.Itoa(int(event.CronID)) + "\nИмя ивента: " + event.Title + "\n"
	}
	bot.sendMessage(message, answer)
	return
}

func (bot *tgBot) HandleDeleteEvent(update tgapi.Update) {
	var message *tgapi.Message
	if update.Message == nil {
		message = update.CallbackQuery.Message
	} else {
		message = update.Message
	}

	var eventCronIdentifier string
	re := regexp.MustCompile(`^/del_event\s+(.*)$`)
	match := re.FindStringSubmatch(message.Text)
	if match == nil || len(match) < 2 {
		bot.sendMessage(message, "Кажется, параметры неверные")
		return
	}

	eventCronIdentifier = match[1]
	chat, _ := bot.repo.GetChatParameters(message.Chat.Title)
	eventID, err := strconv.Atoi(eventCronIdentifier)
	if err != nil {
		bot.sendMessage(message, "Ты передал невалидный id, чувырло")
		return
	}

	delIndex := sort.Search(len(chat.Events), func(i int) bool {
		return int(chat.Events[i].CronID) == eventID
	})

	if delIndex >= len(chat.Events) {
		bot.logger.Warn(fmt.Sprintf("Не найден cron ID <%d> в чате <%s>", eventCronIdentifier, chat.Title))
		bot.sendMessage(message, "Распознал переданный id, но не нашел его в вашем чатике. Сделаем вид, что все удалено)))")
		time.Sleep(time.Second * 3)
		bot.sendMessage(message, "Если не удалилось, то мои полномочия все, пишите бате @OrientPlus")
		return
	}

	bot.cron.Remove(cron.EntryID(eventID))

	err = bot.repo.DeleteEventFromChatByExternalId(int64(eventID), chat.TgID)

	err = bot.repo.DeleteEvent(chat.Events[delIndex])
	if err != nil {
		bot.logger.Error(fmt.Sprintf("ошибка обновления чата %s:%d в БД: %s", chat.Title, chat.TgID, err.Error()))
	}

	chat.Events = append(chat.Events[:delIndex], chat.Events[delIndex+1:]...)
}

func (bot *tgBot) HandleDelAllEvents(update tgapi.Update) {
	var message *tgapi.Message
	if update.Message == nil {
		message = update.CallbackQuery.Message
	} else {
		message = update.Message
	}

	chat, _ := bot.repo.GetChatParameters(message.Chat.Title)
	for _, event := range chat.Events {
		bot.cron.Remove(cron.EntryID(event.CronID))

		err := bot.repo.DeleteEventFromChatByExternalId(event.CronID, event.TgID)
		if err != nil {
			bot.logger.Error(fmt.Sprintf("ошибка удаления ивента из чата в бд: %s", err))
			continue
		}

		err = bot.repo.DeleteEvent(event)
		if err != nil {
			bot.logger.Error(fmt.Sprintf("ошибка удаления ивента из БД: %s", err))
			continue
		}
	}
	chat.Events = nil

	bot.sendMessage(message, "Супер, я удалил все ивенты чата!")
	return
}

func getRandomCatAnswerPhrase() string {
	rand.Seed(time.Now().UnixNano())
	return steelCatPhrases[rand.Intn(len(steelCatPhrases))]
}

func getRandomPesAnswerPhrase() string {
	rand.Seed(time.Now().UnixNano())
	return steelPesPhrases[rand.Intn(len(steelPesPhrases))]
}

func (bot *tgBot) sendMessage(message *tgapi.Message, text string) (tgapi.Message, error) {
	msg := tgapi.NewMessage(message.Chat.ID, text)
	sentMsg, err := bot.botApi.Send(msg)
	if err != nil {
		errorMsg := fmt.Sprintf("не удалось отправить сообщение %s в чат %s; error: %s", text, message.Chat.Title, err)
		bot.logger.Error(errorMsg)
		return tgapi.Message{}, err
	}

	return sentMsg, nil
}
