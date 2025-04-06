package parser

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"regexp"
	"sort"
	"strings"
	"sync"
)

type Lesson struct {
	StartTime string `json:"start_time"`
	Text      string `json:"text"`
}

var (
	fileCache = make(map[int]*excelize.File)
	cacheLock sync.RWMutex
)

// Загружаем файл в кэш
func getCachedFile(course int, fileName string) (*excelize.File, error) {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	if f, exists := fileCache[course]; exists {
		return f, nil
	}

	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия файла: %v", err)
	}

	fileCache[course] = f
	return f, nil
}

func FindCurrentLessons(teacherName string) (string, error) {
	fileName := map[int]string{
		1: "resources/1_course.xlsx",
		2: "resources/2_course.xlsx",
		3: "resources/3_course.xlsx",
		4: "resources/4_course.xlsx",
		5: "resources/СПО_1_course.xlsx",
		6: "resources/СПО_2_course.xlsx",
		7: "resources/СПО_3_course.xlsx",
		8: "resources/СПО_4_course.xlsx",
		9: "resources/5_course.xlsx",
	}

	_, day, week, _ := NowTime()

	if day == 0 {
		day = 1
	}
	weekDays := [...]string{"", "понедельник", "вторник", "среда", "четверг", "пятница", "суббота"}
	if int(day) >= len(weekDays) {
		return "Неверный день недели", nil
	}
	dayStr := weekDays[day]

	lessons := make([]Lesson, 0, 10)
	lessonSet := make(map[string]struct{})

	re := regexp.MustCompile(`\b` + regexp.QuoteMeta(teacherName) + `\b\s*`)

	for i := 1; i <= 9; i++ {
		f, err := getCachedFile(i, fileName[i])
		if err != nil {
			return "", err
		}

		rows, err := f.GetRows(week)
		if err != nil {
			return "", fmt.Errorf("ошибка чтения строк: %v", err)
		}

		var previousStartTime string

		for _, row := range rows {
			if len(row) < 3 {
				continue
			}
			if !strings.Contains(row[0], dayStr) {
				continue
			}

			startTime := row[1]
			if startTime == "" {
				startTime = previousStartTime
			} else {
				previousStartTime = startTime
			}

			for _, str := range row {
				if strings.Contains(str, teacherName) {
					cleanedLesson := strings.TrimSpace(re.ReplaceAllString(str, ""))
					lessonKey := fmt.Sprintf("%s|%s", startTime, cleanedLesson)
					if _, exists := lessonSet[lessonKey]; !exists {
						lessonSet[lessonKey] = struct{}{}
						lessons = append(lessons, Lesson{
							StartTime: startTime,
							Text:      fmt.Sprintf("Время: %s\n%s", startTime, cleanedLesson),
						})
					}
				}
			}
		}
	}

	if len(lessons) == 0 {
		return "У этого преподавателя сегодня нет пар", nil
	}

	sort.Slice(lessons, func(i, j int) bool {
		return lessons[i].StartTime < lessons[j].StartTime
	})

	var result strings.Builder
	result.WriteString("\nПары на сегодня:\n\n")
	for _, lesson := range lessons {
		result.WriteString(lesson.Text + "\n\n")
	}

	return result.String(), nil
}
