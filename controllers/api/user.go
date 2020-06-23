package api

import (
	"fmt"
	"log"
	controllers "models/controllers/middleware"
	"models/models"
	"models/properties"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

func UserProfile(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	//fmt.Println(r.Cookie("token"))
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("404")
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
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
	emp := models.Users{}
	res := []models.Users{}
	userDb, err := db.Query("select email from users where uid=?", claims.Uid)
	if err != nil {
		log.Fatal(err.Error())
	}

	for userDb.Next() {
		var email string
		err = userDb.Scan(&email)
		if err != nil {
			log.Fatal(err.Error())
		}
		emp.Email = email
	}
	defer db.Close()
	projDb, err := db.Query("select projects.id from projects inner join users_projects on users_projects.project_id=projects.id where users_projects.user_id=? and projects.status not in ('private')", claims.Uid)
	if err != nil {
		log.Fatal(err.Error())
	}
	emp.ProjectIDs = []int{}
	var projectID int
	for projDb.Next() {
		err = projDb.Scan(&projectID)
		if err != nil {
			log.Fatal(err.Error())
		}
		emp.ProjectIDs = append(emp.ProjectIDs, projectID)
	}
	defer db.Close()
	allProjDb, err := db.Query("select projects.name from projects inner join users_projects on users_projects.project_id=projects.id where users_projects.user_id=? and projects.status not in ('private')", claims.Uid)
	if err != nil {
		log.Fatal(err.Error())
	}
	emp.Projects = []string{}
	var names string
	for allProjDb.Next() {
		err = allProjDb.Scan(&names)
		if err != nil {
			log.Fatal(err.Error())
		}
		emp.Projects = append(emp.Projects, names)
	}
	defer db.Close()

	for i := 0; i < len(emp.ProjectIDs); i++ {
		query := fmt.Sprintf("select users.email, users.uid from users inner join users_projects on users_projects.user_id=users.uid where users_projects.project_id=%d and  users_projects.user_id not in (%d);", emp.ProjectIDs[i], claims.Uid)
		fmt.Println(query)
		collabsDb, err := db.Query(query)
		if err != nil {
			log.Fatal(err.Error())
		}
		for collabsDb.Next() {
			var emails string
			var uid int
			err = collabsDb.Scan(&emails, &uid)
			if err != nil {
				log.Fatal(err.Error())
			}
			emp.Collaborators = properties.UniqueString(append(emp.Collaborators, emails))
			emp.CollabUids = properties.UniqueInt(append(emp.CollabUids, uid))
		}

	}
	emp.SessionId = []int{}
	codeSessionCollabs, err := db.Query("select code_sessions.id from code_sessions inner join user_code_sessions on user_code_sessions.code_session_id=code_sessions.id where user_code_sessions.user_id=?", claims.Uid)
	if err != nil {
		panic(err.Error())
	}
	for codeSessionCollabs.Next() {
		var id int
		err = codeSessionCollabs.Scan(&id)
		if err != nil {
			panic(err.Error())
		}
		emp.SessionId = append(emp.SessionId, id)
	}
	for i := 0; i < len(emp.SessionId); i++ {
		query := fmt.Sprintf("select users.email, users.uid from users inner join user_code_sessions on user_code_sessions.user_id=users.uid where user_code_sessions.code_session_id=%d and user_code_sessions.user_id not in (%d);", emp.SessionId[i], claims.Uid)
		fmt.Println(query)
		collabsDb, err := db.Query(query)
		if err != nil {
			panic(err.Error())
		}
		for collabsDb.Next() {
			var emails string
			var uid int
			err = collabsDb.Scan(&emails, &uid)
			if err != nil {
				log.Fatal(err.Error())
			}
			emp.Collaborators = properties.UniqueString(append(emp.Collaborators, emails))
			emp.CollabUids = properties.UniqueInt(append(emp.CollabUids, uid))
		}
	}
	defer db.Close()
	fmt.Println(emp.Collaborators)
	res = append(res, emp)
	fmt.Println(res)
	controllers.Tmpl.ExecuteTemplate(w, "Profile", res)
}
func UserSearch(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	//fmt.Println(r.Cookie("token"))
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("404")
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
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
	emp := models.Users{}
	res := []models.Users{}
	nID := r.URL.Query().Get("uid")
	fmt.Println(nID)
	userDb, err := db.Query("select email from users where uid=?", nID)
	if err != nil {
		log.Fatal(err.Error())
	}

	for userDb.Next() {
		var email string
		err = userDb.Scan(&email)
		if err != nil {
			log.Fatal(err.Error())
		}
		emp.Email = email

	}
	defer db.Close()
	projDb, err := db.Query("select projects.id from projects inner join users_projects on users_projects.project_id=projects.id where users_projects.user_id=? and projects.status not in ('private')", nID)
	if err != nil {
		log.Fatal(err.Error())
	}
	emp.ProjectIDs = []int{}
	var projectID int
	for projDb.Next() {
		err = projDb.Scan(&projectID)
		if err != nil {
			log.Fatal(err.Error())
		}
		emp.ProjectIDs = append(emp.ProjectIDs, projectID)
	}

	allProjDb, err := db.Query("select projects.name from projects inner join users_projects on users_projects.project_id=projects.id where users_projects.user_id=? and projects.status not in ('private')", nID)
	if err != nil {
		log.Fatal(err.Error())
	}
	emp.Projects = []string{}
	var names string
	for allProjDb.Next() {
		err = allProjDb.Scan(&names)
		if err != nil {
			log.Fatal(err.Error())
		}
		emp.Projects = append(emp.Projects, names)
	}
	for i := 0; i < len(emp.ProjectIDs); i++ {
		query := fmt.Sprintf("select users.email, users.uid from users inner join users_projects on users_projects.user_id=users.uid where users_projects.project_id=%d and  users_projects.user_id not in (%s);", emp.ProjectIDs[i], nID)
		fmt.Println(query)
		collabsDb, err := db.Query(query)
		if err != nil {
			log.Fatal(err.Error())
		}
		for collabsDb.Next() {
			var emails string
			var uid int
			err = collabsDb.Scan(&emails, &uid)
			if err != nil {
				log.Fatal(err.Error())
			}
			emp.Collaborators = properties.UniqueString(append(emp.Collaborators, emails))
			emp.CollabUids = properties.UniqueInt(append(emp.CollabUids, uid))
		}
	}
	emp.SessionId = []int{}
	codeSessionCollabs, err := db.Query("select code_sessions.id from code_sessions inner join user_code_sessions on user_code_sessions.code_session_id=code_sessions.id where user_code_sessions.user_id=?", nID)
	if err != nil {
		panic(err.Error())
	}
	for codeSessionCollabs.Next() {
		var id int
		err = codeSessionCollabs.Scan(&id)
		if err != nil {
			panic(err.Error())
		}
		emp.SessionId = append(emp.SessionId, id)
	}
	for i := 0; i < len(emp.SessionId); i++ {
		query := fmt.Sprintf("select users.email, users.uid from users inner join user_code_sessions on user_code_sessions.user_id=users.uid where user_code_sessions.code_session_id=%d and user_code_sessions.user_id not in (%s);", emp.SessionId[i], nID)
		fmt.Println(query)
		collabsDb, err := db.Query(query)
		if err != nil {
			panic(err.Error())
		}
		for collabsDb.Next() {
			var emails string
			var uid int
			err = collabsDb.Scan(&emails, &uid)
			if err != nil {
				log.Fatal(err.Error())
			}
			emp.Collaborators = properties.UniqueString(append(emp.Collaborators, emails))
			emp.CollabUids = properties.UniqueInt(append(emp.CollabUids, uid))
		}
	}
	res = append(res, emp)
	fmt.Println(res)
	controllers.Tmpl.ExecuteTemplate(w, "ProfileSearch", res)

}

func Search(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	//fmt.Println(r.Cookie("token"))
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("404")
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
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
	fmt.Println("hello")
	if r.Method == "POST" {
		r.ParseForm()
		search := r.FormValue("search")
		fmt.Println(search)
		db := controllers.DbConn()
		var exists bool
		searchDb, err := db.Query("select exists(select email from users where email=?)", search)
		emp := models.Projects{}
		res := []models.Projects{}
		if err != nil {
			log.Fatal(err.Error())
		}

		for searchDb.Next() {
			err = searchDb.Scan(&exists)
			if err != nil {
				log.Fatal(err.Error())
			}
			if exists {

				emailDb, err := db.Query("select email, uid from users where email=?", search)
				if err != nil {
					log.Fatal(err.Error())
				}
				for emailDb.Next() {
					var email string
					var uid int
					err = emailDb.Scan(&email, &uid)
					if err != nil {
						log.Fatal(err.Error())
					}
					emp.UserEmail = email
					emp.UserId = uid
					res = append(res, emp)
				}
				fmt.Println("ja")
				fmt.Println(res)
				controllers.Tmpl.ExecuteTemplate(w, "SearchResult", res)
			} else if exists == false {
				projDbExist, err := db.Query("select exists(select name, id from projects where name=?)", search)
				if err != nil {
					log.Fatal(err.Error())
				}
				for projDbExist.Next() {
					var projExist bool
					err = projDbExist.Scan(&projExist)
					if err != nil {
						log.Fatal(err.Error())
					}
					if projExist {
						projDb, err := db.Query("select name, id from projects where name=? and status not in ('private')", search)
						if err != nil {
							log.Fatal(err.Error())
						}
						for projDb.Next() {
							var projName string
							var projId int
							err = projDb.Scan(&projName, &projId)
							if err != nil {
								log.Fatal(err.Error())
							}
							emp.ProjectName = projName
							emp.Id = projId
							res = append(res, emp)
						}
						fmt.Println("jedsada")
						fmt.Println(res)
						controllers.Tmpl.ExecuteTemplate(w, "SearchResult", res)
					} else {
						emp.UserId = 0
						emp.Id = 0
						res = append(res, emp)
						controllers.Tmpl.ExecuteTemplate(w, "SearchResult", res)
					}
				}

			} else {
				controllers.Tmpl.ExecuteTemplate(w, "SearchResult", res)
			}
			defer db.Close()
		}

	}
}

func Dashboard(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie("token")
	//fmt.Println(r.Cookie("token"))
	if err != nil {
		if err == http.ErrNoCookie {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Println("404")
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
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
	rows, err := db.Query("select count(*) from projects inner join users_projects on users_projects.project_id=projects.id where users_projects.user_id=?", claims.Uid)
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
		empProj.Dates = properties.UniqueString(append(empProj.Dates, dates))
	}
	defer datesDb.Close()

	for i := 0; i < len(empProj.Dates); i++ {
		issuesPerDate, err := db.Query("select count(*) from issues where date=? and user_id=?", empProj.Dates[i], claims.Uid)
		if err != nil {
			log.Fatal(err)
		}
		var issueDateCount int
		for issuesPerDate.Next() {
			err = issuesPerDate.Scan(&issueDateCount)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(issueDateCount)
			empProj.IssuesPerDate = append(empProj.IssuesPerDate, issueDateCount)
		}
		defer issuesPerDate.Close()

	}

	fmt.Println(empProj.Dates)
	fmt.Println(empProj.IssuesPerDate)
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

	fmt.Println(controllers.GithubAccess)

	controllers.Tmpl.ExecuteTemplate(w, "Dashboard", resProj)
	defer db.Close()
}
