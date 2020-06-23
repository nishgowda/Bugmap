package models

type Code_Sessions struct {
	Id            int
	Name          string
	Url           string
	Language      string
	Collaborators []string
	Email         string
}
