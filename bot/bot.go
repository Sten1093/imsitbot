package bot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"tgbot/database"
	"tgbot/parser"
)

const adminID int64 = 925884466

var awaitingAdminMessage bool
var adminChatID int64

func Bot() {
	err := godotenv.Load("./resources/metadata/token/test_bot_token.env")
	if err != nil {
		log.Fatal("Ошибка загрузки токена: ", err)
	}

	botToken := os.Getenv("BOT_TOKEN")
	const dbPath = "./resources/data/users.sqlite"

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal("Ошибка инициализации бота: ", err)
	}
	log.Printf("Авторизован как %s", bot.Self.UserName)

	updates := configureUpdates(bot)
	userDAO := database.NewUserDAO(dbPath)
	defer userDAO.Close()

	users := make(map[int64]*database.User)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		user := getOrCreateUser(chatID, userDAO, users)
		log.Printf("Сообщение от %s: %s", update.Message.From.UserName, update.Message.Text)

		handleState(bot, update, user, userDAO, users)
	}
}

// Настройка канала обновлений
func configureUpdates(bot *tgbotapi.BotAPI) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	return bot.GetUpdatesChan(u)
}

// Получение или создание пользователя
func getOrCreateUser(chatID int64, userDAO *database.UserDAO, users map[int64]*database.User) *database.User {
	if user, exists := users[chatID]; exists {
		return user
	}

	dbUser, err := userDAO.GetUser(chatID)
	if err == nil && dbUser != nil {
		users[chatID] = dbUser
		return dbUser
	}

	user := &database.User{ID: chatID, State: "hello"}
	if err := userDAO.SaveUser(user); err != nil {
		log.Printf("Ошибка сохранения нового пользователя: %s", err)
	}
	users[chatID] = user
	return user
}

// Обработка состояний
func handleState(bot *tgbotapi.BotAPI, update tgbotapi.Update, user *database.User, userDAO *database.UserDAO, users map[int64]*database.User) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	switch user.State {
	case "hello":
		handleHelloState(bot, update, user, userDAO, users)
	case "waiting_for_education":
		handleEducationState(bot, chatID, text, user, userDAO)
	case "waiting_for_course":
		handleCourseState(bot, chatID, text, user)
	case "select_group":
		handleGroupState(bot, chatID, text, user)
	case "select_format":
		handleFormatState(bot, chatID, text, user, userDAO, users)
	case "waiting_for_return":
		handleReturnState(bot, chatID, text, user, userDAO)
	case "teacher":
		handleTeacherState(bot, chatID, text, user)
	case "corpus_info":
		handleCorpusState(bot, chatID, text, user)
	}
}

// Обработка состояния "hello"
func handleHelloState(bot *tgbotapi.BotAPI, update tgbotapi.Update, user *database.User, userDAO *database.UserDAO, users map[int64]*database.User) {
	chatID := update.Message.Chat.ID
	text := update.Message.Text

	users[chatID] = user
	switch text {
	case "🗓Расписание🗓":
		if user.EducationLevel == "" {
			sendKeyboardMessage(bot, chatID, "Выбери форму обучения", createEducationKeyboard, user, "waiting_for_education")
		} else {
			sendKeyboardMessage(bot, chatID, "Выбери курс:", createCourseKeyboardUp, user, "waiting_for_course")
		}
	case "👱‍♂️Найти препода👱":
		sendKeyboardMessage(bot, chatID, "Напиши фамилию преподавателя в формате (Иванов)", nil, user, "teacher")
	case "🏢Найти корпус🏫":
		sendKeyboardMessage(bot, chatID, "Напиши номер корпуса", createCorpusNum, user, "corpus_info")
	case "/start":
		sendKeyboardMessage(bot, chatID, "Привет, Я бот для помощи тебе в твоем обучении!", createHelloKeyboard, user, "")
	case "/send":
		if update.Message.From.ID != adminID {
			sendKeyboardMessage(bot, chatID, "У вас нет прав для выполнения этой команды.", createHelloKeyboard, user, "")
			return
		}
		awaitingAdminMessage = true
		adminChatID = chatID
		sendKeyboardMessage(bot, chatID, "Пожалуйста, введите сообщение для всех пользователей.", createHelloKeyboard, user, "")
	default:
		if awaitingAdminMessage && update.Message.From.ID == adminID && chatID == adminChatID {
			handleAdminMessage(bot, chatID, text, userDAO)
		} else {
			sendKeyboardMessage(bot, chatID, "Используй клавиатуру", createHelloKeyboard, user, "")
		}
	}
}

// Обработка сообщения администратора
func handleAdminMessage(bot *tgbotapi.BotAPI, chatID int64, messageText string, userDAO *database.UserDAO) {
	if messageText == "/cancel" {
		awaitingAdminMessage = false
		sendKeyboardMessage(bot, chatID, "Рассылка отменена.", createHelloKeyboard, nil, "")
		return
	}

	usersList, err := userDAO.GetAllUsers()
	if err != nil {
		log.Printf("Ошибка получения списка пользователей: %s", err)
		sendKeyboardMessage(bot, chatID, "Ошибка при отправке сообщений.", createHelloKeyboard, nil, "")
		awaitingAdminMessage = false
		return
	}

	for _, u := range usersList {
		if _, err := bot.Send(tgbotapi.NewMessage(u.ID, messageText)); err != nil {
			log.Printf("Ошибка отправки сообщения пользователю %d: %s", u.ID, err)
		}
	}

	sendKeyboardMessage(bot, chatID, "Сообщение успешно отправлено всем пользователям.", createHelloKeyboard, nil, "")
	awaitingAdminMessage = false
}

// Обработка состояния выбора формы образования
func handleEducationState(bot *tgbotapi.BotAPI, chatID int64, text string, user *database.User, userDAO *database.UserDAO) {
	user.EducationLevel = text
	switch text {
	case "Высшее":
		sendKeyboardMessage(bot, chatID, "Выбери курс:", createCourseKeyboardUp, user, "waiting_for_course")
	case "Среднее":
		sendKeyboardMessage(bot, chatID, "Выбери курс:", createCourseKeyboardDown, user, "waiting_for_course")
	case "⬅️Назад":
		sendKeyboardMessage(bot, chatID, "Попробуем снова", createHelloKeyboard, user, "hello")
	default:
		sendKeyboardMessage(bot, chatID, "Используй клавиатуру", createEducationKeyboard, user, "")
		return
	}
	saveUser(user, userDAO, chatID)
}

// Обработка состояния выбора курса
func handleCourseState(bot *tgbotapi.BotAPI, chatID int64, text string, user *database.User) {
	user.Course = text
	switch text {
	case "⬅️Назад":
		sendKeyboardMessage(bot, chatID, "Попробуем еще раз", createHelloKeyboard, user, "hello")
	case "🤓 1 курс", "😎 2 курс", "🧐 3 курс", "🎓 4 курс", "🫠 5 курс":
		sendKeyboardMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel), user, "select_group")
	default:
		keyboard := createCourseKeyboardUp
		if user.EducationLevel == "Среднее" {
			keyboard = createCourseKeyboardDown
		}
		sendKeyboardMessage(bot, chatID, "Нажми кнопочку на клавиатуре", keyboard, user, "")
	}
}

// Обработка состояния выбора группы
func handleGroupState(bot *tgbotapi.BotAPI, chatID int64, text string, user *database.User) {
	user.Group = text
	if text == "⬅️Назад" {
		sendKeyboardMessage(bot, chatID, "Выбери курс:", createCourseKeyboardUp, user, "waiting_for_course")
		return
	}

	if user.Format == "" {
		sendKeyboardMessage(bot, chatID, "Выбери формат вывода", createPrintKeyboard, user, "select_format")
	} else {
		schedule := parser.Tab(user.Group, user.Format, user.EducationLevel)
		sendKeyboardMessage(bot, chatID, schedule, createBackKeyboard, user, "waiting_for_return")
	}
}

// Обработка состояния выбора формата вывода
func handleFormatState(bot *tgbotapi.BotAPI, chatID int64, text string, user *database.User, userDAO *database.UserDAO, users map[int64]*database.User) {
	if text == "⬅️Назад" {
		sendKeyboardMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel), user, "select_group")
		return
	}

	user.Format = text
	schedule := parser.Tab(user.Group, user.Format, user.EducationLevel)
	sendKeyboardMessage(bot, chatID, schedule, createBackKeyboard, user, "waiting_for_return")
	saveUser(user, userDAO, chatID)
	delete(users, chatID)
}

// Обработка состояния ожидания возврата
func handleReturnState(bot *tgbotapi.BotAPI, chatID int64, text string, user *database.User, userDAO *database.UserDAO) {
	saveUserStateOnly(user, userDAO, chatID)
	switch text {
	case "📚 Курс":
		sendKeyboardMessage(bot, chatID, "Выберите курс:", createCourseKeyboardUp, user, "waiting_for_course")
	case "🏫 Группа":
		sendKeyboardMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel), user, "select_group")
	case "📋 Вывод":
		sendKeyboardMessage(bot, chatID, "Выберите формат вывода:", createPrintKeyboard, user, "select_format")
	case "🎓Образование":
		sendKeyboardMessage(bot, chatID, "Выберите форму обучения:", createEducationKeyboard, user, "waiting_for_education")
	case "〽️Начало":
		sendKeyboardMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard, user, "hello")
	default:
		sendKeyboardMessage(bot, chatID, "Нажми кнопку на клавиатуре", createBackKeyboard, user, "")
	}
}

// Обработка состояния поиска преподавателя
func handleTeacherState(bot *tgbotapi.BotAPI, chatID int64, text string, user *database.User) {
	teacher := parser.FindTeacher(text)
	if teacher == nil || teacher.Picture == "" {
		sendKeyboardMessage(bot, chatID, "Преподаватель "+text+" не найден", createHelloKeyboard, user, "hello")
		return
	}

	lesson, _ := parser.FindCurrentLessons(teacher.FileName)
	info := teacher.Surname + teacher.Name + teacher.Text + lesson
	handleMediaGroupInfo(bot, chatID, info, teacher.Picture, "")
	sendKeyboardMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard, user, "hello")
}

// Обработка состояния информации о корпусе
func handleCorpusState(bot *tgbotapi.BotAPI, chatID int64, text string, user *database.User) {
	corpusInfo := map[string]struct {
		description string
		mapImg      string
		corpusImg   string
	}{
		"1": {"Эот наш главный корпус\nОриентиром тут послужит огромная парковка(курилка)\nАдрес: Зиповская, д.5", "./resources/images/corpus/1_map.jpg", "./resources/images/corpus/1_corpus.jpg"},
		"2": {"Второй корпус или (Сбербанк)\nНаходиться на пересечении зиповской и московской. А наш ориентиир это компьютерный клуб Fenix\nАдрес: Зиповская 8", "./resources/images/corpus/2_map.jpg", "./resources/images/corpus/2_corpus.jpg"},
		"3": {"Третий корпус\nНаходиться за трамвайными путями по правой стороне\nАдрес: Зиповская 12", "./resources/images/corpus/3_map.jpg", "./resources/images/corpus/3_corpus.jpg"},
		"4": {"Четвертый корпус\nНаш ориентиир это SubWay а точнее слева от него\nАдрес: Зиповская 5/2", "./resources/images/corpus/4_map.jpg", "./resources/images/corpus/4_corpus.jpg"},
		"5": {"Пятый корпус (школа)\nНаходиться в пристройке бывшей гимназии имсит. Но только не путай наш вход с торца а не главный\nАдрес: Зиповская 8", "./resources/images/corpus/5_map.jpg", "./resources/images/corpus/5_corpus.jpg"},
		"6": {"Шестой корпус(Дизайнеры)\nнаходиьтся за углом от мфц напротив входа в главный корпус\nАдрес: Зиповская 5к1", "./resources/images/corpus/6_map.jpg", "./resources/images/corpus/6_corpus.jpg"},
		"7": {"Седьмой корпус\nНаходиьтся сразу за главным", "./resources/images/corpus/7_map.jpg", "./resources/images/corpus/7_corpus.jpg"},
		"8": {"Восьмой корпус\nнаходитьтся слева от корпуса пять в здании гимназии\nАдрес: Зиповская 3", "./resources/images/corpus/8_map.jpg", "./resources/images/corpus/8_corpus.jpg"},
	}

	if info, exists := corpusInfo[text]; exists {
		handleMediaGroupInfo(bot, chatID, info.description, info.mapImg, info.corpusImg)
	} else if text == "〽️Начало" {
		sendKeyboardMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard, user, "hello")
	} else {
		sendKeyboardMessage(bot, chatID, "Нажми цифру на клавиатуре", createCorpusNum, user, "")
	}
}

// Сохранение пользователя (все данные)
func saveUser(user *database.User, userDAO *database.UserDAO, chatID int64) {
	if err := userDAO.SaveUser(user); err != nil {
		log.Printf("Ошибка сохранения пользователя %d: %s", chatID, err)
	}
}

// Сохранение только состояния пользователя
func saveUserStateOnly(user *database.User, userDAO *database.UserDAO, chatID int64) {
	// Создаем временного пользователя с только ID и State для обновления в БД
	tempUser := &database.User{
		ID:    chatID,
		State: user.State,
	}
	if err := userDAO.SaveUser(tempUser); err != nil {
		log.Printf("Ошибка сохранения состояния пользователя %d: %s", chatID, err)
	}
}

// Существующие функции из вашего кода
func sendKeyboardMessage(bot *tgbotapi.BotAPI, chatID int64, text string, keyboardFunc func() tgbotapi.ReplyKeyboardMarkup, user *database.User, newState string) {
	msg := tgbotapi.NewMessage(chatID, text)
	if keyboardFunc != nil {
		msg.ReplyMarkup = keyboardFunc()
	}

	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
		return
	}

	if newState != "" {
		user.State = newState
	}
}

var groupKeyboards = map[string]map[string]func() tgbotapi.ReplyKeyboardMarkup{
	"Высшее": {
		"🤓 1 курс": func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(1) },
		"😎 2 курс": func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(2) },
		"🧐 3 курс": func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(3) },
		"🎓 4 курс": func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(4) },
		"🫠 5 курс": func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(5) },
	},
	"Среднее": {
		"🤓 1 курс": func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(7) },
		"😎 2 курс": func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(8) },
		"🧐 3 курс": func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(9) },
		"🎓 4 курс": func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(10) },
	},
}

func getGroupKeyboard(course, education string) func() tgbotapi.ReplyKeyboardMarkup {
	if eduMap, ok := groupKeyboards[education]; ok {
		if kbFunc, ok := eduMap[course]; ok {
			return kbFunc
		}
	}
	return nil
}

func handleMediaGroupInfo(api *tgbotapi.BotAPI, chatID int64, text, filePath1, filePath2 string) {
	media1 := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(filePath1))
	media1.Caption = text

	mediaGroup := []interface{}{media1}

	if filePath2 != "" {
		media2 := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(filePath2))
		mediaGroup = append(mediaGroup, media2)
	}

	_, err := api.Send(tgbotapi.NewMediaGroup(chatID, mediaGroup))
	if err != nil {
		log.Printf("Ошибка при отправке медиагруппы: %v", err)
	}
}
