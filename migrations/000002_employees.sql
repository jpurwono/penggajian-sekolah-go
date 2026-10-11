-- 000002: Data pegawai (guru & staff)

CREATE TABLE employees (
  id            BIGSERIAL PRIMARY KEY,
  employee_no   VARCHAR(30)  NOT NULL UNIQUE,
  full_name     VARCHAR(100) NOT NULL,
  gender        VARCHAR(1)   CONSTRAINT chk_employees_gender CHECK (gender IN ('L', 'P')),
  birth_date    DATE,
  address       TEXT,
  phone         VARCHAR(20),
  email         VARCHAR(100),
  employee_type VARCHAR(20)  NOT NULL
                CONSTRAINT chk_employees_type CHECK (employee_type IN ('tetap', 'pppk', 'honorer', 'staff_tu')),
  job_title     VARCHAR(100),
  join_date     DATE         NOT NULL,
  end_date      DATE,
  base_salary   BIGINT       NOT NULL DEFAULT 0
                CONSTRAINT chk_employees_base_salary CHECK (base_salary >= 0),
  bank_name     VARCHAR(50),
  bank_account  VARCHAR(30),
  photo_url     VARCHAR(255),
  is_active     BOOLEAN      NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
  deleted_at    TIMESTAMPTZ
);
CREATE TRIGGER trg_employees_updated_at BEFORE UPDATE ON employees
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_employees_employee_type ON employees(employee_type);
CREATE INDEX idx_employees_is_active ON employees(is_active);
