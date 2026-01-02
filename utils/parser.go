package utils

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"golang.org/x/net/html"
)

type CourseFromPDF struct {
	CourseName   string
	Teacher      string
	Classroom    string
	WeekDay      int
	StartWeek    int
	EndWeek      int
	StartSection int
	EndSection   int
	WeekType     string
}

func ParseCourseHTML(htmlPath string) ([]CourseFromPDF, error) {
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		return nil, fmt.Errorf("读取HTML失败: %w", err)
	}

	doc, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return nil, fmt.Errorf("解析HTML失败: %w", err)
	}

	var courses []CourseFromPDF
	var table *html.Node

	var findTable func(*html.Node)
	findTable = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "table" {
			table = n
			return
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			if table == nil {
				findTable(c)
			}
		}
	}
	findTable(doc)

	if table == nil {
		text, _ := ExtractTextFromHTMLFile(htmlPath)
		return ParseCourseText(text), nil
	}

	var rows []*html.Node
	var findRows func(*html.Node)
	findRows = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "tr" {
			rows = append(rows, n)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			findRows(c)
		}
	}
	findRows(table)

	weekDays := make(map[int]int)
	for _, row := range rows {
		cells := getTableCells(row)
		for i, cell := range cells {
			text := strings.TrimSpace(getNodeText(cell))
			if strings.Contains(text, "星期日") {
				weekDays[i] = 7
			} else if strings.Contains(text, "星期一") {
				weekDays[i] = 1
			} else if strings.Contains(text, "星期二") {
				weekDays[i] = 2
			} else if strings.Contains(text, "星期三") {
				weekDays[i] = 3
			} else if strings.Contains(text, "星期四") {
				weekDays[i] = 4
			} else if strings.Contains(text, "星期五") {
				weekDays[i] = 5
			} else if strings.Contains(text, "星期六") {
				weekDays[i] = 6
			}
		}
		if len(weekDays) > 0 {
			break
		}
	}

	for _, row := range rows {
		cells := getTableCells(row)
		for colIdx, cell := range cells {
			weekDay, hasWeekDay := weekDays[colIdx]
			if !hasWeekDay {
				continue
			}

			text := strings.TrimSpace(getNodeText(cell))
			if text == "" || len(text) < 5 {
				continue
			}

			if !strings.Contains(text, "*") && !strings.Contains(text, "@") &&
				!strings.Contains(text, "#") && !strings.Contains(text, "&") {
				continue
			}

			course := parseCourseCell(text, weekDay)
			if course.CourseName != "" {
				courses = append(courses, course)
			}
		}
	}

	return courses, nil
}

func getTableCells(row *html.Node) []*html.Node {
	var cells []*html.Node
	for c := row.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
			cells = append(cells, c)
		}
	}
	return cells
}

func getNodeText(n *html.Node) string {
	if n.Type == html.TextNode {
		return n.Data
	}
	var text strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		text.WriteString(getNodeText(c))
	}
	return text.String()
}

func parseCourseCell(text string, weekDay int) CourseFromPDF {
	course := CourseFromPDF{
		WeekDay:      weekDay,
		StartSection: 1,
		EndSection:   2,
		StartWeek:    1,
		EndWeek:      18,
		WeekType:     "全周",
	}

	lines := strings.Split(text, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasSuffix(line, "*") || strings.HasSuffix(line, "@") ||
			strings.HasSuffix(line, "#") || strings.HasSuffix(line, "&") {
			course.CourseName = strings.TrimRight(line, "*@#&")
			course.CourseName = strings.TrimSpace(course.CourseName)
		}

		if sectionMatch := regexp.MustCompile(`\((\d+)-(\d+)节\)`).FindStringSubmatch(line); len(sectionMatch) > 2 {
			course.StartSection, _ = strconv.Atoi(sectionMatch[1])
			course.EndSection, _ = strconv.Atoi(sectionMatch[2])
		}

		if weekMatch := regexp.MustCompile(`(\d+)-(\d+)周`).FindStringSubmatch(line); len(weekMatch) > 2 {
			course.StartWeek, _ = strconv.Atoi(weekMatch[1])
			course.EndWeek, _ = strconv.Atoi(weekMatch[2])
		}

		if strings.Contains(line, "周,") {
			if weekSpecialMatch := regexp.MustCompile(`(\d+)周,(\d+)周`).FindStringSubmatch(line); len(weekSpecialMatch) > 2 {
				course.StartWeek, _ = strconv.Atoi(weekSpecialMatch[1])
				course.EndWeek, _ = strconv.Atoi(weekSpecialMatch[2])
				course.WeekType = "指定周"
			}
		}

		if strings.Contains(line, "单周") {
			course.WeekType = "单周"
		} else if strings.Contains(line, "双周") {
			course.WeekType = "双周"
		}

		if strings.Contains(line, "场地") {
			if classroomMatch := regexp.MustCompile(`场地[: ：\s]*([^\s/\n]+)`).FindStringSubmatch(line); len(classroomMatch) > 1 {
				classroom := strings.TrimSpace(classroomMatch[1])
				if idx := strings.Index(classroom, "/"); idx > 0 {
					classroom = strings.TrimSpace(classroom[:idx])
				}
				if len(classroom) > 100 {
					classroom = classroom[:100]
				}
				if classroom != "" {
					course.Classroom = classroom
				}
			}
		}

		if strings.Contains(line, "教师: ") {
			parts := strings.Split(line, "教师:")
			if len(parts) > 1 {
				teacher := strings.TrimSpace(parts[1])
				if idx := strings.Index(teacher, "教学班"); idx > 0 {
					teacher = strings.TrimSpace(teacher[:idx])
				}
				words := strings.Fields(teacher)
				if len(words) > 0 {
					teacher = words[0]
				}
				if len(teacher) > 50 {
					teacher = teacher[:50]
				}
				if teacher != "" {
					course.Teacher = teacher
				}
			}
		}
	}

	return course
}

func ExtractTextFromHTMLFile(htmlPath string) (string, error) {
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		return "", err
	}

	doc, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return "", err
	}

	var text strings.Builder
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
			return
		}
		if n.Type == html.TextNode {
			data := strings.TrimSpace(n.Data)
			if data != "" {
				text.WriteString(data)
				text.WriteString("\n")
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			extractText(c)
		}
	}
	extractText(doc)

	return text.String(), nil
}

func ParseCourseText(text string) []CourseFromPDF {
	var courses []CourseFromPDF
	lines := strings.Split(text, "\n")

	var currentCourse CourseFromPDF
	var inCourseBlock bool
	currentWeekDay := 0
	currentSectionStart := 1
	currentSectionEnd := 2

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.Contains(line, "学年") || strings.Contains(line, "学号") ||
			strings.Contains(line, "时间段") || strings.Contains(line, "打印时间") ||
			strings.Contains(line, "节次") {
			continue
		}

		if regexp.MustCompile(`^\d+-\d+$`).MatchString(line) {
			parts := strings.Split(line, "-")
			if len(parts) == 2 {
				currentSectionStart, _ = strconv.Atoi(parts[0])
				currentSectionEnd, _ = strconv.Atoi(parts[1])
			}
			continue
		}

		if strings.Contains(line, "星期日") {
			currentWeekDay = 7
		} else if strings.Contains(line, "星期一") {
			currentWeekDay = 1
		} else if strings.Contains(line, "星期二") {
			currentWeekDay = 2
		} else if strings.Contains(line, "星期三") {
			currentWeekDay = 3
		} else if strings.Contains(line, "星期四") {
			currentWeekDay = 4
		} else if strings.Contains(line, "星期五") {
			currentWeekDay = 5
		} else if strings.Contains(line, "星期六") {
			currentWeekDay = 6
		}

		if strings.HasSuffix(line, "*") || strings.HasSuffix(line, "@") ||
			strings.HasSuffix(line, "#") || strings.HasSuffix(line, "&") {

			if inCourseBlock && currentCourse.CourseName != "" && currentCourse.WeekDay > 0 {
				courses = append(courses, currentCourse)
			}

			courseName := strings.TrimRight(line, "*@#&")
			courseName = strings.TrimSpace(courseName)

			currentCourse = CourseFromPDF{
				CourseName:   courseName,
				WeekDay:      currentWeekDay,
				StartSection: currentSectionStart,
				EndSection:   currentSectionEnd,
				StartWeek:    1,
				EndWeek:      18,
				WeekType:     "全周",
			}
			inCourseBlock = true
			continue
		}

		if inCourseBlock {
			if strings.Contains(line, "周") {
				if weekMatch := regexp.MustCompile(`(\d+)-(\d+)周`).FindStringSubmatch(line); len(weekMatch) > 2 {
					currentCourse.StartWeek, _ = strconv.Atoi(weekMatch[1])
					currentCourse.EndWeek, _ = strconv.Atoi(weekMatch[2])
				}

				if strings.Contains(line, "周,") {
					if weekSpecialMatch := regexp.MustCompile(`(\d+)周,(\d+)周`).FindStringSubmatch(line); len(weekSpecialMatch) > 2 {
						currentCourse.StartWeek, _ = strconv.Atoi(weekSpecialMatch[1])
						currentCourse.EndWeek, _ = strconv.Atoi(weekSpecialMatch[2])
						currentCourse.WeekType = "指定周"
					}
				}

				if strings.Contains(line, "单周") {
					currentCourse.WeekType = "单周"
				} else if strings.Contains(line, "双周") {
					currentCourse.WeekType = "双周"
				}
			}

			if strings.Contains(line, "场地") {
				parts := strings.Split(line, "场地")
				if len(parts) > 1 {
					classroom := strings.TrimSpace(parts[1])
					classroom = strings.TrimLeft(classroom, ":：")
					classroom = strings.TrimSpace(classroom)

					if idx := strings.Index(classroom, "/"); idx > 0 {
						classroom = strings.TrimSpace(classroom[:idx])
					}
					if idx := strings.Index(classroom, "教师"); idx > 0 {
						classroom = strings.TrimSpace(classroom[:idx])
					}

					words := strings.Fields(classroom)
					if len(words) > 0 {
						classroom = words[0]
					}

					if len(classroom) > 100 {
						classroom = classroom[:100]
					}

					if classroom != "" && classroom != "/" {
						currentCourse.Classroom = classroom
					}
				}
			}

			if strings.Contains(line, "教师:") {
				parts := strings.Split(line, "教师:")
				if len(parts) > 1 {
					teacher := strings.TrimSpace(parts[1])
					if idx := strings.Index(teacher, "教学班"); idx > 0 {
						teacher = strings.TrimSpace(teacher[:idx])
					}
					words := strings.Fields(teacher)
					if len(words) > 0 {
						teacher = words[0]
					}
					if len(teacher) > 50 {
						teacher = teacher[:50]
					}
					currentCourse.Teacher = teacher
				}
			}

			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if nextLine == "" ||
					strings.HasSuffix(nextLine, "*") ||
					strings.HasSuffix(nextLine, "@") ||
					strings.HasSuffix(nextLine, "#") ||
					strings.HasSuffix(nextLine, "&") ||
					regexp.MustCompile(`^\d+-\d+$`).MatchString(nextLine) {

					if currentCourse.CourseName != "" && currentCourse.WeekDay > 0 {
						courses = append(courses, currentCourse)
						inCourseBlock = false
					}
				}
			}
		}
	}

	if inCourseBlock && currentCourse.CourseName != "" && currentCourse.WeekDay > 0 {
		courses = append(courses, currentCourse)
	}

	return courses
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
