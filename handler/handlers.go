package main

import (
	"fmt"
	"log"
	"models/controllers"
	"net/http"
)

func handler() {
	log.Println("Server started on: http://localhost:8080")

	http.HandleFunc("/", controllers.Home)
	http.HandleFunc("/login", controllers.Login)
	http.HandleFunc("/signup", controllers.SingUpPage)
	http.HandleFunc("/register", controllers.Register)
	http.HandleFunc("/logoutpage", controllers.LogoutPage)
	http.HandleFunc("/logout", controllers.Logout)
	http.HandleFunc("/dashboard", controllers.Dashboard)
	http.HandleFunc("/displayprojects", controllers.DisplayProjects)
	http.HandleFunc("/bugs", controllers.DisplayIssues)
	http.HandleFunc("/showproject", controllers.ShowProject)
	http.HandleFunc("/newproject", controllers.NewProject)
	http.HandleFunc("/editproject", controllers.EditProject)
	http.HandleFunc("/insertproject", controllers.InsertProject)
	http.HandleFunc("/updateproject", controllers.UpdateProject)
	http.HandleFunc("/deleteproject", controllers.DeleteProject)
	http.HandleFunc("/index", controllers.Index)
	http.HandleFunc("/show", controllers.Show)
	http.HandleFunc("/new", controllers.New)
	http.HandleFunc("/edit", controllers.Edit)
	http.HandleFunc("/insert", controllers.Insert)
	http.HandleFunc("/update", controllers.Update)
	http.HandleFunc("/delete", controllers.Delete)
	http.ListenAndServe(":8080", nil)
}

func main() {
	fmt.Println("Starting application")
	handler()
}
