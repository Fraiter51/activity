package mpage

import (
	MODMLOG "MyProjectLab_8/mlog"
	"database/sql"
	"fmt"
	_ "github.com/360EntSecGroup-Skylar/excelize"
	_ "github.com/go-sql-driver/mysql"
	"github.com/nguyenthenguyen/docx"
	"net/http"
	"strconv"
	"strings"
	"time"
)

/* ****************************************************** */
// функция построения отчетов
func SearchReport(w http.ResponseWriter, r *http.Request) {
	// ---------------------------------------------------------
	// объявляем переменные
	format := ""
	studentid := ""
	confname := ""
	projectname := ""
	papername := ""
	// ---------------------------------------------------------
	// обрабатываем запрос GET для получения файлов формата xlsx и docx
	if r.Method == "GET" {
		// ---------------------------------------------------------
		MODMLOG.CheckLoginGET(w, r)
		// ---------------------------------------------------------
		query := r.URL.Query()
		format = query.Get("format")
		format = strings.TrimSpace(format)
		// если id студента не задан, выполняем поиск для всех студентов
		studentid = query.Get("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		// если не задано название конференции, проекта или статьи
		// выполняем поиск по всей научной деятельности
		confname = query.Get("confname")
		confname = strings.TrimSpace(confname)
		if confname == "" {
			confname = "%"
		}
		projectname = query.Get("projectname")
		projectname = strings.TrimSpace(projectname)
		if projectname == "" {
			projectname = "%"
		}
		papername = query.Get("papername")
		papername = strings.TrimSpace(papername)
		if papername == "" {
			papername = "%"
		}
	} else {
		// запрос POST используется для вывода HTML отчета
		// ---------------------------------------------------------
		if MODMLOG.CheckLoginPOST(w, r) == 0 {
			fmt.Fprintf(w, "%v", "0####/")
			return
		}
		// ---------------------------------------------------------
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			MODMLOG.CheckErr(err, "Ошибка запроса POST: SearchReport")
		}
		// ---------------------------------------------------------
		format = r.FormValue("format")
		format = strings.TrimSpace(format)
		// если id студента не задан, выполняем поиск для всех студентов
		studentid = r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		// если не задано название конференции, проекта или статьи
		// выполняем поиск по всей научной деятельности
		confname = r.FormValue("confname")
		confname = strings.TrimSpace(confname)
		if confname == "" {
			confname = "%"
		}
		projectname = r.FormValue("projectname")
		projectname = strings.TrimSpace(projectname)
		if projectname == "" {
			projectname = "%"
		}
		papername = r.FormValue("papername")
		papername = strings.TrimSpace(papername)
		if papername == "" {
			papername = "%"
		}
	}
	// ---------------------------------------------------------
	// подготовка запроса к серверу
	stmt, err := Exdbmysqlg.Prepare("SELECT student.fio, " +
		"IFNULL((SELECT SUM(student_conference.point) FROM student_conference " +
		"LEFT JOIN conference ON student_conference.conference_id=conference.id " + "WHERE student_conference.student_id=student.id AND conference.name LIKE? " +
		"),0) as conference_point, " + "IFNULL((SELECT SUM(IFNULL(student_project.point,0)) FROM student_project" +
		"LEFT JOIN project ON student_project.project_id=project.id " +
		"WHERE student_project.student_id=student.id AND project.name LIKE ? " +
		"),0) as project_point, " +
		"IFNULL((SELECT SUM(IFNULL(student_paper.point,0)) FROM student_paper " +
		"LEFT JOIN paper ON student_paper.paper_id=paper.id " +
		"WHERE student_paper.student_id=student.id AND paper.name LIKE ? " +
		"),0) as paper_point " +
		"FROM student " +
		"WHERE student.id LIKE ? " +
		"ORDER BY fio " +
		";")
	MODMLOG.CheckErr(err, "Не могу подготовить запрос к БД activity")
	defer stmt.Close()
	// выполнение запроса
	rows, err := stmt.Query(
		confname+"%",
		projectname+"%",
		papername+"%",
		studentid)
	MODMLOG.CheckErr(err, "Не могу выполнить запрос в БД activity")
	defer rows.Close()
	// ---------------------------------------------------------
	// в зависимости от значения переменной format вызываем различные
	// подпрограммы
	switch format {
	case "HTML":
		ReportHTML(rows, w)
	case "XLSX":
		fmt.Println("XLS error")
		//ReportXLSX(rows, w)
	case "DOCX":
		ReportDOCX(rows, w)
	default:
		ReportHTML(rows, w)
	}
	// ---------------------------------------------------------
}

/* ****************************************************** */
// функция построения отчета в формате html
func ReportHTML(rows *sql.Rows, w http.ResponseWriter) {
	sOut := ""
	// объявляем переменные
	nItogoConf := 0
	nItogoProject := 0
	nItogoPaper := 0
	nItogoStudent := 0
	// объявляем переменные соответствующего типа для считывания полей
	var fio sql.NullString
	var conference_point sql.NullString
	var project_point sql.NullString
	var paper_point sql.NullString
	// набор строк передается как аргумент функции
	for rows.Next() {
		err := rows.Scan(&fio, &conference_point,
			&project_point, &paper_point)
		MODMLOG.CheckErr(err, "Не могу прочесть запись")
		nItogoStudent = 0
		if n, err := strconv.Atoi(conference_point.String); err == nil {
			nItogoConf += n
			nItogoStudent += n
		}
		if n, err := strconv.Atoi(project_point.String); err == nil {
			nItogoProject += n
			nItogoStudent += n
		}
		if n, err := strconv.Atoi(paper_point.String); err == nil {
			nItogoPaper += n
			nItogoStudent += n
		}
		// формируем строки таблицы
		sOut += "<tr><td>" + fio.String + "</td><td>&nbsp;&nbsp;</td>" +
			"<td>" + conference_point.String + "</td><td>&nbsp;&nbsp;</td>" +
			"<td>" + project_point.String + "</td><td>&nbsp;&nbsp;</td>" +
			"<td>" + paper_point.String + "</td><td>&nbsp;&nbsp;</td>" +
			"<td>" + strconv.Itoa(nItogoStudent) + "</td></tr>\n"
		// break
	}
	sOut += "<tr><td><b>Итого:</b></td><td>&nbsp;&nbsp;</td>" +
		"<td>" + strconv.Itoa(nItogoConf) + "</td><td>&nbsp;&nbsp;</td>" +
		"<td>" + strconv.Itoa(nItogoProject) + "</td><td>&nbsp;&nbsp;</td>" +
		"<td>" + strconv.Itoa(nItogoPaper) + "</td><td>&nbsp;&nbsp;</td>" +
		"<td>" + strconv.Itoa(nItogoConf+nItogoProject+nItogoPaper) + "</td></tr>\n"
	// формируем заголовок таблицы с использованием Bootstrap
	sOut = "<table class=\"table table-striped\">\n" +
		" <thead>\n" +
		" <tr>\n" +
		" <th scope=\"col\">ФИО</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Конференции</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Проекты</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Статьи</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Все</th>\n" +
		" </tr>\n" +
		" </thead>\n" +
		" <tbody>\n" +
		sOut +
		" </tbody>" +
		"</table>\n"
	fmt.Fprintf(w, "%v", sOut)
}

/* ****************************************************** */
// функция построения отчета в формате xlsx
//func ReportXLSX(rows *sql.Rows, w http.ResponseWriter) {
//	// объявляем переменные
//	nItogoConf := 0
//	nItogoProject := 0
//	nItogoPaper := 0
//	nItogoStudent := 0
//	// объявляем переменные соответствующего типа для считывания полей
//	var fio sql.NullString
//	var conference_point sql.NullString
//	var project_point sql.NullString
//	var paper_point sql.NullString
//	// ---------------------------------------------------------
//	// задаем номер левой верхней ячейки
//	// задаем корректирующие коэффициенты для осей X и Y
//	nLeftCol := 1
//	nUpperRow := 1
//	sNameCol := ""
//	dCoeffX := 1.018
//	dCoeffY := 1.0
//	// ---------------------------------------------------------
//	xlsx := excelize.NewFile()
//	// создаем новый лист
//	sNameSheet := "Sheet1"
//	index := xlsx.NewSheet(sNameSheet)
//	// ориентация страницы
//	err := xlsx.SetPageLayout(
//		sNameSheet,
//		excelize.PageLayoutOrientation(excelize.OrientationLandscape),
//	)
//	if err != nil {
//		panic(err)
//	}
//	// задаем параметры страницы
//	err = xlsx.SetPageMargins(
//		sNameSheet,
//		excelize.PageMarginBottom(0.2),
//		excelize.PageMarginFooter(0.2),
//		excelize.PageMarginHeader(0.2),
//		excelize.PageMarginLeft(0.2),
//		excelize.PageMarginRight(0.2),
//		excelize.PageMarginTop(0.2),
//	)
//	if err != nil {
//		panic(err)
//	}
//	// ---------------------------------------------------------
//	// задаем границы и шрифт ячейки
//	var vBorder = `[
// { "type": "left", "color": "000000", "style": 1 },
// { "type": "top", "color": "000000", "style": 1 },
// { "type": "bottom", "color": "000000", "style": 1 },
// { "type": "right", "color": "000000", "style": 1 }]`
//	styleTNR12, err := xlsx.NewStyle(`{"font":{"family": "Times New
//Roman","size":12,"bold":false},
// "alignment":{"wrap_text":true, "vertical":"center", "horizontal":"center"},
// "fill":{"type":"pattern","color":["#FFFFFF"],"pattern":1}, "border":
//` + vBorder + `}`)
//	styleTNR12L, err := xlsx.NewStyle(`{"font":{"family": "Times New
//Roman","size":12,"bold":false},
// "alignment":{"wrap_text":true, "vertical":"center", "horizontal":"left"},
// "fill":{"type":"pattern","color":["#FFFFFF"],"pattern":1}, "border":
//` + vBorder + `}`)
//	styleTNR12B, err := xlsx.NewStyle(`{"font":{"family": "Times New
//Roman","size":12,"bold":true},
// "alignment":{"wrap_text":true, "vertical":"center", "horizontal":"center"},
// "fill":{"type":"pattern","color":["#FFFFFF"],"pattern":1}, "border":
//` + vBorder + `}`)
//	styleTNR12BL, err := xlsx.NewStyle(`{"font":{"family": "Times New
//Roman","size":12,"bold":true},
// "alignment":{"wrap_text":true, "vertical":"center", "horizontal":"left"},
// "fill":{"type":"pattern","color":["#FFFFFF"],"pattern":1}, "border":
//` + vBorder + `}`)
//	// ---------------------------------------------------------
//	// задаем стили и название колонок
//	nMeter := 0
//	xlsx.SetRowHeight(sNameSheet, nMeter, 30*dCoeffY)
//	xlsx.MergeCell(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter),
//		ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter))
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol, nUpperRow+nMeter),
//		styleTNR12B)
//	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter),
//		"Отчет")
//	nMeter++
//	xlsx.SetRowHeight(sNameSheet, nMeter, 25*dCoeffY)
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol, nUpperRow+nMeter),
//		styleTNR12B)
//	sNameCol = ExcelX2Name(nLeftCol)
//	xlsx.SetColWidth(sNameSheet, sNameCol,
//		sNameCol, 35*dCoeffX)
//	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter),
//		"ФИО")
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+1,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter),
//		styleTNR12B)
//	sNameCol = ExcelX2Name(nLeftCol + 1)
//	xlsx.SetColWidth(sNameSheet,
//		sNameCol, sNameCol, 20.5*dCoeffX)
//	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol+1,
//		nUpperRow+nMeter), "Конференции")
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+2,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter),
//		styleTNR12B)
//	sNameCol = ExcelX2Name(nLeftCol + 2)
//	xlsx.SetColWidth(sNameSheet,
//		sNameCol, sNameCol, 20.5*dCoeffX)
//	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol+2,
//		nUpperRow+nMeter), "Проекты")
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+3,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter),
//		styleTNR12B)
//	sNameCol = ExcelX2Name(nLeftCol + 3)
//	xlsx.SetColWidth(sNameSheet,
//		sNameCol, sNameCol, 20.5*dCoeffX)
//	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol+3,
//		nUpperRow+nMeter), "Статьи")
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+4,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter),
//		styleTNR12B)
//	sNameCol = ExcelX2Name(nLeftCol + 4)
//	xlsx.SetColWidth(sNameSheet,
//		sNameCol, sNameCol, 20.5*dCoeffX)
//	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol+4,
//		nUpperRow+nMeter), "Все")
//	// ---------------------------------------------------------
//	// формируем строки отчета
//	for rows.Next() {
//		nMeter++
//		err := rows.Scan(&fio, &conference_point,
//			&project_point, &paper_point)
//		MODMLOG.CheckErr(err, "Не могу прочесть запись")
//		err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol,
//			nUpperRow+nMeter), ExcelXY2Name(nLeftCol, nUpperRow+nMeter),
//			styleTNR12L)
//		xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter),
//			fio.String)
//		nItogoStudent = 0
//		if n, err := strconv.Atoi(conference_point.String); err == nil {
//			err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+1,
//				nUpperRow+nMeter), ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter),
//				styleTNR12)
//			xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter),
//				n)
//			nItogoConf += n
//			nItogoStudent += n
//		}
//		if n, err := strconv.Atoi(project_point.String); err == nil {
//			err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+2,
//				nUpperRow+nMeter), ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter),
//				styleTNR12)
//			xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter),
//				n)
//			nItogoProject += n
//			nItogoStudent += n
//		}
//		if n, err := strconv.Atoi(paper_point.String); err == nil {
//			err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+3,
//				nUpperRow+nMeter), ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter),
//				styleTNR12)
//			xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter),
//				n)
//			nItogoPaper += n
//			nItogoStudent += n
//		}
//		err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+4,
//			nUpperRow+nMeter), ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter),
//			styleTNR12)
//		xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter),
//			nItogoStudent)
//	}
//	nMeter++
//	// формируем строку Итого
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol, nUpperRow+nMeter),
//		styleTNR12BL)
//	xlsx.SetCellValue(sNameSheet, ExcelXY2Name(nLeftCol, nUpperRow+nMeter),
//		"Итого:")
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+1,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter),
//		styleTNR12B)
//	xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+1, nUpperRow+nMeter),
//		nItogoConf)
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+2,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter),
//		styleTNR12B)
//	xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+2, nUpperRow+nMeter),
//		nItogoProject)
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+3,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter),
//		styleTNR12B)
//	xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+3, nUpperRow+nMeter),
//		nItogoPaper)
//	err = xlsx.SetCellStyle(sNameSheet, ExcelXY2Name(nLeftCol+4,
//		nUpperRow+nMeter), ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter),
//		styleTNR12B)
//	xlsx.SetCellInt(sNameSheet, ExcelXY2Name(nLeftCol+4, nUpperRow+nMeter),
//		nItogoConf+nItogoProject+nItogoPaper)
//	// ---------------------------------------------------------
//	xlsx.SetActiveSheet(index)
//	file := xlsx
//	sDate := time.Now().Format("2006-01-02")
//	sDate = MODMLOG.DateToRus(sDate)
//	// Устанавливается заголовок для отображения браузером загружаемого файла
//	w.Header().Set("Content-Type", "application/octet-stream")
//	w.Header().Set("Content-Disposition",
//		"attachment;filename="+sDate+"_Отчет.xlsx")
//	w.Header().Set("File-Name", sDate+"_Отчет.xlsx") // userInputData.xlsx
//	w.Header().Set("Content-Transfer-Encoding", "binary")
//	w.Header().Set("Expires", "0")
//	err = file.Write(w)
//	if err != nil {
//		fmt.Println(err)
//	}
//	// ---------------------------------------------------------
//}

/* ****************************************************** */
// функция построения отчета в формате docx
func ReportDOCX(rows *sql.Rows, w http.ResponseWriter) {
	// объявляем переменные
	nItogoConf := 0
	nItogoProject := 0
	nItogoPaper := 0
	// объявляем переменные соответствующего типа для считывания полей
	var fio sql.NullString
	var conference_point sql.NullString
	var project_point sql.NullString
	var paper_point sql.NullString
	// формируем значение строки отчета
	for rows.Next() {
		err := rows.Scan(&fio, &conference_point,
			&project_point, &paper_point)
		MODMLOG.CheckErr(err, "Не могу прочесть запись")
		if n, err := strconv.Atoi(conference_point.String); err == nil {
			nItogoConf += n
		}
		if n, err := strconv.Atoi(project_point.String); err == nil {
			nItogoProject += n
		}
		if n, err := strconv.Atoi(paper_point.String); err == nil {
			nItogoPaper += n
		}
	}
	// ---------------------------------------------------------
	// считываем файл образца
	rWord, err := docx.ReadDocxFile("mpage/report.docx")
	MODMLOG.CheckErr(err, "Не могу получить доступ к report.docx")
	docx1 := rWord.Editable()
	// заменяем в образце паттерны на считанные значения
	docx1.Replace("old_1_1", strconv.Itoa(nItogoConf), -1)
	docx1.Replace("old_1_2", strconv.Itoa(nItogoProject), -1)
	docx1.Replace("old_1_3", strconv.Itoa(nItogoPaper), -1)
	// ---------------------------------------------------------
	file := docx1
	sDate := time.Now().Format("2006-01-02")
	sDate = MODMLOG.DateToRus(sDate)
	// Устанавливается заголовок для отображения браузером загружаемого файла
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", "attachment;filename=WhatsApp"+sDate+".docx")
	w.Header().Set("File-Name", "WhatsApp "+sDate+".docx")
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Expires", "0")
	err = file.Write(w)
	if err != nil {
		fmt.Println(err)
	}
	rWord.Close()
	// ---------------------------------------------------------
}

// *********************************************************
// перевод координат ячеек в имя ячейки
//func ExcelXY2Name(X int, Y int) string {
//	c, err := excelize.CoordinatesToCellName(X, Y)
//	MODMLOG.CheckErr(err, "Ошибка в ExcelXY2Name")
//	return c
//}

// *********************************************************
// перевод номера ячейки по X в имя ячейки по X
//func ExcelX2Name(X int) string {
//	c, err := excelize.ColumnNumberToName(X)
//	MODMLOG.CheckErr(err, "Ошибка в ExcelX2Name")
//	return c
//}
