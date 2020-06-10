package models

type Projects struct {
	Id           int
	UserId       int
	ProjectName  string
	Description  string
	Technologies string
	Users        []string
}
