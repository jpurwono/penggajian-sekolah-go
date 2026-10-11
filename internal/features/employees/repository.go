package employees

import (
	"context"

	"gorm.io/gorm"
)

type ListFilter struct {
	Page         int64
	PerPage      int64
	Q            string
	EmployeeType string
	Active       *bool
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, f ListFilter) ([]*Employee, int64, error) {
	query := r.db.WithContext(ctx).Model(&Employee{})

	if f.Q != "" {
		like := "%" + f.Q + "%"
		query = query.Where(
			"employee_no ILIKE ? OR full_name ILIKE ? OR COALESCE(email, '') ILIKE ?",
			like, like, like,
		)
	}
	if f.EmployeeType != "" {
		query = query.Where("employee_type = ?", f.EmployeeType)
	}
	if f.Active != nil {
		query = query.Where("is_active = ?", *f.Active)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*Employee
	err := query.
		Order("full_name ASC").
		Offset(int((f.Page - 1) * f.PerPage)).
		Limit(int(f.PerPage)).
		Find(&list).Error
	return list, total, err
}

func (r *Repository) Find(ctx context.Context, id int64) (*Employee, error) {
	var e Employee
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		return nil, err
	}
	return &e, nil
}

func (r *Repository) Create(ctx context.Context, e *Employee) error {
	return r.db.WithContext(ctx).Create(e).Error
}

func (r *Repository) Update(ctx context.Context, e *Employee) error {
	return r.db.WithContext(ctx).Model(e).
		Select(
			"employee_no", "full_name", "gender", "birth_date", "address", "phone", "email",
			"employee_type", "job_title", "join_date", "end_date", "base_salary",
			"bank_name", "bank_account", "photo_url", "is_active",
		).
		Updates(e).Error
}

func (r *Repository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&Employee{}, id).Error
}

func (r *Repository) EmployeeNoExists(ctx context.Context, employeeNo string, excludeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&Employee{}).
		Where("employee_no = ? AND id <> ?", employeeNo, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *Repository) CountActive(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Employee{}).Where("is_active = ?", true).Count(&count).Error
	return count, err
}
