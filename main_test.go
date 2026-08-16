package main

import (
	"math"
	"testing"
)

// Helper function untuk membandingkan float dengan toleransi presisi
func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < 0.00001
}

// TestFullTimeEmployee_CalculateSalary menguji perhitungan gaji karyawan tetap
func TestFullTimeEmployee_CalculateSalary(t *testing.T) {
	emp := FullTimeEmployee{
		BaseSalary: 10000000,
		Allowance:  2000000,
		TaxRate:    0.05,
	}

	expected := (10000000.0 + 2000000.0) * (1.0 - 0.05) // 11,400,000
	salary, err := emp.CalculateSalary()
	if err != nil {
		t.Fatalf("CalculateSalary gagal: %v", err)
	}

	if !almostEqual(salary, expected) {
		t.Errorf("Gaji FullTime salah. Ekspektasi: %.2f, Hasil: %.2f", expected, salary)
	}

	// Uji validasi angka negatif
	empInvalid := FullTimeEmployee{BaseSalary: -5000, Allowance: 1000, TaxRate: 0.05}
	_, err = empInvalid.CalculateSalary()
	if err == nil {
		t.Errorf("Harusnya mengembalikan error jika BaseSalary negatif")
	}
}

// TestContractEmployee_CalculateSalary menguji perhitungan gaji karyawan kontrak
func TestContractEmployee_CalculateSalary(t *testing.T) {
	emp := ContractEmployee{
		MonthlyRate:      8000000,
		PerformanceBonus: 1500000,
	}

	expected := 8000000.0 + 1500000.0 // 9,500,000
	salary, err := emp.CalculateSalary()
	if err != nil {
		t.Fatalf("CalculateSalary gagal: %v", err)
	}

	if !almostEqual(salary, expected) {
		t.Errorf("Gaji Contract salah. Ekspektasi: %.2f, Hasil: %.2f", expected, salary)
	}

	// Uji validasi angka negatif
	empInvalid := ContractEmployee{MonthlyRate: 8000000, PerformanceBonus: -500000}
	_, err = empInvalid.CalculateSalary()
	if err == nil {
		t.Errorf("Harusnya mengembalikan error jika PerformanceBonus negatif")
	}
}

// TestFreelancer_CalculateSalary menguji perhitungan gaji freelancer
func TestFreelancer_CalculateSalary(t *testing.T) {
	emp := Freelancer{
		HourlyRate:  100000,
		HoursWorked: 120,
	}

	expected := 100000.0 * 120.0 // 12,000,000
	salary, err := emp.CalculateSalary()
	if err != nil {
		t.Fatalf("CalculateSalary gagal: %v", err)
	}

	if !almostEqual(salary, expected) {
		t.Errorf("Gaji Freelancer salah. Ekspektasi: %.2f, Hasil: %.2f", expected, salary)
	}

	// Uji validasi angka negatif
	empInvalid := Freelancer{HourlyRate: 100000, HoursWorked: -10}
	_, err = empInvalid.CalculateSalary()
	if err == nil {
		t.Errorf("Harusnya mengembalikan error jika HoursWorked negatif")
	}
}

// TestHRIS_RegisterEmployee menguji registrasi karyawan dan seluruh skenario validasi error
func TestHRIS_RegisterEmployee(t *testing.T) {
	hris := NewHRIS()

	emp1 := FullTimeEmployee{BaseSalary: 10000000, Allowance: 2000000, TaxRate: 0.05}
	emp2 := ContractEmployee{MonthlyRate: 8000000, PerformanceBonus: 1500000}

	// 1. Registrasi Sukses
	err := hris.RegisterEmployee("EMP-001", "Budi Santoso", emp1)
	if err != nil {
		t.Fatalf("Registrasi EMP-001 gagal: %v", err)
	}

	// 2. Validasi ID Duplikat
	err = hris.RegisterEmployee("EMP-001", "Budi Baru", emp2)
	if err == nil {
		t.Errorf("Harusnya error saat mendaftarkan ID duplikat")
	} else if err.Error() != "karyawan dengan ID tersebut sudah terdaftar" {
		t.Errorf("Pesan error duplikat ID tidak sesuai: %v", err)
	}

	// 3. Validasi Nama Kosong
	err = hris.RegisterEmployee("EMP-002", "   ", emp2)
	if err == nil {
		t.Errorf("Harusnya error saat nama karyawan kosong")
	} else if err.Error() != "nama karyawan tidak boleh kosong" {
		t.Errorf("Pesan error nama kosong tidak sesuai: %v", err)
	}

	// 4. Validasi Nilai Keuangan Negatif
	empNegative := Freelancer{HourlyRate: -50000, HoursWorked: 10}
	err = hris.RegisterEmployee("EMP-003", "Doni Negatif", empNegative)
	if err == nil {
		t.Errorf("Harusnya error saat komponen keuangan bernilai negatif")
	}
}

// TestHRIS_CalculateTotalPayout menguji perhitungan total anggaran seluruh karyawan
func TestHRIS_CalculateTotalPayout(t *testing.T) {
	hris := NewHRIS()

	emp1 := FullTimeEmployee{BaseSalary: 10000000, Allowance: 2000000, TaxRate: 0.05} // 11,400,000
	emp2 := ContractEmployee{MonthlyRate: 8000000, PerformanceBonus: 1500000}         // 9,500,000
	emp3 := Freelancer{HourlyRate: 100000, HoursWorked: 120}                           // 12,000,000

	_ = hris.RegisterEmployee("EMP-001", "Budi Santoso", emp1)
	_ = hris.RegisterEmployee("EMP-002", "Siti Aminah", emp2)
	_ = hris.RegisterEmployee("EMP-003", "Rian Ardiansyah", emp3)

	expectedTotal := 11400000.0 + 9500000.0 + 12000000.0 // 32,900,000
	totalPayout := hris.CalculateTotalPayout()

	if !almostEqual(totalPayout, expectedTotal) {
		t.Errorf("CalculateTotalPayout salah. Ekspektasi: %.2f, Hasil: %.2f", expectedTotal, totalPayout)
	}
}
