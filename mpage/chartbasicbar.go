package mpage

import (
	MODMLOG "MyProjectLab_8/mlog"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"strings"
	"sync"
)

// создадим карту (ассоциативный массив) для хранения числа баллов,
// полученных каждым студентом за участие в том или ином научном
// мероприятии
var arDC = make(map[int]map[string]int64)

/* ****************************************************** */
// функция построения графиков
func ChartBasicBar(w http.ResponseWriter, r *http.Request) {
	// ---------------------------------------------------------
	arDC[0] = make(map[string]int64)
	arDC[1] = make(map[string]int64)
	arDC[2] = make(map[string]int64)
	arDC[3] = make(map[string]int64)
	// ---------------------------------------------------------
	// объявляем переменную для id студента
	studentid := ""
	// объявляем переменную для выбора типа графика
	activitytype := ""
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
		if MODMLOG.CheckLoginPOST(w, r) == 0 {
			fmt.Fprintf(w, "%v", "0####/")
			return
		}
		// ---------------------------------------------------------
		if err := r.ParseMultipartForm(64 << 20); err != nil {
			fmt.Println("ParseForm() err: ", err)
			fmt.Fprintf(w, "ParseForm() err: %v", err)
			MODMLOG.CheckErr(err, "Ошибка запроса POST: ChartBasicBar")
		}
		// ---------------------------------------------------------
		// если id студента не задан, осуществляем поиск для всех студентов
		studentid = r.FormValue("studentid")
		studentid = strings.TrimSpace(studentid)
		if studentid == "" {
			studentid = "%"
		}
		// из входного запроса получаем тип графика
		activitytype = r.FormValue("activitytype")
		activitytype = strings.TrimSpace(activitytype)
	}
	// ---------------------------------------------------------
	switch activitytype {
	// в зависимости от значения переменной activitytype
	// формируем возвращаемое значение для построения графика
	case "1":
		BasicBarJSON(studentid, 1)
		pagesJson, err := json.Marshal(arDC[1])
		MODMLOG.CheckErr(err, "Cannot encode to JSON ")
		fmt.Fprintf(w, "%s", pagesJson)
	case "2":
		BasicBarJSON(studentid, 2)
		pagesJson, err := json.Marshal(arDC[2])
		MODMLOG.CheckErr(err, "Cannot encode to JSON ")
		fmt.Fprintf(w, "%s", pagesJson)
	case "3":
		BasicBarJSON(studentid, 3)
		pagesJson, err := json.Marshal(arDC[3])
		MODMLOG.CheckErr(err, "Cannot encode to JSON ")
		fmt.Fprintf(w, "%s", pagesJson)
	default:
		// переменная, в которой хранится число вызванных горутин
		var wg sync.WaitGroup
		for i := 1; i <= 3; i++ {
			// к счетчику горутин добавляется единица (Add), создается горутина
			wg.Add(1)
			go func(i int) {
				BasicBarJSON(studentid, i)
				// после выполнения горутины счетчик уменьшается на 1 (Done)
				wg.Done()
			}(i) // передача i как явного аргумента
		}
		// ожидаем выполнения всх горутин
		wg.Wait()
		// поскольку нам нужны итоговые баллы, суммируем все значения баллов
		// за различные мероприятия
		for k := range arDC[1] {
			arDC[0][k] = arDC[1][k] +
				arDC[2][k] + arDC[3][k]
		}
		pagesJson, err := json.Marshal(arDC[0])
		MODMLOG.CheckErr(err, "Cannot encode to JSON ")
		fmt.Fprintf(w, "%s", pagesJson)
	}
	// ---------------------------------------------------------
}

/* ****************************************************** */
// в зависимости от значения переменной activitytype выполняем SQL
//запрос
func BasicBarJSON(studentid string, activitytype int) {
	// ---------------------------------------------------------
	// подготовка запроса в зависимости от значения переменной activitytype
	sSQLPoint := ""
	switch activitytype {
	case 1:
		sSQLPoint = "IFNULL((SELECT SUM(student_conference.point) FROM student_conference " +
			"LEFT JOIN conference ON student_conference.conference_id=conference.id " +
			"WHERE student_conference.student_id=student.id " +
			"),0) "
	case 2:
		sSQLPoint = "IFNULL((SELECT SUM(IFNULL(student_project.point,0)) FROM student_project " +
			"LEFT JOIN project ON student_project.project_id=project.id " +
			"WHERE student_project.student_id=student.id " +
			"),0) "
	case 3:
		sSQLPoint = "IFNULL((SELECT SUM(IFNULL(student_paper.point,0)) FROM student_paper " +
			"LEFT JOIN paper ON student_paper.paper_id=paper.id " +
			"WHERE student_paper.student_id=student.id " +
			"),0) "
	}
	// ---------------------------------------------------------
	// подготовка запроса для выполнения
	stmt, err := Exdbmysqlg.Prepare("SELECT student.fio, " +
		sSQLPoint + "as std_point " +
		"FROM student " +
		"WHERE student.id LIKE ? " +
		"ORDER BY fio " +
		";")
	MODMLOG.CheckErr(err, "Не могу подготовить запрос к БД activity")
	defer stmt.Close()
	// выполняем запрос
	rows, err := stmt.Query(studentid)
	MODMLOG.CheckErr(err, "Не могу выполнить запрос в БД activity")
	defer rows.Close()
	// ---------------------------------------------------------
	var fio sql.NullString
	var std_point sql.NullInt64
	for rows.Next() {
		err := rows.Scan(&fio, &std_point)
		MODMLOG.CheckErr(err, "Не могу прочесть запись")
		arDC[activitytype][fio.String] = std_point.Int64
	}
}
