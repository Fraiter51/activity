package mpage

import (
	"MyProjectLab_8/mlog"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strings"
)

/* ****************************************************** */
func SearchPaper(w http.ResponseWriter, r *http.Request) {
	// функция для поиска и выбора научных статей
	// ---------------------------------------------------------
	sOut := ""
	// проверяем, что у нас не запрос GET, а запрос POST
	// если запрос был искусственно отправлен методом GET, то
	// возвращаем сообщение об этом
	// из программы приходит только запрос POST
	if r.Method == "GET" {
		fmt.Fprintf(w, "%v", sOut+" GET ")
		return
	} else {
		// ---------------------------------------------------------
		// проверяем, если есть проблемы с входом в сессию методом POST,
		// если есть, возвращаем сообщение об ошибке
		if mlog.CheckLoginPOST(w, r) == 0 {
			fmt.Fprintf(w, "%v", "0####/")
			return
		}
		// ---------------------------------------------------------
		// считываем данные формы
		// функция ParseMultipartForm пакета net/http анализирует тело запроса
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			mlog.CheckErr(err, "Ошибка запроса POST: SearchPaper")
		}
		// ---------------------------------------------------------
		// если id студента не задан, осуществляем поиск для всех студентов
		studentid := r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		// fmt.Println("studentid=",studentid)
		if studentid == "" {
			studentid = "%"
		}
		papername := r.FormValue("papername")
		journalname := r.FormValue("journalname")
		publishingname := r.FormValue("publishingname")
		paperdatestart := r.FormValue("paperdatestart")
		paperdatestart = strings.TrimSpace(paperdatestart)
		paperdateend := r.FormValue("paperdateend")
		paperdateend = strings.TrimSpace(paperdateend)
		if paperdateend == "" {
			paperdateend = "9999-99-99"
		}
		// fmt.Fprintf(w, "%v", confname)
		// return
		// ---------------------------------------------------------
		// обращаемся к базе данных, готовим текст запроса
		stmt, err := Exdbmysqlg.Prepare("SELECT " +
			"paper.name, " +
			"scientific_journal.name, " +
			"scientific_journal.publishing, " +
			"scientific_journal.date " +
			"FROM paper " +
			"LEFT JOIN scientific_journal ON paper.scientific_journal_id=scientific_journal.id " +
			"WHERE EXISTS(SELECT * FROM student_paper WHERE student_paper.paper_id=paper.id AND student_paper.student_id LIKE ? OR ?='%') " +
			"AND paper.name LIKE ? " +
			"AND scientific_journal.name LIKE ? " +
			"AND scientific_journal.publishing LIKE ? " +
			"AND IFNULL(scientific_journal.date,'') BETWEEN ? AND ? " +
			";")

		// проверяем, что удалось сформировать запрос
		mlog.CheckErr(err, "Не могу подготовить запрос к БД activity")
		defer stmt.Close()
		fmt.Println("paperdatestart=", paperdatestart, "paperdateend=",
			paperdateend)
		// выполняем запрос
		rows, err := stmt.Query(studentid, studentid,
			papername+"%",
			journalname+"%",
			publishingname+"%",
			paperdatestart, paperdateend)
		// проверяем, что удалось выполнить запрос
		mlog.CheckErr(err, "Не могу выполнить запрос в БД activity")
		defer rows.Close()
		// объявляем переменные соответствующего типа для считывания полей
		var papername2 sql.NullString
		var journalname2 sql.NullString
		var publishingname2 sql.NullString
		var scientific_journaldate2 sql.NullString
		// считывание строки
		for rows.Next() {
			err = rows.Scan(&papername2, &journalname2, &publishingname2,
				&scientific_journaldate2)
			mlog.CheckErr(err, "Не могу прочесть запись")
			scientific_journaldate2.String =
				mlog.DateToRus(scientific_journaldate2.String)
			sOut += "<tr><td>" + papername2.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + journalname2.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + publishingname2.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + scientific_journaldate2.String + "</td></tr>\n"
		}
	} // POST
	// формируем заголовок таблицы с использованием Bootstrap
	sOut = "<table class=\"table table-striped\">\n" +
		" <thead>\n" +
		" <tr>\n" +
		" <th scope=\"col\">Название статьи</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Название журнала</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Название издательства</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Дата публикации</th>\n" +
		" </tr>\n" +
		" </thead>\n" +
		" <tbody>\n" +
		sOut +
		" </tbody>" +
		"</table>\n"
	fmt.Fprintf(w, "%v", sOut)
}
