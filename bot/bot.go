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
		log.Fatal("Ошибка загрузки токена")
	}

	botToken := os.Getenv("BOT_TOKEN")
	dbPath := "./resources/data/users.sqlite"

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)
	userDAO := database.NewUserDAO(dbPath)
	defer userDAO.Close()

	users := make(map[int64]*database.User) // Локальный кеш пользователей

	for update := range updates {
		if update.Message == nil {
			continue
		}

		chatID := update.Message.Chat.ID
		user, exists := users[chatID]

		if !exists {
			dbUser, err := userDAO.GetUser(chatID)
			if err == nil && dbUser != nil {
				user = dbUser
			} else {
				user = &database.User{ID: chatID, State: "hello"}
				err := userDAO.SaveUser(user)
				if err != nil {
					log.Printf("Ошибка при сохранении нового пользователя: %s", err)
				}
			}
			users[chatID] = user
		}

		log.Printf("Получено сообщение от %s: %s", update.Message.From.UserName, update.Message.Text)

		switch user.State {

		// выбор расписания/корпуса/препода
		case "hello":
			users[chatID] = user
			if update.Message.Text == "🗓Расписание🗓" {
				if user.EducationLevel == "" {
					sendKeyboardMessage(bot, chatID, "Выбери форму обучения", createEducationKeyboard, user, "waiting_for_education")
				} else {
					sendKeyboardMessage(bot, chatID, "Выбери курс:", createCourseKeyboardUp, user, "waiting_for_course")
				}
			} else if update.Message.Text == "👱‍♂️Найти препода👱" {
				sendKeyboardMessage(bot, chatID, "Напиши фамилию преподоваеля в формате (Иванов)", nil, user, "teacher")
			} else if update.Message.Text == "🏢Найти корпус🏫" {
				sendKeyboardMessage(bot, chatID, "Напиши номер корпуса", createCorpusNum, user, "corpus_info")
			} else if update.Message.Text == "/start" {
				sendKeyboardMessage(bot, chatID, "Привет, Я бот для помощи тебе в твоем обучении!", createHelloKeyboard, user, "")
			} else if update.Message.Text == "/send" {
				if update.Message.From.ID != adminID {
					sendKeyboardMessage(bot, chatID, "У вас нет прав для выполнения этой команды.", createHelloKeyboard, user, "")
					break
				}
				awaitingAdminMessage = true
				adminChatID = chatID
				sendKeyboardMessage(bot, chatID, "Пожалуйста, введите сообщение, которое хотите отправить всем пользователям.", createHelloKeyboard, user, "")
			} else if awaitingAdminMessage && update.Message.From.ID == adminID && chatID == adminChatID {
				messageText := update.Message.Text

				// Проверка на команду /cancel внутри блока отправки
				if messageText == "/cancel" {
					awaitingAdminMessage = false
					sendKeyboardMessage(bot, chatID, "Рассылка отменена.", createHelloKeyboard, user, "")
				} else {
					usersList, err := userDAO.GetAllUsers()
					if err != nil {
						log.Printf("Ошибка получения списка пользователей: %s", err)
						sendKeyboardMessage(bot, chatID, "Ошибка при отправке сообщений.", createHelloKeyboard, user, "")
						awaitingAdminMessage = false
						break
					}

					for _, u := range usersList {
						_, err := bot.Send(tgbotapi.NewMessage(u.ID, messageText))
						if err != nil {
							log.Printf("Ошибка отправки сообщения пользователю %d: %s", u.ID, err)
						}
					}

					sendKeyboardMessage(bot, chatID, "Сообщение успешно отправлено всем пользователям.", createHelloKeyboard, user, "")
					awaitingAdminMessage = false
				}
			} else {
				sendKeyboardMessage(bot, chatID, "Используй клавиатуру", createHelloKeyboard, user, "")
			}
		// выбор формы образования
		case "waiting_for_education":
			user.EducationLevel = update.Message.Text
			switch user.EducationLevel {
			case "Высшее":

				sendKeyboardMessage(bot, chatID, "Выбери курс:", createCourseKeyboardUp, user, "waiting_for_course")
			case "Среднее":
				sendKeyboardMessage(bot, chatID, "Выбери курс:", createCourseKeyboardDown, user, "waiting_for_course")
			case "⬅️Назад":
				sendKeyboardMessage(bot, chatID, "Попробуем снова", createHelloKeyboard, user, "hello")
			default:
				sendKeyboardMessage(bot, chatID, "Используй для этого клавиатуру", createEducationKeyboard, user, "")
			}
			// Сохраняем пользователя в базе данных после изменения образования
			user.UserName = update.Message.From.UserName
			err := userDAO.SaveUser(user)
			if err != nil {
				log.Printf("Ошибка при сохранении пользователя в БД: %s", err)
			}
			users[chatID] = user
		//выбор курса
		case "waiting_for_course":
			user.Course = update.Message.Text
			if update.Message.Text == "⬅️Назад" {
				sendKeyboardMessage(bot, chatID, "Попробуем еще раз", createHelloKeyboard, user, "hello")
			} else if user.Course == "🤓 1 курс" || user.Course == "😎 2 курс" || user.Course == "🧐 3 курс" || user.Course == "🎓 4 курс" || user.Course == "🫠 5 курс" {
				sendKeyboardMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel), user, "select_group")
			} else {
				sendKeyboardMessage(bot, chatID, "Нажми кнопочку на клавиатуре", createCourseKeyboardUp, user, "")
			}
		// выбор группы
		case "select_group":
			user.Group = update.Message.Text

			if update.Message.Text == "⬅️Назад" {
				sendKeyboardMessage(bot, chatID, "Выбери курс:", createCourseKeyboardUp, user, "waiting_for_course")
			} else {
				if user.Format == "" {
					sendKeyboardMessage(bot, chatID, "Выбери формат вывода", createPrintKeyboard, user, "select_format")
				} else {
					schedule := parser.Tab(user.Group, user.Format, user.EducationLevel)
					sendKeyboardMessage(bot, chatID, schedule, createBackKeyboard, user, "waiting_for_return")
				}

			}
		// выбор формата вывода
		case "select_format":
			if update.Message.Text == "⬅️Назад" {
				sendKeyboardMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel), user, "select_group")
			} else {
				user.Format = update.Message.Text
				schedule := parser.Tab(user.Group, user.Format, user.EducationLevel)
				sendKeyboardMessage(bot, chatID, schedule, createBackKeyboard, user, "waiting_for_return")
				user.UserName = update.Message.From.UserName

				err := userDAO.SaveUser(user)
				if err != nil {
					log.Printf("Ошибка при сохранении пользователя в БД: %s", err)
				}
				// Удаляем пользователя из кеша, чтобы не хранить лишние данные
				delete(users, chatID)
			}
		// ожидание выбора возврата
		case "waiting_for_return":
			switch update.Message.Text {
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
		// вывод учителя и его расписания
		case "teacher":
			surname := update.Message.Text
			// получение учителя из списка
			teacher := parser.FindTeacher(surname)

			if teacher == nil || teacher.Picture == "" {
				sendKeyboardMessage(bot, chatID, "Преподаватель "+surname+" не найден", createHelloKeyboard, user, "hello")
				users[chatID] = user
				break
			}
			// получение его пары на данный момент времени
			lesson, _ := parser.FindCurrentLessons(teacher.FileName)

			handleMediaGroupInfo(bot, chatID, teacher.Surname+teacher.Name+teacher.Text+lesson, teacher.Picture, "")
			sendKeyboardMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard, user, "hello")
		// нужны фотки и описание корпусов
		case "corpus_info":
			switch update.Message.Text {
			case "1":
				handleMediaGroupInfo(bot, chatID, "Эот наш главный корпус\nОриентиром тут послужит огромная парковка(курилка)\nАдрес: Зиповская, д.5", "./resources/images/corpus/1_map.jpg", "./resources/images/corpus/1_corpus.jpg") // 1 карта 2 корпус
			case "2":
				handleMediaGroupInfo(bot, chatID, "Второй корпус или (Сбербанк)\nНаходиться на пересечении зиповской и московской. А наш ориентиир это компьютерный клуб Fenix\nАдрес: Зиповская 8", "./resources/images/corpus/2_map.jpg", "./resources/images/corpus/2_corpus.jpg")
			case "3":
				handleMediaGroupInfo(bot, chatID, "Третий корпус\nНаходиться за трамвайными путями по правой стороне\nАдрес: Зиповская 12", "./resources/images/corpus/3_map.jpg", "./resources/images/corpus/3_corpus.jpg")
			case "4":
				handleMediaGroupInfo(bot, chatID, "Четвертый корпус\nНаш ориентиир это SubWay а точнее слева от него\nАдрес: Зиповская 5/2", "./resources/images/corpus/4_map.jpg", "./resources/images/corpus/4_corpus.jpg")
			case "5":
				handleMediaGroupInfo(bot, chatID, "Пятый корпус (школа)\nНаходиться в пристройке бывшей гимназии имсит. Но только не путай наш вход с торца а не главный\nАдрес: Зиповская 8", "./resources/images/corpus/5_map.jpg", "./resources/images/corpus/5_corpus.jpg")
			case "6":
				handleMediaGroupInfo(bot, chatID, "Шестой корпус(Дизайнеры)\nнаходиьтся за углом от мфц напротив входа в главный корпус\nАдрес: Зиповская 5к1", "./resources/images/corpus/6_map.jpg", "./resources/images/corpus/6_corpus.jpg")
			case "7":
				handleMediaGroupInfo(bot, chatID, "Седьмой корпус\nНаходиьтся сразу за главным", "./resources/images/corpus/7_map.jpg", "./resources/images/corpus/7_corpus.jpg")
			case "8":
				handleMediaGroupInfo(bot, chatID, "Восьмой корпус\nнаходитьтся слева от корпуса пять в здании гимназии\nАдрес: Зиповская 3", "./resources/images/corpus/8_map.jpg", "./resources/images/corpus/8_corpus.jpg")
			case "〽️Начало":
				sendKeyboardMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard, user, "hello")
			default:
				sendKeyboardMessage(bot, chatID, "Нажми цифру на клавиатуре", nil, user, "")
			}
		}
	}
}

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
