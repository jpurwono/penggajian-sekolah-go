-- 000003: Penggajian — komponen, komponen pegawai, periode, slip & item slip

CREATE TABLE pay_components (
  id          BIGSERIAL PRIMARY KEY,
  code        VARCHAR(20)  NOT NULL UNIQUE,
  name        VARCHAR(100) NOT NULL,
  type        VARCHAR(10)  NOT NULL
              CONSTRAINT chk_pay_components_type CHECK (type IN ('earning', 'deduction')),
  description TEXT,
  is_active   BOOLEAN      NOT NULL DEFAULT true,
  created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
  deleted_at  TIMESTAMPTZ
);
CREATE TRIGGER trg_pay_components_updated_at BEFORE UPDATE ON pay_components
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE employee_components (
  id           BIGSERIAL PRIMARY KEY,
  employee_id  BIGINT      NOT NULL REFERENCES employees(id),
  component_id BIGINT      NOT NULL REFERENCES pay_components(id),
  amount       BIGINT      NOT NULL
               CONSTRAINT chk_employee_components_amount CHECK (amount >= 0),
  note         TEXT,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_employee_components UNIQUE (employee_id, component_id)
);
CREATE TRIGGER trg_employee_components_updated_at BEFORE UPDATE ON employee_components
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_employee_components_component_id ON employee_components(component_id);

CREATE TABLE payroll_periods (
  id           BIGSERIAL PRIMARY KEY,
  name         VARCHAR(50) NOT NULL,
  period_month SMALLINT    NOT NULL
               CONSTRAINT chk_periods_month CHECK (period_month BETWEEN 1 AND 12),
  period_year  SMALLINT    NOT NULL,
  status       VARCHAR(15) NOT NULL DEFAULT 'draft'
               CONSTRAINT chk_periods_status CHECK (status IN ('draft', 'approved', 'paid', 'cancelled')),
  paid_at      TIMESTAMPTZ,
  note         TEXT,
  created_by   BIGINT REFERENCES users(id),
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at   TIMESTAMPTZ,
  CONSTRAINT uq_periods_year_month UNIQUE (period_year, period_month)
);
CREATE TRIGGER trg_payroll_periods_updated_at BEFORE UPDATE ON payroll_periods
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE payslips (
  id              BIGSERIAL PRIMARY KEY,
  period_id       BIGINT      NOT NULL REFERENCES payroll_periods(id),
  employee_id     BIGINT      NOT NULL REFERENCES employees(id),
  slip_no         VARCHAR(30) NOT NULL UNIQUE,
  total_earning   BIGINT      NOT NULL DEFAULT 0,
  total_deduction BIGINT      NOT NULL DEFAULT 0,
  take_home_pay   BIGINT      NOT NULL DEFAULT 0,
  note            TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT uq_payslips_period_employee UNIQUE (period_id, employee_id)
);
CREATE TRIGGER trg_payslips_updated_at BEFORE UPDATE ON payslips
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_payslips_employee_id ON payslips(employee_id);

CREATE TABLE payslip_items (
  id            BIGSERIAL PRIMARY KEY,
  payslip_id    BIGINT      NOT NULL REFERENCES payslips(id) ON DELETE CASCADE,
  component_id  BIGINT      REFERENCES pay_components(id),
  code_snapshot VARCHAR(20) NOT NULL,
  name_snapshot VARCHAR(100) NOT NULL,
  type          VARCHAR(10) NOT NULL
                CONSTRAINT chk_payslip_items_type CHECK (type IN ('earning', 'deduction')),
  amount        BIGINT      NOT NULL,
  note          TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_payslip_items_updated_at BEFORE UPDATE ON payslip_items
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_payslip_items_payslip_id ON payslip_items(payslip_id);
CREATE INDEX idx_payslip_items_component_id ON payslip_items(component_id);
