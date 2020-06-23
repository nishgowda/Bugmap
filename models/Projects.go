package models

type Projects struct {
	Id            int
	UserId        int
	ProjectName   string
	Description   string
	Technologies  string
	Status        string
	Owner         int
	Users         []string
	UserEmail     string
	Viewer        int
	CollabUids    []int
	GrantedAccess bool
}
