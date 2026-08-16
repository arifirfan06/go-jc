package main

import (
	"errors"
	"fmt"
	"strings"
)

// ============================================================================
// 1. INTERFACE & CONTRACT
// ============================================================================

// PayrollCalculator adalah interface yang mendefinisikan kontrak metode
// untuk seluruh jenis penggajian karyawan di dalam sistem HRIS.
type PayrollCalculator interface {
	// CalculateSalary menghitung total gaji bulanan bersih karyawan.
	CalculateSalary() (float64, error)
	// GetEmployeeType mengembalikan tipe string dari tipe karyawan (misal: "FullTime", "Contract", "Freelance").
	GetEmployeeType() string
}

// ============================================================================
// 2. CONCRETE STRUCTS (IMPLEMENTASI INTERFACE PAYROLLCALCULATOR)
// ============================================================================

// FullTimeEmployee merepresentasikan karyawan tetap.
type FullTimeEmployee struct {
	BaseSalary float64 // Gaji Pokok
	Allowance  float64 // Tunjangan
	TaxRate    float64 // Persentase Pajak (contoh: 0.05 untuk 5%)
}

// GetEmployeeType mengembalikan jenis tipe karyawan.
func (f FullTimeEmployee) GetEmployeeType() string {
	return "FullTime"
}

// CalculateSalary menghitung gaji bersih Karyawan Tetap.
// Rumus: (BaseSalary + Allowance) * (1.0 - TaxRate)
func (f FullTimeEmployee) CalculateSalary() (float64, error) {
	// Validasi komponen nilai keuangan
	if f.BaseSalary < 0 || f.Allowance < 0 || f.TaxRate < 0 {
		return 0, errors.New("komponen keuangan karyawan FullTime tidak boleh bernilai negatif")
	}
	grossSalary := f.BaseSalary + f.Allowance
	netSalary := grossSalary * (1.0 - f.TaxRate)
	return netSalary, nil
}

// ContractEmployee merepresentasikan karyawan kontrak.
type ContractEmployee struct {
	MonthlyRate      float64 // Tarif Bulanan
	PerformanceBonus float64 // Bonus Kinerja
}

// GetEmployeeType mengembalikan jenis tipe karyawan.
func (c ContractEmployee) GetEmployeeType() string {
	return "Contract"
}

// CalculateSalary menghitung gaji bersih Karyawan Kontrak.
// Rumus: MonthlyRate + PerformanceBonus
func (c ContractEmployee) CalculateSalary() (float64, error) {
	// Validasi komponen nilai keuangan
	if c.MonthlyRate < 0 || c.PerformanceBonus < 0 {
		return 0, errors.New("komponen keuangan karyawan Contract tidak boleh bernilai negatif")
	}
	return c.MonthlyRate + c.PerformanceBonus, nil
}

// Freelancer merepresentasikan karyawan pekerja lepas.
type Freelancer struct {
	HourlyRate  float64 // Tarif Per Jam
	HoursWorked int     // Total Jam Kerja dalam Sebulan
}

// GetEmployeeType mengembalikan jenis tipe karyawan.
func (fr Freelancer) GetEmployeeType() string {
	return "Freelance"
}

// CalculateSalary menghitung gaji bersih Pekerja Lepas (Freelancer).
// Rumus: HourlyRate * float64(HoursWorked)
func (fr Freelancer) CalculateSalary() (float64, error) {
	// Validasi komponen nilai keuangan
	if fr.HourlyRate < 0 || fr.HoursWorked < 0 {
		return 0, errors.New("komponen keuangan Freelancer tidak boleh bernilai negatif")
	}
	return fr.HourlyRate * float64(fr.HoursWorked), nil
}

// ============================================================================
// 3. DYNAMIC COLLECTION & HRIS STRUCT
// ============================================================================

// HRIS adalah struct utama untuk mengelola data kepegawaian dan penggajian secara dinamis.
type HRIS struct {
	Employees map[string]string            // Mapping: [EmployeeID] -> Nama Karyawan
	Payrolls  map[string]PayrollCalculator // Mapping: [EmployeeID] -> Interface PayrollCalculator
}

// NewHRIS adalah constructor untuk menginisialisasi HRIS beserta map internalnya.
func NewHRIS() *HRIS {
	return &HRIS{
		Employees: make(map[string]string),
		Payrolls:  make(map[string]PayrollCalculator),
	}
}

// ============================================================================
// 4. METHODS PADA HRIS (POINTER RECEIVER)
// ============================================================================

// RegisterEmployee mendaftarkan karyawan baru ke dalam sistem HRIS.
// ALUR LOGIKA VALIDASI:
// 1. Pastikan map internal sudah diinisialisasi (mencegah panic pada nil map).
// 2. Cek apakah nama karyawan kosong -> return error "nama karyawan tidak boleh kosong".
// 3. Cek apakah EmployeeID sudah ada di map -> return error "karyawan dengan ID tersebut sudah terdaftar".
// 4. Cek nilai keuangan karyawan (gaji/rate/jam kerja/pajak) -> return error jika ada bernilai negatif.
// 5. Simpan data ke map Employees dan Payrolls jika semua validasi lolos.
func (h *HRIS) RegisterEmployee(id string, name string, payroll PayrollCalculator) error {
	// Inisialisasi map jika pointer receiver belum meng-inisiasi map
	if h.Employees == nil {
		h.Employees = make(map[string]string)
	}
	if h.Payrolls == nil {
		h.Payrolls = make(map[string]PayrollCalculator)
	}

	// Validasi 1: Nama tidak boleh kosong
	if strings.TrimSpace(name) == "" {
		return errors.New("nama karyawan tidak boleh kosong")
	}

	// Validasi 2: ID tidak boleh duplikat
	if _, exists := h.Employees[id]; exists {
		return errors.New("karyawan dengan ID tersebut sudah terdaftar")
	}

	// Validasi 3: Cek nilai keuangan bernilai negatif melalui Type Switch & CalculateSalary
	if payroll == nil {
		return errors.New("payroll calculator tidak boleh nil")
	}

	switch v := payroll.(type) {
	case FullTimeEmployee:
		if v.BaseSalary < 0 || v.Allowance < 0 || v.TaxRate < 0 {
			return errors.New("nilai keuangan (gaji pokok/tunjangan/pajak) tidak boleh negatif")
		}
	case *FullTimeEmployee:
		if v.BaseSalary < 0 || v.Allowance < 0 || v.TaxRate < 0 {
			return errors.New("nilai keuangan (gaji pokok/tunjangan/pajak) tidak boleh negatif")
		}
	case ContractEmployee:
		if v.MonthlyRate < 0 || v.PerformanceBonus < 0 {
			return errors.New("nilai keuangan (rate bulanan/bonus) tidak boleh negatif")
		}
	case *ContractEmployee:
		if v.MonthlyRate < 0 || v.PerformanceBonus < 0 {
			return errors.New("nilai keuangan (rate bulanan/bonus) tidak boleh negatif")
		}
	case Freelancer:
		if v.HourlyRate < 0 || v.HoursWorked < 0 {
			return errors.New("nilai keuangan (rate per jam/jam kerja) tidak boleh negatif")
		}
	case *Freelancer:
		if v.HourlyRate < 0 || v.HoursWorked < 0 {
			return errors.New("nilai keuangan (rate per jam/jam kerja) tidak boleh negatif")
		}
	}

	// Verifikasi tambahan lewat kalkulasi
	_, err := payroll.CalculateSalary()
	if err != nil {
		return err
	}

	// Simpan ke dalam kedua map
	h.Employees[id] = name
	h.Payrolls[id] = payroll
	return nil
}

// CalculateTotalPayout menghitung total anggaran gaji seluruh karyawan aktif terdaftar.
// ALUR LOGIKA:
// 1. Inisialisasi akumulator total = 0.
// 2. Iterasi map Payrolls menggunakan `for range`.
// 3. Panggil method CalculateSalary() pada masing-masing objek payroll.
// 4. Tambahkan gaji bersih ke total akumulator.
// 5. Kembalikan total nilai pengeluaran gaji.
func (h *HRIS) CalculateTotalPayout() float64 {
	totalPayout := 0.0
	for _, payroll := range h.Payrolls {
		salary, err := payroll.CalculateSalary()
		if err == nil {
			totalPayout += salary
		}
	}
	return totalPayout
}

// PrintPayrollReport mencetak slip laporan keuangan penggajian secara terperinci.
// ALUR LOGIKA:
// 1. Iterasi map Employees menggunakan `for range` untuk mendapatkan ID dan Nama Karyawan.
// 2. Ambil objek PayrollCalculator dari map Payrolls berdasarkan ID.
// 3. Deteksi tipe konkret karyawan menggunakan TYPE SWITCH.
// 4. Cetak detail parameter spesifik sesuai tipe karyawan:
//    - FullTimeEmployee: Cetak Gaji Pokok, Tunjangan, dan Pajak (%)
//    - ContractEmployee: Cetak Rate Bulanan dan Bonus Kinerja
//    - Freelancer: Cetak Tarif Per Jam dan Jumlah Jam Kerja
// 5. Cetak Hasil Akhir Total Gaji Bersih Karyawan.
// 6. Cetak ringkasan Total Anggaran Gaji Seluruh Karyawan (Total Payout).
func (h *HRIS) PrintPayrollReport() {
	fmt.Println("==========================================================================")
	fmt.Println("                       LAPORAN PENGGAJIAN HRIS                            ")
	fmt.Println("==========================================================================")

	if len(h.Employees) == 0 {
		fmt.Println("Belum ada data karyawan yang terdaftar.")
		fmt.Println("==========================================================================")
		return
	}

	for id, name := range h.Employees {
		payroll, exists := h.Payrolls[id]
		if !exists {
			continue
		}

		fmt.Printf("\nID Karyawan   : %s\n", id)
		fmt.Printf("Nama Karyawan : %s\n", name)
		fmt.Printf("Tipe Karyawan : %s\n", payroll.GetEmployeeType())
		fmt.Println("--------------------------------------------------------------------------")

		// TYPE SWITCH: Mendeteksi tipe konkret struct untuk mencetak parameter spesifik
		switch emp := payroll.(type) {
		case FullTimeEmployee:
			fmt.Printf("  • Gaji Pokok    : Rp %.2f\n", emp.BaseSalary)
			fmt.Printf("  • Tunjangan     : Rp %.2f\n", emp.Allowance)
			fmt.Printf("  • Potongan Pajak: %.1f%% (Rate: %.2f)\n", emp.TaxRate*100, emp.TaxRate)
		case *FullTimeEmployee:
			fmt.Printf("  • Gaji Pokok    : Rp %.2f\n", emp.BaseSalary)
			fmt.Printf("  • Tunjangan     : Rp %.2f\n", emp.Allowance)
			fmt.Printf("  • Potongan Pajak: %.1f%% (Rate: %.2f)\n", emp.TaxRate*100, emp.TaxRate)
		case ContractEmployee:
			fmt.Printf("  • Rate Bulanan  : Rp %.2f\n", emp.MonthlyRate)
			fmt.Printf("  • Bonus Kinerja : Rp %.2f\n", emp.PerformanceBonus)
		case *ContractEmployee:
			fmt.Printf("  • Rate Bulanan  : Rp %.2f\n", emp.MonthlyRate)
			fmt.Printf("  • Bonus Kinerja : Rp %.2f\n", emp.PerformanceBonus)
		case Freelancer:
			fmt.Printf("  • Tarif Per Jam : Rp %.2f\n", emp.HourlyRate)
			fmt.Printf("  • Jam Kerja     : %d Jam\n", emp.HoursWorked)
		case *Freelancer:
			fmt.Printf("  • Tarif Per Jam : Rp %.2f\n", emp.HourlyRate)
			fmt.Printf("  • Jam Kerja     : %d Jam\n", emp.HoursWorked)
		default:
			fmt.Println("  • Parameter spesifik tidak diketahui.")
		}

		netSalary, err := payroll.CalculateSalary()
		if err != nil {
			fmt.Printf("  • Error Kalkulasi: %v\n", err)
		} else {
			fmt.Printf("  • TOTAL GAJI BERSIH: Rp %.2f\n", netSalary)
		}
		fmt.Println("--------------------------------------------------------------------------")
	}

	fmt.Println("==========================================================================")
	fmt.Printf("TOTAL ANGGARAN PENGELUARAN GAJI BULANAN (TOTAL PAYOUT): Rp %.2f\n", h.CalculateTotalPayout())
	fmt.Println("==========================================================================")
}

// ============================================================================
// MAIN FUNCTION & SIMULASI PENGGUNAAN
// ============================================================================
func main() {
	fmt.Println("=== SIMULASI SISTEM MANAJEMEN KEPEGAWAIAN & PAYROLL HRIS ===")

	// Inisialisasi HRIS
	hris := NewHRIS()

	// 1. Registrasi Karyawan Full-Time
	emp1 := FullTimeEmployee{
		BaseSalary: 10000000,
		Allowance:  2000000,
		TaxRate:    0.05, // Pajak 5%
	}
	err := hris.RegisterEmployee("EMP-001", "Budi Santoso", emp1)
	if err != nil {
		fmt.Printf("[Gagal Registrasi] %v\n", err)
	} else {
		fmt.Println("[Sukses] Registrasi EMP-001 (Budi Santoso)")
	}

	// 2. Registrasi Karyawan Contract
	emp2 := ContractEmployee{
		MonthlyRate:      8000000,
		PerformanceBonus: 1500000,
	}
	err = hris.RegisterEmployee("EMP-002", "Siti Aminah", emp2)
	if err != nil {
		fmt.Printf("[Gagal Registrasi] %v\n", err)
	} else {
		fmt.Println("[Sukses] Registrasi EMP-002 (Siti Aminah)")
	}

	// 3. Registrasi Freelancer
	emp3 := Freelancer{
		HourlyRate:  100000, // Rp 100.000 / jam
		HoursWorked: 120,    // 120 jam
	}
	err = hris.RegisterEmployee("EMP-003", "Rian Ardiansyah", emp3)
	if err != nil {
		fmt.Printf("[Gagal Registrasi] %v\n", err)
	} else {
		fmt.Println("[Sukses] Registrasi EMP-003 (Rian Ardiansyah)")
	}

	// 4. Simulasi Validasi Error: ID Duplikat
	fmt.Println("\n--- Testing Validasi 1: ID Duplikat ---")
	err = hris.RegisterEmployee("EMP-001", "Budi Baru", emp1)
	if err != nil {
		fmt.Printf("Uji Coba Duplikat ID -> Error Sesuai Ekspektasi: %v\n", err)
	}

	// 5. Simulasi Validasi Error: Nama Kosong
	fmt.Println("\n--- Testing Validasi 2: Nama Kosong ---")
	err = hris.RegisterEmployee("EMP-004", "   ", emp2)
	if err != nil {
		fmt.Printf("Uji Coba Nama Kosong -> Error Sesuai Ekspektasi: %v\n", err)
	}

	// 6. Simulasi Validasi Error: Nilai Keuangan Negatif
	fmt.Println("\n--- Testing Validasi 3: Nilai Keuangan Negatif ---")
	empInvalid := Freelancer{
		HourlyRate:  -50000,
		HoursWorked: 10,
	}
	err = hris.RegisterEmployee("EMP-005", "Doni Negatif", empInvalid)
	if err != nil {
		fmt.Printf("Uji Coba Nilai Negatif -> Error Sesuai Ekspektasi: %v\n", err)
	}

	// 7. Cetak Laporan HRIS
	fmt.Println()
	hris.PrintPayrollReport()
}
