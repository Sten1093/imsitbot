package bot

import (
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"log"
	"os"
	"tgbot/database"
	"tgbot/parser"
)

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
					user.State = "waiting_for_education"
					sendMessage(bot, chatID, "Выбери форму обучения", createEducationKeyboard)
				} else {
					user.State = "waiting_for_course"
					sendMessage(bot, chatID, "Выбери курс:", createCourseKeyboardUp)
				}
			} else if update.Message.Text == "👱‍♂️Найти препода👱" {
				user.State = "teacher"
				sendMessage(bot, chatID, "Напиши фамилию преподоваеля в формате (Иванов)", nil)
			} else if update.Message.Text == "🏢Найти корпус🏫" {
				user.State = "corpus_info"
				sendMessage(bot, chatID, "Напиши номер корпуса", createCorpusNum)
			} else if update.Message.Text == "/start" {
				sendMessage(bot, chatID, "Привет, Я бот для помощи тебе в твоем обучении!", createHelloKeyboard)
			} else {
				sendMessage(bot, chatID, "Используй клавиатуру", createHelloKeyboard)
			}

		case "waiting_for_education":
			user.EducationLevel = update.Message.Text
			switch user.EducationLevel {
			case "Высшее":
				user.State = "waiting_for_course"
				sendMessage(bot, chatID, "Выбери курс:", createCourseKeyboardUp)
			case "Среднее":
				user.State = "waiting_for_course"
				sendMessage(bot, chatID, "Выбери курс:", createCourseKeyboardDown)
			case "⬅️Назад":
				user.State = "hello"
				sendMessage(bot, chatID, "Попробуем снова", createHelloKeyboard)
			default:
				sendMessage(bot, chatID, "Используй для этого клавиатуру", createEducationKeyboard)
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
				sendMessage(bot, chatID, "Попробуем еще раз", createHelloKeyboard)
				user.State = "hello"
			} else if user.Course == "🤓 1 курс" || user.Course == "😎 2 курс" || user.Course == "🧐 3 курс" || user.Course == "🎓 4 курс" || user.Course == "🫠 5 курс" {
				sendMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel))
				user.State = "select_group"
			} else {
				sendMessage(bot, chatID, "Нажми кнопочку на клавиатуре", createCourseKeyboardUp)
			}
		// выбор группы
		case "select_group":
			user.Group = update.Message.Text

			if update.Message.Text == "⬅️Назад" {
				user.State = "waiting_for_course"
				sendMessage(bot, chatID, "Выбери курс:", createCourseKeyboardUp)
			} else {
				if user.Format == "" {
					sendMessage(bot, chatID, "Выбери формат вывода", createPrintKeyboard)
					user.State = "select_format"
				} else {
					schedule := parser.Tab(user.Group, user.Format, user.EducationLevel)
					sendMessage(bot, chatID, schedule, createBackKeyboard)
					user.State = "waiting_for_return"
				}

			}
		// выбор формата вывода
		case "select_format":
			if update.Message.Text == "⬅️Назад" {
				user.State = "select_group"
				sendMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel))
			} else {
				user.Format = update.Message.Text
				schedule := parser.Tab(user.Group, user.Format, user.EducationLevel)
				sendMessage(bot, chatID, schedule, createBackKeyboard)
				user.State = "waiting_for_return"
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
				user.State = "waiting_for_course"
				sendMessage(bot, chatID, "Выберите курс:", createCourseKeyboardUp)
			case "🏫 Группа":
				user.State = "select_group"
				sendMessage(bot, chatID, "Выберите группу:", getGroupKeyboard(user.Course, user.EducationLevel))
			case "📋 Вывод":
				user.State = "select_format"
				sendMessage(bot, chatID, "Выберите формат вывода:", createPrintKeyboard)
			case "🎓Образование":
				user.State = "waiting_for_education"
				sendMessage(bot, chatID, "Выберите форму обучения:", createEducationKeyboard)
			case "〽️Начало":
				user.State = "hello"
				sendMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard)
			default:
				sendMessage(bot, chatID, "Нажми кнопку на клавиатуре", createBackKeyboard)
			}

		// вывод учителя и его расписания
		case "teacher":
			surname := update.Message.Text
			// получение учителя из списка
			teacher := parser.FindTeacher(surname)

			if teacher == nil || teacher.Picture == "" {
				user.State = "hello"
				users[chatID] = user
				sendMessage(bot, chatID, "Преподователь "+surname+" не найден", createHelloKeyboard)
				break
			}
			// получение его пары на данный момент времени
			lesson, _ := parser.FindCurrentLessons(teacher.FileName)

			handleMediaGroupInfo(bot, chatID, teacher.Surname+teacher.Name+teacher.Text+lesson, teacher.Picture, "")
			sendMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard)
			user.State = "hello"
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
				user.State = "hello"
				sendMessage(bot, chatID, "Чем еще помочь?", createHelloKeyboard)
			default:
				sendMessage(bot, chatID, "Нажми цифру на клавиатуре", nil)
			}
		}
	}
}

func sendMessage(api *tgbotapi.BotAPI, chatID int64, text string, keyboardFunc func() tgbotapi.ReplyKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	if keyboardFunc != nil {
		msg.ReplyMarkup = keyboardFunc()
	}
	_, err := api.Send(msg)
	if err != nil {
		log.Printf("Ошибка при отправке сообщения: %v", err)
	}
}

func getGroupKeyboard(course, education string) func() tgbotapi.ReplyKeyboardMarkup {
	switch education {
	case "Высшее":
		switch course {
		case "🤓 1 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(1)
			}
		case "😎 2 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(2)
			}
		case "🧐 3 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(3)
			}
		case "🎓 4 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(4)
			}
		case "🫠 5 курс":
			return func() tgbotapi.ReplyKeyboardMarkup { return createGroupKeyboardCourseById(5) }
		}

	case "Среднее":
		switch course {
		case "🤓 1 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(7)
			}
		case "😎 2 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(8)
			}
		case "🧐 3 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(9)
			}
		case "🎓 4 курс":
			return func() tgbotapi.ReplyKeyboardMarkup {
				return createGroupKeyboardCourseById(10)
			}
		}

	default:
		return nil
	}
	return nil
}

func handleMediaGroupInfo(api *tgbotapi.BotAPI, chatID int64, text, filePath1, filePath2 string) {
	log.Println("Начало выполнения handleMediaGroupInfo")

	media1 := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(filePath1))
	media1.Caption = text

	var mediaGroup []interface{}
	mediaGroup = append(mediaGroup, media1)

	log.Printf("Добавлено изображение: %s", filePath1)

	// Проверяем, передан ли второй файл
	if filePath2 != "" {
		media2 := tgbotapi.NewInputMediaPhoto(tgbotapi.FilePath(filePath2))
		mediaGroup = append(mediaGroup, media2)
		log.Printf("Добавлено изображение: %s", filePath2)
	} else {
		log.Println("Второй файл не передан")
	}

	mediaGroupMsg := tgbotapi.NewMediaGroup(chatID, mediaGroup)

	log.Println("Отправка медиагруппы в чат...")
	_, err := api.Send(mediaGroupMsg)
	if err != nil {
		log.Printf("Ошибка при отправке медиагруппы: %v", err)
	} else {
		log.Println("Медиагруппа успешно отправлена")
	}
}
