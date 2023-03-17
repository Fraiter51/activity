package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// определим массив [] типа byte
	// общий вид: passwd := []byte("password")
	// в нашем случае пусть паролем будет слово "",
	// Внимание! В случае использования кириллицы
	// возможно некорректное срабатывание хеш-функции
	passwd := []byte("parol")
	// функция GenerateFromPassword возвращает хеш пароля в формате
	// bcrypt, в случае использования bcrypt.MinCost число итераций равно 4
	// можно число итераций задать явно, вместо bcrypt.MinCost написать
	// целое число из диапазона от 4 до 31, например, 6
	hashedPassword, err := bcrypt.GenerateFromPassword(passwd,
		bcrypt.MinCost)
	if err != nil {
		panic(err)
	}
	// печать строки, которая генерируется хеш-функцией
	fmt.Printf("The hashed password is : %s\n", string(hashedPassword))
	// Сравнение пароля с хешем
	// функция bcrypt.CompareHashAndPassword принимает на вход
	// хешированный пароль и сам пароль и сравнивает их
	// возвращает ноль в случае успеха или ошибку в случае неудачи
	hashedPassword = []byte("$2a$04$ACbHpRWOUckvoZz/4/jFhOOpqvpYdZKBGvjadr0y6uG10IEeeQmde")
	passwd = []byte("parol")
	err = bcrypt.CompareHashAndPassword(hashedPassword, passwd)
	// если напечатано "nil", пароль совпадает с хешем из БД
	fmt.Println(err)
}
