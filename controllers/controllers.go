package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"models/models"
	"net/http"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
	"golang.org/x/crypto/bcrypt"

	jparse "github.com/nishgowda/Jparse"
	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github" // with go modules enabled (GO111MODULE=on or outside GOPATH)
	"golang.org/x/oauth2/google"
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

var (
	googleOauthConfig *oauth2.Config
)

func init() {
	initGithub()
	initGoogle()
}
func initGoogle() {
	jsonFile, err := os.Open("../googlesecret.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened Json file")
	defer jsonFile.Close()

	body, err := ioutil.ReadAll(jsonFile)
	js := string(body)
	var result map[string]interface{}
	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(js), &result)
	clientID := fmt.Sprint(result["client_id"])
	clientSecret := fmt.Sprint(result["client_secret"])
	redirectURL := fmt.Sprint(result["redirect_url"])
	fmt.Println(clientID)
	fmt.Println(clientSecret)
	googleOauthConfig = &oauth2.Config{
		RedirectURL:  redirectURL,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
		Endpoint:     google.Endpoint,
	}
}

var (
	githubOauthConfig *oauth2.Config
)

func initGithub() {
	jsonFile, err := os.Open("../githubsecret.json")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Successfully Opened Json file")
	defer jsonFile.Close()

	body, err := ioutil.ReadAll(jsonFile)
	js := string(body)
	var result map[string]interface{}
	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(js), &result)
	clientID := fmt.Sprint(result["client_id"])
	clientSecret := fmt.Sprint(result["client_secret"])
	redirectURL := fmt.Sprint(result["redirect_url"])
	githubOauthConfig = &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{"user:email"},
		Endpoint:     githuboauth.Endpoint,
	}
}

// Generate a random string of A-Z chars with len = l

var randState = "random"

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(randState)
	http.Redirect(w, r, url, 301)
}

func HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := githubOauthConfig.AuthCodeURL(randState, oauth2.AccessTypeOnline)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}
func HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != randState {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", randState, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	code := r.FormValue("code")
	token, err := githubOauthConfig.Exchange(oauth2.NoContext, code)
	if err != nil {
		fmt.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	oauthClient := githubOauthConfig.Client(oauth2.NoContext, token)
	client := github.NewClient(oauthClient)
	user, _, err := client.Users.Get(oauth2.NoContext, "")
	if err != nil {
		fmt.Printf("client.Users.Get() faled with '%s'\n", err)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Printf("Logged in as GitHub user: %s\n", *user.Email)
	email := *user.Email
	db := dbConn()
	var exists bool
	existsDb, err := db.Query("select exists(select email from users where email=?)", email)
	if err != nil {
		fmt.Println(err)
	}
	//emp := models.Users{}
	//res := []models.Users{}
	for existsDb.Next() {
		err = existsDb.Scan(&exists)
		if err != nil {
			fmt.Println(err.Error())
		}
		if exists {
			selDb, err := db.Query("SELECT uid from users where email=?", email)
			if err != nil {
				fmt.Println(err)
			}
			empExist := models.Users{}
			resExist := []models.Users{}
			for selDb.Next() {
				err = selDb.Scan(&uid)
				if err != nil {
					panic(err.Error())
				}
				fmt.Println(uid)
				empExist.Uid = uid
				resExist = append(resExist, empExist)
			}
		} else {
			insForm, err := db.Prepare("INSERT INTO Users(email) VALUES(?)")
			if err != nil {
				fmt.Println(err.Error())
			}
			insForm.Exec(email)
			log.Println("INSERT: Email" + email)
			defer db.Close()
		}
	}
	singedIn = true
	defer db.Close()
	fmt.Println("dashboard?")
	http.Redirect(w, r, "/dashboard", 301)

}

func HandleCallback(w http.ResponseWriter, r *http.Request) {
	if r.FormValue("state") != randState {
		fmt.Println("state is not valid")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	token, err := googleOauthConfig.Exchange(oauth2.NoContext, r.FormValue("code"))
	if err != nil {
		fmt.Println("couldn't get token", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	resp, err := http.Get("https://www.googleapis.com/oauth2/v3/userinfo?access_token=" + token.AccessToken)
	if err != nil {
		fmt.Println("couldn't get request", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	defer resp.Body.Close()
	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("couldn't parse response", err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	fmt.Println(string(content))
	//values := []string{""}
	embValues := []string{"email", "name"}

	//embObj := []string{"response"}
	userEmail := jparse.SimpleParse(embValues, string(content))
	var email string
	for i := range userEmail {
		email = userEmail[i]
	}
	fmt.Println(email)
	db := dbConn()
	var exists bool
	existsDb, err := db.Query("select exists(select email from users where email=?)", email)
	if err != nil {
		fmt.Println(err)
	}
	//emp := models.Users{}
	//res := []models.Users{}
	for existsDb.Next() {
		err = existsDb.Scan(&exists)
		if err != nil {
			fmt.Println(err.Error())
		}
		if exists {
			selDb, err := db.Query("SELECT uid from users where email=?", email)
			if err != nil {
				fmt.Println(err)
			}
			empExist := models.Users{}
			resExist := []models.Users{}
			for selDb.Next() {
				err = selDb.Scan(&uid)
				if err != nil {
					panic(err.Error())
				}
				fmt.Println(uid)
				empExist.Uid = uid
				resExist = append(resExist, empExist)
			}
		} else {
			insForm, err := db.Prepare("INSERT INTO Users(email) VALUES(?)")
			if err != nil {
				fmt.Println(err.Error())
			}
			insForm.Exec(email)
			log.Println("INSERT: Email" + email)
			defer db.Close()
		}
	}
	singedIn = true
	defer db.Close()
	fmt.Println("dashboard?")
	http.Redirect(w, r, "/dashboard", 301)
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
func uniqueInt(intSlice []int) []int {
	keys := make(map[int]bool)
	list := []int{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func uniqueString(stringSlice []string) []string {
	keys := make(map[string]bool)
	list := []string{}
	for _, entry := range stringSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	if singedIn == true {
		db := dbConn()
		rows, err := db.Query("SELECT COUNT(*) FROM Projects where user_id=?", uid)
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()
		empProj := models.Totals{}
		resProj := []models.Totals{}
		var countProjects int

		for rows.Next() {
			if err := rows.Scan(&countProjects); err != nil {
				log.Fatal(err)
			}
		}

		allRows, errs := db.Query("SELECT COUNT(*) FROM Issues where user_id=?", uid)
		if errs != nil {
			log.Fatal(errs)
		}
		defer allRows.Close()
		var countIssues int
		for allRows.Next() {
			if errs := allRows.Scan(&countIssues); errs != nil {
				log.Fatal(err)
			}
		}

		lowPriorityRows, errs := db.Query("Select count(*) from issues where priority='Low' and user_id=?", uid)
		if errs != nil {
			log.Fatal(errs)
		}
		defer lowPriorityRows.Close()
		var LowPriorityCount int
		for lowPriorityRows.Next() {
			if errs := lowPriorityRows.Scan(&LowPriorityCount); errs != nil {
				log.Fatal(err)
			}
		}
		medPriorityRows, errs := db.Query("Select count(*) from issues where priority='Medium' and user_id=?", uid)
		if errs != nil {
			log.Fatal(errs)
		}
		defer medPriorityRows.Close()
		var MedPriorityCount int
		for medPriorityRows.Next() {
			if errs := medPriorityRows.Scan(&MedPriorityCount); errs != nil {
				log.Fatal(err)
			}
		}

		highPriorityRows, errs := db.Query("Select count(*) from issues where priority='High' and user_id=?", uid)
		if errs != nil {
			log.Fatal(errs)
		}
		defer highPriorityRows.Close()
		var HighPriorityCount int
		for highPriorityRows.Next() {
			if errs := highPriorityRows.Scan(&HighPriorityCount); errs != nil {
				log.Fatal(err)
			}
		}

		critPriorityRows, errs := db.Query("Select count(*) from issues where priority='Critical' and user_id=?", uid)
		if errs != nil {
			log.Fatal(errs)
		}
		defer highPriorityRows.Close()
		var CriticalPriorityCount int
		for critPriorityRows.Next() {
			if errs := critPriorityRows.Scan(&CriticalPriorityCount); errs != nil {
				log.Fatal(err)
			}
		}

		empProj.NumLow = LowPriorityCount
		empProj.NumMedium = MedPriorityCount
		empProj.NumHigh = HighPriorityCount
		empProj.NumCritical = CriticalPriorityCount
		empProj.NumIssues = countIssues
		empProj.NumProjects = countProjects

		//emp := models.Totals{}
		//res := []models.Totals{}
		var dates string
		dateRows, errs := db.Query("Select Date from issues where user_id=?", uid)
		if errs != nil {
			log.Fatal(errs)
		}
		for dateRows.Next() {

			errs = dateRows.Scan(&dates)
			if errs != nil {
				log.Fatal(errs)
			}
			empProj.Dates = uniqueString(append(empProj.Dates, dates))
		}
		var issuesPerDateCount int
		for i := 0; i < len(empProj.Dates); i++ {
			issuesPerDate, errs := db.Query("Select count(*) from issues where date=?", empProj.Dates[i])
			if errs != nil {
				log.Fatal(errs)
			}
			defer issuesPerDate.Close()
			for issuesPerDate.Next() {
				if errs := issuesPerDate.Scan(&issuesPerDateCount); errs != nil {
					log.Fatal(err)
				}
				empProj.IssuesPerDate = uniqueInt(append(empProj.IssuesPerDate, issuesPerDateCount))

			}
		}
		fmt.Println(empProj.Dates)
		fmt.Println(empProj.IssuesPerDate)

		for _, value := range empProj.Dates {
			fmt.Println(value)
		}
		for _, value := range empProj.IssuesPerDate {
			fmt.Println(value)
		}
		resProj = append(resProj, empProj)
		fmt.Println(resProj)
		tmpl.ExecuteTemplate(w, "Dashboard", resProj)
		defer db.Close()
	} else {
		http.Redirect(w, r, "/", 301)
	}
}

func DisplayProjects(w http.ResponseWriter, r *http.Request) {
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
			var name, description, technologies string
			err = selDB.Scan(&id, &name, &description, &user_id, &technologies)
			if err != nil {
				panic(err.Error())
			}
			emp.Id = id
			emp.UserId = user_id
			emp.ProjectName = name
			emp.Description = description
			emp.Technologies = technologies
			res = append(res, emp)
		}
		//fmt.Println(uid)
		fmt.Println(res)
		tmpl.ExecuteTemplate(w, "DisplayProjects", res)
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
			name, description, technologies := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("technologies")
			insForm, err := db.Prepare("INSERT INTO Projects(name, description, user_id, technologies) VALUES(?,?,?, ?)")
			if err != nil {
				panic(err.Error())
			}
			insForm.Exec(name, description, uid, technologies)
			log.Println("INSERT: Name: " + name + " | Description: " + description + " | Technologies: " + technologies)
		}
		defer db.Close()
		http.Redirect(w, r, "/displayprojects", 301)
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
			var name, description, technologies string
			err = selDB.Scan(&id, &name, &description, &user_id, &technologies)
			if err != nil {
				panic(err.Error())
			}
			emp.Id = id
			emp.UserId = user_id
			emp.ProjectName = name
			emp.Description = description
			emp.Technologies = technologies
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
			var name, description, technologies string
			err = selDB.Scan(&id, &name, &description, &user_id, &technologies)
			if err != nil {
				panic(err.Error())
			}
			emp.Id = id
			project_id = emp.Id
			emp.UserId = user_id
			emp.ProjectName = name
			emp.Description = description
			emp.Technologies = technologies
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
			name, description, technologies := r.PostFormValue("name"), r.PostFormValue("description"), r.PostFormValue("technologies")
			fmt.Println(project_id)
			insForm, err := db.Prepare("UPDATE Projects SET name=?, description=?, technologies=? WHERE id=?")
			if err != nil {
				fmt.Println(err.Error())
			}
			insForm.Exec(name, description, technologies, project_id)
			log.Println("UPDATE: Name: " + name + " | Description: " + description + " | Technologies: " + technologies)
		}
		defer db.Close()
		http.Redirect(w, r, "/displayprojects", 301)
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
		http.Redirect(w, r, "/displayprojects", 301)
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
		selDB, err := db.Query("SELECT * FROM Issues WHERE id=? and project_id=?", nId, project_id)
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
