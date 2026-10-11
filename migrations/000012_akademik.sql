-- 000012: Akademik — mapel, kelas, penugasan mengajar, & relasi user↔pegawai

CREATE TABLE subjects (
  id         BIGSERIAL PRIMARY KEY,
  code       VARCHAR(20)  NOT NULL UNIQUE,
  name       VARCHAR(100) NOT NULL,
  created_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ  NOT NULL DEFAULT now(),
  deleted_at TIMESTAMPTZ
);
CREATE TRIGGER trg_subjects_updated_at BEFORE UPDATE ON subjects
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE school_classes (
  id          BIGSERIAL PRIMARY KEY,
  name        VARCHAR(30) NOT NULL UNIQUE,
  level       VARCHAR(20),
  description TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  deleted_at  TIMESTAMPTZ
);
CREATE TRIGGER trg_school_classes_updated_at BEFORE UPDATE ON school_classes
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();

CREATE TABLE teaching_assignments (
  id          BIGSERIAL PRIMARY KEY,
  employee_id BIGINT NOT NULL REFERENCES employees(id),
  subject_id  BIGINT NOT NULL REFERENCES subjects(id),
  class_id    BIGINT NOT NULL REFERENCES school_classes(id),
  day_of_week SMALLINT NOT NULL
              CONSTRAINT chk_teaching_day CHECK (day_of_week BETWEEN 1 AND 7),
  start_time  TIME    NOT NULL,
  end_time    TIME    NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT chk_teaching_time CHECK (end_time > start_time)
);
CREATE TRIGGER trg_teaching_assignments_updated_at BEFORE UPDATE ON teaching_assignments
  FOR EACH ROW EXECUTE FUNCTION set_updated_at();
CREATE INDEX idx_teaching_assignments_employee_id ON teaching_assignments(employee_id);
CREATE INDEX idx_teaching_assignments_class_id ON teaching_assignments(class_id);
CREATE INDEX idx_teaching_assignments_subject_id ON teaching_assignments(subject_id);

ALTER TABLE users ADD COLUMN employee_id BIGINT REFERENCES employees(id);
CREATE INDEX idx_users_employee_id ON users(employee_id);
