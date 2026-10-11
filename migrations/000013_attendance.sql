-- 000013: Tabel absensi + koordinat sekolah untuk validasi geofence

CREATE TABLE attendance_records (
    id             BIGSERIAL PRIMARY KEY,
    employee_id    BIGINT       NOT NULL REFERENCES employees(id),
    jenis          VARCHAR(10)  NOT NULL
                   CONSTRAINT chk_attendance_jenis CHECK (jenis IN ('hadir', 'kelas')),
    teaching_id    BIGINT       REFERENCES teaching_assignments(id),
    photo_filename VARCHAR(255),
    latitude       DOUBLE PRECISION NOT NULL,
    longitude      DOUBLE PRECISION NOT NULL,
    accuracy_m     FLOAT,
    is_mock        BOOLEAN      NOT NULL DEFAULT false,
    distance_m     FLOAT,
    device_info    TEXT,
    recorded_at    TIMESTAMPTZ  NOT NULL,
    created_at     TIMESTAMPTZ  NOT NULL DEFAULT now()
);
CREATE INDEX idx_attendance_employee_id ON attendance_records(employee_id);
CREATE INDEX idx_attendance_recorded_at ON attendance_records(recorded_at);
CREATE INDEX idx_attendance_jenis ON attendance_records(jenis);
CREATE INDEX idx_attendance_teaching_id ON attendance_records(teaching_id);

-- Koordinat sekolah + toleransi geofence di profil sekolah
ALTER TABLE school_profiles ADD COLUMN latitude DOUBLE PRECISION;
ALTER TABLE school_profiles ADD COLUMN longitude DOUBLE PRECISION;
ALTER TABLE school_profiles ADD COLUMN geofence_radius_m INT NOT NULL DEFAULT 200;
