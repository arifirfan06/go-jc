package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Config merepresentasikan konfigurasi aplikasi yang dimuat dari environment variable.
type Config struct {
	AppPort   string
	DBHost    string
	DBPort    string
	DBUser    string
	DBPass    string
	DBName    string
	SSLMode   string
	JWTSecret string
}

// LoadConfig membaca berkas .env dan mengembalikan struct Config.
func LoadConfig() *Config {
	// FLOW: Memuat file .env jika ada (abaikan error jika file tidak ditemukan, gunakan env OS)
	if err := godotenv.Load(); err != nil {
		log.Println("[INFO] File .env tidak ditemukan, menggunakan environment variable bawaan OS")
	}

	// FLOW: Inisialisasi struct Config dengan fallback default value
	cfg := &Config{
		AppPort:   getEnv("PORT", "8080"),
		DBHost:    getEnv("DB_HOST", "localhost"),
		DBPort:    getEnv("DB_PORT", "5432"),
		DBUser:    getEnv("DB_USER", "postgres"),
		DBPass:    getEnv("DB_PASSWORD", "postgres"),
		DBName:    getEnv("DB_NAME", "hris_payroll_db"),
		SSLMode:   getEnv("DB_SSLMODE", "disable"),
		JWTSecret: getEnv("JWT_SECRET", "secret-key-default-hris"),
	}

	return cfg
}

// InitDB membuat koneksi ke PostgreSQL menggunakan GORM.
func InitDB(cfg *Config) (*gorm.DB, error) {
	// FLOW: Membuat Data Source Name (DSN) PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=Asia/Jakarta",
		cfg.DBHost, cfg.DBUser, cfg.DBPass, cfg.DBName, cfg.DBPort, cfg.SSLMode,
	)

	// FLOW: Membuka koneksi GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("gagal terhubung ke database PostgreSQL: %w", err)
	}

	log.Println("[SUCCESS] Berhasil terhubung ke database PostgreSQL")
	return db, nil
}

// getEnv adalah fungsi pembantu untuk mengambil variabel lingkungan dengan nilai default.
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists && value != "" {
		return value
	}
	return fallback
}
