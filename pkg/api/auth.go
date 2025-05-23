package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

// AnswerAuth структура для формирования ответа на запрос
type AnswerAuth struct {
	Token string `json:"token,omitempty"`
	Err   string `json:"error,omitempty"`
}

// Password структура
type Password struct {
	InputPass string `json:"password"`
}

// authHandler служит для установки пароля
func authHandler(w http.ResponseWriter, r *http.Request) {

	var ans AnswerAuth
	var buf bytes.Buffer
	var input Password

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		ans.Err = fmt.Sprintf("невозможно прочитать тело запроса %v", err.Error())
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &input)
	if err != nil {
		ans.Err = fmt.Sprintf("невозможно десериализовать тело запроса с паролем %v", err.Error())
		WriterJSON(w, http.StatusBadRequest, ans)
		return
	}

	if len(input.InputPass) > 0 {

		err = os.Setenv("TODO_PASSWORD", input.InputPass)
		if err != nil {
			ans.Err = fmt.Sprintf("ошибка Setenv: %v", err.Error())
			WriterJSON(w, http.StatusBadRequest, ans)
			return
		}
	}

	hash := sha256.Sum256([]byte(input.InputPass))
	hashString := hex.EncodeToString(hash[:])
	ans.Token = hashString

	WriterJSON(w, http.StatusOK, ans)

}

// auth проверяет соответствие пароля хэшу в cookie
func auth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		pass := os.Getenv("TODO_PASSWORD")

		if len(pass) > 0 {

			var jwt string
			cookie, err := r.Cookie("token")
			if err == nil {
				jwt = cookie.Value
			}

			hashedPassword := sha256.Sum256([]byte(pass))
			hashStringPassword := hex.EncodeToString(hashedPassword[:])

			if hashStringPassword != jwt {
				http.Error(w, "Authentification required", http.StatusUnauthorized)
				return
			}
		}
		next(w, r)
	})
}
