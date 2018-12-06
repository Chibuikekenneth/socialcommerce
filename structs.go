package main

import "fmt"

type Business struct {
	Logo string
	Slogan string
	DbName string
	DbUser string
	DbPassword string
	UserTable string
	Message string
	Data string
}

func (b Business) dbInfo() string {
	return fmt.Sprintf("user=%s password=%s dbname=%s sslmode=disable",
            		   b.DbUser, b.DbPassword, b.DbName)
}

type ClientContext struct {
	ServerContext Business
	CurrentUser string
	Error string
	HiddenError string
}