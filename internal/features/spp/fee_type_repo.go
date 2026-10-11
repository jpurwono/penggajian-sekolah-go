package spp

import (
	"context"

	"gorm.io/gorm"
)

type FeeTypeRepository struct{ db *gorm.DB }

func NewFeeTypeRepository(db *gorm.DB) *FeeTypeRepository {
	return &FeeTypeRepository{db: db}
}

func (r *FeeTypeRepository) List(ctx context.Context, q string, active *bool) ([]*FeeTypeResponse, error) {
	where := "WHERE ft.deleted_at IS NULL"
	args := []any{}
	if q != "" {
		where += " AND (ft.code ILIKE ? OR ft.name ILIKE ?)"
		like := "%" + q + "%"
		args = append(args, like, like)
	}
	if active != nil {
		where += " AND ft.is_active = ?"
		args = append(args, *active)
	}
	rows := []*FeeTypeResponse{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT ft.id, ft.code, ft.name, ft.frequency, ft.default_amount,
		       ft.description, ft.is_active, ft.created_at,
		       CASE ft.frequency WHEN 'monthly' THEN 'Bulanan' ELSE 'Tahunan' END AS frequency_label
		FROM spp_fee_types ft `+where+` ORDER BY ft.frequency ASC, ft.name ASC`,
		args...).Scan(&rows).Error
	return rows, err
}

func (r *FeeTypeRepository) Find(ctx context.Context, id int64) (*FeeType, error) {
	var ft FeeType
	if err := r.db.WithContext(ctx).First(&ft, id).Error; err != nil {
		return nil, err
	}
	return &ft, nil
}

func (r *FeeTypeRepository) Create(ctx context.Context, ft *FeeType) error {
	return r.db.WithContext(ctx).Create(ft).Error
}

func (r *FeeTypeRepository) Update(ctx context.Context, ft *FeeType) error {
	return r.db.WithContext(ctx).Model(ft).
		Select("code", "name", "frequency", "default_amount", "description", "is_active").
		Updates(ft).Error
}

func (r *FeeTypeRepository) SoftDelete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Delete(&FeeType{}, id).Error
}

func (r *FeeTypeRepository) CodeExists(ctx context.Context, code string, excludeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&FeeType{}).
		Where("code = ? AND id <> ?", code, excludeID).Count(&count).Error
	return count > 0, err
}

func (r *FeeTypeRepository) CountBillsByType(ctx context.Context, feeTypeID int64) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Bill{}).Where("fee_type_id = ?", feeTypeID).Count(&count).Error
	return count, err
}

func (r *FeeTypeRepository) BillExists(ctx context.Context, studentID int64, feeTypeID int64, month *int, year int) (bool, error) {
	query := r.db.WithContext(ctx).Model(&Bill{}).Where("student_id = ? AND fee_type_id = ? AND period_year = ?", studentID, feeTypeID, year)
	if month != nil {
		query = query.Where("period_month = ?", *month)
	} else {
		query = query.Where("period_month IS NULL")
	}
	var count int64
	err := query.Count(&count).Error
	return count > 0, err
}
