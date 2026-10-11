-- 000006: SPP — data siswa, tagihan, dan pembayaran

CREATE TABLE students (
  id            BIGSERIAL PRIMARY KEY,
  nis           VARCHAR(30)  NOT NULL UNIQUE,
  nisn          VARCHAR(30),
  full_name     VARCHAR(100) NOT NULL,
  class_name    VARCHAR(30),
  gender        VARCHAR(1)   CONSTRAINT chk_students_gender CHECK (gender IN ('L', 'P')),
  guardian_name VARCHAR(100),
  phone         VARCHAR(20),
  address       TEXT,
  is_active     BOOLEAN      NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
  deleted_at    TIMESTAMPTZ
);
CREATE TRIGGER trg_students_updated_at BEFORE UPDATE ON students
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_students_class_name ON students(class_name);
CREATE INDEX idx_students_is_active ON students(is_active);

CREATE TABLE spp_bills (
  id          BIGSERIAL PRIMARY KEY,
  bill_no     VARCHAR(30)  NOT NULL UNIQUE,
  student_id  BIGINT       NOT NULL REFERENCES students(id),
  name        VARCHAR(100) NOT NULL,
  period_month SMALLINT    CONSTRAINT chk_spp_bills_month CHECK (period_month BETWEEN 1 AND 12),
  period_year SMALLINT,
  amount      BIGINT       NOT NULL CONSTRAINT chk_spp_bills_amount CHECK (amount >= 0),
  paid_amount BIGINT       NOT NULL DEFAULT 0,
  status      VARCHAR(10)  NOT NULL DEFAULT 'unpaid'
              CONSTRAINT chk_spp_bills_status CHECK (status IN ('unpaid', 'partial', 'paid')),
  due_date    DATE,
  note        TEXT,
  created_by  BIGINT REFERENCES users(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_spp_bills_updated_at BEFORE UPDATE ON spp_bills
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_spp_bills_student_id ON spp_bills(student_id);
CREATE INDEX idx_spp_bills_period ON spp_bills(period_year, period_month);
CREATE INDEX idx_spp_bills_status ON spp_bills(status);

CREATE TABLE spp_payments (
  id          BIGSERIAL PRIMARY KEY,
  payment_no  VARCHAR(30) NOT NULL UNIQUE,
  bill_id     BIGINT      NOT NULL REFERENCES spp_bills(id),
  amount      BIGINT      NOT NULL CONSTRAINT chk_spp_payments_amount CHECK (amount > 0),
  paid_date   DATE        NOT NULL,
  method      VARCHAR(15) NOT NULL DEFAULT 'cash'
              CONSTRAINT chk_spp_payments_method CHECK (method IN ('cash', 'transfer', 'lainnya')),
  note        TEXT,
  received_by BIGINT REFERENCES users(id),
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE TRIGGER trg_spp_payments_updated_at BEFORE UPDATE ON spp_payments
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_spp_payments_bill_id ON spp_payments(bill_id);
CREATE INDEX idx_spp_payments_paid_date ON spp_payments(paid_date);
