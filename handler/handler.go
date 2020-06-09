package handler

import (
	"log"
	"models/controllers/api"
	controllers "models/controllers/middleware"

	"net/http"
)

func HandlerFunc() {
	log.Println("Server started on: http://localhost:8080")
	http.HandleFunc("/", controllers.Home)
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/refresh", controllers.RefreshToken)
	http.HandleFunc("/googlelogin", controllers.HandleLogin)
	http.HandleFunc("/callback", controllers.HandleCallback)
	http.HandleFunc("/githublogin", controllers.HandleGitHubLogin)
	http.HandleFunc("/callback/github", controllers.HandleGitHubCallback)
	http.HandleFunc("/signup", controllers.SingUpPage)
	http.HandleFunc("/register", controllers.Register)
	http.HandleFunc("/logoutpage", controllers.LogoutPage)
	http.HandleFunc("/logout", controllers.Logout)
	http.HandleFunc("/dashboard", api.Dashboard)
	http.HandleFunc("/displayprojects", api.DisplayProjects)
	http.HandleFunc("/invite", api.Invite)
	http.HandleFunc("/inviteuser", api.InviteUser)
	http.HandleFunc("/bugs", api.DisplayIssues)
	http.HandleFunc("/showproject", api.ShowProject)
	http.HandleFunc("/newproject", api.NewProject)
	http.HandleFunc("/importrepos", api.ImportRepos)
	http.HandleFunc("/editproject", api.EditProject)
	http.HandleFunc("/insertproject", api.InsertProject)
	http.HandleFunc("/updateproject", api.UpdateProject)
	http.HandleFunc("/deleteproject", api.DeleteProject)
	http.HandleFunc("/index", api.Index)
	http.HandleFunc("/show", api.Show)
	http.HandleFunc("/new", api.New)
	http.HandleFunc("/edit", api.Edit)
	http.HandleFunc("/insert", api.Insert)
	http.HandleFunc("/update", api.Update)
	http.HandleFunc("/delete", api.Delete)
	http.ListenAndServe(":8080", nil)
}
