package utils

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

const TokenEXP = time.Hour * 3
const SecretKEY = "*Hello-World*"

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

func BuildJWTString(userID int) (string, error) {
	// создаём новый токен с алгоритмом подписи HS256 и утверждениями — Claims

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			// когда создан токен
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(TokenEXP)),
		},
		// собственное утверждение
		UserID: userID,
	})

	// создаём строку токена
	tokenString, err := token.SignedString([]byte(SecretKEY))
	if err != nil {
		return "", err
	}

	// возвращаем строку токена
	return tokenString, nil
}

func GetUserID(tokenString string) int {
	// создаём экземпляр структуры с утверждениями
	claims := &Claims{}
	// парсим из строки токена tokenString в структуру claims
	jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(SecretKEY), nil
	})

	// возвращаем ID пользователя в читаемом виде
	return claims.UserID
}

func CheckOrderByLuna(orderNum string) bool {
	sum := 0
	//var digits = make([]int, len(orderNum))
	//for i := 0; i < len(orderNum); i++ {
	pos := 0
	for i := len(orderNum) - 1; i >= 0; i-- {
		digit, _ := strconv.Atoi(string(orderNum[i]))
		if pos%2 != 0 {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		pos++
		// digits[i] = digit
		sum += digit
	}
	//fmt.Println("tt", digits, sum, int(sum%10) == 0)
	return int(sum%10) == 0
}

func SillyGenerateOrderNumberLuhna(length int) string {
	order := ""
	for {
		order += fmt.Sprintf("%d", rand.Intn(9))
		if len(order) == length {
			if CheckOrderByLuna(order) {
				return order
			} else {
				order = ""
			}
		}
	}
}
