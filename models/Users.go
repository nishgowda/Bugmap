package models

import "github.com/dgrijalva/jwt-go"

type Users struct {
	Uid      int
	Username string
	Email    string
	Password string
}

type Claims struct {
	Email string
	Uid   int
	jwt.StandardClaims
}
