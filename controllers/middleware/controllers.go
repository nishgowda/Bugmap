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

var Tmpl = template.Must(template.ParseGlob("./views/*"))
var Uid int
var Project_id int
var SingedIn = false

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

var (
	googleOauthConfig *oauth2.Config
)

func init() {
	initGithub()
	initGoogle()

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
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email"},
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
	db := DbConn()
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
				err = selDb.Scan(&Uid)
				if err != nil {
					panic(err.Error())
				}
				fmt.Println(Uid)
				empExist.Uid = Uid
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
	defer db.Close()
	SingedIn = true
	fmt.Println(email)
	expirationTime := time.Now().Add(30 * time.Minute)
	claims := &models.Claims{
		Email: email,
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
	email := fmt.Sprint(result["email"])
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
				err = selDb.Scan(&Uid)
				if err != nil {
					panic(err.Error())
				}
				fmt.Println(Uid)
				empExist.Uid = Uid
				resExist = append(resExist, empExist)
			}
		} else {
			insForm, err := db.Prepare("INSERT IGNORE INTO Users(email, password) VALUES(?,?)")
			if err != nil {
				fmt.Println(err.Error())
			}
			password := RandStringBytes(14)
			hash, _ := HashPassword(password)
			insForm.Exec(email, hash)
			log.Println("INSERT: Email" + email + " | Password: " + hash)
			defer db.Close()
		}
	}
	SingedIn = true
	expirationTime := time.Now().Add(30 * time.Minute)
	claims := &models.Claims{
		Email: email,
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

	defer db.Close()
	fmt.Println("dashboard?")
	http.Redirect(w, r, "/dashboard", 301)
}

func Login(w http.ResponseWriter, r *http.Request) {
	db := DbConn()
	if r.Method == "POST" {
		r.ParseForm()
		email, password := r.PostFormValue("email"), r.PostFormValue("password")
		ogPassword := password
		selDb, err := db.Query("SELECT uid, email, password FROM USERS WHERE email=?", email)
		if err != nil {
			http.Redirect(w, r, "/", 301)
		}
		emp := models.Users{}
		res := []models.Users{}
		for selDb.Next() {
			err = selDb.Scan(&Uid, &email, &password)
			if err != nil {
				panic(err.Error())
			}
			emp.Uid = Uid
			emp.Password = password
			emp.Email = email
			if Uid != 0 {
				if CheckPasswordHash(ogPassword, emp.Password) == true {
					res = append(res, emp)
					//fmt.Println(emp.Password)
					fmt.Println("succesfully logged in as " + email)

					expirationTime := time.Now().Add(30 * time.Minute)
					claims := &models.Claims{
						Email: email,
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
					SingedIn = true
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
	Tmpl.ExecuteTemplate(w, "Logout", nil)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	SingedIn = false
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
		email, password := r.PostFormValue("email"), r.PostFormValue("password")
		insForm, err := db.Prepare("INSERT IGNORE INTO Users(email, password ) VALUES(?,?)")
		if err != nil {
			fmt.Println(err.Error)
		}
		hash, _ := HashPassword(password)
		insForm.Exec(email, hash)
		log.Println("INSERT: Email: " + email + " | Password: " + string(hash))
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

	if time.Unix(claims.ExpiresAt, 0).Sub(time.Now()) > 30*time.Second {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	expirationTime := time.Now().Add(5 * time.Minute)
	claims.ExpiresAt = expirationTime.Unix()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the new token as the users `session_token` cookie
	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   tokenString,
		Expires: expirationTime,
	})

}
