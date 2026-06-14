package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"coupon-service/internal/constants"
	"coupon-service/internal/model"
	"coupon-service/internal/pkg/logger"
)

type Repository struct {
	db *sql.DB
}

func New(dbPath string) (*Repository, error) {
	db, err := sql.Open("sqlite3", dbPath+"?_journal=WAL&_timeout=5000&_fk=1&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open db failed: %w", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL; PRAGMA busy_timeout=5000;"); err != nil {
		return nil, fmt.Errorf("set pragma failed: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db failed: %w", err)
	}
	repo := &Repository{db: db}
	if err := repo.initTables(); err != nil {
		return nil, fmt.Errorf("init tables failed: %w", err)
	}
	return repo, nil
}

func (r *Repository) DB() *sql.DB {
	return r.db
}

func (r *Repository) initTables() error {
	sqls := []string{
		`CREATE TABLE IF NOT EXISTS coupon_template (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			type INTEGER NOT NULL DEFAULT 0,
			value REAL NOT NULL,
			threshold REAL NOT NULL DEFAULT 0,
			total_count INTEGER NOT NULL,
			remaining_count INTEGER NOT NULL,
			per_user_limit INTEGER NOT NULL DEFAULT 1,
			valid_from DATETIME NOT NULL,
			valid_to DATETIME NOT NULL,
			applicable_level INTEGER NOT NULL DEFAULT 0,
			category INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS coupon_record (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			template_id INTEGER NOT NULL,
			user_id INTEGER NOT NULL,
			status INTEGER NOT NULL DEFAULT 0,
			value REAL NOT NULL,
			threshold REAL NOT NULL DEFAULT 0,
			valid_from DATETIME NOT NULL,
			valid_to DATETIME NOT NULL,
			used_at DATETIME,
			order_id TEXT,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS user (
			id INTEGER PRIMARY KEY,
			level INTEGER NOT NULL DEFAULT 0,
			new_user_gift_sent INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_coupon_record_user_template ON coupon_record(user_id, template_id)`,
		`CREATE INDEX IF NOT EXISTS idx_coupon_record_status ON coupon_record(status)`,
		`CREATE INDEX IF NOT EXISTS idx_coupon_record_valid_to ON coupon_record(valid_to)`,
	}
	for _, s := range sqls {
		if _, err := r.db.Exec(s); err != nil {
			return fmt.Errorf("exec sql [%s] failed: %w", s, err)
		}
	}
	return nil
}

func (r *Repository) CreateTemplate(tpl *model.CouponTemplate) error {
	now := logger.Now()
	tpl.CreatedAt = now
	tpl.UpdatedAt = now
	tpl.RemainingCount = tpl.TotalCount
	res, err := r.db.Exec(`
		INSERT INTO coupon_template
		(name, type, value, threshold, total_count, remaining_count, per_user_limit,
		 valid_from, valid_to, applicable_level, category, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tpl.Name, tpl.Type, tpl.Value, tpl.Threshold, tpl.TotalCount, tpl.RemainingCount,
		tpl.PerUserLimit, tpl.ValidFrom, tpl.ValidTo, tpl.ApplicableLevel, tpl.Category,
		tpl.CreatedAt, tpl.UpdatedAt,
	)
	if err != nil {
		return err
	}
	tpl.ID, err = res.LastInsertId()
	return err
}

func (r *Repository) GetTemplate(id int64) (*model.CouponTemplate, error) {
	row := r.db.QueryRow(`
		SELECT id, name, type, value, threshold, total_count, remaining_count, per_user_limit,
		       valid_from, valid_to, applicable_level, category, created_at, updated_at
		FROM coupon_template WHERE id = ?`, id)
	tpl := &model.CouponTemplate{}
	err := row.Scan(&tpl.ID, &tpl.Name, &tpl.Type, &tpl.Value, &tpl.Threshold,
		&tpl.TotalCount, &tpl.RemainingCount, &tpl.PerUserLimit, &tpl.ValidFrom,
		&tpl.ValidTo, &tpl.ApplicableLevel, &tpl.Category, &tpl.CreatedAt, &tpl.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return tpl, err
}

func (r *Repository) ListTemplates(page, size int) ([]*model.CouponTemplate, int64, error) {
	var total int64
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM coupon_template`).Scan(&total); err != nil {
		return nil, 0, err
	}
	offset := (page - 1) * size
	rows, err := r.db.Query(`
		SELECT id, name, type, value, threshold, total_count, remaining_count, per_user_limit,
		       valid_from, valid_to, applicable_level, category, created_at, updated_at
		FROM coupon_template ORDER BY id DESC LIMIT ? OFFSET ?`, size, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	var list []*model.CouponTemplate
	for rows.Next() {
		tpl := &model.CouponTemplate{}
		if err := rows.Scan(&tpl.ID, &tpl.Name, &tpl.Type, &tpl.Value, &tpl.Threshold,
			&tpl.TotalCount, &tpl.RemainingCount, &tpl.PerUserLimit, &tpl.ValidFrom,
			&tpl.ValidTo, &tpl.ApplicableLevel, &tpl.Category, &tpl.CreatedAt, &tpl.UpdatedAt); err != nil {
			return nil, 0, err
		}
		list = append(list, tpl)
	}
	return list, total, nil
}

func (r *Repository) GetTemplateForUpdate(tx *sql.Tx, id int64) (*model.CouponTemplate, error) {
	row := tx.QueryRow(`
		SELECT id, name, type, value, threshold, total_count, remaining_count, per_user_limit,
		       valid_from, valid_to, applicable_level, category, created_at, updated_at
		FROM coupon_template WHERE id = ? LIMIT 1`, id)
	tpl := &model.CouponTemplate{}
	err := row.Scan(&tpl.ID, &tpl.Name, &tpl.Type, &tpl.Value, &tpl.Threshold,
		&tpl.TotalCount, &tpl.RemainingCount, &tpl.PerUserLimit, &tpl.ValidFrom,
		&tpl.ValidTo, &tpl.ApplicableLevel, &tpl.Category, &tpl.CreatedAt, &tpl.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return tpl, err
}

func (r *Repository) DecreaseTemplateRemaining(tx *sql.Tx, id int64) (int64, error) {
	res, err := tx.Exec(`
		UPDATE coupon_template
		SET remaining_count = remaining_count - 1, updated_at = ?
		WHERE id = ?`, logger.Now(), id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *Repository) CountUserRecords(userID, templateID int64) (int64, error) {
	var cnt int64
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM coupon_record
		WHERE user_id = ? AND template_id = ?`, userID, templateID).Scan(&cnt)
	return cnt, err
}

func (r *Repository) GetLastClaimTime(userID, templateID int64) (*time.Time, error) {
	var t time.Time
	err := r.db.QueryRow(`
		SELECT MAX(created_at) FROM coupon_record
		WHERE user_id = ? AND template_id = ?`, userID, templateID).Scan(&t)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if t.IsZero() {
		return nil, nil
	}
	return &t, nil
}

func (r *Repository) CreateRecordTx(tx *sql.Tx, rec *model.CouponRecord) error {
	now := logger.Now()
	rec.CreatedAt = now
	rec.UpdatedAt = now
	rec.Status = constants.CouponStatusUnused
	res, err := tx.Exec(`
		INSERT INTO coupon_record
		(template_id, user_id, status, value, threshold, valid_from, valid_to, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		rec.TemplateID, rec.UserID, rec.Status, rec.Value, rec.Threshold,
		rec.ValidFrom, rec.ValidTo, rec.CreatedAt, rec.UpdatedAt,
	)
	if err != nil {
		return err
	}
	rec.ID, err = res.LastInsertId()
	return err
}

func (r *Repository) GetRecord(id int64) (*model.CouponRecord, error) {
	row := r.db.QueryRow(`
		SELECT id, template_id, user_id, status, value, threshold, valid_from, valid_to,
		       used_at, order_id, created_at, updated_at
		FROM coupon_record WHERE id = ?`, id)
	rec := &model.CouponRecord{}
	err := row.Scan(&rec.ID, &rec.TemplateID, &rec.UserID, &rec.Status, &rec.Value,
		&rec.Threshold, &rec.ValidFrom, &rec.ValidTo, &rec.UsedAt, &rec.OrderID,
		&rec.CreatedAt, &rec.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return rec, err
}

func (r *Repository) UseRecord(id int64, orderID string) error {
	now := logger.Now()
	res, err := r.db.Exec(`
		UPDATE coupon_record
		SET status = ?, used_at = ?, order_id = ?, updated_at = ?
		WHERE id = ? AND status = ?`,
		constants.CouponStatusUsed, now, orderID, now, id, constants.CouponStatusUnused)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return errors.New("update failed, record not unused")
	}
	return nil
}

func (r *Repository) ListUserRecords(userID int64, status *constants.CouponStatus) ([]*model.CouponRecord, error) {
	var rows *sql.Rows
	var err error
	if status != nil {
		rows, err = r.db.Query(`
			SELECT id, template_id, user_id, status, value, threshold, valid_from, valid_to,
			       used_at, order_id, created_at, updated_at
			FROM coupon_record WHERE user_id = ? AND status = ? ORDER BY id DESC`,
			userID, *status)
	} else {
		rows, err = r.db.Query(`
			SELECT id, template_id, user_id, status, value, threshold, valid_from, valid_to,
			       used_at, order_id, created_at, updated_at
			FROM coupon_record WHERE user_id = ? ORDER BY id DESC`, userID)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []*model.CouponRecord
	for rows.Next() {
		rec := &model.CouponRecord{}
		if err := rows.Scan(&rec.ID, &rec.TemplateID, &rec.UserID, &rec.Status, &rec.Value,
			&rec.Threshold, &rec.ValidFrom, &rec.ValidTo, &rec.UsedAt, &rec.OrderID,
			&rec.CreatedAt, &rec.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, rec)
	}
	return list, nil
}

func (r *Repository) BatchMarkExpired(beforeTime time.Time) (int64, error) {
	res, err := r.db.Exec(`
		UPDATE coupon_record
		SET status = ?, updated_at = ?
		WHERE status = ? AND valid_to < ?`,
		constants.CouponStatusExpired, logger.Now(), constants.CouponStatusUnused, beforeTime)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (r *Repository) GetOrCreateUser(userID int64, level constants.UserLevel) (*model.User, bool, error) {
	row := r.db.QueryRow(`SELECT id, level, new_user_gift_sent, created_at, updated_at FROM user WHERE id = ?`, userID)
	u := &model.User{}
	var sentInt int
	err := row.Scan(&u.ID, &u.Level, &sentInt, &u.CreatedAt, &u.UpdatedAt)
	if err == nil {
		u.NewUserGiftSent = sentInt == 1
		return u, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return nil, false, err
	}
	now := logger.Now()
	u = &model.User{
		ID:              userID,
		Level:           level,
		NewUserGiftSent: false,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	_, err = r.db.Exec(`
		INSERT INTO user (id, level, new_user_gift_sent, created_at, updated_at) VALUES (?, ?, 0, ?, ?)`,
		userID, level, now, now)
	if err != nil {
		return nil, false, err
	}
	return u, true, nil
}

func (r *Repository) MarkNewUserGiftSent(userID int64) error {
	_, err := r.db.Exec(`UPDATE user SET new_user_gift_sent = 1, updated_at = ? WHERE id = ?`,
		logger.Now(), userID)
	return err
}

func (r *Repository) IsNewUserGiftSent(userID int64) (bool, error) {
	var sent int
	err := r.db.QueryRow(`SELECT new_user_gift_sent FROM user WHERE id = ?`, userID).Scan(&sent)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return sent == 1, err
}

func (r *Repository) GetNewUserTemplate() (*model.CouponTemplate, error) {
	now := logger.Now()
	row := r.db.QueryRow(`
		SELECT id, name, type, value, threshold, total_count, remaining_count, per_user_limit,
		       valid_from, valid_to, applicable_level, category, created_at, updated_at
		FROM coupon_template
		WHERE type = ? AND remaining_count > 0 AND valid_from <= ? AND valid_to >= ?
		ORDER BY id DESC LIMIT 1`, constants.CouponTypeNewUser, now, now)
	tpl := &model.CouponTemplate{}
	err := row.Scan(&tpl.ID, &tpl.Name, &tpl.Type, &tpl.Value, &tpl.Threshold,
		&tpl.TotalCount, &tpl.RemainingCount, &tpl.PerUserLimit, &tpl.ValidFrom,
		&tpl.ValidTo, &tpl.ApplicableLevel, &tpl.Category, &tpl.CreatedAt, &tpl.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return tpl, err
}
