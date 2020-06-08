package models

type Totals struct {
	NumIssues   int
	NumProjects int
	NumCritical int
	NumHigh     int
	NumMedium   int
	NumLow      int
	NumFeature  int
	NumIssue    int
	NumNote     int
}

type Ratios struct {
	Dates         string
	IssuesPerDate int
}
