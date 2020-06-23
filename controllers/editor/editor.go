package editor

import (
	"fmt"
	"log"
	controllers "models/controllers/middleware"
	"models/models"
	"models/properties"
	"net/http"
	"strings"

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
	emp := models.Code_Sessions{}
	res := []models.Code_Sessions{}
	for userDb.Next() {
		var email string
		err = userDb.Scan(&email)
		if err != nil {
			log.Fatal(err.Error())
		}
		emp.Email = email
		emp.Language = language
	}
	query := fmt.Sprintf(`select email from users where email not in ('%s');`, emp.Email)
	collabDb, err := db.Query(query)
	if err != nil {
		panic(err.Error())
	}
	fmt.Println("first dont work")
	emp.Collaborators = []string{}
	var otherEmails string
	for collabDb.Next() {
		err = collabDb.Scan(&otherEmails)
		if err != nil {
			panic(err.Error())
		}
		emp.Collaborators = append(emp.Collaborators, otherEmails)
	}
	fmt.Println(emp.Collaborators)
	fmt.Println(emp.Language)
	res = append(res, emp)
	fmt.Println(res)
	defer db.Close()

	controllers.Tmpl.ExecuteTemplate(w, "Editor", res)

}

var sessionID int
var sessionURL string

func InviteSession(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == "POST" {
		log.Println("helloooooo")
		r.ParseForm()
		sessionURL = r.PostFormValue("url")
		email := r.PostFormValue("email")
		log.Println(sessionURL)
		codeSessionDb, err := db.Prepare("Update code_sessions set url=? where id=?")
		if err != nil {
			panic(err)
		}
		fmt.Println(sessionURL)
		codeSessionDb.Exec(sessionURL, sessionID)
		sessionDb, err := db.Query("select uid from users where email=?", email)
		if err != nil {
			log.Fatal(err.Error())
		}
		var invitedUser string
		for sessionDb.Next() {
			err = sessionDb.Scan(&invitedUser)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
		fmt.Println(sessionID)
		fmt.Println(invitedUser)
		fmt.Println(email)
		inviteDb, err := db.Prepare("insert into user_code_sessions(user_id, code_session_id) values(?,?)")
		if err != nil {
			panic(err)
		}
		inviteDb.Exec(invitedUser, sessionID)
		log.Println("UPDATING")
	}
	http.Redirect(w, r, sessionURL, 301)

}

func DisplaySessions(w http.ResponseWriter, r *http.Request) {
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
	selDB, err := db.Query("select code_sessions.* from code_sessions inner join user_code_sessions on user_code_sessions.code_session_id=code_sessions.id where user_code_sessions.user_id=?", claims.Uid)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Code_Sessions{}
	res := []models.Code_Sessions{}
	for selDB.Next() {
		var id int
		var name, url string
		err = selDB.Scan(&id, &url, &name, &language)
		if err != nil {
			log.Fatal(err.Error())
		}
		emp.Id = id
		emp.Url = url
		emp.Name = name
		sessionID = id
		emp.Language = language
		emp.Collaborators = []string{}
		sessionDb, err := db.Query("select users.email from users inner join user_code_sessions on user_code_sessions.user_id=users.uid where user_code_sessions.code_session_id=?", emp.Id)
		if err != nil {
			log.Fatal(err.Error())
		}
		for sessionDb.Next() {
			var emails string
			err = sessionDb.Scan(&emails)
			if err != nil {
				log.Fatal(err.Error())
			}
			emp.Collaborators = append(emp.Collaborators, emails)
		}
		res = append(res, emp)
	}
	controllers.Tmpl.ExecuteTemplate(w, "Sessions", res)
	defer db.Close()
}
func NewSession(w http.ResponseWriter, r *http.Request) {
	controllers.Tmpl.ExecuteTemplate(w, "NewSession", nil)
}

var language string

func InsertSession(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == "POST" {
		r.ParseForm()
		name := r.PostFormValue("name")
		lang := r.PostFormValue("language")
		lang = strings.ToLower(lang)
		fmt.Println(name)
		fmt.Println(lang)
		url := "/editor?id=" + string(sessionID)
		insForm, err := db.Prepare("insert into code_sessions(url,name, language) values(?, ?, ?)")
		if err != nil {
			log.Fatal(err.Error())
		}
		insForm.Exec(url, name, lang)
		collabSession, err := db.Prepare("INSERT INTO user_code_sessions(user_id, code_session_id) values(?,LAST_INSERT_ID())")
		if err != nil {
			log.Fatal(err.Error())
		}
		collabSession.Exec(claims.Uid)
		language = lang
	}
	http.Redirect(w, r, "/sessions", 301)
}

func DeleteSession(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/", 301)
			return
		}
		http.Redirect(w, r, "/", 301)
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
	emp := r.URL.Query().Get("id")
	fmt.Println(emp)
	delSession, err := db.Prepare("delete from code_sessions where id=?")
	if err != nil {
		panic(err.Error())
	}
	delSession.Exec(emp)
	log.Println("DELTED SESSION")
	defer db.Close()
	http.Redirect(w, r, "/sessions", 301)
}
