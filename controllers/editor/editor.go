package editor

import (
	"log"
	controllers "models/controllers/middleware"
	"models/models"
	"models/properties"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = properties.JwtKey()

func CodeEditor(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	tknStr := c.Value
	claims := &models.Claims{}

	tkn, err := jwt.ParseWithClaims(tknStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	db := controllers.DbConn()
	userDb, err := db.Query("select email from users where uid=?", claims.Uid)
	if err != nil {
		log.Fatal(err.Error())
	}
	emp := models.Users{}
	res := []models.Users{}
	for userDb.Next() {
		var email string
		err = userDb.Scan(&email)
		if err != nil {
			log.Fatal(err.Error())
		}
		emp.Email = email
		res = append(res, emp)
	}
	controllers.Tmpl.ExecuteTemplate(w, "Editor", res)
}
