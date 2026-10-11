package database

import (
	"log"

	"gorm.io/gorm"

	"penggajian/server/internal/platform/security"
)

type seedPermission struct {
	Code   string
	Name   string
	Module string
}

// Katalog permission per menu sesuai PRD.
var permissionCatalog = []seedPermission{
	{"dashboard.read", "Lihat dashboard", "dashboard"},
	{"pegawai.read", "Lihat pegawai", "pegawai"},
	{"pegawai.create", "Tambah pegawai", "pegawai"},
	{"pegawai.update", "Ubah pegawai", "pegawai"},
	{"pegawai.delete", "Hapus pegawai", "pegawai"},
	{"penggajian.read", "Lihat penggajian", "penggajian"},
	{"penggajian.create", "Tambah penggajian", "penggajian"},
	{"penggajian.update", "Ubah penggajian", "penggajian"},
	{"penggajian.delete", "Hapus penggajian", "penggajian"},
	{"penggajian.approve", "Setujui penggajian", "penggajian"},
	{"inventory.read", "Lihat inventory", "inventory"},
	{"inventory.create", "Tambah inventory", "inventory"},
	{"inventory.update", "Ubah inventory", "inventory"},
	{"inventory.delete", "Hapus inventory", "inventory"},
	{"akademik.read", "Lihat akademik", "akademik"},
	{"akademik.create", "Tambah akademik", "akademik"},
	{"akademik.update", "Ubah akademik", "akademik"},
	{"akademik.delete", "Hapus akademik", "akademik"},
	{"laporan.read", "Lihat laporan", "laporan"},
	{"spp.read", "Lihat SPP", "spp"},
	{"spp.create", "Tambah SPP", "spp"},
	{"spp.update", "Ubah SPP", "spp"},
	{"spp.delete", "Hapus SPP", "spp"},
	{"pengaturan.read", "Lihat pengaturan", "pengaturan"},
	{"pengaturan.create", "Tambah pengaturan", "pengaturan"},
	{"pengaturan.update", "Ubah pengaturan", "pengaturan"},
	{"pengaturan.delete", "Hapus pengaturan", "pengaturan"},
}

const (
	defaultAdminUsername = "admin"
	defaultAdminPassword = "admin123"
	defaultAdminEmail    = "admin@sekolah.sch.id"
)

// Seed mengisi data awal: katalog permission, role Admin, user admin,
// dan profil sekolah. Aman dipanggil berulang (idempoten).
func Seed(db *gorm.DB) error {
	for _, p := range permissionCatalog {
		if err := db.Exec(
			`INSERT INTO permissions (code, name, module) VALUES (?, ?, ?) ON CONFLICT (code) DO NOTHING`,
			p.Code, p.Name, p.Module,
		).Error; err != nil {
			return err
		}
	}

	if err := db.Exec(
		`INSERT INTO roles (name, description) VALUES ('Admin', 'Administrator sistem (akses penuh)') ON CONFLICT (name) DO NOTHING`,
	).Error; err != nil {
		return err
	}

	if err := db.Exec(
		`INSERT INTO role_permissions (role_id, permission_id)
		 SELECT r.id, p.id FROM roles r CROSS JOIN permissions p
		 WHERE r.name = 'Admin'
		 ON CONFLICT DO NOTHING`,
	).Error; err != nil {
		return err
	}

	hash, err := security.HashPassword(defaultAdminPassword)
	if err != nil {
		return err
	}
	res := db.Exec(
		`INSERT INTO users (username, email, password_hash, full_name, role_id)
		 SELECT ?, ?, ?, 'Administrator', r.id FROM roles r WHERE r.name = 'Admin'
		 ON CONFLICT (username) DO NOTHING`,
		defaultAdminUsername, defaultAdminEmail, hash,
	)
	if err := res.Error; err != nil {
		return err
	}
	if res.RowsAffected > 0 {
		log.Printf("[seed] user admin dibuat (username: %s / password: %s) — segera ganti password setelah login", defaultAdminUsername, defaultAdminPassword)
	}

	if err := db.Exec(
		`INSERT INTO school_profiles (id, school_name) VALUES (1, 'Sekolah Kita') ON CONFLICT (id) DO NOTHING`,
	).Error; err != nil {
		return err
	}

	defaultComponents := []struct{ Code, Name, Type, Description string }{
		{"GP", "Gaji Pokok", "earning", "Gaji pokok otomatis dari data pegawai"},
		{"T_JABATAN", "Tunjangan Jabatan", "earning", ""},
		{"T_MAKAN", "Uang Makan", "earning", ""},
		{"T_TRANSPORT", "Tunjangan Transport", "earning", ""},
		{"T_KEHADIRAN", "Tunjangan Kehadiran", "earning", ""},
		{"BONUS", "Bonus / Insentif", "earning", ""},
		{"P_ALPHA", "Potongan Alpha", "deduction", ""},
		{"P_IZIN", "Potongan Izin", "deduction", ""},
		{"P_TELAT", "Potongan Terlambat", "deduction", ""},
		{"KAS_BON", "Kas Bon / Pinjaman", "deduction", ""},
		{"BPJS", "BPJS", "deduction", ""},
		{"PPH21", "PPh 21", "deduction", ""},
		{"LAIN2", "Potongan Lain-lain", "deduction", ""},
	}
	for _, comp := range defaultComponents {
		if err := db.Exec(
			`INSERT INTO pay_components (code, name, type, description) VALUES (?, ?, ?, NULLIF(?, ''))
			 ON CONFLICT (code) DO NOTHING`,
			comp.Code, comp.Name, comp.Type, comp.Description,
		).Error; err != nil {
			return err
		}
	}

	defaultSubjects := []struct{ Code, Name string }{
		{"MTK", "Matematika"},
		{"IPA", "Ilmu Pengetahuan Alam"},
		{"IPS", "Ilmu Pengetahuan Sosial"},
		{"BIN", "Bahasa Indonesia"},
		{"BIG", "Bahasa Inggris"},
		{"PAI", "Pendidikan Agama Islam"},
		{"PJK", "PJOK"},
	}
	for _, sub := range defaultSubjects {
		if err := db.Exec(
			`INSERT INTO subjects (code, name) VALUES (?, ?) ON CONFLICT (code) DO NOTHING`,
			sub.Code, sub.Name,
		).Error; err != nil {
			return err
		}
	}

	defaultClasses := []struct{ Name, Level string }{
		{"7A", "7"}, {"7B", "7"},
		{"8A", "8"}, {"8B", "8"},
		{"9A", "9"}, {"9B", "9"},
	}
	for _, cls := range defaultClasses {
		if err := db.Exec(
			`INSERT INTO school_classes (name, level) VALUES (?, ?) ON CONFLICT (name) DO NOTHING`,
			cls.Name, cls.Level,
		).Error; err != nil {
			return err
		}
	}

	return nil
}
