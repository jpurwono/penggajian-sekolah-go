package akademik

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"gorm.io/gorm"

	"penggajian/server/internal/platform/httpx"
)

var ErrNotFound = errors.New("data tidak ditemukan")

var timePattern = regexp.MustCompile(`^([01]\d|2[0-3]):[0-5]\d$`)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func mapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

// ---------- Mapel ----------

func (s *Service) ListSubjects(ctx context.Context, page, per int64, q string) (*SubjectListResult, error) {
	list, total, err := s.repo.ListSubjects(ctx, q, per, (page-1)*per)
	if err != nil {
		return nil, err
	}
	resp := make([]*SubjectResponse, 0, len(list))
	for _, sub := range list {
		resp = append(resp, &SubjectResponse{ID: sub.ID, Code: sub.Code, Name: sub.Name, CreatedAt: sub.CreatedAt})
	}
	return &SubjectListResult{Subjects: resp, Meta: httpx.NewMeta(page, per, total)}, nil
}

func (s *Service) validateSubject(ctx context.Context, req SaveSubjectRequest, excludeID int64) (*Subject, error) {
	code := strings.ToUpper(strings.TrimSpace(req.Code))
	name := strings.TrimSpace(req.Name)
	if code == "" || name == "" {
		return nil, errors.New("kode & nama mapel wajib diisi")
	}
	if exists, err := s.repo.SubjectCodeExists(ctx, code, excludeID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("kode mapel sudah digunakan")
	}
	return &Subject{Code: code, Name: name}, nil
}

func (s *Service) CreateSubject(ctx context.Context, req SaveSubjectRequest) (*SubjectResponse, error) {
	sub, err := s.validateSubject(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateSubject(ctx, sub); err != nil {
		return nil, err
	}
	return &SubjectResponse{ID: sub.ID, Code: sub.Code, Name: sub.Name, CreatedAt: sub.CreatedAt}, nil
}

func (s *Service) UpdateSubject(ctx context.Context, id int64, req SaveSubjectRequest) (*SubjectResponse, error) {
	if _, err := s.repo.FindSubject(ctx, id); err != nil {
		return nil, mapNotFound(err)
	}
	sub, err := s.validateSubject(ctx, req, id)
	if err != nil {
		return nil, err
	}
	sub.ID = id
	if err := s.repo.UpdateSubject(ctx, sub); err != nil {
		return nil, err
	}
	updated, err := s.repo.FindSubject(ctx, id)
	if err != nil {
		return nil, err
	}
	return &SubjectResponse{ID: updated.ID, Code: updated.Code, Name: updated.Name, CreatedAt: updated.CreatedAt}, nil
}

func (s *Service) DeleteSubject(ctx context.Context, id int64) error {
	if _, err := s.repo.FindSubject(ctx, id); err != nil {
		return mapNotFound(err)
	}
	if count, err := s.repo.CountTeachingsBySubject(ctx, id); err != nil {
		return err
	} else if count > 0 {
		return errors.New("mapel masih dipakai di penugasan mengajar")
	}
	return s.repo.SoftDeleteSubject(ctx, id)
}

// ---------- Kelas ----------

func (s *Service) ListClasses(ctx context.Context, page, per int64, q string) (*ClassListResult, error) {
	list, total, err := s.repo.ListClasses(ctx, q, per, (page-1)*per)
	if err != nil {
		return nil, err
	}
	resp := make([]*ClassResponse, 0, len(list))
	for _, cls := range list {
		resp = append(resp, &ClassResponse{ID: cls.ID, Name: cls.Name, Level: cls.Level, Description: cls.Description, CreatedAt: cls.CreatedAt})
	}
	return &ClassListResult{Classes: resp, Meta: httpx.NewMeta(page, per, total)}, nil
}

func (s *Service) validateClass(ctx context.Context, req SaveClassRequest, excludeID int64) (*SchoolClass, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.New("nama kelas wajib diisi")
	}
	if exists, err := s.repo.ClassNameExists(ctx, name, excludeID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("nama kelas sudah digunakan")
	}
	return &SchoolClass{Name: name, Level: cleanStr(req.Level), Description: cleanStr(req.Description)}, nil
}

func (s *Service) CreateClass(ctx context.Context, req SaveClassRequest) (*ClassResponse, error) {
	cls, err := s.validateClass(ctx, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateClass(ctx, cls); err != nil {
		return nil, err
	}
	return &ClassResponse{ID: cls.ID, Name: cls.Name, Level: cls.Level, Description: cls.Description, CreatedAt: cls.CreatedAt}, nil
}

func (s *Service) UpdateClass(ctx context.Context, id int64, req SaveClassRequest) (*ClassResponse, error) {
	if _, err := s.repo.FindClass(ctx, id); err != nil {
		return nil, mapNotFound(err)
	}
	cls, err := s.validateClass(ctx, req, id)
	if err != nil {
		return nil, err
	}
	cls.ID = id
	if err := s.repo.UpdateClass(ctx, cls); err != nil {
		return nil, err
	}
	updated, err := s.repo.FindClass(ctx, id)
	if err != nil {
		return nil, err
	}
	return &ClassResponse{ID: updated.ID, Name: updated.Name, Level: updated.Level, Description: updated.Description, CreatedAt: updated.CreatedAt}, nil
}

func (s *Service) DeleteClass(ctx context.Context, id int64) error {
	if _, err := s.repo.FindClass(ctx, id); err != nil {
		return mapNotFound(err)
	}
	if count, err := s.repo.CountTeachingsByClass(ctx, id); err != nil {
		return err
	} else if count > 0 {
		return errors.New("kelas masih dipakai di penugasan mengajar")
	}
	return s.repo.SoftDeleteClass(ctx, id)
}

// ---------- Penugasan Mengajar ----------

func (s *Service) ListTeachings(ctx context.Context, employeeID int64) ([]*TeachingResponse, error) {
	rows, err := s.repo.ListTeachings(ctx, employeeID)
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		row.DayName = DayNames[row.DayOfWeek]
	}
	return rows, nil
}

func (s *Service) validateTeaching(ctx context.Context, employeeID int64, req SaveTeachingRequest, excludeID int64) (*TeachingAssignment, error) {
	if employeeID <= 0 {
		return nil, errors.New("pegawai wajib dipilih")
	}
	if req.SubjectID <= 0 {
		return nil, errors.New("mapel wajib dipilih")
	}
	if req.ClassID <= 0 {
		return nil, errors.New("kelas wajib dipilih")
	}
	if !IsValidDay(req.DayOfWeek) {
		return nil, errors.New("hari harus 1 (Senin) s/d 7 (Minggu)")
	}
	start := strings.TrimSpace(req.StartTime)
	end := strings.TrimSpace(req.EndTime)
	if !timePattern.MatchString(start) || !timePattern.MatchString(end) {
		return nil, errors.New("format jam harus HH:MM (contoh 07:30)")
	}
	sh, _ := time.Parse("15:04", start)
	eh, _ := time.Parse("15:04", end)
	if !eh.After(sh) {
		return nil, errors.New("jam selesai harus setelah jam mulai")
	}
	if _, err := s.repo.FindSubject(ctx, req.SubjectID); err != nil {
		return nil, errors.New("mapel tidak ditemukan")
	}
	if _, err := s.repo.FindClass(ctx, req.ClassID); err != nil {
		return nil, errors.New("kelas tidak ditemukan")
	}
	if exists, err := s.repo.TeachingExists(ctx, employeeID, req, excludeID); err != nil {
		return nil, err
	} else if exists {
		return nil, errors.New("penugasan sama (mapel, kelas, hari, jam) sudah ada")
	}
	return &TeachingAssignment{
		EmployeeID: employeeID, SubjectID: req.SubjectID, ClassID: req.ClassID,
		DayOfWeek: req.DayOfWeek, StartTime: start, EndTime: end,
	}, nil
}

func (s *Service) CreateTeaching(ctx context.Context, employeeID int64, req SaveTeachingRequest) (*TeachingResponse, error) {
	t, err := s.validateTeaching(ctx, employeeID, req, 0)
	if err != nil {
		return nil, err
	}
	if err := s.repo.CreateTeaching(ctx, t); err != nil {
		return nil, err
	}
	return s.GetTeaching(ctx, t.ID)
}

func (s *Service) GetTeaching(ctx context.Context, id int64) (*TeachingResponse, error) {
	row, err := s.repo.FindTeaching(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	row.DayName = DayNames[row.DayOfWeek]
	return row, nil
}

func (s *Service) UpdateTeaching(ctx context.Context, id int64, req SaveTeachingRequest) (*TeachingResponse, error) {
	existing, err := s.repo.FindTeaching(ctx, id)
	if err != nil {
		return nil, mapNotFound(err)
	}
	t, err := s.validateTeaching(ctx, existing.EmployeeID, req, id)
	if err != nil {
		return nil, err
	}
	t.ID = id
	if err := s.repo.UpdateTeaching(ctx, t); err != nil {
		return nil, err
	}
	return s.GetTeaching(ctx, id)
}

func (s *Service) DeleteTeaching(ctx context.Context, id int64) error {
	if _, err := s.repo.FindTeaching(ctx, id); err != nil {
		return mapNotFound(err)
	}
	return s.repo.DeleteTeaching(ctx, id)
}
