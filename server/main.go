package main

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"text/template"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

//Issues model
type Issues struct {
	Id          int
	Name        string
	Description string
	Priority    string
	Date        string
}

type Users struct {
	Uid      int
	Username string
	Password string
}

//ObtainDatabaseName obtains database credentials for the name of the database
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

func Login(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "Login", nil)
}

func Register(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	fmt.Println("Working?")
	if r.Method == "POST" {
		r.ParseForm()
		username, password := r.PostFormValue("username"), r.PostFormValue("password")
		insForm, err := db.Prepare("INSERT INTO Users(username, password) VALUES(?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(username, password)
		log.Println("INSERT: Username: " + username + " | Password: " + password)
	}
	defer db.Close()
	http.Redirect(w, r, "/index", 301)
}

//Index routes to index template, returns all available issues
func Index(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	selDB, err := db.Query("SELECT * FROM Issues ORDER BY id DESC")
	if err != nil {
		panic(err.Error())
	}
	emp := Issues{}
	res := []Issues{}
	for selDB.Next() {
		var id int
		var name, description, priority, date string
		err = selDB.Scan(&id, &name, &description, &priority, &date)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
		emp.Date = date
		res = append(res, emp)
	}
	tmpl.ExecuteTemplate(w, "Index", res)
	defer db.Close()
}

// Show is a function that routes to View template
func Show(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")

	selDB, err := db.Query("SELECT * FROM Issues WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}
	emp := Issues{}
	for selDB.Next() {
		var id int
		var name, description, priority, date string
		err = selDB.Scan(&id, &name, &description, &priority, &date)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
		emp.Date = date
	}
	tmpl.ExecuteTemplate(w, "Show", emp)
	defer db.Close()
}

// New is a router to create a new Blip
func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

// Edit is a route to UPDATE an existing Blip
func Edit(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	fmt.Println(r.Method)
	selDB, err := db.Query("SELECT * FROM Issues WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}
	emp := Issues{}
	for selDB.Next() {
		var id int
		var name, description, priority, date string
		err = selDB.Scan(&id, &name, &description, &priority, &date)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
		emp.Date = date
	}
	tmpl.ExecuteTemplate(w, "Edit", emp)
	defer db.Close()
}

// Insert is a router function that creates the new Blip
func Insert(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		r.ParseForm()
		name, description, priority := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("priority")
		date := time.Now().Format("01-02-2006 15:04")
		insForm, err := db.Prepare("INSERT INTO Issues(name, description, priority, date) VALUES(?,?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, description, priority, date)
		log.Println("INSERT: Name: " + name + " | Description: " + description + " | Priority: " + priority + " | Date: " + date)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

// Update is the router function that updates an existing Blip in the database
func Update(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		r.ParseForm()
		name, description, priority := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("priority")
		id := r.FormValue("uid")
		date := time.Now().Format("01-02-2006 15:04")
		insForm, err := db.Prepare("UPDATE Issues SET name=?, description=?, priority=?, date=? WHERE id=?")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, description, priority, date, id)
		log.Println("UPDATE: Name: " + name + " | Description: " + description + " | Priority: " + priority + " | Date: " + date)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

// Delete is a function that removes the Blip from the database
func Delete(w http.ResponseWriter, r *http.Request) {
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
	http.Redirect(w, r, "/", 301)
}
func main() {
	log.Println("Server started on: http://localhost:8080")
	http.HandleFunc("/", (Login))
	http.HandleFunc("/login", (Register))
	http.HandleFunc("/index", (Index))
	http.HandleFunc("/show", (Show))
	http.HandleFunc("/new", (New))
	http.HandleFunc("/edit", (Edit))
	http.HandleFunc("/insert", (Insert))
	http.HandleFunc("/update", (Update))
	http.HandleFunc("/delete", (Delete))
	http.ListenAndServe(":8080", nil)
}
