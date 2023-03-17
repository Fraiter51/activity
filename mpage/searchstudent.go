package mpage

import (
	"MyProjectLab_8/mlog"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
)

// переменная которой было в пакете main файла activity.go присвоено
// значение, необходимое для работы с базой данных
var Exdbmysqlg *sql.DB

/* ****************************************************** */
// функция для поиска и выбора студента
// выбор студента осуществляется двойным щелчком мыши по записи
func SearchStudent(w http.ResponseWriter, r *http.Request) {
	// ---------------------------------------------------------
	// проверяем, что у нас не запрос GET, а запрос POST
	// в данной программе ограничиваемся этими двумя запросами
	sOut := ""
	if r.Method == "GET" {
		// если запрос был искусственно отправлен методом GET, то
		// возвращаем сообщение об этом
		// из программы приходит только запрос POST
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
		// как multipart/form-data
		// multipart/form-data чаще всего используется для отправки HTML-форм с
		// бинарными данными методом POST протокола HTTP
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			mlog.CheckErr(err, "Ошибка запроса POST: SearchStudent")
		}
		// ---------------------------------------------------------
		// считываем данные из текстового поля
		searchstr := r.FormValue("searchstr")
		// ---------------------------------------------------------
		// обращаемся к базе данных, готовим текст запроса
		stmt, err := Exdbmysqlg.Prepare("SELECT student.id as studentid, " +
			"student.fio as studentfio, " +
			"university.name as universityname, " +
			"faculty.name as facultyname, " +
			"speciality.name as specialitytyname, " +
			"student.contacts as studentcontacts " +
			"FROM student " +
			"LEFT JOIN university_faculty_speciality ON " +
			"student.university_faculty_speciality_id=university_faculty_speciality.id " +
			"LEFT JOIN university ON " +
			"university_faculty_speciality.university_id=university.id " +
			"LEFT JOIN faculty ON university_faculty_speciality.faculty_id=faculty.id " +
			"LEFT JOIN speciality ON university_faculty_speciality.speciality_id=speciality.id " +
			"WHERE student.fio LIKE ? OR student.contacts LIKE ? ;")
		// проверяем, что удалось сформировать запрос
		mlog.CheckErr(err, "Не могу подготовить запрос к БД activity")
		defer stmt.Close()
		// выполняем запрос
		rows, err := stmt.Query(searchstr+"%", searchstr+"%")
		// проверяем, что удалось выполнить запрос
		mlog.CheckErr(err, "Не могу выполнить запрос в БД activity")
		defer rows.Close()
		// объявляем переменные соответствующего типа для считывания полей
		var studentid sql.NullInt64
		var studentfio sql.NullString
		var universityname sql.NullString
		var facultyname sql.NullString
		var specialityname sql.NullString
		var studentcontacts sql.NullString
		// цикл по всем возвращенным строкам
		for rows.Next() {
			// считывание строки
			err = rows.Scan(&studentid, &studentfio,
				&universityname, &facultyname, &specialityname,
				&studentcontacts)
			// проверяем, что записи считались
			mlog.CheckErr(err, "Не могу прочесть запись")
			// формируем из считанных полей БД строку таблицы HTML
			sOut += "<tr onDblClick=\"ChooseStudent(" + strconv.FormatInt(studentid.Int64, 10) + ", '" + studentfio.String +
				"'); return false;\"><td>" + studentfio.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + universityname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + facultyname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + specialityname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + studentcontacts.String + "</td></tr>\n"

		}
	} // POST
	// формируем заголовок таблицы с использованием Bootstrap
	sOut = "<table class=\"table table-striped\">\n" +
		" <thead>\n" +
		" <tr>\n" +
		" <th scope=\"col\">ФИО</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">ВУЗ</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Факультет</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Специальность</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Контакт</th>\n" +
		" </tr>\n" +
		" </thead>\n" +
		" <tbody>\n" +
		sOut +
		" </tbody>" +
		"</table>\n"
	fmt.Fprintf(w, "%v", sOut)
}
