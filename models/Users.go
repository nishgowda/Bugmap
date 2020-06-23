package models

import "github.com/dgrijalva/jwt-go"

type Users struct {
	Uid           int
	Username      string
	Email         string
	Password      string
	ProjectIDs    []int
	Projects      []string
	Collaborators []string
	CollabUids    []int
	SessionId     []int
}

type Claims struct {
	Email    string
	Username string
	Uid      int
	jwt.StandardClaims
}
