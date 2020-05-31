package controllers

import (
	"database/sql"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"models/models"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
)

func ObtainDatabaseName() string {

	file, er := os.Open("../secretDbName.txt")
	if er != nil {
		log.Fatal(er)
	}
	defer file.Close()
	b, er := ioutil.ReadAll(file)
	dName := string(b)
	return dName
}

//ObtainDatabasePassword obtains database credentials for the pword of the database
func ObtainDatabasePassword() string {
	files, ers := os.Open("../secretDbPass.txt")
	if ers != nil {
		log.Fatal(ers)
	}
	defer files.Close()
	body, ers := ioutil.ReadAll(files)
	dPass := string(body)
	return dPass
}
func dbConn() (db *sql.DB) {

	dbDriver := "mysql"
	dbUser := "root"
	dbPass := ObtainDatabasePassword()
	dbName := ObtainDatabaseName()
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

var tmpl = template.Must(template.ParseGlob("../views/*"))
var uid int
var project_id int
var singedIn = false

func Home(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Login", nil)
}
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
func Login(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		r.ParseForm()
		username, password := r.PostFormValue("username"), r.PostFormValue("password")
		ogPassword := password
		selDb, err := db.Query("SELECT uid, username, password FROM USERS WHERE username=?", username)
		if err != nil {
			http.Redirect(w, r, "/", 301)
		}
		emp := models.Users{}
		res := []models.Users{}
		for selDb.Next() {
			err = selDb.Scan(&uid, &username, &password)
			if err != nil {
				panic(err.Error())
			}
			emp.Uid = uid
			emp.Password = password
			emp.Username = username
			if uid != 0 {
				if CheckPasswordHash(ogPassword, emp.Password) == true {
					res = append(res, emp)
					//fmt.Println(emp.Password)
					fmt.Println("succesfully logged in as " + username)
					singedIn = true
				} else {
					http.Redirect(w, r, "/", 301) // ---> Figure out a work around for this superfluous response.WriteHeader call from main.Login (main.go:129)
				}
			}

		}

	}
	defer db.Close()
	http.Redirect(w, r, "/dashboard", 301)
}

func LogoutPage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Logout", nil)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	singedIn = false
	http.Redirect(w, r, "/", 301)
}

func SingUpPage(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Register", nil)
}
func Register(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	fmt.Println("Working?")
	if r.Method == "POST" {
		r.ParseForm()
		username, password, email := r.PostFormValue("username"), r.PostFormValue("password"), r.PostFormValue("email")
		insForm, err := db.Prepare("INSERT INTO Users(username, password, email) VALUES(?,?, ?)")
		if err != nil {
			fmt.Println(err.Error)
		}
		hash, _ := HashPassword(password)
		insForm.Exec(username, hash, email)
		log.Println("INSERT: Username: " + username + " | Password: " + string(hash))
	}

	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		selDB, err := db.Query("SELECT * FROM Projects WHERE user_id=? ORDER BY id DESC", uid)
		if err != nil {
			panic(err.Error())
		}
		emp := models.Projects{}
		res := []models.Projects{}
		for selDB.Next() {
			var id, user_id int
			var name, description string
			err = selDB.Scan(&id, &name, &description, &user_id)
			if err != nil {
				panic(err.Error())
			}
			emp.Id = id
			emp.UserId = user_id
			emp.ProjectName = name
			emp.Description = description
			res = append(res, emp)
		}
		//fmt.Println(uid)
		tmpl.ExecuteTemplate(w, "Dashboard", res)
		defer db.Close()
	} else {
		http.Redirect(w, r, "/", 301)
	}
}

func NewProject(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "NewProject", nil)
}
func InsertProject(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		if r.Method == "POST" {
			r.ParseForm()
			name, description := r.PostFormValue("name"), r.PostFormValue("description")
			insForm, err := db.Prepare("INSERT INTO Projects(name, description, user_id) VALUES(?,?,?)")
			if err != nil {
				panic(err.Error())
			}
			insForm.Exec(name, description, uid)
			log.Println("INSERT: Name: " + name + " | Description: " + description)
		}
		defer db.Close()
		http.Redirect(w, r, "/dashboard", 301)
	} else {
		http.Redirect(w, r, "/", 301)
	}
}

func ShowProject(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		nId := r.URL.Query().Get("id")
		selDB, err := db.Query("SELECT * FROM Projects WHERE id=? and user_id=?", nId, uid)
		if err != nil {
			panic(err.Error())
		}
		emp := models.Projects{}
		for selDB.Next() {
			var id, user_id int
			var name, description string
			err = selDB.Scan(&id, &name, &description, &user_id)
			if err != nil {
				panic(err.Error())
			}
			emp.Id = id
			emp.UserId = user_id
			emp.ProjectName = name
			emp.Description = description
			project_id = emp.Id
			fmt.Println("Project id is " + string(project_id))
		}
		tmpl.ExecuteTemplate(w, "ShowProject", emp)
		defer db.Close()
	} else {
		http.Redirect(w, r, "/", 301)
	}
}

func EditProject(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		nId := r.URL.Query().Get("id")
		fmt.Println(r.Method)
		selDB, err := db.Query("SELECT * FROM Projects WHERE id=? and user_id=?", nId, uid)
		if err != nil {
			panic(err.Error())
		}
		emp := models.Projects{}
		for selDB.Next() {
			var id, user_id int
			var name, description string
			err = selDB.Scan(&id, &name, &description, &user_id)
			if err != nil {
				panic(err.Error())
			}
			emp.Id = id
			emp.UserId = user_id
			emp.ProjectName = name
			emp.Description = description
		}
		tmpl.ExecuteTemplate(w, "EditProject", emp)
		defer db.Close()
	} else {
		http.Redirect(w, r, "/", 301)
	}
}

func UpdateProject(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		if r.Method == "POST" {
			r.ParseForm()
			name, description := r.PostFormValue("name"), r.PostFormValue("description")
			id := r.FormValue("uid")
			insForm, err := db.Prepare("UPDATE Projects SET name=?, description=? WHERE id=?")
			if err != nil {
				fmt.Println(err.Error())
			}
			insForm.Exec(name, description, id)
			log.Println("UPDATE: Name: " + name + " | Description: " + description)
		}
		defer db.Close()
		http.Redirect(w, r, "/dashboard", 301)
	} else {
		http.Redirect(w, r, "/", 301)
	}

}
func DeleteProject(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
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
		http.Redirect(w, r, "/dashboard", 301)
	} else {
		http.Redirect(w, r, "/", 301)
	}
}

func DisplayIssues(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		selDB, err := db.Query("SELECT * FROM Issues WHERE user_id=? ORDER BY id DESC", uid)
		if err != nil {
			panic(err.Error())
		}
		emp := models.Issues{}
		res := []models.Issues{}
		for selDB.Next() {
			var id, user_id int
			var name, description, priority, date string
			err = selDB.Scan(&id, &name, &description, &priority, &date, &project_id, &user_id)
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
		fmt.Print("display users id is ")
		fmt.Println(uid)
		tmpl.ExecuteTemplate(w, "Issues", res)
		defer db.Close()
	} else {
		http.Redirect(w, r, "/", 301)
	}

}

//Index routes to index template, returns all available issues
func Index(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		selDB, err := db.Query("SELECT * FROM Issues WHERE project_id=? and user_id=? ORDER BY id DESC", project_id, uid)
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
		tmpl.ExecuteTemplate(w, "Index", res)
		defer db.Close()
	} else {
		fmt.Println(uid)
		http.Redirect(w, r, "/", 301)
	}
}

// Show is a function that routes to View template
func Show(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		nId := r.URL.Query().Get("id")
		fmt.Println(nId)
		selDB, err := db.Query("SELECT * FROM Issues WHERE id=? and project_id=? and user_id=?", nId, project_id, uid)
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
		tmpl.ExecuteTemplate(w, "Show", emp)
		defer db.Close()
	} else {
		http.Redirect(w, r, "/", 301)
	}
}

// New is a router to create a new Blip
func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

// Edit is a route to UPDATE an existing Blip
func Edit(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		nId := r.URL.Query().Get("id")
		fmt.Println(r.Method)
		selDB, err := db.Query("SELECT * FROM Issues WHERE id=? and project_id=? and user_id", nId, project_id, uid)
		if err != nil {
			panic(err.Error())
		}
		emp := models.Issues{}
		for selDB.Next() {
			var id, project_id int
			var name, description, priority, date string
			err = selDB.Scan(&id, &name, &description, &priority, &date, &project_id, &uid)
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
		tmpl.ExecuteTemplate(w, "Edit", emp)
		defer db.Close()
	} else {
		fmt.Println(uid)
		http.Redirect(w, r, "/", 301)
	}
}

// Insert is a router function that creates the new Blip
func Insert(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
		if r.Method == "POST" {
			r.ParseForm()
			name, description, priority := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("priority")
			date := time.Now().Format("01-02-2006")
			insForm, err := db.Prepare("INSERT INTO Issues(name, description, priority, date, project_id, user_id) VALUES(?,?,?,?,?, ?)")
			if err != nil {
				panic(err.Error())
			}
			insForm.Exec(name, description, priority, date, project_id, uid)
			log.Println("INSERT: Name: " + name + " | Description: " + description + " | Priority: " + priority + " | Date: " + date)
		}
		defer db.Close()
		http.Redirect(w, r, "/index", 301)
	} else {
		http.Redirect(w, r, "/", 301)
	}
}

// Update is the router function that updates an existing Blip in the database
func Update(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
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
	} else {
		http.Redirect(w, r, "/", 301)
	}
}

// Delete is a function that removes the Blip from the database
func Delete(w http.ResponseWriter, r *http.Request) {
	if singedIn != false {
		db := dbConn()
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
	} else {
		http.Redirect(w, r, "/", 301)
	}
}
