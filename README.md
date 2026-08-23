# Mini HRIS — Backend REST API

Backend **Human Resource Information System** sederhana yang dibangun dengan **Go**, **Gin Gonic**, **GORM**, dan **SQLite**. Aplikasi menangani data master kepegawaian, pencatatan kehadiran, pengajuan cuti, hingga perhitungan payroll bulanan secara dinamis.

---

## Cara Menjalankan

```bash
go mod tidy
go run .
```

Server berjalan di `http://localhost:8080`. Saat pertama kali dijalankan, aplikasi otomatis:

1. Membuat berkas database `hris.db`
2. Menjalankan migrasi 6 tabel beserta relasinya
3. Mengisi data awal (seeder) agar API langsung bisa diuji

Variabel lingkungan yang tersedia:

| Variabel   | Bawaan    | Keterangan                                |
| ---------- | --------- | ----------------------------------------- |
| `PORT`     | `8080`    | Port server HTTP                          |
| `DB_NAME`  | `hris.db` | Nama berkas database SQLite               |
| `GIN_MODE` | `debug`   | Isi `release` untuk menyembunyikan log Gin |

> **Catatan driver:** proyek ini memakai `github.com/glebarez/sqlite`, driver SQLite **murni Go**. Berbeda dengan `mattn/go-sqlite3`, driver ini tidak memerlukan CGO maupun compiler C, sehingga aplikasi tetap bisa dikompilasi di Windows tanpa memasang MinGW/gcc.

Untuk mengulang dari nol, hapus `hris.db` lalu jalankan ulang aplikasi.

---

## Struktur Folder

```
ch2/
├── config/           Koneksi dan migrasi database SQLite
│   └── database.go
├── models/           Definisi 6 struct entitas + DTO validasi
│   ├── department.go
│   ├── position.go
│   ├── employee.go
│   ├── attendance.go
│   ├── leave.go
│   └── salary.go
├── controllers/      Logika handler setiap endpoint API
│   ├── helper.go
│   ├── department_controller.go
│   ├── position_controller.go
│   ├── employee_controller.go
│   ├── attendance_controller.go
│   ├── leave_controller.go
│   └── salary_controller.go
├── routes/           Registrasi perutean Gin
│   └── routes.go
├── seeders/          Data awal untuk pengujian
│   └── seeder.go
├── utils/            Helper response API dan perhitungan tanggal
│   ├── response.go
│   └── date.go
└── main.go           Titik masuk utama aplikasi
```

---

## Skema Database

| Tabel         | Relasi                                              |
| ------------- | --------------------------------------------------- |
| `departments` | —                                                   |
| `positions`   | —                                                   |
| `employees`   | FK → `departments`, FK → `positions`                |
| `attendances` | FK → `employees`                                    |
| `leaves`      | FK → `employees`                                    |
| `salaries`    | FK → `employees`                                    |

**Aturan integritas:**

- `PRAGMA foreign_keys` diaktifkan lewat DSN, sehingga Foreign Key benar-benar ditegakkan SQLite.
- Departemen/jabatan **tidak dapat dihapus** selama masih dipakai karyawan (`OnDelete:RESTRICT`).
- Menghapus karyawan **ikut menghapus** absensi, cuti, dan slip gajinya (`OnDelete:CASCADE`).
- `attendances` punya index unik `(employee_id, date)` — mencegah check-in ganda.
- `salaries` punya index unik `(employee_id, period)` — mencegah slip gaji ganda.

**Kolom tambahan di luar spesifikasi minimum:**

| Kolom                                                          | Alasan                                                                     |
| -------------------------------------------------------------- | -------------------------------------------------------------------------- |
| `employees.leave_quota`, `employees.leave_balance`              | Menyimpan jatah dan sisa cuti, sesuai kebutuhan "kurangi saldo jatah cuti"  |
| `leaves.total_days`                                             | Durasi cuti, agar saldo tidak perlu menghitung ulang selisih tanggal        |
| `salaries.present_days`, `late_days`, `absent_days`, `leave_days`, `leave_balance` | Rincian asal-usul angka gaji agar slip dapat diaudit tanpa query tambahan |

---

## Daftar Endpoint

Semua response memakai kerangka seragam:

```json
{ "success": true, "message": "...", "data": {}, "meta": {}, "errors": {} }
```

### A. Data Master

| Method   | Endpoint                | Keterangan                                    |
| -------- | ----------------------- | --------------------------------------------- |
| `POST`   | `/api/departments`      | Tambah departemen                             |
| `GET`    | `/api/departments`      | Daftar departemen (`?search=`)                |
| `GET`    | `/api/departments/:id`  | Detail departemen                             |
| `PUT`    | `/api/departments/:id`  | Ubah departemen                               |
| `DELETE` | `/api/departments/:id`  | Hapus departemen (ditolak bila masih dipakai) |
| `POST`   | `/api/positions`        | Tambah jabatan                                |
| `GET`    | `/api/positions`        | Daftar jabatan (`?search=`)                   |
| `GET`    | `/api/positions/:id`    | Detail jabatan                                |
| `PUT`    | `/api/positions/:id`    | Ubah jabatan                                  |
| `DELETE` | `/api/positions/:id`    | Hapus jabatan (ditolak bila masih dipakai)    |

### B. Manajemen Karyawan

| Method   | Endpoint             | Keterangan                                        |
| -------- | -------------------- | ------------------------------------------------- |
| `POST`   | `/api/employees`     | Tambah karyawan (validasi FK departemen & jabatan) |
| `GET`    | `/api/employees`     | Daftar karyawan + detail relasi, pencarian & filter |
| `GET`    | `/api/employees/:id` | Detail karyawan                                   |
| `PUT`    | `/api/employees/:id` | Ubah karyawan, termasuk pindah departemen/jabatan  |
| `DELETE` | `/api/employees/:id` | Hapus karyawan                                    |

Filter pada `GET /api/employees` (dapat digabung):

```
?search=budi&department_id=1&position_id=2&status=ACTIVE
```

### C. Kehadiran & Cuti

| Method  | Endpoint                          | Keterangan                                |
| ------- | --------------------------------- | ----------------------------------------- |
| `POST`  | `/api/attendances`                | Check-in (status LATE otomatis > 08:00)   |
| `PATCH` | `/api/attendances/:id/checkout`   | Catat jam pulang                          |
| `GET`   | `/api/attendances`                | `?employee_id=&period=&date=&status=`     |
| `POST`  | `/api/leaves`                     | Ajukan cuti (status awal `PENDING`)       |
| `GET`   | `/api/leaves`                     | `?employee_id=&status=`                   |
| `PATCH` | `/api/leaves/:id/approve`         | HR menyetujui / menolak cuti              |

### D. Payroll

| Method | Endpoint                        | Keterangan                     |
| ------ | ------------------------------- | ------------------------------ |
| `POST` | `/api/salaries/calculate`       | Hitung & simpan gaji satu periode |
| `GET`  | `/api/salaries/period/:period`  | Daftar slip gaji satu bulan    |

---

## Aturan Perhitungan Payroll

```
BasicSalary = Position.BaseSalary karyawan
Allowance   = jumlah hari PRESENT  × Rp  50.000
Deductions  = jumlah hari LATE     × Rp  20.000
            + jumlah hari ABSENT   × Rp 100.000
NetSalary   = BasicSalary + Allowance − Deductions      (minimum 0)
```

**Perlakuan cuti:** hari cuti berstatus `APPROVED` tidak menghasilkan baris kehadiran, sehingga tidak terkena potongan absen. Saldo jatah cuti dihitung ulang dengan rumus `LeaveQuota − total hari cuti APPROVED pada tahun berjalan`, bukan sekadar dikurangi — sehingga endpoint aman dijalankan berulang kali (idempoten).

**Cakupan:** tanpa `employee_id`, payroll dihitung untuk seluruh karyawan `ACTIVE` saja. Menjalankan ulang periode yang sama akan **memperbarui** slip gaji lama, bukan membuat duplikat.

---

## Contoh Pengujian

```bash
# 1. Cek server
curl http://localhost:8080/health

# 2. Daftar karyawan beserta departemen & jabatannya
curl "http://localhost:8080/api/employees"

# 3. Pencarian + filter gabungan
curl "http://localhost:8080/api/employees?search=budi&status=ACTIVE"

# 4. Tambah karyawan
curl -X POST http://localhost:8080/api/employees \
  -H "Content-Type: application/json" \
  -d '{"nik":"EMP-010","full_name":"Nur Aisyah","email":"nur.aisyah@ptdika.co.id","department_id":1,"position_id":1}'

# 5. Check-in dan check-out
curl -X POST http://localhost:8080/api/attendances \
  -H "Content-Type: application/json" \
  -d '{"employee_id":2,"date":"2026-08-24","check_in":"08:40"}'

curl -X PATCH http://localhost:8080/api/attendances/1/checkout \
  -H "Content-Type: application/json" -d '{"check_out":"17:30"}'

# 6. Ajukan cuti, lalu setujui
curl -X POST http://localhost:8080/api/leaves \
  -H "Content-Type: application/json" \
  -d '{"employee_id":4,"start_date":"2026-08-26","end_date":"2026-08-28","reason":"Cuti tahunan keluarga"}'

curl -X PATCH http://localhost:8080/api/leaves/4/approve \
  -H "Content-Type: application/json" -d '{"status":"APPROVED"}'

# 7. Jalankan payroll dan lihat hasilnya
curl -X POST http://localhost:8080/api/salaries/calculate \
  -H "Content-Type: application/json" -d '{"period":"2026-08"}'

curl http://localhost:8080/api/salaries/period/2026-08
```

---

## Validasi & Kode Status

Kegagalan binding JSON selalu dibalas **400 Bad Request** dengan rincian per field dalam bahasa Indonesia:

```json
{
  "success": false,
  "message": "Validasi input gagal",
  "errors": {
    "email": "Field 'email' harus berupa alamat email yang valid",
    "department_id": "Field 'department_id' wajib diisi",
    "status": "Field 'status' hanya boleh bernilai: ACTIVE, SUSPENDED, TERMINATED"
  }
}
```

| Kode  | Dipakai untuk                                                          |
| ----- | ---------------------------------------------------------------------- |
| `200` | Permintaan berhasil                                                    |
| `201` | Data baru berhasil dibuat                                              |
| `400` | Validasi binding gagal, atau Foreign Key menunjuk data yang tidak ada   |
| `404` | Data tidak ditemukan                                                   |
| `409` | Bentrok data: duplikat NIK/email/kode, check-in ganda, cuti sudah diproses |
| `422` | Aturan bisnis dilanggar: karyawan non-aktif absen, saldo cuti kurang    |
| `500` | Kesalahan internal server                                              |

---

## Data Awal (Seeder)

Seeder hanya berjalan bila tabel `departments` masih kosong, sehingga menjalankan ulang aplikasi tidak menggandakan data.

- **4 departemen** — IT, HR, Finance, Marketing
- **5 jabatan** — Software Engineer sampai Marketing Executive (Rp 7 jt – Rp 14 jt)
- **6 karyawan** — 5 `ACTIVE` dan 1 `SUSPENDED` (agar filter status dan aturan "karyawan non-aktif tidak ikut payroll" langsung dapat diuji)
- **Kehadiran bulan berjalan** — hari kerja dari tanggal 1 sampai hari ini, dengan pola `PRESENT`/`LATE`/`ABSENT` yang deterministik sehingga hasil payroll dapat diverifikasi ulang
- **3 pengajuan cuti** — masing-masing berstatus `APPROVED`, `PENDING`, dan `REJECTED`
