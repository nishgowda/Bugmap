package api

import (
	"fmt"
	"io/ioutil"
	"log"
	controllers "models/controllers/middleware"
	"models/models"
	"net/http"

	jparse "github.com/nishgowda/Jparse"

	"github.com/dgrijalva/jwt-go"
)

var jwtKey = controllers.JwtKey()

func Dashboard(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	fmt.Println(r.Cookie("token"))
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/", 301)
			fmt.Println("404")
			return
		}
		http.Redirect(w, r, "/", 301)
		fmt.Println("401")
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
	fmt.Println(claims.Email)
	fmt.Println(claims.Uid)
	empProj := models.Totals{}
	resProj := []models.Totals{}
	rows, err := db.Query("SELECT COUNT(*) FROM Projects where user_id=?", claims.Uid)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var countProjects int

	for rows.Next() {
		if err := rows.Scan(&countProjects); err != nil {
			log.Fatal(err)
		}
	}
	allRows, errs := db.Query("SELECT COUNT(*) FROM Issues where user_id=?", claims.Uid)
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

	lowPriorityRows, errs := db.Query("Select count(*) from issues where priority='Low' and user_id=?", claims.Uid)
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
	medPriorityRows, errs := db.Query("Select count(*) from issues where priority='Medium' and user_id=?", claims.Uid)
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

	highPriorityRows, errs := db.Query("Select count(*) from issues where priority='High' and user_id=?", claims.Uid)
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

	critPriorityRows, errs := db.Query("Select count(*) from issues where priority='Critical' and user_id=?", claims.Uid)
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

	numFeatureDb, err := db.Query("Select count(*) from issues where kind='Feature' and user_id=?", claims.Uid)
	if err != nil {
		log.Fatal(err)
	}
	defer numFeatureDb.Close()
	var featureCount int
	for numFeatureDb.Next() {
		if errs := numFeatureDb.Scan(&featureCount); errs != nil {
			log.Fatal(err)
		}

	}

	numIssueDb, err := db.Query("Select count(*) from issues where kind='Issue' and user_id=?", claims.Uid)
	if err != nil {
		log.Fatal(err)
	}
	defer numIssueDb.Close()
	var issueCount int
	for numIssueDb.Next() {
		if errs := numIssueDb.Scan(&issueCount); errs != nil {
			log.Fatal(err)
		}
	}

	numNoteDb, err := db.Query("Select count(*) from issues where kind='Note' and user_id=?", claims.Uid)
	if err != nil {
		log.Fatal(err)
	}
	defer numNoteDb.Close()
	var noteCount int
	for numNoteDb.Next() {
		if errs := numNoteDb.Scan(&noteCount); errs != nil {
			log.Fatal(err)
		}
	}
	emp := models.Ratios{}
	res := []models.Ratios{}
	datesDb, err := db.Query("Select date from issues where user_id=?", claims.Uid)
	if err != nil {
		log.Fatal(err)
	}
	var dates string
	for datesDb.Next() {
		err = datesDb.Scan(&dates)
		if err != nil {
			log.Fatal(err)
		}
		emp.Dates = dates
		res = append(res, emp)
	}
	defer datesDb.Close()
	issuesPerDate, err := db.Query("select count(*) from issues where date=? and user_id=?", emp.Dates, claims.Uid)
	if err != nil {
		log.Fatal(err)
	}
	defer datesDb.Close()
	var issueDateCount int
	for issuesPerDate.Next() {
		err = issuesPerDate.Scan(&issueDateCount)
		if err != nil {
			log.Fatal(err)
		}
	}
	emp.IssuesPerDate = issueDateCount
	res = append(res, emp)
	empProj.NumLow = LowPriorityCount
	empProj.NumMedium = MedPriorityCount
	empProj.NumHigh = HighPriorityCount
	empProj.NumCritical = CriticalPriorityCount
	empProj.NumIssues = countIssues
	empProj.NumProjects = countProjects
	empProj.NumFeature = featureCount
	empProj.NumIssue = issueCount
	empProj.NumNote = noteCount

	//emp := models.Totals{}
	//res := []models.Totals{}

	resProj = append(resProj, empProj)
	fmt.Println(resProj)
	fmt.Println(res)
	fmt.Println(controllers.GithubAccess)

	controllers.Tmpl.ExecuteTemplate(w, "Dashboard", resProj)
	controllers.Tmpl.ExecuteTemplate(w, "Charts", res)
	defer db.Close()
}

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
	selDB, err := db.Query("SELECT * FROM Projects WHERE user_id=? ORDER BY id DESC", claims.Uid)
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
	//fmt.Println(res)
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
		insForm, err := db.Prepare("INSERT INTO Projects(name, description, user_id, technologies) VALUES(?,?,?, ?)")
		if err != nil {
			panic(err.Error())
		}
		insForm.Exec(name, description, claims.Uid, technologies)
		log.Println("INSERT: Name: " + name + " | Description: " + description + " | Technologies: " + technologies)
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
	selDB, err := db.Query("SELECT * FROM Projects WHERE id=? and user_id=?", nId, claims.Uid)
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
		controllers.Project_id = emp.Id
		fmt.Println("Project id is " + string(controllers.Project_id))
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
	selDB, err := db.Query("SELECT * FROM Projects WHERE id=? and user_id=?", nId, claims.Uid)
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
		controllers.Project_id = emp.Id
		emp.UserId = user_id
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
