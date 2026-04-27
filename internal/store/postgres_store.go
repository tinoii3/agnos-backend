package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(pool *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{pool: pool}
}

func (s *PostgresStore) CreateStaff(ctx context.Context, username, hashedPassword, hospital string) error {
	_, err := s.pool.Exec(ctx, `
		INSERT INTO staffs (username, password, hospital)
		VALUES ($1, $2, $3)
	`, username, hashedPassword, hospital)
	if err != nil {
		return fmt.Errorf("insert staff: %w", err)
	}
	return nil
}

func (s *PostgresStore) GetStaffByUsername(ctx context.Context, username string) (Staff, error) {
	var staff Staff
	err := s.pool.QueryRow(ctx, `
		SELECT id, username, password, hospital
		FROM staffs
		WHERE username = $1
	`, username).Scan(&staff.ID, &staff.Username, &staff.HashedPassword, &staff.Hospital)
	if err != nil {
		if err == pgx.ErrNoRows {
			return Staff{}, err
		}
		return Staff{}, fmt.Errorf("query staff: %w", err)
	}
	return staff, nil
}

func (s *PostgresStore) SearchPatients(ctx context.Context, hospital string, filters PatientFilters) ([]Patient, error) {
	base := `
		SELECT
			id,
			COALESCE(first_name_th, '') AS first_name_th,
			COALESCE(middle_name_th, '') AS middle_name_th,
			COALESCE(last_name_th, '') AS last_name_th,
			COALESCE(first_name_en, '') AS first_name_en,
			COALESCE(middle_name_en, '') AS middle_name_en,
			COALESCE(last_name_en, '') AS last_name_en,
			COALESCE(date_of_birth::text, '') AS date_of_birth,
			COALESCE(patient_hn, '') AS patient_hn,
			COALESCE(national_id, '') AS nation_id,
			COALESCE(passport_id, '') AS passport_id,
			COALESCE(phone_number, '') AS phone_number,
			COALESCE(email, '') AS email,
			COALESCE(gender::text, '') AS gender,
			hospital
		FROM patients
		WHERE hospital = $1
	`
	args := []any{hospital}
	nextArg := 2

	appendFilter := func(field, value string) {
		if value == "" {
			return
		}
		base += fmt.Sprintf(" AND %s ILIKE $%d", field, nextArg)
		args = append(args, "%"+value+"%")
		nextArg++
	}

	if filters.ID != "" {
		base += fmt.Sprintf(" AND (national_id ILIKE $%d OR passport_id ILIKE $%d)", nextArg, nextArg+1)
		like := "%" + filters.ID + "%"
		args = append(args, like, like)
		nextArg += 2
	}

	appendFilter("national_id", filters.NationalID)
	appendFilter("passport_id", filters.PassportID)
	appendFilter("COALESCE(first_name_th, '') || ' ' || COALESCE(first_name_en, '')", filters.FirstName)
	appendFilter("COALESCE(middle_name_th, '') || ' ' || COALESCE(middle_name_en, '')", filters.MiddleName)
	appendFilter("COALESCE(last_name_th, '') || ' ' || COALESCE(last_name_en, '')", filters.LastName)
	appendFilter("date_of_birth::text", filters.DateOfBirth)
	appendFilter("email", filters.Email)
	appendFilter("phone_number", filters.PhoneNumber)

	base += " ORDER BY id"
	rows, err := s.pool.Query(ctx, base, args...)
	if err != nil {
		return nil, fmt.Errorf("query patients: %w", err)
	}
	defer rows.Close()

	patients := make([]Patient, 0)
	for rows.Next() {
		var patient Patient
		if err := rows.Scan(
			&patient.ID,
			&patient.FirstNameTH,
			&patient.MiddleNameTH,
			&patient.LastNameTH,
			&patient.FirstNameEN,
			&patient.MiddleNameEN,
			&patient.LastNameEN,
			&patient.DateOfBirth,
			&patient.PatientHN,
			&patient.NationID,
			&patient.PassportID,
			&patient.PhoneNumber,
			&patient.Email,
			&patient.Gender,
			&patient.Hospital,
		); err != nil {
			return nil, fmt.Errorf("scan patient: %w", err)
		}
		patients = append(patients, patient)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows err: %w", err)
	}

	return patients, nil
}

var _ Store = (*PostgresStore)(nil)

func IsUniqueViolation(err error) bool {
	if err == nil {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}

	return false
}
