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
	projects, err := db.Query("select projects.id from projects inner join users_projects on users_projects.project_id=projects.id where users_projects.user_id=?", claims.Uid)
	if err != nil {
		panic(err.Error())
	}
	allProjects := []int{}
	var projectId int
	for projects.Next() {
		err = projects.Scan(&projectId)
		if err != nil {
			panic(err.Error())
		}
		allProjects = append(allProjects, projectId)
	}
	emp := models.Issues{}
	res := []models.Issues{}
	for i := 0; i < len(allProjects); i++ {
		selDB, err := db.Query("SELECT * FROM Issues WHERE project_id=? ORDER BY id DESC", allProjects[i])
		if err != nil {
			panic(err.Error())
		}
		for selDB.Next() {
			var id, user_id int
			var name, description, priority, date, kind string
			var projectName string
			err = selDB.Scan(&id, &name, &description, &priority, &kind, &controllers.Project_id, &user_id, &date)
			if err != nil {
				panic(err.Error())
			}
			emp.Id = id
			emp.User_id = user_id
			emp.Name = name
			emp.Description = description
			emp.Priority = priority
			emp.Date = date
			emp.Kind = kind
			emp.Project_id = controllers.Project_id

			projDb, err := db.Query("Select name from projects where id=?", emp.Project_id)
			if err != nil {
				panic(err.Error())
			}
			for projDb.Next() {
				err = projDb.Scan(&projectName)
				if err != nil {
					panic(err.Error())
				}
				emp.ProjectName = projectName
			}
			res = append(res, emp)
		}
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
	projDB, err := db.Query("Select name from projects where id=?", controllers.Project_id)
	if err != nil {
		panic(err.Error())
	}
	var projectName string
	for projDB.Next() {
		err = projDB.Scan(&projectName)
		if err != nil {
			panic(err.Error())
		}
	}

	selDB, err := db.Query("SELECT * FROM Issues WHERE project_id=? ORDER BY id DESC", controllers.Project_id)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Issues{}
	res := []models.Issues{}
	for selDB.Next() {
		var id, project_id, user_id int
		var name, description, priority, date, kind string
		err = selDB.Scan(&id, &name, &description, &priority, &kind, &project_id, &user_id, &date)
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
		emp.Kind = kind
		emp.Users = []string{}
		userDb, err := db.Query("select email from users where uid=?", emp.User_id)
		if err != nil {
			log.Fatal(err.Error())
		}
		for userDb.Next() {
			var emails string
			err = userDb.Scan(&emails)
			if err != nil {
				log.Fatal(err.Error())
			}
			emp.Users = append(emp.Users, emails)
		}
		res = append(res, emp)
	}
	fmt.Println(projectName)

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
	selDB, err := db.Query("SELECT * FROM Issues WHERE id=? and project_id=?", nId, controllers.Project_id)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Issues{}
	for selDB.Next() {
		var id, project_id, user_id int
		var name, description, priority, date, kind string
		err = selDB.Scan(&id, &name, &description, &priority, &kind, &project_id, &user_id, &date)
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
		emp.Kind = kind
	}
	controllers.Tmpl.ExecuteTemplate(w, "Show", emp)
	defer db.Close()

}

// New is a router to create a new Blip
func New(w http.ResponseWriter, r *http.Request) {
	controllers.Tmpl.ExecuteTemplate(w, "New", nil)
}

var issueId string

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
	fmt.Println(nId)
	issueId = nId
	fmt.Println(r.Method)
	selDB, err := db.Query("SELECT * FROM Issues WHERE id=? and project_id=?", nId, controllers.Project_id)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Issues{}
	for selDB.Next() {
		var id, project_id int
		var name, description, priority, date, kind string
		err = selDB.Scan(&id, &name, &description, &priority, &kind, &project_id, &claims.Uid, &date)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Project_id = project_id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
		emp.Date = date
		emp.Kind = kind
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
		name, description, priority, kind := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("priority"), r.PostFormValue("kind")
		date := time.Now().Format("01-02-2006")
		insForm, err := db.Prepare("INSERT INTO Issues(name, description, priority, kind , project_id, user_id, date ) VALUES(?,?,?,?,?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, description, priority, kind, controllers.Project_id, claims.Uid, date)
		log.Println("INSERT: Name: " + name + " | Description: " + description + " | Priority: " + priority + " | Date: " + date + " | Kind: " + kind)
	}
	defer db.Close()
	http.Redirect(w, r, "/project/tickets", 301)

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
		name, description, priority, kind := r.FormValue("name"), r.FormValue("description"), r.FormValue("priority"), r.FormValue("kind")
		id := issueId
		fmt.Println(id)
		date := time.Now().Format("01-02-2006")
		insForm, err := db.Prepare("UPDATE Issues SET name=?, description=?, priority=?,kind=? , project_id=?, user_id=?, date=? WHERE id=?")
		if err != nil {
			fmt.Println(err.Error())
		}
		insForm.Exec(name, description, priority, kind, controllers.Project_id, claims.Uid, date, id)
		log.Println("UPDATE: Name: " + name + " | Description: " + description + " | Priority: " + priority + " | Date: " + date + " | Kind: " + kind)
	}
	defer db.Close()
	http.Redirect(w, r, "/project/tickets", 301)

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
	http.Redirect(w, r, "/project/tickets", 301)
}
