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

// ParseCourseHTML 直接解析HTML表格
// ParseCourseHTML 直接解析HTML表格
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

	// 查找课程表格
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
		fmt.Println("⚠️ [DEBUG] 未找到 <table> 标签，使用增强文本解析")
		// 使用增强的文本解析
		text, _ := ExtractTextFromHTMLFile(htmlPath)
		return ParseCourseText(text), nil
	}

	// 解析表格
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

	fmt.Printf("✅ [DEBUG] 找到 %d 行\n", len(rows))

	// 提取星期标题行
	weekDays := make(map[int]int) // 列索引 -> 星期几
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
			fmt.Printf("✅ [DEBUG] 找到星期映射: %v\n", weekDays)
			break
		}
	}

	// 解析课程数据行
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

			// 🔍 调试：检测是否包含课程标记
			if strings.Contains(text, "*") || strings.Contains(text, "@") ||
				strings.Contains(text, "#") || strings.Contains(text, "&") {
				fmt.Printf("\n🎯 [DEBUG] 找到课程单元格（列%d，星期%d）\n", colIdx, weekDay)
				previewLen := Min(200, len(text))
				fmt.Printf("📝 [DEBUG] 文本前%d字符: %s\n", previewLen, text[:previewLen])
			}

			// 检测课程标记（改用 Contains 而不是 HasSuffix）
			if !strings.Contains(text, "*") && !strings.Contains(text, "@") &&
				!strings.Contains(text, "#") && !strings.Contains(text, "&") {
				continue
			}

			// 解析课程
			fmt.Printf("🔄 [DEBUG] 开始解析.. .\n")
			course := parseCourseCell(text, weekDay)
			fmt.Printf("✅ [DEBUG] 解析结果: 课程=%s | 教师=%s | 教室=%s | 节次=%d-%d\n",
				course.CourseName, course.Teacher, course.Classroom,
				course.StartSection, course.EndSection)

			if course.CourseName != "" {
				courses = append(courses, course)
			}
		}
	}

	return courses, nil
}

// / 获取表格单元格
func getTableCells(row *html.Node) []*html.Node {
	var cells []*html.Node
	for c := row.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && (c.Data == "td" || c.Data == "th") {
			cells = append(cells, c)
		}
	}
	return cells
}

// 获取节点文本
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

// 解析单个课程单元格
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

		// 课程名
		if strings.HasSuffix(line, "*") || strings.HasSuffix(line, "@") ||
			strings.HasSuffix(line, "#") || strings.HasSuffix(line, "&") {
			course.CourseName = strings.TrimRight(line, "*@#&")
			course.CourseName = strings.TrimSpace(course.CourseName)
			fmt.Printf("    📚 [DEBUG] 提取课程名: %s\n", course.CourseName)
		}

		// 节次：(3-4节)
		if sectionMatch := regexp.MustCompile(`\((\d+)-(\d+)节\)`).FindStringSubmatch(line); len(sectionMatch) > 2 {
			course.StartSection, _ = strconv.Atoi(sectionMatch[1])
			course.EndSection, _ = strconv.Atoi(sectionMatch[2])
			fmt.Printf("    ⏰ [DEBUG] 提取节次: %d-%d\n", course.StartSection, course.EndSection)
		}

		// 周次：1-16周
		if weekMatch := regexp.MustCompile(`(\d+)-(\d+)周`).FindStringSubmatch(line); len(weekMatch) > 2 {
			course.StartWeek, _ = strconv.Atoi(weekMatch[1])
			course.EndWeek, _ = strconv.Atoi(weekMatch[2])
			fmt.Printf("    📅 [DEBUG] 提取周次: %d-%d\n", course.StartWeek, course.EndWeek)
		}

		// 特殊周次：8周,14周
		if strings.Contains(line, "周,") {
			if weekSpecialMatch := regexp.MustCompile(`(\d+)周,(\d+)周`).FindStringSubmatch(line); len(weekSpecialMatch) > 2 {
				course.StartWeek, _ = strconv.Atoi(weekSpecialMatch[1])
				course.EndWeek, _ = strconv.Atoi(weekSpecialMatch[2])
				course.WeekType = "指定周"
			}
		}

		// 单双周
		if strings.Contains(line, "单周") {
			course.WeekType = "单周"
		} else if strings.Contains(line, "双周") {
			course.WeekType = "双周"
		}

		// 教室：场地: 文渊607
		if strings.Contains(line, "场地: ") {
			parts := strings.Split(line, "场地:")
			if len(parts) > 1 {
				classroom := strings.TrimSpace(parts[1])
				// 去掉 "/" 后的内容
				if idx := strings.Index(classroom, "/"); idx > 0 {
					classroom = strings.TrimSpace(classroom[:idx])
				}
				// 限制长度
				if len(classroom) > 100 {
					classroom = classroom[:100]
				}
				if classroom != "" {
					course.Classroom = classroom
					fmt.Printf("    🏫 [DEBUG] 提取教室: %s\n", classroom)
				}
			}
		}

		// 教师：教师:王显珉
		if strings.Contains(line, "教师: ") {
			fmt.Printf("    🔍 [DEBUG] 原始教师行: %s\n", line)

			parts := strings.Split(line, "教师:")
			if len(parts) > 1 {
				teacher := strings.TrimSpace(parts[1])
				fmt.Printf("    🔍 [DEBUG] 分割后:  %s\n", teacher)

				// 去掉"教学班"后的所有内容
				if idx := strings.Index(teacher, "教学班"); idx > 0 {
					teacher = strings.TrimSpace(teacher[:idx])
					fmt.Printf("    🔍 [DEBUG] 去除教学班后: %s\n", teacher)
				}

				// 只取第一个词（教师名）
				words := strings.Fields(teacher)
				if len(words) > 0 {
					teacher = words[0]
					fmt.Printf("    🔍 [DEBUG] 取第一个词: %s\n", teacher)
				}

				// 限制长度
				if len(teacher) > 50 {
					teacher = teacher[:50]
				}

				if teacher != "" {
					course.Teacher = teacher
					fmt.Printf("    👨‍🏫 [DEBUG] 最终教师名: %s\n", teacher)
				}
			}
		}
	}

	return course
}

// ParseCoursePDF 解析 PDF（暂不可用，返回错误提示）
func ParseCoursePDF(pdfPath string) ([]CourseFromPDF, error) {
	return nil, fmt.Errorf("PDF直接解析功能暂不可用，请使用HTML上传或文本粘贴方式")
}

// ExtractTextFromHTMLFile 从HTML文件提取纯文本
func ExtractTextFromHTMLFile(htmlPath string) (string, error) {
	// 读取HTML
	htmlBytes, err := os.ReadFile(htmlPath)
	if err != nil {
		return "", fmt.Errorf("读取HTML文件失败: %w", err)
	}

	// 解析HTML
	doc, err := html.Parse(strings.NewReader(string(htmlBytes)))
	if err != nil {
		return "", fmt.Errorf("解析HTML失败:  %w", err)
	}

	// 提取文本
	var text strings.Builder
	var extractText func(*html.Node)
	extractText = func(n *html.Node) {
		// 跳过 script 和 style 标签
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

// ParseCourseText 解析课程表文本
// ParseCourseText 解析课程表文本
func ParseCourseText(text string) []CourseFromPDF {
	var courses []CourseFromPDF

	lines := strings.Split(text, "\n")

	var currentCourse CourseFromPDF
	var inCourseBlock bool
	currentWeekDay := 0
	currentSectionStart := 1 // 当前节次起始
	currentSectionEnd := 2   // 当前节次结束

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 跳过表头和无关信息
		if strings.Contains(line, "学年") || strings.Contains(line, "学号") ||
			strings.Contains(line, "时间段") || strings.Contains(line, "打印时间") ||
			strings.Contains(line, "节次") {
			continue
		}

		// 🆕 检测节次行（纯数字格式：1-2, 3-4, 5-6 等）
		if regexp.MustCompile(`^\d+-\d+$`).MatchString(line) {
			parts := strings.Split(line, "-")
			if len(parts) == 2 {
				currentSectionStart, _ = strconv.Atoi(parts[0])
				currentSectionEnd, _ = strconv.Atoi(parts[1])
				fmt.Printf("🔍 [DEBUG] 检测到节次: %d-%d\n", currentSectionStart, currentSectionEnd)
			}
			continue
		}

		// 检测星期
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

		// 检测课程名称（以 * @ # & 结尾）
		if strings.HasSuffix(line, "*") || strings.HasSuffix(line, "@") ||
			strings.HasSuffix(line, "#") || strings.HasSuffix(line, "&") {

			// 保存之前的课程
			if inCourseBlock && currentCourse.CourseName != "" && currentCourse.WeekDay > 0 {
				courses = append(courses, currentCourse)
			}

			// 开始新课程
			courseName := strings.TrimRight(line, "*@#&")
			courseName = strings.TrimSpace(courseName)

			currentCourse = CourseFromPDF{
				CourseName:   courseName,
				WeekDay:      currentWeekDay,
				StartSection: currentSectionStart, // 使用当前检测到的节次
				EndSection:   currentSectionEnd,   // 使用当前检测到的节次
				StartWeek:    1,
				EndWeek:      18,
				WeekType:     "全周",
			}
			fmt.Printf("📚 [DEBUG] 创建课程: %s, 星期%d, 节次: %d-%d\n",
				courseName, currentWeekDay, currentSectionStart, currentSectionEnd)
			inCourseBlock = true
			continue
		}

		// 如果在课程块中，解析详细信息
		if inCourseBlock {
			// 提取周次：1-16周
			if strings.Contains(line, "周") {
				if weekMatch := regexp.MustCompile(`(\d+)-(\d+)周`).FindStringSubmatch(line); len(weekMatch) > 2 {
					currentCourse.StartWeek, _ = strconv.Atoi(weekMatch[1])
					currentCourse.EndWeek, _ = strconv.Atoi(weekMatch[2])
					fmt.Printf("    📅 [DEBUG] 提取周次: %d-%d周\n", currentCourse.StartWeek, currentCourse.EndWeek)
				}

				// 特殊周次：8周,14周
				if strings.Contains(line, "周,") {
					if weekSpecialMatch := regexp.MustCompile(`(\d+)周,(\d+)周`).FindStringSubmatch(line); len(weekSpecialMatch) > 2 {
						currentCourse.StartWeek, _ = strconv.Atoi(weekSpecialMatch[1])
						currentCourse.EndWeek, _ = strconv.Atoi(weekSpecialMatch[2])
						currentCourse.WeekType = "指定周"
					}
				}

				// 检测单双周
				if strings.Contains(line, "单周") {
					currentCourse.WeekType = "单周"
				} else if strings.Contains(line, "双周") {
					currentCourse.WeekType = "双周"
				}
			}

			// 提取教室：/场地: 文渊607
			if strings.Contains(line, "场地:") {
				if classroomMatch := regexp.MustCompile(`场地[:  ：]([^/\n]+)`).FindStringSubmatch(line); len(classroomMatch) > 1 {
					classroom := strings.TrimSpace(classroomMatch[1])

					// 去掉 "/" 后的内容
					if idx := strings.Index(classroom, "/"); idx > 0 {
						classroom = strings.TrimSpace(classroom[:idx])
					}

					// 去掉"教师:"后的内容
					if idx := strings.Index(classroom, "教师:"); idx > 0 {
						classroom = strings.TrimSpace(classroom[:idx])
					}

					// 限制长度
					if len(classroom) > 100 {
						classroom = classroom[:100]
					}

					currentCourse.Classroom = classroom
				}
			}

			// 提取教师：/教师: 赵崇和
			if strings.Contains(line, "教师:") {
				if teacherMatch := regexp.MustCompile(`教师[:：]([^/\n]+)`).FindStringSubmatch(line); len(teacherMatch) > 1 {
					teacher := strings.TrimSpace(teacherMatch[1])

					// 只保留教师名（去掉"教学班"后面的内容）
					if idx := strings.Index(teacher, "教学班"); idx > 0 {
						teacher = strings.TrimSpace(teacher[:idx])
					}

					// 只取第一个词（教师名）
					words := strings.Fields(teacher)
					if len(words) > 0 {
						teacher = words[0]
					}

					// 限制长度
					if len(teacher) > 50 {
						teacher = teacher[:50]
					}

					currentCourse.Teacher = teacher
				}
			}

			// 检测课程块结束
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				// 如果下一行是新课程标记、空行、节次标记或时间标记
				if nextLine == "" ||
					strings.HasSuffix(nextLine, "*") ||
					strings.HasSuffix(nextLine, "@") ||
					strings.HasSuffix(nextLine, "#") ||
					strings.HasSuffix(nextLine, "&") ||
					regexp.MustCompile(`^\d+-\d+$`).MatchString(nextLine) ||
					regexp.MustCompile(`^\d+$`).MatchString(nextLine) ||
					(strings.Contains(nextLine, ": ") && len(nextLine) < 10) {

					if currentCourse.CourseName != "" && currentCourse.WeekDay > 0 {
						courses = append(courses, currentCourse)
						inCourseBlock = false
					}
				}
			}
		}
	}

	// 保存最后一门课程
	if inCourseBlock && currentCourse.CourseName != "" && currentCourse.WeekDay > 0 {
		courses = append(courses, currentCourse)
	}

	return courses
}

// Min 返回两个整数中的较小值
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
