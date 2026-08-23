package config

import (
	"fmt"
	"log"
	"os"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"go-jc-challange2/models"
)

// ============================================================================
// KONFIGURASI KONEKSI DATABASE SQLITE
// ============================================================================

// DB adalah instance koneksi GORM yang dipakai bersama oleh seluruh controller.
// Variabel ini diisi satu kali saat aplikasi start melalui ConnectDatabase().
var DB *gorm.DB

// defaultDBName adalah nama berkas database SQLite bawaan.
// Berkas ini dibuat otomatis di direktori kerja bila belum ada.
const defaultDBName = "hris.db"

// ConnectDatabase membuka koneksi ke database SQLite dan menyimpannya ke DB.
//
// Driver yang dipakai adalah github.com/glebarez/sqlite, yaitu driver SQLite
// murni Go (tanpa CGO). Pilihan ini penting agar aplikasi tetap bisa dikompilasi
// di mesin Windows yang tidak memiliki compiler C (gcc/MinGW), berbeda dengan
// driver populer mattn/go-sqlite3 yang mensyaratkan CGO_ENABLED=1.
func ConnectDatabase() {
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = defaultDBName
	}

	// Parameter _pragma pada DSN dieksekusi tepat setelah koneksi terbuka:
	//   - foreign_keys(1)     : mengaktifkan penegakan Foreign Key. SQLite
	//                           mematikannya secara bawaan, sehingga tanpa baris
	//                           ini relasi antar tabel tidak akan dijaga.
	//   - busy_timeout(5000)  : menunggu maksimal 5 detik bila database terkunci
	//                           oleh proses lain, alih-alih langsung gagal.
	dsn := fmt.Sprintf("%s?_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)", dbName)

	database, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		// Menampilkan query SQL yang dijalankan ke terminal agar mudah ditelusuri
		// saat pengembangan maupun saat demo pengujian API.
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("[DATABASE] Gagal membuka koneksi SQLite: %v", err)
	}

	DB = database
	log.Printf("[DATABASE] Koneksi SQLite berhasil dibuka (berkas: %s)", dbName)
}

// MigrateDatabase menjalankan AutoMigrate GORM untuk keenam tabel entitas.
//
// Urutan pemanggilan penting: tabel induk (departments, positions) dibuat lebih
// dulu, baru tabel yang menyimpan Foreign Key ke tabel tersebut (employees),
// lalu tabel turunan karyawan (attendances, leaves, salaries).
func MigrateDatabase() {
	err := DB.AutoMigrate(
		&models.Department{}, // 1. Master departemen
		&models.Position{},   // 2. Master jabatan
		&models.Employee{},   // 3. Karyawan (FK -> departments, positions)
		&models.Attendance{}, // 4. Kehadiran (FK -> employees)
		&models.Leave{},      // 5. Pengajuan cuti (FK -> employees)
		&models.Salary{},     // 6. Slip gaji (FK -> employees)
	)
	if err != nil {
		log.Fatalf("[DATABASE] Gagal menjalankan migrasi tabel: %v", err)
	}

	log.Println("[DATABASE] Migrasi 6 tabel berhasil dijalankan")
}
