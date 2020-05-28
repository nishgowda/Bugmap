package main

import (
	"database/sql"
	"log"
	"net/http"
	"text/template"

	_ "github.com/go-sql-driver/mysql"
)

type Issues struct {
	Id          int
	Name        string
	Description string
	Priority    string
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := "dbpword"
	dbName := "dbname"
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	return db
}

var tmpl = template.Must(template.ParseGlob("form/*"))

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
		var name, description, priority string
		err = selDB.Scan(&id, &name, &description, &priority)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
		res = append(res, emp)
	}
	tmpl.ExecuteTemplate(w, "Index", res)
	defer db.Close()
}

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
		var name, description, priority string
		err = selDB.Scan(&id, &name, &description, &priority)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
	}
	tmpl.ExecuteTemplate(w, "Show", emp)
	defer db.Close()
}

func New(w http.ResponseWriter, r *http.Request) {
	tmpl.ExecuteTemplate(w, "New", nil)
}

func Edit(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	nId := r.URL.Query().Get("id")
	selDB, err := db.Query("SELECT * FROM Issues WHERE id=?", nId)
	if err != nil {
		panic(err.Error())
	}
	emp := Issues{}
	for selDB.Next() {
		var id int
		var name, description, priority string
		err = selDB.Scan(&id, &name, &description, &priority)
		if err != nil {
			panic(err.Error())
		}
		emp.Id = id
		emp.Name = name
		emp.Description = description
		emp.Priority = priority
	}
	tmpl.ExecuteTemplate(w, "Edit", emp)
	defer db.Close()
}

func Insert(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		name := r.FormValue("name")
		description := r.FormValue("description")
		priority := r.FormValue("priority")
		insForm, err := db.Prepare("INSERT INTO Issues(name, description, priority) VALUES(?,?,?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, description, priority)
		log.Println("INSERT: Name: " + name + " | Description: " + description + " | Priority: " + priority)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}

<<<<<<< HEAD
func Update(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	if r.Method == "POST" {
		name := r.FormValue("name")
		description := r.FormValue("description")
		priority := r.FormValue("priority")
		id := r.FormValue("uid")
		insForm, err := db.Prepare("UPDATE Issues SET name=?, description=?, priority=? WHERE id=?")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, description, priority, id)
		log.Println("UPDATE: Name: " + name + " | Description: " + description + " | Priority: " + priority)
	}
	defer db.Close()
	http.Redirect(w, r, "/", 301)
}
=======
func main() {
	db, err = gorm.Open("mysql", "root:pword@tcp(127.0.0.1:3306)/IssueTracker?charset=utf8&parseTime=True")
>>>>>>> 0285abb81b16968b7595c6678cd87fa6e3c8217e

func Delete(w http.ResponseWriter, r *http.Request) {
	db := dbConn()
	emp := r.URL.Query().Get("id")
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
	http.HandleFunc("/", Index)
	http.HandleFunc("/show", Show)
	http.HandleFunc("/new", New)
	http.HandleFunc("/edit", Edit)
	http.HandleFunc("/insert", Insert)
	http.HandleFunc("/update", Update)
	http.HandleFunc("/delete", Delete)
	http.ListenAndServe(":8080", nil)
}
