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
const adminTag string = "@OrientPlus"

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
	chats        *entity.ChatsQueue
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

	bot.chats = entity.NewQueue()

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

	var chat entity.Chat
	var err error
	if message.Chat.Type != "private" {
		tx, _ := bot.repo.BeginTx()
		chat, err = bot.repo.GetChat(tx, message.Chat.ID)
		bot.repo.CommitTx(tx)
		if err != nil {
			bot.sendMessage(message, "Упс, внутренняя ошибочка вышла... мяу")
			time.Sleep(1 * time.Second)
			bot.sendMessage(message, fmt.Sprintf("Жаловаться(баг репортить) тудой: %s", adminTag))
			return
		}
	}

	if update.ChatMember != nil {
		if update.ChatMember.OldChatMember.WasKicked() {
			//TODO
			// implement and check WasKicked method!
		}
	}

	if message.Chat.Type != "private" && !chat.OpPerTime.Allow() {
		return
	}
	if message != nil {
		text := message.Text
		if bot.reelRegex.MatchString(text) {
			bot.handleCommandInstReel(update)
		}
		if strings.Contains(message.Text, "/start") && (chat.LuckyPesLimiter.Allow() || chat.LuckyCatLimiter.Allow()) {
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
		if message.Text == "/reset_chat_users" {
			bot.handleResetUsers(update)
		}
		if message.Text == "/reset_chat" {
			bot.handleResetChat(update)
		}
		if strings.Contains(message.Text, "/admin_message") {
			bot.handleSendAdminMessageToAllChats(update)
		}
		if strings.Contains(message.Text, "/admin_message_") {
			bot.handleSendAdminMessage(update)
		}
		if strings.Contains(message.Text, "/chat_list") {
			bot.handleChatList(update)
		}
		if strings.Contains(message.Text, "/ban_chat") {
			bot.handleBanChat(update)
		}

		return
	}
	if update.CallbackQuery != nil {
		bot.handleCommandButtonLuckyPet(update)
		return
	}
}

func (bot *tgBot) getRandomUser(chat entity.Chat) entity.User {
	rand.Seed(time.Now().UnixNano())
	usersCount := len(chat.Members)
	if usersCount < 2 {
		return entity.User{TgID: -1}
	}

	var luckyUser entity.User
	for range 10 {
		luckyUser = chat.Members[rand.Intn(usersCount)]
		if luckyUser.TgID == chat.LastCat.TgID || luckyUser.TgID == chat.LastPes.TgID {
			continue
		} else {
			break
		}
	}

	return luckyUser
}

func (bot *tgBot) checkUser(update tgapi.Update) {
	message := entity.GetMessage(update)
	if message.From.IsBot {
		return
	}

	tx, _ := bot.repo.BeginTx()
	var err error
	var chatId int64
	chat, exist := bot.chats.Exist(message.Chat.ID)
	if !exist {
		chat, err = bot.repo.GetChat(tx, message.Chat.ID)
		if err != nil {
			chat = entity.NewChat(message.Chat)
			chatId, err = bot.repo.AddChat(tx, &chat)
			if err != nil {
				bot.logger.Error(fmt.Sprintf("ошибка добавления чата %s:%d в бд: %s",
					message.Chat.Title, message.Chat.ID, err.Error()))
				bot.repo.RollbackTx(tx)
				return
			}
		}
	}

	var user entity.User
	user.TgID = -1
	for _, curUser := range chat.Members {
		if curUser.TgID == message.From.ID {
			user = curUser
		}
	}
	if user.TgID == -1 {
		user = entity.NewUser(message.From)
		chat.Members = append(chat.Members, user)

		userId, err := bot.repo.AddUser(tx, &user)
		if err != nil {
			bot.logger.Error(fmt.Sprintf("ошибка добавления юзера в бд: %s", err.Error()))
			chat.Members = append(chat.Members[:len(chat.Members)-1])
			if !exist {
				bot.chats.Push(chat)
			}
			bot.repo.RollbackTx(tx)
			return
		}

		_, err = bot.repo.AddUserInChat(tx, userId, chatId)
		if err != nil {
			bot.logger.Error(fmt.Sprintf("ошибка добавления юзера чат в бд: %s", err.Error()))
			chat.Members = append(chat.Members[:len(chat.Members)-1])
			if !exist {
				bot.chats.Push(chat)
			}
			bot.repo.RollbackTx(tx)
			return
		}
	}

	_, err = bot.repo.UpdateChat(tx, &chat)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("ошибка обновления чата %s;%d в бд %s", chat.Title, chat.TgID, err.Error()))
		bot.repo.RollbackTx(tx)
	} else {
		bot.repo.CommitTx(tx)
	}
	if !exist {
		bot.chats.Push(chat)
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
	bot.logger.Info(fmt.Sprintf("распознана команда: %s; User: %s", text, message.From.UserName))

	if message.Chat.Type == "private" || message.Chat.Type == "channel" {
		bot.logger.Info(fmt.Sprintf("команда '/start' проигнорирована. User: %s; Name: %s", message.From.UserName, message.From.FirstName))
		return
	}
	msg := tgapi.NewMessage(update.Message.Chat.ID, "Нажми кнопку, чтобы выбрать котика или пса дня!")

	// Создаем inline-кнопку
	buttonCat := tgapi.NewInlineKeyboardButtonData("Выбрать Котеночка дня", "choose_kitten")
	buttonPes := tgapi.NewInlineKeyboardButtonData("Выбрать Псину дня", "choose_pes")
	buttonsGroup := tgapi.NewInlineKeyboardMarkup(tgapi.NewInlineKeyboardRow(buttonCat), tgapi.NewInlineKeyboardRow(buttonPes))

	msg.ReplyMarkup = buttonsGroup
	_, err := bot.botApi.Send(msg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось отправить сообщение %s: %s", msg.Text, err.Error()))
	}
}

func (bot *tgBot) handleCommandButtonLuckyPet(update tgapi.Update) {
	defer func() {
		callback := tgapi.NewCallback(update.CallbackQuery.ID, "")
		if _, err := bot.botApi.Request(callback); err != nil {
			bot.logger.Error(fmt.Sprintf("Ошибка при отправке CallbackQuery ответа в чат %s:%d: %s",
				update.FromChat().Title, update.FromChat().ID, err.Error()))
		}
	}()
	message := entity.GetMessage(update)
	if message.Chat.Type == "private" || message.Chat.Type == "channel" {
		bot.logger.Info(fmt.Sprintf("нажатие кнопки проигнорировано для типа чата %s. User: %s; Name: %s",
			message.Chat.Type, message.From.UserName, message.From.FirstName))
		return
	}

	tx, _ := bot.repo.BeginTx()
	chat, err := bot.repo.GetChat(tx, message.Chat.ID)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("ошибка получения чата %s:%d из бд: %s",
			message.Chat.Title, message.Chat.ID, err.Error()))
		bot.sendCrushMessage(message)
		bot.repo.RollbackTx(tx)
		return
	}

	if update.CallbackQuery.Data == "choose_kitten" {
		if !chat.LuckyCatLimiter.Allow() {
			//TODO fix log span here
			bot.logger.Warn(fmt.Sprintf("юзер %s(%d) из чата %s(%d) дудосит бота",
				message.From.UserName, message.From.ID, message.Chat.Title, message.Chat.ID))
			bot.repo.CommitTx(tx)
			return
		}
		bot.logger.Info(fmt.Sprintf("нажата кнопка 'choose_kitten'; User %s:%d; Chat: %s:%d",
			message.From.UserName, message.From.ID, message.Chat.Title, message.Chat.ID))

		randomUser := bot.getRandomUser(chat)
		if randomUser.TgID == -1 {
			bot.logger.Warn(fmt.Sprintf("в чате %s:%d обнаружено слишком мало участников для выполнения команды",
				message.Chat.Title, message.Chat.ID))
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
				"Пока что я знаю мало людей в чате, чтобы выбрать котеночка( Попробуй позже")
			bot.botApi.Send(msg)
			bot.repo.CommitTx(tx)
			return
		}

		if isNextDay(chat.LastCatChoice) == false {
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, getRandomCatAnswerPhrase())
			bot.botApi.Send(msg)
			bot.repo.CommitTx(tx)
			return
		}

		chat.LastCatChoice = time.Now()
		chat.LastCat = &randomUser
		_, err = bot.repo.UpdateChat(tx, &chat)
		if err != nil {
			bot.logger.Error(fmt.Sprintf("ошибка обновления чата[name:%s;ID:%d] в бд: %s",
				chat.Title, chat.TgID, err.Error()))
			bot.sendCrushMessage(message)
			bot.repo.RollbackTx(tx)
		} else {
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
				fmt.Sprintf("@%s, поздравляю, ты котеночек дня! Чмок в пупок!", randomUser.UserName))
			bot.botApi.Send(msg)
			bot.repo.CommitTx(tx)
		}
	} else if update.CallbackQuery.Data == "choose_pes" {
		if !chat.LuckyPesLimiter.Allow() {
			// TODO fix log span here
			bot.logger.Warn(fmt.Sprintf("юзер %s(%d) из чата %s(%d) дудосит бота",
				message.From.UserName, message.From.ID, message.Chat.Title, message.Chat.ID))
			bot.repo.CommitTx(tx)
			return
		}
		bot.logger.Info(fmt.Sprintf("нажата кнопка 'choose_pes'; User %s:%d; Chat: %s:%d",
			message.From.UserName, message.From.ID, message.Chat.Title, message.Chat.ID))

		randomUser := bot.getRandomUser(chat)
		if randomUser.TgID == -1 {
			bot.logger.Warn(fmt.Sprintf("в чате %s:%d обнаружено слишком мало участников для выполнения команды",
				message.Chat.Title, message.Chat.ID))
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "Пока что я знаю мало людей в чате, чтобы выбрать псину(( Попробуй позже")
			bot.botApi.Send(msg)
			bot.repo.CommitTx(tx)
			return
		}

		if isNextDay(chat.LastPesChoice) == false {
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID, getRandomPesAnswerPhrase())
			bot.botApi.Send(msg)
			bot.repo.CommitTx(tx)
			return
		}

		chat.LastPesChoice = time.Now()
		chat.LastPes = &randomUser
		_, err = bot.repo.UpdateChat(tx, &chat)
		if err != nil {
			bot.logger.Error(fmt.Sprintf("ошибка обновления чата[name:%s;ID:%d] в бд: %s",
				chat.Title, chat.TgID, err.Error()))
			bot.sendCrushMessage(message)
			bot.repo.RollbackTx(tx)
		} else {
			msg := tgapi.NewMessage(update.CallbackQuery.Message.Chat.ID,
				fmt.Sprintf("@%s, поздравляю, ты пЭс этого дня! ", randomUser.UserName))
			bot.botApi.Send(msg)
			bot.repo.CommitTx(tx)
		}
	}
}

func (bot *tgBot) handleCommandInstReel(update tgapi.Update) {
	message := entity.GetMessage(update)
	text := message.Text
	bot.logger.Info(fmt.Sprintf("распознана ссылка %s в чате %s:%d",
		message.From.UserName, message.From.ID, message.Chat.Title))

	if bot.instModule == nil {
		bot.sendMessage(message, "Модуль инстаграма временно недоступен((")
		return
	}

	answerMsg, err := bot.sendMessage(message, "Вижу ссыль, обрабатываю!")
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось отправить сообщение в чат %s:%d: %s",
			message.From.UserName, message.From.ID, err.Error()))
		return
	}
	defer func() {
		deleteMsg := tgapi.DeleteMessageConfig{
			ChatID:    answerMsg.Chat.ID,
			MessageID: answerMsg.MessageID,
		}
		_, err = bot.botApi.Request(deleteMsg)
		if err != nil {
			bot.logger.Warn(fmt.Sprintf("не удалось удалить сообщение в чате %s:%d : %s",
				message.From.UserName, message.From.ID, err.Error()))
		}
	}()

	link := bot.reelRegex.FindString(text)
	if link == "" {
		bot.logger.Error("ссылка не отработана" + text)
		bot.sendMessage(message, "Я не смог распознать ссыль((( У меня лапки...")
		return
	}

	// Скачиваем видео
	videoPath, err := bot.instModule.DownloadReelFastdl(link)
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
		bot.logger.Error(fmt.Sprintf("не удалось отправить видео в чат %s:%d; error: %s",
			message.Chat.Title, message.Chat.ID, err))
		return
	}

	deleteMsg := tgapi.DeleteMessageConfig{
		ChatID:    message.Chat.ID,
		MessageID: message.MessageID,
	}
	_, err = bot.botApi.Request(deleteMsg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("не удалось удалить сообщение в чате %s:%d : %s",
			message.Chat.Title, message.Chat.ID, err.Error()))
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
		bot.logger.Error(fmt.Sprintf("не удалось отправить сообщение: %s в чат %s:%d; From: %s:%s",
			err.Error(), message.Chat.Title, message.Chat.ID, message.From.UserName, message.From.FirstName))
	}

}

func (bot *tgBot) handleCommandRandom(update tgapi.Update) {
	// TODO implement this method

}

func (bot *tgBot) handleChatList(update tgapi.Update) {

}

func (bot *tgBot) handleSendAdminMessage(update tgapi.Update) {
	message := entity.GetMessage(update)

	tx, _ := bot.repo.BeginTx()
	isAdmin, err := bot.repo.IsAdmin(tx, message.From.ID)
	if err != nil {
		bot.repo.RollbackTx(tx)
		bot.logger.Error(fmt.Sprintf("ошибка определения статуса админа для юзера %s:%s:%d: %s",
			message.From.UserName, message.From.FirstName, message.From.ID, err.Error()))
	}
	if !isAdmin {
		bot.logger.Error(fmt.Sprintf("попытка выполнения команды '/admin_message_<>' не админом %s:%s:%d",
			message.From.UserName, message.From.FirstName, message.From.ID))
		return
	}

	var chatId int64
	var adminMessage string
	_, err = fmt.Sscanf(message.Text, "/admin_message_%d %s", &chatId, &adminMessage)
	if err != nil {
		bot.repo.RollbackTx(tx)
		bot.logger.Error(fmt.Sprintf("ошибка парсинга команды '/admin_message_<chat id>': %s", err.Error()))
		return
	}

	msg := tgapi.NewMessage(chatId, adminMessage)
	_, err = bot.botApi.Send(msg)
	if err != nil {
		bot.logger.Error(fmt.Sprintf("ошибка отправки сообщения админа в чат %s:%d",
			message.Chat.Title, message.Chat.ID))
		bot.sendMessage(message, fmt.Sprintf("Не удалось отправить сообщение в чат %s:%d : %s",
			message.Chat.Title, message.Chat.ID, err.Error()))
	} else {
		bot.sendMessage(message, fmt.Sprintf("Сообщение отправлено в чат %s:%d",
			message.Chat.Title, message.Chat.ID))
	}
	bot.repo.CommitTx(tx)
}

func (bot *tgBot) handleSendAdminMessageToAllChats(update tgapi.Update) {
	message := entity.GetMessage(update)

	tx, _ := bot.repo.BeginTx()
	isAdmin, err := bot.repo.IsAdmin(tx, message.From.ID)
	if err != nil {
		bot.repo.RollbackTx(tx)
		bot.logger.Error(fmt.Sprintf("ошибка определения статуса админа для юзера %s:%s:%d: %s",
			message.From.UserName, message.From.FirstName, message.From.ID, err.Error()))
	}
	if !isAdmin {
		bot.logger.Error(fmt.Sprintf("попытка выполнения команды '/admin_message' не админом %s:%s:%d",
			message.From.UserName, message.From.FirstName, message.From.ID))
		return
	}

	adminMessage := strings.TrimPrefix(message.Text, "/admin_message ")
	chats, err := bot.repo.GetAllChats(tx)
	if err != nil {
		bot.repo.RollbackTx(tx)
		bot.logger.Error("ошибка получения всех чатов из бд")
		return
	}
	for _, chat := range chats {
		msg := tgapi.NewMessage(chat.TgID, adminMessage)
		_, err = bot.botApi.Send(msg)
		if err != nil {
			bot.logger.Error(fmt.Sprintf("ошибка отправки сообщения админа в чат %s:%d",
				message.Chat.Title, message.Chat.ID))
			bot.sendMessage(message, fmt.Sprintf("Не удалось отправить сообщение в чат %s:%d : %s",
				message.Chat.Title, message.Chat.ID, err.Error()))
			continue
		}
		bot.sendMessage(message, fmt.Sprintf("Сообщение отправлено в чат %s:%d",
			message.Chat.Title, message.Chat.ID))
	}
	bot.repo.CommitTx(tx)
}

func (bot *tgBot) handleBanChat(update tgapi.Update) {

}

func (bot *tgBot) handleResetUsers(update tgapi.Update) {
	message := entity.GetMessage(update)
	if message.Chat.Type == "private" || message.Chat.Type == "channel" {
		bot.logger.Info(fmt.Sprintf("команда '/reset_chat_users' проигнорирована для типа чата %s. User: %s:%s:%d",
			message.Chat.Type, message.From.UserName, message.From.FirstName, message.From.ID))
		return
	}

	tx, _ := bot.repo.BeginTx()
	chat, err := bot.repo.GetChat(tx, message.Chat.ID)
	if err != nil {
		bot.repo.RollbackTx(tx)
		bot.logger.Error(fmt.Sprintf("ошибка получения чата %s:%d из бд: %s",
			message.Chat.Title, message.Chat.ID, err.Error()))

		bot.sendCrushMessage(message)
		return
	}

	for _, user := range chat.Members {
		err = bot.repo.DeleteUser(tx, &user)
		if err != nil {
			bot.repo.RollbackTx(tx)
			bot.logger.Error(fmt.Sprintf("ошибка удаления юзера %s:%d из бд: %s",
				message.Chat.Title, message.Chat.ID, err.Error()))
			bot.sendCrushMessage(message)
			return
		}
		err = bot.repo.DeleteUserFromChat(tx, user.TgID, chat.TgID)
		if err != nil {
			bot.repo.RollbackTx(tx)
			bot.logger.Error(fmt.Sprintf("ошибка удаления юзера %s:%d из бд: %s",
				message.Chat.Title, message.Chat.ID, err.Error()))
			bot.sendCrushMessage(message)
			return
		}
	}

	schat, ex := bot.chats.Exist(chat.TgID)
	if ex {
		schat.Members = nil
	}

	chat.LastCat = nil
	chat.LastPes = nil
	chat.Members = nil

	_, err = bot.repo.UpdateChat(tx, &chat)
	if err != nil {
		bot.repo.RollbackTx(tx)
		bot.sendCrushMessage(message)
		return
	}
	bot.repo.CommitTx(tx)
	bot.sendMessage(message, "Готово, ботик-котик не знает ни одного участника из чата")
	return
}

func (bot *tgBot) handleResetChat(update tgapi.Update) {

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

func (bot *tgBot) sendCrushMessage(message *tgapi.Message) {
	// TODO send report to admin

	msg := tgapi.NewMessage(message.Chat.ID, "Упс, случился локальный краш, кажись шото сломалось, сейчпукс ищйгун m4ur!№@!##... мяу")
	bot.botApi.Send(msg)
	time.Sleep(1 * time.Second)
	msg = tgapi.NewMessage(message.Chat.ID, fmt.Sprintf("Жаловаться на краш тудой: %s", adminTag))
	bot.botApi.Send(msg)
}
