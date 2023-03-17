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
// функция для поиска и выбора конференции
func SearchСonference(w http.ResponseWriter, r *http.Request) {
	// ---------------------------------------------------------
	// проверяем, что у нас не запрос GET, а запрос POST
	sOut := ""
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
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			mlog.CheckErr(err, "Ошибка запроса POST: SearchСonference")
		}
		// ---------------------------------------------------------
		// если id студента не задан, осуществляем поиск для всех студентов
		studentid := r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		// считываем данные из текстовых полей
		confname := r.FormValue("confname")
		confcity := r.FormValue("confcity")
		confcity = strings.TrimSpace(confcity)
		if confcity == "" {
			confcity = "%"
		}
		confdatestart := r.FormValue("confdatestart")
		confdatestart = strings.TrimSpace(confdatestart)
		confdateend := r.FormValue("confdateend")
		confdateend = strings.TrimSpace(confdateend)
		// если не указана конечная дата, делаем ее максимально большой
		// даже если такой не существует
		if confdateend == "" {
			confdateend = "9999-99-99"
		}
		// ---------------------------------------------------------
		// обращаемся к базе данных, готовим текст запроса
		stmt, err := Exdbmysqlg.Prepare("SELECT conference.name as conferencename, " +
			"city.name as cityname, " +
			"conference.date_start as conferencedate_start, " +
			"conference.date_end as conferencedate_end " +
			"FROM conference " +
			"LEFT JOIN city ON conference.city_id=city.id " +
			"WHERE EXISTS(SELECT * FROM student_conference WHERE student_conference.conference_id=conference.id AND student_conference.student_id LIKE ? OR ?='%') AND " +
			"conference.name LIKE ? AND city.id LIKE ? AND " +
			"( " +
			"(\"\"=? AND \"9999-99-99\"=?) OR " +
			"(IFNULL(conference.date_start,\"\") BETWEEN ? AND ?) OR " +
			"(IFNULL(conference.date_end,\"\") BETWEEN ? AND ?) OR " +
			"(IFNULL(conference.date_start,\"\")<=? AND ?<=IFNULL(conference.date_end,\"\")) OR " +
			"(IFNULL(conference.date_start,\"\")<=? AND ?<=IFNULL(conference.date_end,\"\")) " +
			") " +
			";")
		// проверяем, что удалось сформировать запрос
		mlog.CheckErr(err, "Не могу подготовить запрос к БД activity")
		defer stmt.Close()
		// выполняем запрос
		rows, err := stmt.Query(studentid, studentid,
			confname+"%",
			confcity+"%",
			confdatestart, confdateend, // "min max"
			confdatestart, confdateend,
			confdatestart, confdateend,
			confdatestart, confdatestart,
			confdateend, confdateend)
		// проверяем, что удалось выполнить запрос
		mlog.CheckErr(err, "Не могу выполнить запрос в БД activity")
		defer rows.Close()
		// объявляем переменные соответствующего типа для считывания полей
		var conferencename sql.NullString
		var cityname sql.NullString
		var conferencedate_start sql.NullString
		var conferencedate_end sql.NullString
		for rows.Next() {
			// считывание строки
			err = rows.Scan(&conferencename, &cityname,
				&conferencedate_start, &conferencedate_end)
			// проверяем, что записи считались
			mlog.CheckErr(err, "Не могу прочесть запись")
			// переводим дату в соответствующий формат
			conferencedate_start.String =
				mlog.DateToRus(conferencedate_start.String)
			conferencedate_end.String =
				mlog.DateToRus(conferencedate_end.String)
			// формируем из считанных полей БД строку таблицы HTML
			sOut += "<tr><td>" + conferencename.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + cityname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + conferencedate_start.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + conferencedate_end.String + "</td></tr>\n"
		}
	} // POST
	// формируем заголовок таблицы с использованием Bootstrap
	sOut = "<table class=\"table table-striped\">\n" +
		" <thead>\n" +
		" <tr>\n" +
		" <th scope=\"col\">Название конференции</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Город</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Дата начала</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Дата окончания</th>\n" +
		" </tr>\n" +
		" </thead>\n" +
		" <tbody>\n" +
		sOut +
		" </tbody>" +
		"</table>\n"
	fmt.Fprintf(w, "%v", sOut)
}
