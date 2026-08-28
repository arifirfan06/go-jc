package domain

import "errors"

// Definisi Standar Custom Error Domain
var (
	ErrUnauthorized        = errors.New("akses tidak diizinkan: token tidak valid atau telah logout")
	ErrForbidden           = errors.New("akses ditolak: Anda tidak memiliki hak akses untuk resource ini")
	ErrNotFound            = errors.New("data tidak ditemukan")
	ErrEmailAlreadyExists  = errors.New("email sudah terdaftar di sistem")
	ErrInvalidCredentials  = errors.New("email atau password salah")
	ErrInvalidEmailDomain  = errors.New("email harus diakhiri dengan @company.co.id")
	ErrInsufficientBudget  = errors.New("proses penggajian gagal: sisa anggaran departemen tidak mencukupi")
	ErrInternalServerError = errors.New("terjadi kesalahan internal pada server")
)
