package models

type Workshop struct {
    Creator_Id  string
	Creation_Timestamp string
	Title  string
    Description  string
	Location  string
	Vacancies  int64
	Attendees  []string
	Registration_Deadline string
	Start_Timestamp string
}