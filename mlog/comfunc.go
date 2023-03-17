package mlog

import (
	"bufio"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// *********************************************************
func CheckErr(err error, mes string) {
	if err != nil {
		fmt.Println(mes)
		panic(err)
	}
}

// *********************************************************
func InfoMyConn(filename string) (string, string, string) {
	// ---------------------------------------------------------
	host, user, pass := "", "", ""
	file, err := os.Open(filename) // открываем файл на чтение
	// в случае ошибки печатаем сообщение и завершаем выполнение
	CheckErr(err, "Не могу открыть файл "+filename)
	defer file.Close()
	// создаем новую переменную типа Scanner
	scanner := bufio.NewScanner(file)
	// считываем последовательно текст
	for scanner.Scan() {
		// удаляем начальные и конечные пробельные символы
		s := strings.TrimSpace(scanner.Text())
		if s != "" {
			// в соответствии с указанным разделителем (двоеточие)
			// разбиваем фрагмент строки на подстроки (хост, пользователь, пароль)
			ar := strings.Split(s, ":")
			host, user, pass = ar[0], ar[1], ar[2]
			break
		}
	}
	err = scanner.Err()
	CheckErr(err, "Не могу считать строку в файле "+filename)
	// ---------------------------------------------------------
	return host, user, pass
}

// *********************************************************
func MyReadFile(filename string) string {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal(err)
	}
	// Конвертируем []byte в строку и возвращаем строку
	return string(content)
}
