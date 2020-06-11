package api

import (
	"fmt"
	"io/ioutil"
	"log"
	controllers "models/controllers/middleware"
	"models/models"
	"models/properties"
	"net/http"

	jparse "github.com/nishgowda/Jparse"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = properties.JwtKey()

func DisplayProjects(w http.ResponseWriter, r *http.Request) {
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
	selDB, err := db.Query("select projects.* from projects inner join users_projects on users_projects.project_id=projects.id where users_projects.user_id=?", claims.Uid)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Projects{}
	res := []models.Projects{}

	for selDB.Next() {
		var id int
		var name, description, technologies string
		err = selDB.Scan(&id, &name, &description, &technologies)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.ProjectName = name
		emp.Description = description
		emp.Technologies = technologies
		emp.Users = []string{}
		projDb, err := db.Query("select users.email from users inner join users_projects on users_projects.user_id=users.uid where users_projects.project_id=?", emp.Id)
		if err != nil {
			log.Fatal(err.Error())
		}
		for projDb.Next() {
			var userEmails string
			err = projDb.Scan(&userEmails)
			if err != nil {
				log.Fatal(err.Error())
			}
			emp.Users = append(emp.Users, userEmails)
		}
		res = append(res, emp)

	}
	//fmt.Println(uid)
	fmt.Println(res)

	controllers.Tmpl.ExecuteTemplate(w, "DisplayProjects", res)
	defer db.Close()
}

func ImportRepos(w http.ResponseWriter, r *http.Request) {
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

	res, err := http.Get(controllers.JsonURL)
	if err != nil {
		log.Fatal(err)
	}

	body, err := ioutil.ReadAll(res.Body)
	js := string(body)
	names := jparse.SimpleArrayParse([]string{"name"}, js)
	description := jparse.SimpleArrayParse([]string{"description"}, js)
	languages := jparse.SimpleArrayParse([]string{"language"}, js)

	db := controllers.DbConn()
	for i := 0; i < len(names); i++ {
		nameDb, err := db.Prepare("insert ignore into projects(name, description, user_id, technologies) VALUES(?,?,?,?)")
		if err != nil {
			log.Fatal(err)
		}
		nameDb.Exec(names[i], description[i], claims.Uid, languages[i])
		log.Println("INSERT: Name: " + names[i] + " | Description: " + description[i] + " | Technologies: " + languages[i])
	}

	defer db.Close()
	http.Redirect(w, r, "/dashboard", 301)

}

func NewProject(w http.ResponseWriter, r *http.Request) {
	controllers.Tmpl.ExecuteTemplate(w, "NewProject", nil)
}

func InsertProject(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == "POST" {
		r.ParseForm()
		name, description, technologies := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("technologies")
		insForm, err := db.Prepare("INSERT INTO Projects(name, description, technologies) VALUES(?, ?, ?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, description, technologies)
		var project_id int
		fmt.Println(project_id)
		fmt.Println(claims.Uid)
		log.Println("INSERT: Name: " + name + " | Description: " + description + " | Technologies: " + technologies)
		teamForm, err := db.Prepare("INSERT INTO users_projects(user_id, project_id) values(?,LAST_INSERT_ID())")
		if err != nil {
			panic(err.Error())
		}
		teamForm.Exec(claims.Uid)

	}
	defer db.Close()
	http.Redirect(w, r, "/displayprojects", 301)

}

func ShowProject(w http.ResponseWriter, r *http.Request) {
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
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM projects WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Projects{}
	for selDB.Next() {
		var id int
		var name, description, technologies string
		err = selDB.Scan(&id, &name, &description, &technologies)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.ProjectName = name
		emp.Description = description
		emp.Technologies = technologies
		controllers.Project_id = emp.Id
		fmt.Println(controllers.Project_id)
		emp.Users = []string{}
		projDb, err := db.Query("select users.email from users inner join users_projects on users_projects.user_id=users.uid where users_projects.project_id=?", emp.Id)
		if err != nil {
			log.Fatal(err.Error())
		}
		for projDb.Next() {
			var userEmails string
			err = projDb.Scan(&userEmails)
			if err != nil {
				log.Fatal(err.Error())
			}
			emp.Users = append(emp.Users, userEmails)
		}
	}
	controllers.Tmpl.ExecuteTemplate(w, "ShowProject", emp)
	defer db.Close()

}

func EditProject(w http.ResponseWriter, r *http.Request) {
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
	nId := r.URL.Query().Get("id")
	fmt.Println(r.Method)
	selDB, err := db.Query("SELECT * FROM Projects WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}
	emp := models.Projects{}
	for selDB.Next() {
		var id int
		var name, description, technologies string
		err = selDB.Scan(&id, &name, &description, &technologies)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		controllers.Project_id = emp.Id
		emp.ProjectName = name
		emp.Description = description
		emp.Technologies = technologies
	}
	controllers.Tmpl.ExecuteTemplate(w, "EditProject", emp)
	defer db.Close()

}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == "POST" {
		r.ParseForm()
		name, description, technologies := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("technologies")
		fmt.Println(controllers.Project_id)
		insForm, err := db.Prepare("UPDATE Projects SET name=?, description=?, technologies=? WHERE id=?")
		if err != nil {
			fmt.Println(err.Error())
		}
		insForm.Exec(name, description, technologies, controllers.Project_id)
		log.Println("UPDATE: Name: " + name + " | Description: " + description + " | Technologies: " + technologies)
	}
	defer db.Close()
	http.Redirect(w, r, "/displayprojects", 301)

}
func Invite(w http.ResponseWriter, r *http.Request) {
	controllers.Tmpl.ExecuteTemplate(w, "Invite", nil)
}

func InviteUser(w http.ResponseWriter, r *http.Request) {
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
	if r.Method == "POST" {
		r.ParseForm()
		email := r.FormValue("email")
		validate := properties.ValidateEmail(email)
		if validate == false {
			fmt.Println("bad email")
		} else {
			selDb, err := db.Query("select uid from users where email=?", email)
			if err != nil {
				log.Fatal(err.Error())
			}

			var invitedUser string
			for selDb.Next() {
				err = selDb.Scan(&invitedUser)
				if err != nil {
					log.Fatal(err.Error())
				}
			}
			inviteDb, err := db.Prepare("insert ignore into users_projects(user_id, project_id) values(?,?)")
			if err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println(invitedUser)
			fmt.Println(controllers.Project_id)
			inviteDb.Exec(invitedUser, controllers.Project_id)
			fmt.Println("Inserted")
		}
		defer db.Close()
		http.Redirect(w, r, "/displayprojects", 301)
	}
}

func DeleteProject(w http.ResponseWriter, r *http.Request) {
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
	fmt.Println(r.Method)
	fmt.Println(emp)
	delForm, err := db.Prepare("DELETE FROM Projects WHERE id=?")
	if err != nil {
		panic(err.Error())
	}
	delForm.Exec(emp)
	log.Println("DELETE")
	defer db.Close()
	http.Redirect(w, r, "/displayprojects", 301)

}
