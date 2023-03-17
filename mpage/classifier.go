package mpage

import (
	"MyProjectLab_8/mlog"
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strconv"
	"strings"
)

/* ****************************************************** */
func CityClassifier(w http.ResponseWriter, r *http.Request) {
	// функция для загрузки справочника городов
	// ---------------------------------------------------------
	sOut := ""
	// проверяем, что у нас не запрос GET, а запрос POST
	// если запрос был искусственно отправлен методом GET, то
	// возвращаем сообщение об этом
	// из программы приходит только запрос POST
	if r.Method == "GET" {
		fmt.Fprintf(w, "%v", "")
		return
	} else {
		// ---------------------------------------------------------
		// проверяем, если есть проблемы с входом в сессию методом POST,
		// если есть, возвращаем сообщение об ошибке
		if mlog.CheckLoginPOST(w, r) == 0 {
			fmt.Fprintf(w, "%v", "")
			return
		}
		// ---------------------------------------------------------
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			mlog.CheckErr(err, "Ошибка запроса POST: CityClassifier")
		}
		// ---------------------------------------------------------
		// в зависимости от значения параметра numvariety
		// отображаем в видимой части раскрывающегося списка
		// то или иное поле
		numvariety := r.FormValue("numvariety")
		numvariety = strings.TrimSpace(numvariety)
		sNameField := ""
		if numvariety == "1" {
			sNameField = "name"
		}
		// ---------------------------------------------------------
		sOut = "<option value=\"\"></option>"
		// ---------------------------------------------------------
		// обращаемся к базе данных, готовим текст запроса
		stmt, err := Exdbmysqlg.Prepare("SELECT id, " + sNameField + " " +
			"FROM city " +
			";")
		// проверяем, что удалось сформировать запрос
		mlog.CheckErr(err, "Не могу подготовить запрос к БД activity")
		defer stmt.Close()
		// выполняем запрос
		rows, err := stmt.Query()
		// проверяем, что удалось выполнить запрос
		mlog.CheckErr(err, "Не могу выполнить запрос в БД activity")
		defer rows.Close()
		// объявляем переменные соответствующего типа для считывания полей
		var id sql.NullInt64
		var name sql.NullString
		for rows.Next() {
			err = rows.Scan(&id, &name)
			mlog.CheckErr(err, "Не могу прочесть запись")
			sId := strconv.FormatInt(id.Int64, 10)
			// формируем содержимое раскрывающегося списка
			sOut += "<option value=\"" + sId + "\">" + name.String + "</option>\n"
		}
	} // POST
	fmt.Fprintf(w, "%v", sOut)
}
