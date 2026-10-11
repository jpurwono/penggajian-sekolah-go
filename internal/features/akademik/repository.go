package akademik

import (
	"context"
	"strings"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) dbCtx(ctx context.Context) *gorm.DB { return r.db.WithContext(ctx) }

// ---------- Mapel ----------

func (r *Repository) ListSubjects(ctx context.Context, q string, limit, offset int64) ([]*Subject, int64, error) {
	query := r.dbCtx(ctx).Model(&Subject{})
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("code ILIKE ? OR name ILIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*Subject
	err := query.Order("name ASC").Offset(int(offset)).Limit(int(limit)).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindSubject(ctx context.Context, id int64) (*Subject, error) {
	var s Subject
	if err := r.dbCtx(ctx).First(&s, id).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *Repository) CreateSubject(ctx context.Context, s *Subject) error {
	return r.dbCtx(ctx).Create(s).Error
}

func (r *Repository) UpdateSubject(ctx context.Context, s *Subject) error {
	return r.dbCtx(ctx).Model(s).Select("code", "name").Updates(s).Error
}

func (r *Repository) SoftDeleteSubject(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&Subject{}, id).Error
}

func (r *Repository) SubjectCodeExists(ctx context.Context, code string, excludeID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Unscoped().Model(&Subject{}).
		Where("code = ? AND id <> ?", code, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CountTeachingsBySubject(ctx context.Context, subjectID int64) (int64, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&TeachingAssignment{}).Where("subject_id = ?", subjectID).Count(&count).Error
	return count, err
}

// ---------- Kelas ----------

func (r *Repository) ListClasses(ctx context.Context, q string, limit, offset int64) ([]*SchoolClass, int64, error) {
	query := r.dbCtx(ctx).Model(&SchoolClass{})
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("name ILIKE ? OR COALESCE(level, '') ILIKE ?", like, like)
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var list []*SchoolClass
	err := query.Order("name ASC").Offset(int(offset)).Limit(int(limit)).Find(&list).Error
	return list, total, err
}

func (r *Repository) FindClass(ctx context.Context, id int64) (*SchoolClass, error) {
	var c SchoolClass
	if err := r.dbCtx(ctx).First(&c, id).Error; err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *Repository) CreateClass(ctx context.Context, c *SchoolClass) error {
	return r.dbCtx(ctx).Create(c).Error
}

func (r *Repository) UpdateClass(ctx context.Context, c *SchoolClass) error {
	return r.dbCtx(ctx).Model(c).Select("name", "level", "description").Updates(c).Error
}

func (r *Repository) SoftDeleteClass(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&SchoolClass{}, id).Error
}

func (r *Repository) ClassNameExists(ctx context.Context, name string, excludeID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Unscoped().Model(&SchoolClass{}).
		Where("name = ? AND id <> ?", name, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CountTeachingsByClass(ctx context.Context, classID int64) (int64, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&TeachingAssignment{}).Where("class_id = ?", classID).Count(&count).Error
	return count, err
}

// ---------- Penugasan Mengajar ----------

const teachingSelect = `
SELECT ta.id, ta.employee_id, ta.subject_id, s.name AS subject_name,
       ta.class_id, c.name AS class_name,
       ta.day_of_week, to_char(ta.start_time, 'HH24:MI') AS start_time,
       to_char(ta.end_time, 'HH24:MI') AS end_time
FROM teaching_assignments ta
JOIN subjects s ON s.id = ta.subject_id
JOIN school_classes c ON c.id = ta.class_id `

func (r *Repository) ListTeachings(ctx context.Context, employeeID int64) ([]*TeachingResponse, error) {
	rows := []*TeachingResponse{}
	err := r.dbCtx(ctx).Raw(
		teachingSelect+"WHERE ta.employee_id = ? ORDER BY ta.day_of_week ASC, ta.start_time ASC",
		employeeID,
	).Scan(&rows).Error
	return rows, err
}

func (r *Repository) FindTeaching(ctx context.Context, id int64) (*TeachingResponse, error) {
	rows := []*TeachingResponse{}
	err := r.dbCtx(ctx).Raw(teachingSelect+"WHERE ta.id = ?", id).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return rows[0], nil
}

func (r *Repository) CreateTeaching(ctx context.Context, t *TeachingAssignment) error {
	return r.dbCtx(ctx).Create(t).Error
}

func (r *Repository) UpdateTeaching(ctx context.Context, t *TeachingAssignment) error {
	return r.dbCtx(ctx).Model(t).
		Select("subject_id", "class_id", "day_of_week", "start_time", "end_time").
		Updates(t).Error
}

func (r *Repository) DeleteTeaching(ctx context.Context, id int64) error {
	return r.dbCtx(ctx).Delete(&TeachingAssignment{}, id).Error
}

func (r *Repository) TeachingExists(ctx context.Context, employeeID int64, req SaveTeachingRequest, excludeID int64) (bool, error) {
	var count int64
	err := r.dbCtx(ctx).Model(&TeachingAssignment{}).
		Where(
			"employee_id = ? AND subject_id = ? AND class_id = ? AND day_of_week = ? AND start_time = ? AND id <> ?",
			employeeID, req.SubjectID, req.ClassID, req.DayOfWeek, req.StartTime, excludeID,
		).Count(&count).Error
	return count > 0, err
}

func (r *Repository) ListTeachingsByEmployeeIDs(ctx context.Context) (map[int64][]*TeachingResponse, error) {
	rows := []*TeachingResponse{}
	err := r.dbCtx(ctx).Raw(
		teachingSelect + "ORDER BY ta.day_of_week ASC, ta.start_time ASC",
	).Scan(&rows).Error
	out := map[int64][]*TeachingResponse{}
	for _, row := range rows {
		out[row.EmployeeID] = append(out[row.EmployeeID], row)
	}
	return out, err
}

func cleanStr(v *string) *string {
	if v == nil {
		return nil
	}
	s := strings.TrimSpace(*v)
	if s == "" {
		return nil
	}
	return &s
}
