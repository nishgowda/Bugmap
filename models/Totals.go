package models

type Totals struct {
	NumIssues     int
	NumProjects   int
	NumCritical   int
	NumHigh       int
	NumMedium     int
	NumLow        int
	IssuesPerDate []int
	Dates         []string
}
