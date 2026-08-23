package utils

import (
	"errors"
	"time"
)

// ============================================================================
// HELPER TANGGAL & WAKTU
// ============================================================================
// Aplikasi menyimpan tanggal/jam sebagai string (YYYY-MM-DD, HH:MM, YYYY-MM)
// sesuai ketentuan skema database. Helper di bawah dipakai untuk menghitung
// selisih hari, rentang periode, dan irisan tanggal cuti dengan aman.

// Layout tanggal dan waktu yang dipakai di seluruh aplikasi.
const (
	LayoutDate   = "2006-01-02" // Format tanggal: YYYY-MM-DD
	LayoutTime   = "15:04"      // Format jam: HH:MM (24 jam)
	LayoutPeriod = "2006-01"    // Format periode gaji: YYYY-MM
)

// ParseDate mengurai string tanggal berformat YYYY-MM-DD menjadi time.Time.
func ParseDate(value string) (time.Time, error) {
	return time.Parse(LayoutDate, value)
}

// ParsePeriod mengurai string periode berformat YYYY-MM menjadi time.Time
// yang menunjuk ke tanggal 1 pada bulan tersebut.
func ParsePeriod(value string) (time.Time, error) {
	return time.Parse(LayoutPeriod, value)
}

// PeriodRange mengembalikan tanggal awal dan tanggal akhir sebuah periode gaji.
// Contoh: "2026-02" menghasilkan ("2026-02-01", "2026-02-28").
//
// Tanggal akhir dihitung dengan menambah satu bulan lalu mundur satu hari,
// sehingga bulan Februari pada tahun kabisat pun tetap akurat.
func PeriodRange(period string) (string, string, error) {
	first, err := ParsePeriod(period)
	if err != nil {
		return "", "", errors.New("periode harus mengikuti format YYYY-MM")
	}
	last := first.AddDate(0, 1, -1)
	return first.Format(LayoutDate), last.Format(LayoutDate), nil
}

// CountDaysInclusive menghitung jumlah hari antara dua tanggal secara inklusif,
// artinya tanggal mulai dan tanggal selesai ikut dihitung.
// Cuti 2026-03-02 sampai 2026-03-04 bernilai 3 hari, bukan 2.
func CountDaysInclusive(startDate, endDate string) (int, error) {
	start, err := ParseDate(startDate)
	if err != nil {
		return 0, errors.New("tanggal mulai harus mengikuti format YYYY-MM-DD")
	}
	end, err := ParseDate(endDate)
	if err != nil {
		return 0, errors.New("tanggal selesai harus mengikuti format YYYY-MM-DD")
	}
	if end.Before(start) {
		return 0, errors.New("tanggal selesai tidak boleh mendahului tanggal mulai")
	}
	return int(end.Sub(start).Hours()/24) + 1, nil
}

// CountOverlapDays menghitung berapa hari rentang [startA, endA] beririsan
// dengan rentang [startB, endB]. Dipakai saat payroll untuk mencari jumlah hari
// cuti yang benar-benar jatuh di dalam periode gaji yang sedang dihitung.
//
// Contoh: cuti 2026-01-30 s/d 2026-02-03 terhadap periode Februari
// menghasilkan 3 hari (tanggal 1, 2, dan 3 Februari).
func CountOverlapDays(startA, endA, startB, endB string) int {
	aStart, err1 := ParseDate(startA)
	aEnd, err2 := ParseDate(endA)
	bStart, err3 := ParseDate(startB)
	bEnd, err4 := ParseDate(endB)
	if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
		return 0
	}

	// Ambil batas terdalam dari kedua rentang.
	start := aStart
	if bStart.After(start) {
		start = bStart
	}
	end := aEnd
	if bEnd.Before(end) {
		end = bEnd
	}

	// Tidak ada irisan sama sekali.
	if end.Before(start) {
		return 0
	}
	return int(end.Sub(start).Hours()/24) + 1
}

// IsTimeAfter membandingkan dua jam berformat HH:MM dan mengembalikan true
// bila `value` lebih lambat daripada `threshold`.
// Dipakai untuk menentukan apakah seorang karyawan datang terlambat.
func IsTimeAfter(value, threshold string) bool {
	t1, err1 := time.Parse(LayoutTime, value)
	t2, err2 := time.Parse(LayoutTime, threshold)
	if err1 != nil || err2 != nil {
		return false
	}
	return t1.After(t2)
}

// YearOfPeriod mengambil bagian tahun dari sebuah periode gaji.
// Contoh: "2026-08" menghasilkan "2026".
func YearOfPeriod(period string) string {
	if len(period) < 4 {
		return ""
	}
	return period[:4]
}
