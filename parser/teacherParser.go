package parser

import (
	"strings"
)

const (
	path = "./resources/images/teacher/"
)

type Teacher struct {
	Surname  string `json:"surname"`
	FileName string `json:"fileName"`
	Name     string `json:"name"`
	Text     string `json:"text"`
	Picture  string `json:"picture"`
}

func GetTeacher() []Teacher {
	return []Teacher{
		{"Алферова", "Алферова В.В.", " Виктория Владимировна", "\n", path + "Алферова Виктория Владимировна.jpg"},                                                                 // AgACAgIAAxkBAAIOHme8ytmQ1f6dRjoigmVnB-azy2q-AAL56TEbC8foSdNXDDhCe2yLAQADAgADeAADNgQ
		{"Горшунова", "Горшунова И.В.", " Ирина Викторовна", "\n", path + "Горшунова Ирина Викторовна.jpg"},                                                                        //AgACAgIAAxkBAAIOHGe8ytWh3tvkFVcQTDj6etlRQIjFAAL46TEbC8foSdmGXPnGCO9rAQADAgADbQADNgQ
		{"Грицык", "Грицык Е.А.", " Екатерина Анатольевна", "\n", path + "грицык.jpg"},                                                                                             //AgACAgIAAxkBAAIOXGfBnKJ_8CP1tcQMuwtNS5UvuyYbAAQyG3x4CEpRY5rjNSC1RgEAAwIAA3gAAzYE
		{"Ермишина", "Ермишина Е.Б.", " Елена Борисовна", "\n", path + "Ермишина Елена Борисовна.jpg"},                                                                             //AgACAgIAAxkBAAIOGme8ytG6vNxgm0H6_BKb1C4D1_jjAAL36TEbC8foSUx-MBsodlbNAQADAgADbQADNgQ
		{"Исикова", "Исикова Н.П.", " Наталья Павловна", "\n", path + "Исикова Наталья Павловна.jpg"},                                                                              //AgACAgIAAxkBAAIMNmezmT7YAAEqdtrdnsyxthVXcIpXkQACG-oxGxJVmUmH2790Gsk02QEAAwIAA3gAAzYE
		{"Капустин", "Капустин С.А.", " Сергей Алимович", "\n", path + "Капустин Сергей Алимович.jpg"},                                                                             //AgACAgIAAxkBAAIMOGezmUNy7-P1qPUbzX-zvJZpCDlXAAIc6jEbElWZSUrTgjBDMLcyAQADAgADeAADNgQ
		{"Клинов", "Клинов А.С.", " Анатолий Сергеевич", "\n", path + "Клинов Анатолий Сергеевич.jpg"},                                                                             //AgACAgIAAxkBAAIOMGe8yuipJqfXy6HHyp1usA-xrN_zAAID6jEbC8foSYvLDEW8SUTLAQADAgADbQADNgQ
		{"Корольков", "Корольков Р.А.", " Роман Александрович", "\n", path + "Корольков Роман Александрович.jpg"},                                                                  //AgACAgIAAxkBAAIMOmezmUZ-z617n98jMe5eWb96E9cRAAId6jEbElWZSXFSc6GY1FlbAQADAgADeAADNgQ
		{"Леонова", "Леонова И.В.", " Ирина Васильевна", "\n", path + "Леонова Ирина Васильевна.jpg"},                                                                              //AgACAgIAAxkBAAIMPGezmUksjTkGpbk3P7wCkKjsqiq1AAJv7TEblnyYSUK42azDmqI_AQADAgADbQADNgQ
		{"Леонтьев", "Леонтьев Н.А.", " Николай Александрович", "", path + "Леонтьев Николай Александрович.jpg"},                                                                   //AgACAgIAAxkBAAIMPmezmU3Ue3s6Prhr5NOU8XP7ua_xAAIe6jEbElWZSbr6QpAnQXhNAQADAgADbQADNgQ
		{"Лисин", "Лисин Д.А. ", " Денис Александрович", "\n", path + "Лисин Денис Александрович.jpg"},                                                                             //AgACAgIAAxkBAAIMQGezmU9JRyFR5q5la9Q0y1BJ4YbaAAIf6jEbElWZSd6hNsrCiQABBAEAAwIAA3gAAzYE
		{"Лихачева", "Лихачева О.Н.", " Ольга Николаевна", "\n", path + "Лихачева Ольга Николаевна.jpg"},                                                                           //AgACAgIAAxkBAAIOMme8yukxrkZ2dWSR19I-mRsH1x_xAAIE6jEbC8foSSAHwNPqwSjGAQADAgADbQADNgQ
		{"Мадатова", "Мадатова О.В.", " Оксана Владимировна", "\n", path + "Мадатова Оксана Владимировна.jpg"},                                                                     //AgACAgIAAxkBAAIOLme8yuZRbgJjke0QrGX-fBi3ilSGAAIB6jEbC8foSVwsJ_64JO_vAQADAgADbQADNgQ
		{"Мироненко", "Мироненко Д.С.", " Дмитрий Сергеевич", "", path + "мироненко.jpg"},                                                                                          //AgACAgIAAxkBAAIOXmfBnKdPJ-v6JT4F74i-3UrVEctiAAIBAAEyG3x4CEr7fmiYatCHrgEAAwIAA3kAAzYE
		{"Нигматов", "Нигматов В.А.", " Вадим Азамович", "\n«Сомнение — мой верный спутник, оно помогает мне не разочаровываться в людях».", path + "Нигматов Вадим Азамович.jpg"}, //AgACAgIAAxkBAAIMRGezmVeY978Wpwwy1iCoxQFtBwJsAAIh6jEbElWZSRvymxV2Fx-HAQADAgADeAADNgQ
		{"Нестерова", "Нестерова Н.С.", " Нонна Семеновна", "\n", path + "Нестерова Нонна Семеновна.jpg"},                                                                          //AgACAgIAAxkBAAIMQmezmVKygv4Y_SK-bvl3U6cK-T2WAAIg6jEbElWZSWI2mjpgjbaIAQADAgADbQADNgQ
		{"Обухова", "Обухова Ю.А.", " Юлия Александровна", "\n", path + "Обухова Юлия Александровна.jpg"},                                                                          //AgACAgIAAxkBAAIOYWfBnT2PSjbpAyGPzhvkYh3wZhTaAAL37DEbfHgQSpJq7W3984CKAQADAgADeAADNgQ
		{"Пальников", "Пальников А.В.", " Александр Валерьевич", "\n", path + "Пальников Александр Валерьевич.jpeg"},                                                               //AgACAgIAAxkBAAIMRmezmVpLm_wlh25TL6zTZBiDdouMAAIi6jEbElWZSc_EMTQqesWjAQADAgADeAADNgQ
		{"Пархоменко", "Пархоменко А.А.", " Алина Андреевна", "\n", path + "Пархоменко Алина Андреевна.jpg"},                                                                       //AgACAgIAAxkBAAIOLGe8yuS3-5eCJdhF5P6Vox-u_BeDAAPqMRsLx-hJ8S4zzs8j0qwBAAMCAANtAAM2BA
		{"Петров", "Петров И.Ф.", " Игорь Федорович", "\n", path + "Петров Игорь Федорович.png"},                                                                                   //AgACAgIAAxkBAAIOKme8yuKei1zpaU8WWpDvRKahSq_OAAL_6TEbC8foSRe92_iLrvkvAQADAgADbQADNgQ
		{"Петрова", "Петрова С.И.", " Софья Игоревна", "\n", path + "Петрова Софья Игоревна.png"},                                                                                  //AgACAgIAAxkBAAIOKGe8yuHDFwa6q32z1AnnVTGyLAI0AAL-6TEbC8foSW6n4KllkgABYgEAAwIAA20AAzYE
		{"Петросян", "Петросян А.М.", " Арутюн Микаэлович", "\nХарактер Скверный.\nНе женат.\n", path + "Petr.jpg"},                                                                // AgACAgIAAxkBAAIMSGezmV0vPmKQlgEFZO6ozLKx297NAAIj6jEbElWZSXPVRgYbsQNRAQADAgADeAADNgQ
		{"Рассоха", "Рассоха Е.В.", " Евгений Викторович", "\n", path + "Рассоха Евгений Викторович.png"},                                                                          //AgACAgIAAxkBAAIOJme8yt8yQxuoKj5dEXTIbvokwgtZAAL96TEbC8foSTx1D4hM8L8wAQADAgADeAADNgQ
		{"Саакян", "Саакян Р.Р.", " Рустам Рафикович", "\n", path + "Саакян Рустам Рафикович.jpg"},                                                                                 //AgACAgIAAxkBAAIMSmezmWBaaE7RNDpmwAqtKhFiPN9VAAIk6jEbElWZSRBfFV1TAljKAQADAgADeAADNgQ
		{"Сапунов", "Сапунов А.В.", " Андрей Владимирович", "\n", path + "Сапунов Андрей Владимирович.png"},                                                                        //AgACAgIAAxkBAAIOJGe8yt7WB9PQ0sr5u1GCaPb9SufkAAL86TEbC8foSWwOHpsMrf6cAQADAgADbQADNgQ
		{"Сорокина", "Сорокина В.В.", " Виктория Владимировна", "\n", path + "Сорокина Виктория Владимировна.jpg"},                                                                 // AgACAgIAAxkBAAIMTGezmWRAxRePB2mwIkmcEZ1wPY0sAAIl6jEbElWZSQr0LfHFrR0SAQADAgADeAADNgQ
		{"Субачев", "Субачев С.Ю. ", " Сергей Юрьевич", "\n", path + "Субачев Сергей Юрьевич.png"},                                                                                 //AgACAgIAAxkBAAIOIme8ytz0K9oLZp__VTJy-PLhfq0UAAL76TEbC8foST4rSYi18YZCAQADAgADbQADNgQ
		{"Тиньгаев", "Тиньгаев Е.Г.", " Евгений Геннадьевич", "\n", path + "Тиньгаев Евгений Геннадьевич.png"},                                                                     //AgACAgIAAxkBAAIOIGe8ytuW-QABLhGbI6nD55peX07Y3QAC-ukxGwvH6Em5q3kqB2ZcIQEAAwIAA20AAzYE
		{"Цебренко", "Цебренко К.Н.", " Константин Николаевич", "\n", path + "Цебренко Константин Николаевич.jpg"},                                                                 //AgACAgIAAxkBAAIMTmezmWhi_-hUIPuPqtpbpXvnY_90AAIm6jEbElWZSRe1fE_tcLobAQADAgADbQADNgQ
		{"Шепель", "Шепель Э.В.", " Элона Вячеславна", "\n", path + "Шепель Элона Вячеславна.png"},                                                                                 //AgACAgIAAxkBAAIMUGezmWuDiv8jn3RDLS6DBs-YDBQlAAIn6jEbElWZSQVKZPG2xV7jAQADAgADbQADNgQ
		{"Шпехт", "Шпехт И.А.", " Ирина Александровна", "\n", path + "Шпехт Ирина Александровна.jpg"},                                                                              //AgACAgIAAxkBAAIMUmezmW6A4N4dH-AFyJOvvdaSMFBVAAIo6jEbElWZSS1D7zxsY7koAQADAgADeAADNgQ
		{"Мулько", "Мулько Г.А.", " Герман Андреевич", "\n", path + "Мулько Герман Андреевич.jpg"},
		{"Кочура", "Кочура А.Н.", " Алексей Николаевич", "\n", path + "sil.jpg"},
	}
}

func FindTeacher(surname string) *Teacher {
	for _, teacher := range GetTeacher() {
		if strings.EqualFold(teacher.Surname, surname) { // EqualFold учитывает регистр автоматически
			return &teacher
		}
	}
	return nil
}
