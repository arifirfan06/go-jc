// Package main adalah titik masuk utama aplikasi backend Mini HRIS
// (Human Resource Information System).
//
// Aplikasi dibangun menggunakan:
//   - Go             : bahasa pemrograman utama
//   - Gin Gonic      : framework HTTP dan routing
//   - GORM + SQLite  : ORM dan basis data (driver murni Go, tanpa CGO)
//
// Struktur folder proyek:
//
//	config/      koneksi dan migrasi database SQLite
//	models/      definisi struct entitas beserta aturan validasi
//	controllers/ logika handler untuk setiap endpoint API
//	routes/      registrasi perutean Gin
//	seeders/     pengisian data awal untuk pengujian
//	utils/       helper response API dan perhitungan tanggal
//	main.go      titik masuk utama aplikasi
package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"

	"go-jc-challange2/config"
	"go-jc-challange2/routes"
	"go-jc-challange2/seeders"
)

// defaultPort adalah port bawaan tempat server HTTP mendengarkan request.
// Nilainya dapat ditimpa lewat variabel lingkungan PORT.
const defaultPort = "8080"

func main() {
	// Mode Gin dapat diubah menjadi "release" lewat variabel lingkungan
	// GIN_MODE untuk menyembunyikan log debug saat aplikasi dipakai serius.
	if mode := os.Getenv("GIN_MODE"); mode != "" {
		gin.SetMode(mode)
	}

	// --- Tahap 1: buka koneksi database ------------------------------------
	config.ConnectDatabase()

	// --- Tahap 2: bangun struktur 6 tabel beserta relasinya -----------------
	config.MigrateDatabase()

	// --- Tahap 3: isi data awal agar API langsung bisa diuji ----------------
	seeders.Run()

	// --- Tahap 4: daftarkan seluruh endpoint --------------------------------
	router := routes.SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	log.Printf("[SERVER] Mini HRIS API siap diakses di http://localhost:%s", port)
	log.Printf("[SERVER] Daftar endpoint tersedia di http://localhost:%s/", port)

	// Run memblokir proses hingga server dihentikan atau terjadi kesalahan fatal.
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("[SERVER] Gagal menjalankan server pada port %s: %v", port, err)
	}
}
