package store

import "context"

type Staff struct {
	ID             int64
	Username       string
	HashedPassword string
	Hospital       string
}

type Patient struct {
	ID           int64  `json:"id"`
	FirstNameTH  string `json:"first_name_th"`
	MiddleNameTH string `json:"middle_name_th"`
	LastNameTH   string `json:"last_name_th"`
	FirstNameEN  string `json:"first_name_en"`
	MiddleNameEN string `json:"middle_name_en"`
	LastNameEN   string `json:"last_name_en"`
	DateOfBirth  string `json:"date_of_birth"`
	PatientHN    string `json:"patient_hn"`
	NationID     string `json:"nation_id"`
	PassportID   string `json:"passport_id"`
	PhoneNumber  string `json:"phone_number"`
	Email        string `json:"email"`
	Gender       string `json:"gender"`
	Hospital     string `json:"-"`
}

type PatientFilters struct {
	ID          string
	NationalID  string
	PassportID  string
	FirstName   string
	MiddleName  string
	LastName    string
	DateOfBirth string
	PhoneNumber string
	Email       string
}

type Store interface {
	CreateStaff(ctx context.Context, username, hashedPassword, hospital string) error
	GetStaffByUsername(ctx context.Context, username string) (Staff, error)
	SearchPatients(ctx context.Context, hospital string, filters PatientFilters) ([]Patient, error)
}
