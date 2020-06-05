package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"models/models"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/go-github/github"
	"golang.org/x/crypto/bcrypt"

	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github" // with go modules enabled (GO111MODULE=on or outside GOPATH)
	"golang.org/x/oauth2/google"
)

var Tmpl = template.Must(template.ParseGlob("./views/*"))

var Project_id int

var (
	googleOauthConfig *oauth2.Config
)

func init() {
	initGithub()
	initGoogle()
}

func initGoogle() {
	jsonFile, err := os.Open("./googlesecret.json")
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
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

var (
	githubOauthConfig *oauth2.Config
)

func initGithub() {
	jsonFile, err := os.Open("./githubsecret.json")
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
var jwtKey = JwtKey()

func DbConn() (db *sql.DB) {
	file, err := os.Open("./dbSecret.json")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	body, err := ioutil.ReadAll(file)
	js := string(body)
	var result map[string]interface{}
	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(js), &result)
	DbName := fmt.Sprint(result["dbName"])
	DbPassword := fmt.Sprint(result["dbPassword"])
	dbDriver := "mysql"
	dbUser := "root"
	dbPass := DbPassword
	dbName := DbName
	db, errs := sql.Open(dbDriver, dbUser+":"+dbPass+"@/"+dbName)
	if errs != nil {
		panic(err.Error())
	}
	return db
}

func Home(w http.ResponseWriter, r *http.Request) {
	Tmpl.ExecuteTemplate(w, "Login", nil)
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func JwtKey() []byte {
	jsonFile, err := os.Open("./secretKey.json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	body, err := ioutil.ReadAll(jsonFile)
	js := string(body)
	var result map[string]interface{}
	json.Unmarshal([]byte(js), &result)
	secretKey := fmt.Sprint(result["secret_key"])
	var jwtKey = []byte(secretKey)
	return jwtKey
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	url := googleOauthConfig.AuthCodeURL(randState)
	http.Redirect(w, r, url, 301)
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
	js := string(content)
	//values := []string{""}
	var result map[string]interface{}
	// Unmarshal or Decode the JSON to the interface.
	json.Unmarshal([]byte(js), &result)
	name := fmt.Sprint(result["name"])
	email := fmt.Sprint(result["email"])
	fmt.Println(email)
	fmt.Println(name)

	db := DbConn()
	var exists bool
	existsDb, err := db.Query("select exists(select email from users where email=?)", email)
	if err != nil {
		fmt.Println(err)
	}
	//emp := models.Users{}
	//res := []models.Users{}

	for existsDb.Next() {
		var uid int
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
			insForm, err := db.Prepare("INSERT IGNORE INTO Users(email) VALUES(?)")
			if err != nil {
				fmt.Println(err.Error())
			}
			insForm.Exec(email)
			log.Println("INSERT: Email" + email)
			defer db.Close()
		}
		expirationTime := time.Now().Add(1 * time.Minute)
		claims := &models.Claims{
			Email: email,
			Uid:   uid,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}
		tokens := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := tokens.SignedString(jwtKey)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
			Path:    "/",
		})
		fmt.Println(tokenString)
		fmt.Println(claims.Email)
		fmt.Println(claims.Uid)
	}

	defer db.Close()
	fmt.Println("dashboard?")
	http.Redirect(w, r, "/dashboard", 301)
}

func HandleGitHubLogin(w http.ResponseWriter, r *http.Request) {
	url := githubOauthConfig.AuthCodeURL(randState)
	http.Redirect(w, r, url, 301)
}
func HandleGitHubCallback(w http.ResponseWriter, r *http.Request) {
	state := r.FormValue("state")
	if state != randState {
		fmt.Printf("invalid oauth state, expected '%s', got '%s'\n", randState, state)
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	token, err := githubOauthConfig.Exchange(oauth2.NoContext, r.FormValue("code"))
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
		fmt.Println("couldn't get request", err.Error())
		http.Redirect(w, r, "/", 301)
		return
	}
	fmt.Printf("Logged in as GitHub user: %s\n", *user.Email)
	email := *user.Email
	fmt.Println(email)
	db := DbConn()
	var exists bool
	existsDb, err := db.Query("select exists(select email from users where email=?)", email)
	if err != nil {
		fmt.Println(err)
	}
	//emp := models.Users{}
	//res := []models.Users{}

	for existsDb.Next() {
		var uid int
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
			insForm, err := db.Prepare("INSERT IGNORE INTO Users(email) VALUES(?)")
			if err != nil {
				fmt.Println(err.Error())
			}

			password := RandStringBytes(14)
			hash, _ := HashPassword(password)
			insForm.Exec(email, hash)
			log.Println("INSERT: Email" + email)
			defer db.Close()
		}
		expirationTime := time.Now().Add(1 * time.Minute)
		claims := &models.Claims{
			Email: email,
			Uid:   uid,
			StandardClaims: jwt.StandardClaims{
				ExpiresAt: expirationTime.Unix(),
			},
		}
		tokens := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := tokens.SignedString(jwtKey)
		fmt.Println(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:    "token",
			Value:   tokenString,
			Expires: expirationTime,
			Path:    "/",
		})

		fmt.Println(claims.Email + " UID ")
		fmt.Println(claims.Uid)
	}
	defer db.Close()
	fmt.Println(email)
	fmt.Println("dashboard?")
	http.Redirect(w, r, "/dashboard", 301)

}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
func Login(w http.ResponseWriter, r *http.Request) {
	db := DbConn()
	if r.Method == "POST" {
		r.ParseForm()
		username, password := r.PostFormValue("username"), r.PostFormValue("password")
		ogPassword := password
		currentData, err := db.Query("Select username from users")

		var allUsersNames []string
		var storedUsernames string
		for currentData.Next() {
			err = currentData.Scan(&storedUsernames)
			if err != nil {
				panic(err.Error())
			}
			allUsersNames = append(allUsersNames, storedUsernames)

		}
		fmt.Println(allUsersNames)
		if contains(allUsersNames, username) == false {
			Message := "Failed to Login"
			Tmpl.ExecuteTemplate(w, "Login", Message)
			return
		}
		selDb, err := db.Query("SELECT uid, username, password FROM USERS WHERE username=?", username)
		if err != nil {
			http.Redirect(w, r, "/", 301)
		}
		emp := models.Users{}
		res := []models.Users{}
		var uid int
		for selDb.Next() {
			err = selDb.Scan(&uid, &username, &password)
			if err != nil {
				panic(err.Error())
			}
			emp.Uid = uid
			emp.Password = password
			emp.Username = username
			if CheckPasswordHash(ogPassword, emp.Password) == true {
				res = append(res, emp)
				//fmt.Println(emp.Password)
				fmt.Println("succesfully logged in as " + username)

				expirationTime := time.Now().Add(1 * time.Minute)
				claims := &models.Claims{
					Username: username,
					Uid:      uid,
					StandardClaims: jwt.StandardClaims{
						ExpiresAt: expirationTime.Unix(),
					},
				}
				tokens := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
				tokenString, err := tokens.SignedString(jwtKey)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				http.SetCookie(w, &http.Cookie{
					Name:    "token",
					Value:   tokenString,
					Expires: expirationTime,
				})

			} else {
				Message := "Failed to Login"
				Tmpl.ExecuteTemplate(w, "Login", Message) // ---> Figure out a work around for this superfluous response.WriteHeader call from main.Login (main.go:129)
			}

		}

	}

	defer db.Close()
	http.Redirect(w, r, "/dashboard", 301)
}
func FailedLogin(w http.ResponseWriter, r *http.Request) {
	Tmpl.ExecuteTemplate(w, "FailedLogin", nil)

}

func LogoutPage(w http.ResponseWriter, r *http.Request) {
	Tmpl.ExecuteTemplate(w, "Logout", nil)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", 301)
}

func SingUpPage(w http.ResponseWriter, r *http.Request) {
	Tmpl.ExecuteTemplate(w, "Register", nil)
}
func Register(w http.ResponseWriter, r *http.Request) {
	db := DbConn()
	fmt.Println("Working?")
	if r.Method == "POST" {
		r.ParseForm()
		username, password, first_name, last_name := r.PostFormValue("username"), r.PostFormValue("password"), r.PostFormValue("first_name"), r.PostFormValue("last_name")
		currentData, err := db.Query("Select username from users")
		if err != nil {
			panic(err.Error())
		}
		var allUsersNames []string
		var storedUsernames string
		for currentData.Next() {
			err = currentData.Scan(&storedUsernames)
			if err != nil {
				panic(err.Error())
			}
			allUsersNames = append(allUsersNames, storedUsernames)

		}
		fmt.Println(allUsersNames)
		for i := 0; i < len(allUsersNames); i++ {
			if allUsersNames[i] == username {
				Message := "Username is already taken"
				Tmpl.ExecuteTemplate(w, "Register", Message)
				return
			}
		}
		insForm, err := db.Prepare("INSERT IGNORE INTO Users(username, password, first_name, last_name ) VALUES(?,?,?,?)")
		if err != nil {
			fmt.Println(err.Error)
		}
		hash, _ := HashPassword(password)
		insForm.Exec(username, hash, first_name, last_name)
		log.Println("INSERT: Username: " + username + " | Password: " + string(hash) + " | First Name: " + first_name + " | Last Name : " + last_name)

	}

	defer db.Close()
	http.Redirect(w, r, "/", 301)
}
func UniqueInt(intSlice []int) []int {
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

func UniqueString(stringSlice []string) []string {
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

func RefreshToken(w http.ResponseWriter, r *http.Request) {
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
	if !tkn.Valid {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if err != nil {
		if err == jwt.ErrSignatureInvalid {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 5*time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	expirationTime := time.Now().Add(30 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	fmt.Println("new token: " + tokenString)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Set the new token as the users `session_token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
		Path:    "/",
	})
}
