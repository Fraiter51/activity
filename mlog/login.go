package mlog

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/sessions"
	"golang.org/x/crypto/bcrypt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

// encryptionKey должен быть сложным и уникальным, его используют для
// шифрования данных Cookie
var encryptionKey = "13OtdSecret"

// Сначала мы инициализируем хранилище сеансов, вызывая
// NewCookieStore()
// и передавая секретный ключ, используемый для аутентификации сеанса.
// Затем мы устанавливаем некоторые значения сеанса в session.Values,
// который является
// картой [interface{}] interface{}. И, наконец, мы вызываем session.Save(),
// чтобы сохранить данные сеанса.
var LoggedUserSession = sessions.NewCookieStore([]byte(encryptionKey))

// переменная для отслеживания перезагрузки сервера
// в случае перезагрузки пользователь должен заново войти в систему
var ExSignLogin = 0

// объявим переменную для создания объекта БД
var Exdbmysqlg *sql.DB

// Имя пользователя, прошедшего аутентификацию
var sGDisplayName = ""

// создадим структуру для хранения данных пользователя
type Page struct {
	Date        string
	Username    string
	Displayname string
}

// нам нужно отобразить страницу для ввода логина и пароля
// для этого храним необходимый код в файле login.html
// присваиваем переменной logUserPage его содержимое как строку
var logUserPage = MyReadFile("public/html/login.html")

// template.New()создает новый экземпляр HTML-шаблона, который будет
// выведен на экран
// Must - помощник, который завершает вызов функции, возвращающей
// (* Template, error), и вызывает panic, если ошибка не равна нулю,
// предназначен для использования при инициализации переменных
// Parse () анализирует текст
// таким образом в переменной logUserTemplate будет находится HTML
// шаблон для отображения страницы входа по паролю
var logUserTemplate = template.Must(template.New("").Parse(logUserPage))

/* ****************************************************** */
// сравнения введенных в окне регистрации данных пользователя и
// данных пользователя, хранящихся в БД для осуществления
// аутентификации
func LoginPageHandler(w http.ResponseWriter, r *http.Request) {
	// map[string]interface{} - карта, где ключ имеет тип string, а значение
	// тип interface{},
	// interface{} - это пустой интерфейс
	// последние "{}" это инициализация карты без значений
	// в карте conditionsMap хранятся данные о сессии
	conditionsMap := map[string]interface{}{}
	// проверяем, что сессия активна
	// в ассоциативный массив считывем имя пользователем
	session, _ := LoggedUserSession.Get(r, "my-user-session")
	if session != nil {
		conditionsMap["Username"] = session.Values["username"]
	}
	// проверяем имя и пароль
	if r.FormValue("Login") != "" && r.FormValue("Username") != "" {
		// считываем из формы логин, пароль
		username := r.FormValue("Username")
		password := r.FormValue("Password")
		// в переменную sUP считываем пароль из БД
		var sUP = ""
		sUP, sGDisplayName = GetPassDisplayName(fmt.Sprintf("%v", username))
		// представляем строку как байтовый массив
		hashedPasswordFromDatabase := []byte(sUP)
		// сравниваем пароли (из БД и введенный)
		// функция bcrypt.CompareHashAndPassword сравнивает пароли
		// в случае ошибки выдает сообщение "Either username or password is wrong"
		// и заносим в карту наличие ошибки: conditionsMap["LoginError"] = true
		// в случае верного ввода создается новая сессия для пользователя
		// в карту заносим имя пользователя и conditionsMap["LoginError"] = false
		if err := bcrypt.CompareHashAndPassword(hashedPasswordFromDatabase,
			[]byte(password)); err != nil {
			// сообщение в лог, если пароли не совпали
			log.Println("Either username or password is wrong")
			conditionsMap["LoginError"] = true
		} else {
			// если пароли соврали, создаем новую сессию с именем пользователя
			//Username
			// заносим сообщение в лог и данные в карту conditionsMap
			log.Println("Logged in :", username)
			conditionsMap["Username"] = username
			conditionsMap["LoginError"] = false
			session, _ := LoggedUserSession.New(r, "my-user-session")
			session.Values["username"] = username
			// ---------------------------------------------------------
			// признак того, что вошли в сессию
			ExSignLogin = 1
			// ---------------------------------------------------------
			err := session.Save(r, w)
			if err != nil {
				log.Println(err)
			}
			// если пароль и логин совпали и сессия создана, перейдем
			// по маршруту /index
			http.Redirect(w, r, "/index", http.StatusFound)
		}
	}
	// если не удалось перейти по маршруту /index, переходим на страницу
	//автризации
	if err := logUserTemplate.Execute(w, conditionsMap); err != nil {
		log.Println(err)
	}
}

/* ****************************************************** */
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// считываение параметров сессии
	session, _ := LoggedUserSession.Get(r, "my-user-session")
	// присваиваем пустое значение переменной session.Values["username"]
	session.Values["username"] = ""
	// если значение ошибки не nil, выводится сообщение
	err := session.Save(r, w)
	if err != nil {
		log.Println(err)
	}
	// возврат в исходное окно регистрации
	http.Redirect(w, r, "/", http.StatusFound)
}

/* ****************************************************** */
func GetPassDisplayName(un string) (string, string) {
	// ---------------------------------------------------------
	// запрос на поиск в БД по логину зашифрованного пароля и
	// ФИО пользователя ИС для отображения на экране
	// вызывается из функции LoginPageHandler
	stmt, err := Exdbmysqlg.Prepare("SELECT user_pass, user_display_name " +
		"FROM users " +
		"WHERE user_name=?;")
	CheckErr(err, "Не могу подготовить запрос к БД activity")
	defer stmt.Close()
	res, err := stmt.Query(un)
	CheckErr(err, "Не могу выполнить запрос в БД activity")
	defer res.Close()
	var user_pass sql.NullString
	var user_displayname sql.NullString
	for res.Next() {
		err = res.Scan(&user_pass, &user_displayname)
		CheckErr(err, "Не могу прочесть запись")
		break
	}
	return user_pass.String, user_displayname.String
}

// *********************************************************
func CheckLoginGET(w http.ResponseWriter, r *http.Request) {
	// ---------------------------------------------------------
	// проверка входа в сессию
	session, err := LoggedUserSession.Get(r, "my-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}
	if session.Values["username"] == "" || ExSignLogin == 0 {
		// если логин пользователя пустой, переходим на начальную старницу
		// это не позволяет использовать путь в адресной строке браузера
		http.Redirect(w, r, "/logout", http.StatusFound)
	}
	// ---------------------------------------------------------
}

// *********************************************************
func CheckLoginPOST(w http.ResponseWriter, r *http.Request) int {
	// ---------------------------------------------------------
	session, err := LoggedUserSession.Get(r, "my-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}
	if session.Values["username"] == "" || ExSignLogin == 0 {
		return 0
	}
	// возвращаем в вызывающую функцию 0 или 1
	// в случае 0 переходим на стороне клиента в окно входа
	// в случае 1 продолжаем выполнение вызывающей функции
	return 1
	// ---------------------------------------------------------
}

/* ****************************************************** */
func Index(w http.ResponseWriter, r *http.Request) {
	// ---------------------------------------------------------
	conditionsMap := map[string]interface{}{}
	// проверка, что для данного пользователя сессия активизирована,
	// то есть пользователь вошел со своим логином и паролем
	session, err := LoggedUserSession.Get(r, "my-user-session")
	if err != nil {
		log.Println("Unable to retrieve session data!", err)
	}
	// печать лог имени сессии и пр.
	log.Println("Session name : ", session.Name())
	log.Println("Username : ", session.Values["username"])
	// ---------------------------------------------------------
	if session.Values["username"] == "" || sGDisplayName == "" {
		http.Redirect(w, r, "/logout", http.StatusFound)
	}
	// ---------------------------------------------------------
	// в карту conditionsMap звносится значение имени пользователя
	conditionsMap["Username"] = session.Values["username"]
	// если вызывается неправильный путь, возвращаем ошибку
	if r.URL.Path != "/index" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}
	// формируем дату и имя пользователя для отображения на экране
	year, month, day := time.Now().Date()
	curdate := fmt.Sprintf("%02d.%02d.%d", day, month, year)
	username := fmt.Sprintf("%v", session.Values["username"])
	p := &Page{
		Date:        curdate,
		Username:    username,
		Displayname: sGDisplayName,
	}
	// отображаем главную страницу
	t := template.Must(template.ParseFiles("public/html/index.html"))
	t.Execute(w, p)
}

/* ****************************************************** */
// Функция предназначена для перевода даты формата 2021-11-25
// в формат 25.11.2021
func DateToRus(date string) string {
	date = strings.TrimSpace(date)
	if date == "" {
		return date
	}
	ar := strings.Split(date, "-")
	year, month, day := "", "", ""
	year, month, day = ar[0], ar[1], ar[2]
	sAux := day + "." + month + "." + year
	return sAux
}
