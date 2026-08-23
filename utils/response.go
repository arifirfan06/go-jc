package utils

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ============================================================================
// HELPER RESPONSE API
// ============================================================================
// Seluruh endpoint memakai bentuk response yang seragam supaya klien
// (Postman, frontend, dsb.) cukup menulis satu parser untuk semua kasus.

// APIResponse adalah kerangka response JSON standar aplikasi.
type APIResponse struct {
	Success bool        `json:"success"`          // true bila request berhasil
	Message string      `json:"message"`          // pesan singkat berbahasa Indonesia
	Data    interface{} `json:"data,omitempty"`   // isi data (opsional)
	Meta    interface{} `json:"meta,omitempty"`   // informasi tambahan, mis. jumlah data
	Errors  interface{} `json:"errors,omitempty"` // rincian error validasi (opsional)
}

// Success mengirim response sukses beserta datanya.
func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, APIResponse{Success: true, Message: message, Data: data})
}

// SuccessWithMeta mengirim response sukses lengkap dengan metadata,
// misalnya jumlah baris yang dikembalikan pada endpoint listing.
func SuccessWithMeta(c *gin.Context, status int, message string, data interface{}, meta interface{}) {
	c.JSON(status, APIResponse{Success: true, Message: message, Data: data, Meta: meta})
}

// Error mengirim response gagal tanpa rincian tambahan.
func Error(c *gin.Context, status int, message string) {
	c.JSON(status, APIResponse{Success: false, Message: message})
}

// ErrorWithDetail mengirim response gagal beserta rincian penyebabnya.
func ErrorWithDetail(c *gin.Context, status int, message string, errors interface{}) {
	c.JSON(status, APIResponse{Success: false, Message: message, Errors: errors})
}

// ============================================================================
// PENERJEMAH ERROR VALIDASI
// ============================================================================

// HandleBindingError menangani kegagalan ShouldBindJSON dan selalu membalas
// dengan status 400 Bad Request, sesuai syarat teknis pengumpulan.
//
// Bila error berasal dari validator, pesan teknis bawaan library diterjemahkan
// lebih dulu ke bahasa Indonesia agar mudah dibaca penguji API.
func HandleBindingError(c *gin.Context, err error) {
	var validationErrors validator.ValidationErrors
	// errors.As tidak dipakai di sini karena Gin mengembalikan tipe konkret
	// validator.ValidationErrors, sehingga type assertion sudah memadai.
	if ve, ok := err.(validator.ValidationErrors); ok {
		validationErrors = ve
	}

	// Error non-validasi, misalnya JSON rusak atau tipe data tidak cocok.
	if validationErrors == nil {
		ErrorWithDetail(c, http.StatusBadRequest,
			"Format JSON tidak valid atau tipe data tidak sesuai", err.Error())
		return
	}

	details := make(map[string]string, len(validationErrors))
	for _, fieldErr := range validationErrors {
		details[toSnakeCase(fieldErr.Field())] = translateValidationError(fieldErr)
	}

	ErrorWithDetail(c, http.StatusBadRequest, "Validasi input gagal", details)
}

// translateValidationError mengubah satu error validator menjadi kalimat
// berbahasa Indonesia sesuai jenis tag yang gagal.
func translateValidationError(fe validator.FieldError) string {
	field := toSnakeCase(fe.Field())

	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("Field '%s' wajib diisi", field)
	case "email":
		return fmt.Sprintf("Field '%s' harus berupa alamat email yang valid", field)
	case "min":
		return fmt.Sprintf("Field '%s' minimal %s karakter/nilai", field, fe.Param())
	case "max":
		return fmt.Sprintf("Field '%s' maksimal %s karakter/nilai", field, fe.Param())
	case "len":
		return fmt.Sprintf("Field '%s' harus tepat %s karakter", field, fe.Param())
	case "gt":
		return fmt.Sprintf("Field '%s' harus lebih besar dari %s", field, fe.Param())
	case "gte":
		return fmt.Sprintf("Field '%s' minimal bernilai %s", field, fe.Param())
	case "lte":
		return fmt.Sprintf("Field '%s' maksimal bernilai %s", field, fe.Param())
	case "oneof":
		return fmt.Sprintf("Field '%s' hanya boleh bernilai: %s",
			field, strings.ReplaceAll(fe.Param(), " ", ", "))
	case "uppercase":
		return fmt.Sprintf("Field '%s' harus ditulis dengan huruf kapital semua", field)
	case "datetime":
		return fmt.Sprintf("Field '%s' harus mengikuti format %s", field, humanizeLayout(fe.Param()))
	default:
		return fmt.Sprintf("Field '%s' tidak lolos aturan validasi '%s'", field, fe.Tag())
	}
}

// humanLayouts memetakan layout waktu ala Go ke pola yang lazim dipahami
// pengguna. Pemetaan dibuat eksplisit karena angka penyusun layout Go bersifat
// ambigu bila diterjemahkan satu per satu: "04" berarti menit, sedangkan "01"
// berarti bulan, padahal keduanya sama-sama ditulis "MM" oleh pengguna.
var humanLayouts = map[string]string{
	LayoutDate:   "YYYY-MM-DD",
	LayoutPeriod: "YYYY-MM",
	LayoutTime:   "HH:MM",
}

// humanizeLayout menerjemahkan layout waktu ala Go menjadi pola yang lazim
// dipahami pengguna, misalnya "2006-01-02" menjadi "YYYY-MM-DD".
func humanizeLayout(layout string) string {
	if human, ok := humanLayouts[layout]; ok {
		return human
	}
	return layout
}

// toSnakeCase mengubah nama field Go (PascalCase) menjadi snake_case supaya
// nama field pada pesan error sama persis dengan nama field pada body JSON.
//
// Akronim yang berderet tidak ikut dipecah, sehingga "DepartmentID" menghasilkan
// "department_id" (bukan "department_i_d") dan "NIK" tetap menjadi "nik".
func toSnakeCase(s string) string {
	runes := []rune(s)

	var builder strings.Builder
	for i, r := range runes {
		if !isUpper(r) {
			builder.WriteRune(r)
			continue
		}

		// Garis bawah hanya disisipkan pada batas kata yang sesungguhnya, yaitu
		// ketika huruf sebelumnya bukan kapital (akhir kata biasa, mis. "tID"),
		// atau ketika huruf berikutnya kecil (awal kata baru setelah akronim,
		// mis. "IDNumber" pada bagian "DNu").
		if i > 0 && (!isUpper(runes[i-1]) || (i+1 < len(runes) && isLower(runes[i+1]))) {
			builder.WriteByte('_')
		}
		builder.WriteRune(r + ('a' - 'A'))
	}

	return builder.String()
}

// SentenceCase mengubah huruf pertama sebuah kalimat menjadi kapital.
//
// Pesan galat yang dihasilkan paket lain ditulis dengan huruf kecil mengikuti
// konvensi Go. Helper ini dipakai tepat di perbatasan response, agar pesan yang
// sampai ke klien tetap seragam dengan pesan lain yang diawali huruf kapital.
func SentenceCase(message string) string {
	if message == "" {
		return message
	}

	runes := []rune(message)
	if isLower(runes[0]) {
		runes[0] -= 'a' - 'A'
	}
	return string(runes)
}

// isUpper memeriksa apakah sebuah rune merupakan huruf kapital A-Z.
func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

// isLower memeriksa apakah sebuah rune merupakan huruf kecil a-z.
func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}
