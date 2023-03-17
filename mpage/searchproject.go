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
// функция для поиска и выбора научного проекта
func SearchProject(w http.ResponseWriter, r *http.Request) {
	// ---------------------------------------------------------
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
			mlog.CheckErr(err, "Ошибка запроса POST: SearchProject")
		}
		// ---------------------------------------------------------
		// если id студента не задан, осуществляем поиск для всех студентов
		studentid := r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		// считываем данные из текстовых полей
		projectname := r.FormValue("projectname")
		projectdatestart := r.FormValue("projectdatestart")
		projectdatestart = strings.TrimSpace(projectdatestart)
		projectdateend := r.FormValue("projectdateend")
		projectdateend = strings.TrimSpace(projectdateend)
		// если не указана конечная дата, делаем ее максимально большой
		// даже если такой не существует
		if projectdateend == "" {
			projectdateend = "9999-99-99"
		}
		projectfio := r.FormValue("projectfio")
		projectfio = strings.TrimSpace(projectfio)
		if projectfio == "" {
			projectfio = "%"
		}
		projectcity := r.FormValue("projectcity")
		projectcity = strings.TrimSpace(projectcity)
		if projectcity == "" {
			projectcity = "%"
		}
		projectorganization := r.FormValue("projectorganization")
		projectorganization = strings.TrimSpace(projectorganization)
		if projectorganization == "" {
			projectorganization = "%"
		}
		projectcontacts := r.FormValue("projectcontacts")
		projectcontacts = strings.TrimSpace(projectcontacts)
		if projectcontacts == "" {
			projectcontacts = "%"
		}
		projectposition := r.FormValue("projectposition")
		projectposition = strings.TrimSpace(projectposition)
		if projectposition == "" {
			projectposition = "%"
		}
		// ---------------------------------------------------------
		// обращаемся к базе данных, готовим текст запроса
		stmt, err := Exdbmysqlg.Prepare(`
	SELECT project.name as projectname, 
		project.date_start as projectdate_start, 
		project.date_end as projectdate_end, 
		project_manager.fio as project_managerfio, 
		city.name as cityname, 
		organization.name as organizationname, 
		project_manager.contacts as project_managercontacts, 
		project_manager.position as project_managerposition 
	FROM project 
	LEFT JOIN project_manager ON project.project_manager_id=project_manager.id 
	LEFT JOIN organization ON project_manager.organization_id=organization.id
	LEFT JOIN city ON organization.city_id=city.id 
	WHERE 
		EXISTS (
			SELECT * 
			FROM student_project 
			WHERE student_project.project_id=project.id 
			AND student_project.student_id LIKE ? OR ?='%'
		) 
		AND project.name LIKE ? 
		AND (
			(""=? AND "9999-99-99"=?) 
			OR (IFNULL(project.date_start,"") BETWEEN ? AND ?) 
			OR (IFNULL(project.date_end,"") BETWEEN ? AND ?) 
			OR (IFNULL(project.date_start,"")<=? AND ?<=IFNULL(project.date_end,"")) 
			OR (IFNULL(project.date_start,"")<=? AND ?<=IFNULL(project.date_end,""))
		) 
		AND project_manager.fio LIKE ? 
		AND city.id LIKE ? 
		AND organization.name LIKE ? 
		AND project_manager.contacts LIKE ? 
		AND project_manager.position LIKE ? 
	;
`)

		// проверяем, что удалось сформировать запрос
		mlog.CheckErr(err, "Не могу подготовить запрос к БД activity")
		defer stmt.Close()
		fmt.Println("projectdatestart=", projectdatestart, "projectdateend=",
			projectdateend)
		// выполняем запрос
		rows, err := stmt.Query(studentid, studentid,
			projectname+"%",
			projectdatestart, projectdateend, // "min max"
			projectdatestart, projectdateend,
			projectdatestart, projectdateend,
			projectdatestart, projectdatestart,
			projectdateend, projectdateend,
			projectfio+"%",
			projectcity+"%",
			projectorganization+"%",
			projectcontacts+"%",
			projectposition+"%")
		// проверяем, что удалось выполнить запрос
		mlog.CheckErr(err, "Не могу выполнить запрос в БД activity")
		defer rows.Close()
		// объявляем переменные соответствующего типа для считывания полей
		var projectname2 sql.NullString
		var projectdate_start sql.NullString
		var projectdate_end sql.NullString
		var project_managerfio sql.NullString
		var cityname sql.NullString
		var organizationname sql.NullString
		var project_managercontacts sql.NullString
		var project_managerposition sql.NullString
		for rows.Next() {
			// считывание строки
			err = rows.Scan(&projectname2,
				&projectdate_start, &projectdate_end,
				&project_managerfio, &cityname, &organizationname,
				&project_managercontacts, &project_managerposition)
			// проверяем, что записи считались
			mlog.CheckErr(err, "Не могу прочесть запись")
			// переводим дату в соответствующий формат
			projectdate_start.String = mlog.DateToRus(projectdate_start.String)
			projectdate_end.String = mlog.DateToRus(projectdate_end.String)
			// формируем из считанных полей БД строку таблицы HTML
			sOut += "<tr><td>" + projectname2.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + projectdate_start.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + projectdate_end.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + project_managerfio.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + cityname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + organizationname.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + project_managercontacts.String + "</td><td>&nbsp;&nbsp;</td>" +
				"<td>" + project_managerposition.String + "</td></tr>\n"
		}
	} // POST
	// формируем заголовок таблицы с использованием Bootstrap
	sOut = "<table class=\"table table-striped\">\n" +
		" <thead>\n" +
		" <tr>\n" +
		" <th scope=\"col\">Название проекта</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Дата начала</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\"> Дата окончания </th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Руководитель</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Город</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Организация</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Контакт</th>\n" +
		" <th scope=\"col\">&nbsp;</th>\n" +
		" <th scope=\"col\">Должность</th>\n" +
		" </tr>\n" +
		" </thead>\n" +
		" <tbody>\n" +
		sOut +
		" </tbody>" +
		"</table>\n"
	fmt.Fprintf(w, "%v", sOut)
}
