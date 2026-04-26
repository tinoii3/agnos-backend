CREATE TABLE IF NOT EXISTS staffs (
    id          BIGSERIAL PRIMARY KEY,
    username    VARCHAR(100) NOT NULL UNIQUE,
    password    TEXT        NOT NULL,
    hospital    VARCHAR(100) NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS patients (
    id              BIGSERIAL PRIMARY KEY,
    first_name_th   VARCHAR(100),
    middle_name_th  VARCHAR(100),
    last_name_th    VARCHAR(100),
    first_name_en   VARCHAR(100),
    middle_name_en  VARCHAR(100),
    last_name_en    VARCHAR(100),
    date_of_birth   DATE,
    patient_hn      VARCHAR(50),
    national_id     VARCHAR(20),
    passport_id     VARCHAR(20),
    phone_number    VARCHAR(20),
    email           VARCHAR(100),
    gender          CHAR(1) CHECK (gender IN ('M', 'F')),
    hospital        VARCHAR(100) NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_patients_hospital    ON patients (hospital);
CREATE INDEX IF NOT EXISTS idx_patients_national_id ON patients (national_id);
CREATE INDEX IF NOT EXISTS idx_patients_passport_id ON patients (passport_id);


INSERT INTO staffs (username, password, hospital) VALUES
('staff123',   '10176e7b7b24d317acfcf8d2064cfd2f24e154f7b5a96603077d5ef813d6a6b6', 'hospital-a'),
('staff456',   '8f5dada329d6ade1fdba5e207b5a81b312ae838801ca287a00e9428620808dce', 'hospital-b'),
('staff789',   '571eeecee49f17206833749f9cd3415b936317dd3d19e74e2218507134a4e2a0', 'hospital-c');

INSERT INTO patients (first_name_th, middle_name_th, last_name_th, first_name_en, last_name_en, date_of_birth, patient_hn, national_id, passport_id, phone_number, email, gender, hospital) VALUES
('สมชาย',   NULL,      'ใจดี',   'Somchai',   'Jaidee',   '1990-01-15', 'HN-A-001', '1100100100001', NULL,          '0812345671', 'somchai@email.com',   'M', 'hospital-a'),
('สมหญิง',  NULL,      'รักดี',  'Somying',   'Rakdee',   '1985-06-20', 'HN-A-002', '1100100100002', NULL,          '0812345672', 'somying@email.com',   'F', 'hospital-a'),
('วิชัย',   'กล้า',   'มีสุข',  'Wichai',    'Meesuk',   '1978-11-30', 'HN-B-001', '1100100100003', NULL,          '0812345673', 'wichai@email.com',    'M', 'hospital-b'),
('นภา',     NULL,      'สดใส',   'Napa',      'Sodsai',   '1995-03-08', 'HN-B-002', NULL,            'PA1234567',  '0812345674', 'napa@email.com',      'F', 'hospital-b'),
('ประสิทธิ์','ชัย',   'เจริญ',  'Prasit',    'Charoen',  '1982-09-12', 'HN-C-001', '1100100100005', NULL,          '0812345675', 'prasit@email.com',    'M', 'hospital-c');