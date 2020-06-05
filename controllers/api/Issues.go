package api

import (
	"fmt"
	"log"
	controllers "models/controllers/middleware"
	"models/models"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
)

func DisplayIssues(w http.ResponseWriter, r *http.Request) {
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
	selDB, err := db.Query("SELECT * FROM Issues WHERE user_id=? ORDER BY id DESC", claims.Uid)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Issues{}
	res := []models.Issues{}
	for selDB.Next() {
		var id, user_id int
		var name, description, priority, date string
		err = selDB.Scan(&id, &name, &description, &priority, &date, &controllers.Project_id, &user_id)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.User_id = user_id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
		emp.Date = date
		res = append(res, emp)
	}
	//fmt.Print("display users id is ")
	//fmt.Println(uid)
	controllers.Tmpl.ExecuteTemplate(w, "Issues", res)
	defer db.Close()

}

//Index routes to index template, returns all available issues
func Index(w http.ResponseWriter, r *http.Request) {
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
	selDB, err := db.Query("SELECT * FROM Issues WHERE project_id=? and user_id=? ORDER BY id DESC", controllers.Project_id, claims.Uid)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Issues{}
	res := []models.Issues{}
	for selDB.Next() {
		var id, project_id, user_id int
		var name, description, priority, date string
		err = selDB.Scan(&id, &name, &description, &priority, &date, &project_id, &user_id)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Project_id = project_id
		emp.User_id = user_id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
		emp.Date = date
		res = append(res, emp)
	}
	//fmt.Print("user id is ")
	//fmt.Println(uid)
	controllers.Tmpl.ExecuteTemplate(w, "Index", res)
	defer db.Close()

}

// Show is a function that routes to View template
func Show(w http.ResponseWriter, r *http.Request) {
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
	nId := r.URL.Query().Get("id")
	fmt.Println(nId)
	selDB, err := db.Query("SELECT * FROM Issues WHERE id=? and project_id=? and user_id=?", nId, controllers.Project_id, claims.Uid)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Issues{}
	for selDB.Next() {
		var id, project_id, user_id int
		var name, description, priority, date string
		err = selDB.Scan(&id, &name, &description, &priority, &date, &project_id, &user_id)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Project_id = project_id
		emp.User_id = user_id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
		emp.Date = date
	}
	controllers.Tmpl.ExecuteTemplate(w, "Show", emp)
	defer db.Close()

}

// New is a router to create a new Blip
func New(w http.ResponseWriter, r *http.Request) {
	controllers.Tmpl.ExecuteTemplate(w, "New", nil)
}

// Edit is a route to UPDATE an existing Blip
func Edit(w http.ResponseWriter, r *http.Request) {
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
	nId := r.URL.Query().Get("id")
	fmt.Println(r.Method)
	selDB, err := db.Query("SELECT * FROM Issues WHERE id=? and project_id=?", nId, controllers.Project_id)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Issues{}
	for selDB.Next() {
		var id, project_id int
		var name, description, priority, date string
		err = selDB.Scan(&id, &name, &description, &priority, &date, &project_id, &claims.Uid)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Project_id = project_id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
		emp.Date = date
	}
	controllers.Tmpl.ExecuteTemplate(w, "Edit", emp)
	defer db.Close()

}

// Insert is a router function that creates the new Blip
func Insert(w http.ResponseWriter, r *http.Request) {
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
		name, description, priority := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("priority")
		date := time.Now().Format("01-02-2006")
		insForm, err := db.Prepare("INSERT INTO Issues(name, description, priority, date, project_id, user_id) VALUES(?,?,?,?,?, ?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, description, priority, date, controllers.Project_id, claims.Uid)
		log.Println("INSERT: Name: " + name + " | Description: " + description + " | Priority: " + priority + " | Date: " + date)
	}
	defer db.Close()
	http.Redirect(w, r, "/index", 301)

}

// Update is the router function that updates an existing Blip in the database
func Update(w http.ResponseWriter, r *http.Request) {
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
		name, description, priority := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("priority")
		id := r.FormValue("project_id")
		date := time.Now().Format("01-02-2006")
		insForm, err := db.Prepare("UPDATE Issues SET name=?, description=?, priority=?, date=? WHERE id=?")
		if err != nil {
			fmt.Println(err.Error())
		}
		insForm.Exec(name, description, priority, date, id)
		log.Println("UPDATE: Name: " + name + " | Description: " + description + " | Priority: " + priority + " | Date: " + date)
	}
	defer db.Close()
	http.Redirect(w, r, "/index", 301)

}

// Delete is a function that removes the Blip from the database
func Delete(w http.ResponseWriter, r *http.Request) {
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
	emp := r.URL.Query().Get("id")
	fmt.Println(r.Method)
	delForm, err := db.Prepare("DELETE FROM Issues WHERE id=?")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(emp)
	log.Println("DELETE")
	defer db.Close()
	http.Redirect(w, r, "/index", 301)
}
