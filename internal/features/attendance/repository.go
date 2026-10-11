package attendance

import (
	"context"
	"fmt"
	"math"
	"time"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

const recordSelect = `
SELECT ar.id, ar.employee_id, e.full_name AS employee_name, e.employee_no,
       ar.jenis, ar.teaching_id, ar.latitude, ar.longitude, ar.accuracy_m,
       ar.is_mock, ar.distance_m, ar.recorded_at,
       ar.photo_filename IS NOT NULL AS has_photo,
       ta.class_id, sc.name AS class_name,
       ta.subject_id, su.name AS subject_name,
       ta.day_of_week
FROM attendance_records ar
JOIN employees e ON e.id = ar.employee_id
LEFT JOIN teaching_assignments ta ON ta.id = ar.teaching_id
LEFT JOIN school_classes sc ON sc.id = ta.class_id
LEFT JOIN subjects su ON su.id = ta.subject_id `

type recordRow struct {
	AttendanceRecord
	EmployeeName string  `gorm:"column:employee_name"`
	EmployeeNo   string  `gorm:"column:employee_no"`
	HasPhoto     bool    `gorm:"column:has_photo"`
	ClassName    *string `gorm:"column:class_name"`
	SubjectName  *string `gorm:"column:subject_name"`
	DayOfWeek    *int    `gorm:"column:day_of_week"`
}

type Filter struct {
	EmployeeID *int64
	Jenis      string
	DateFrom   string
	DateTo     string
}

func (r *Repository) List(ctx context.Context, f Filter, limit, offset int64) ([]*AttendanceResponse, int64, error) {
	where := "WHERE 1=1"
	args := []any{}
	if f.EmployeeID != nil {
		where += " AND ar.employee_id = ?"
		args = append(args, *f.EmployeeID)
	}
	if f.Jenis != "" {
		where += " AND ar.jenis = ?"
		args = append(args, f.Jenis)
	}
	if f.DateFrom != "" {
		where += " AND ar.recorded_at >= ?::date"
		args = append(args, f.DateFrom)
	}
	if f.DateTo != "" {
		where += " AND ar.recorded_at < (?::date + INTERVAL '1 day')"
		args = append(args, f.DateTo)
	}

	var total int64
	if err := r.db.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM attendance_records ar JOIN employees e ON e.id = ar.employee_id "+where,
		args...,
	).Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	rows := []*recordRow{}
	err := r.db.WithContext(ctx).Raw(
		recordSelect+where+" ORDER BY ar.recorded_at DESC LIMIT ? OFFSET ?",
		append(args, limit, offset)...,
	).Scan(&rows).Error
	if err != nil {
		return nil, 0, err
	}

	resp := make([]*AttendanceResponse, 0, len(rows))
	for _, row := range rows {
		resp = append(resp, rowToResponse(row))
	}
	return resp, total, nil
}

func (r *Repository) Find(ctx context.Context, id int64) (*recordRow, error) {
	rows := []*recordRow{}
	err := r.db.WithContext(ctx).Raw(recordSelect+"WHERE ar.id = ?", id).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return rows[0], nil
}

func (r *Repository) Create(ctx context.Context, rec *AttendanceRecord) error {
	return r.db.WithContext(ctx).Create(rec).Error
}

func (r *Repository) HadirToday(ctx context.Context, employeeID int64, date time.Time) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&AttendanceRecord{}).
		Where("employee_id = ? AND jenis = ? AND recorded_at::date = ?::date", employeeID, JenisHadir, date).
		Count(&count).Error
	return count > 0, err
}

// SchoolGeo mengembalikan koordinat sekolah + radius toleransi.
func (r *Repository) SchoolGeo(ctx context.Context) (lat, lon float64, radiusM int, err error) {
	row := struct {
		Latitude  *float64
		Longitude *float64
		Radius    int
	}{}
	err = r.db.WithContext(ctx).Raw(
		`SELECT latitude, longitude, geofence_radius_m FROM school_profiles WHERE id = 1`,
	).Scan(&row).Error
	if err != nil {
		return 0, 0, 0, err
	}
	if row.Latitude != nil {
		lat = *row.Latitude
	}
	if row.Longitude != nil {
		lon = *row.Longitude
	}
	radiusM = row.Radius
	if radiusM <= 0 {
		radiusM = 200
	}
	return lat, lon, radiusM, nil
}

func (r *Repository) DailySummary(ctx context.Context, date string) ([]*DailySummaryRow, error) {
	rows := []*DailySummaryRow{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT e.id AS employee_id, e.full_name AS employee_name, e.employee_no,
		       COALESCE(SUM(CASE WHEN ar.jenis = 'hadir' THEN 1 ELSE 0 END), 0) AS hadir_count,
		       COALESCE(SUM(CASE WHEN ar.jenis = 'kelas' THEN 1 ELSE 0 END), 0) AS kelas_count,
		       MAX(CASE WHEN ar.jenis = 'hadir' THEN ar.recorded_at END) AS last_hadir,
		       MAX(CASE WHEN ar.jenis = 'kelas' THEN ar.recorded_at END) AS last_kelas
		FROM employees e
		LEFT JOIN attendance_records ar
		  ON ar.employee_id = e.id AND ar.recorded_at::date = ?::date
		WHERE e.deleted_at IS NULL AND e.is_active = true
		GROUP BY e.id, e.full_name, e.employee_no
		ORDER BY e.full_name ASC
	`, date).Scan(&rows).Error
	return rows, err
}

func (r *Repository) SavePhoto(ctx context.Context, id int64, filename string) error {
	return r.db.WithContext(ctx).Exec(
		`UPDATE attendance_records SET photo_filename = ? WHERE id = ?`, filename, id,
	).Error
}

func (r *Repository) PhotoPath(ctx context.Context, id int64) (string, error) {
	var filename *string
	err := r.db.WithContext(ctx).Raw(
		`SELECT photo_filename FROM attendance_records WHERE id = ?`, id,
	).Scan(&filename).Error
	if err != nil || filename == nil {
		return "", err
	}
	return *filename, nil
}

var dayNames = map[int]string{1: "Senin", 2: "Selasa", 3: "Rabu", 4: "Kamis", 5: "Jumat", 6: "Sabtu", 7: "Minggu"}

func rowToResponse(row *recordRow) *AttendanceResponse {
	resp := &AttendanceResponse{
		ID:           row.ID,
		EmployeeID:   row.EmployeeID,
		EmployeeName: row.EmployeeName,
		EmployeeNo:   row.EmployeeNo,
		Jenis:        row.Jenis,
		JenisLabel:   map[string]string{"hadir": "Absen Kehadiran", "kelas": "Absen Kelas"}[row.Jenis],
		TeachingID:   row.TeachingID,
		Latitude:     row.Latitude,
		Longitude:    row.Longitude,
		AccuracyM:    row.AccuracyM,
		IsMock:       row.IsMock,
		DistanceM:    row.DistanceM,
		RecordedAt:   row.RecordedAt,
		HasPhoto:     row.HasPhoto,
	}
	if row.ClassName != nil {
		resp.ClassName = row.ClassName
		resp.SubjectName = row.SubjectName
		if row.DayOfWeek != nil {
			dn := dayNames[*row.DayOfWeek]
			resp.DayName = &dn
		}
	}
	return resp
}

func HaversineMeters(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLon/2)*math.Sin(dLon/2)
	return 2 * R * math.Asin(math.Sqrt(a))
}

var _ = fmt.Sprintf
