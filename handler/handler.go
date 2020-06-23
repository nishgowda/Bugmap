package handler

import (
	"log"
	"models/controllers/api"
	"models/controllers/editor"
	controllers "models/controllers/middleware"

	"net/http"
)

func HandlerFunc() {
	log.Println("Server started on: http://localhost:8080")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
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
	http.HandleFunc("/userprofile", api.UserProfile)
	http.HandleFunc("/profilesearch", api.UserSearch)
	http.HandleFunc("/search", api.Search)
	http.HandleFunc("/dashboard", api.Dashboard)
	http.HandleFunc("/projects", api.DisplayProjects)
	http.HandleFunc("/project/invite", api.Invite)
	http.HandleFunc("/project/inviteuser", api.InviteUser)
	http.HandleFunc("/tickets", api.DisplayIssues)
	http.HandleFunc("/project", api.ShowProject)
	http.HandleFunc("/project/new", api.NewProject)
	http.HandleFunc("/importrepos", api.ImportRepos)
	http.HandleFunc("/project/edit", api.EditProject)
	http.HandleFunc("/project/insert", api.InsertProject)
	http.HandleFunc("/project/update", api.UpdateProject)
	http.HandleFunc("/project/delete", api.DeleteProject)
	http.HandleFunc("/project/tickets", api.Index)
	http.HandleFunc("/project/ticket", api.Show)
	http.HandleFunc("/project/ticket/new", api.New)
	http.HandleFunc("/project/ticket/edit", api.Edit)
	http.HandleFunc("/project/ticket/insert", api.Insert)
	http.HandleFunc("/project/ticket/update", api.Update)
	http.HandleFunc("/project/ticket/delete", api.Delete)
	http.HandleFunc("/sessions", editor.DisplaySessions)
	http.HandleFunc("/editor", editor.CodeEditor)
	http.HandleFunc("/sharesession", editor.InviteSession)
	http.HandleFunc("/session/new", editor.NewSession)
	http.HandleFunc("/session/insert", editor.InsertSession)
	http.HandleFunc("/session/delete", editor.DeleteSession)
	http.ListenAndServe(":8080", nil)
}
