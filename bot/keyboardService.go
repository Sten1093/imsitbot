package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"tgbot/parser"
)

// Константы для названий курсов и кнопок
const (
	FirstCourse  = "🤓 1 курс"
	SecondCourse = "😎 2 курс"
	ThirdCourse  = "🧐 3 курс"
	FourthCourse = "🎓 4 курс"
	FifthCourse  = "🫠 5 курс"
	BackButton   = "⬅️Назад"
	StartButton  = "〽️Начало"
)

// KeyboardConfig определяет конфигурацию клавиатуры
type KeyboardConfig struct {
	Buttons       []string
	ButtonsPerRow int
	AddBackButton bool
}

// createKeyboard создает клавиатуру на основе конфигурации
func createKeyboard(config KeyboardConfig) tgbotapi.ReplyKeyboardMarkup {
	var keyboard [][]tgbotapi.KeyboardButton
	var row []tgbotapi.KeyboardButton

	for _, btn := range config.Buttons {
		row = append(row, tgbotapi.NewKeyboardButton(btn))
		if len(row) == config.ButtonsPerRow {
			keyboard = append(keyboard, row)
			row = []tgbotapi.KeyboardButton{}
		}
	}

	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}

	if config.AddBackButton {
		keyboard = append(keyboard, tgbotapi.NewKeyboardButtonRow(tgbotapi.NewKeyboardButton(BackButton)))
	}

	return tgbotapi.NewReplyKeyboard(keyboard...)
}

// createHelloKeyboard создает начальную клавиатуру
func createHelloKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return createKeyboard(KeyboardConfig{
		Buttons:       []string{"🗓Расписание🗓", "👱‍♂️Найти препода👱", "🏢Найти корпус🏫"},
		ButtonsPerRow: 2,
	})
}

// createBackKeyboard создает клавиатуру возврата
func createBackKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return createKeyboard(KeyboardConfig{
		Buttons:       []string{"📚 Курс", "🏫 Группа", "📋 Вывод", "🎓Образование", StartButton},
		ButtonsPerRow: 3,
	})
}

// createPrintKeyboard создает клавиатуру выбора формата вывода
func createPrintKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return createKeyboard(KeyboardConfig{
		Buttons:       []string{"🌞 День 🙋‍♂️", "📅 Неделя"},
		ButtonsPerRow: 2,
		AddBackButton: true,
	})
}

// createCourseKeyboardUp создает клавиатуру курсов для высшего образования
func createCourseKeyboardUp() tgbotapi.ReplyKeyboardMarkup {
	return createKeyboard(KeyboardConfig{
		Buttons:       []string{FirstCourse, SecondCourse, ThirdCourse, FourthCourse, FifthCourse},
		ButtonsPerRow: 2,
		AddBackButton: true,
	})
}

// createCourseKeyboardDown создает клавиатуру курсов для среднего образования
func createCourseKeyboardDown() tgbotapi.ReplyKeyboardMarkup {
	return createKeyboard(KeyboardConfig{
		Buttons:       []string{FirstCourse, SecondCourse, ThirdCourse, FourthCourse},
		ButtonsPerRow: 2,
		AddBackButton: true,
	})
}

// createCorpusNum создает клавиатуру номеров корпусов
func createCorpusNum() tgbotapi.ReplyKeyboardMarkup {
	return createKeyboard(KeyboardConfig{
		Buttons:       []string{"1", "2", "3", "4", "5", "6", "7", "8", StartButton},
		ButtonsPerRow: 3,
	})
}

// createGroupKeyboardCourseById создает клавиатуру групп по ID курса
func createGroupKeyboardCourseById(id int) tgbotapi.ReplyKeyboardMarkup {
	groups := parser.GetGroups()
	var groupButtons []string

	for _, group := range groups {
		if group.ID == id {
			groupButtons = append(groupButtons, group.TGName)
		}
	}

	return createKeyboard(KeyboardConfig{
		Buttons:       groupButtons,
		ButtonsPerRow: 3,
		AddBackButton: true,
	})
}

// createEducationKeyboard создает клавиатуру выбора образования
func createEducationKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return createKeyboard(KeyboardConfig{
		Buttons:       []string{"Высшее", "Среднее"},
		ButtonsPerRow: 2,
		AddBackButton: true,
	})
}
