package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var db *gorm.DB
var err error

type Issue struct {
	Id          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to HomePage!")
	fmt.Println("Endpoint Hit: HomePage")
}
func handleRequests() {
	myRouter := mux.NewRouter().StrictSlash(true)
	myRouter.HandleFunc("/", homePage)
	myRouter.HandleFunc("/all-issues", returnAllIssues)
	myRouter.HandleFunc("/new-issue", createNewIssue).Methods("POST")
	myRouter.HandleFunc("/issue/{id}", returnSingleIssue)
	log.Println("Starting development server at http://127.0.0.1:10000/")
	log.Println("Quit the server with CONTROL-C.")
	log.Fatal(http.ListenAndServe(":10000", myRouter))
}

func returnAllIssues(w http.ResponseWriter, r *http.Request) {
	issues := []Issue{}
	db.Find(&issues)
	fmt.Println("Endpoint Hit: returnAllIssues")
	json.NewEncoder(w).Encode(issues)
}
func createNewIssue(w http.ResponseWriter, r *http.Request) {
	// get the body of our POST request
	// return the string response containing the request body
	reqBody, _ := ioutil.ReadAll(r.Body)
	var issue Issue
	json.Unmarshal(reqBody, &issue)
	db.Create(&issue)
	fmt.Println("Endpoint Hit: Creating New Issue")
	json.NewEncoder(w).Encode(issue)
}

func returnSingleIssue(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	key := vars["id"]
	issues := []Issue{}
	db.Find(&issues)

	for _, issue := range issues {
		// string to int
		s, err := strconv.Atoi(key)
		if err == nil {
			if issue.Id == s {
				fmt.Println(issue)
				fmt.Println("Endpoint Hit: Issue No:", key)
				json.NewEncoder(w).Encode(issue)
			}
		}
	}
}

func main() {
	db, err = gorm.Open("mysql", "root:2douglas@tcp(127.0.0.1:3306)/IssueTracker?charset=utf8&parseTime=True")

	if err != nil {
		log.Println("Connection failed to open")
	} else {
		log.Println("Connection Established")
	}

	db.AutoMigrate(&Issue{})
	handleRequests()
}
