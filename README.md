# Go HRIS & Payroll System REST API (Clean Architecture)

Proyek ini adalah sistem backend REST API mandiri untuk **HRIS (Human Resource Information System) & Payroll** yang dibangun menggunakan **Go**, framework **Gin**, ORM **GORM**, dan database **PostgreSQL**. Aplikasi dirancang dengan **Clean Architecture**, proteksi **IDOR**, otorisasi **RBAC (HRD vs EMPLOYEE)**, **Token Blacklist Database**, serta konsistensi transaksi keuangan menggunakan **Database Transactions (`db.Transaction`)** dan **Row-Level Locking (`FOR UPDATE`)**.

---

## 📌 Daftar Isi
- [Struktur Proyek (Clean Architecture)](#-struktur-proyek-clean-architecture)
- [Fitur Utama & Keamanan](#-fitur-utama--keamanan)
- [Prasyarat & Cara Instalasi](#-prasyarat--cara-instalasi)
- [Konfigurasi Environment (.env)](#-konfigurasi-environment-env)
- [Menjalankan Aplikasi](#-menjalankan-aplikasi)
- [Dokumentasi Endpoint API & Contoh Request](#-dokumentasi-endpoint-api--contoh-request)
  - [1. Autentikasi & Akun](#1-autentikasi--akun)
  - [2. Manajemen Penggajian (Payroll System)](#2-manajemen-penggajian-payroll-system)
  - [3. Pengajuan Cuti (Attendance & Proteksi IDOR)](#3-pengajuan-cuti-attendance--proteksi-idor)
- [Pengujian Otomatis (Unit Testing)](#-pengujian-otomatis-unit-testing)

---

## 🏗 Struktur Proyek (Clean Architecture)

```
go-hris-payroll-system/
├── cmd/
│   └── api/
│       └── main.go                  # Entrypoint, Auto Migration, Seeder & Dependency Injection
├── config/
│   └── config.go                    # Membaca .env & inisialisasi koneksi GORM PostgreSQL
├── domain/                          # Entitas murni & interface kontrak
│   ├── user.go                      # Entitas User & Interface UserRepository & AuthUsecase
│   ├── department.go                # Entitas Department (department_budgets) & Interface
│   ├── payroll.go                   # Entitas Payroll & Interface
│   ├── leave.go                     # Entitas Leave & Interface kueri defensif IDOR
│   ├── token_blacklist.go           # Entitas TokenBlacklist & Interface
│   └── errors.go                    # Definisi custom error domain
├── dto/                             # Data Transfer Objects & format response JSON terstandar
│   ├── auth_dto.go                  # Request & Response Autentikasi
│   ├── payroll_dto.go               # Request & Response Penggajian
│   ├── leave_dto.go                 # Request & Response Pengajuan Cuti
│   └── response.go                  # Helper format JSON standar (Success & Error)
├── repository/                      # Implementasi GORM Data Access Layer
│   ├── user_repository.go           # CRUD User
│   ├── department_repository.go     # Query Department & Row-Level Locking (FOR UPDATE)
│   ├── payroll_repository.go        # Save Payroll Transaction
│   ├── leave_repository.go         # Kueri defensif IDOR untuk Cuti
│   └── token_blacklist_repository.go# Simpan & Cek Token Blacklist di DB
├── usecase/                         # Business Logic Layer
│   ├── auth_usecase.go              # Logika Register, Login, Logout
│   ├── payroll_usecase.go           # Transaksi Penggajian (db.Transaction, Row Lock & Budget Deduct)
│   └── leave_usecase.go             # Logika Cuti & Validasi Hak Akses IDOR
├── delivery/                        # HTTP Delivery Layer (Gin)
│   ├── http/
│   │   ├── handler/                 # HTTP Handlers (Auth, Payroll, Leave)
│   │   ├── middleware/              # Middleware JWT Auth & RBAC (RequireRole)
│   │   └── router.go                # Registrasi Route & Custom Validator Gin
├── pkg/                             # Helper & Utilities
│   ├── utils/                       # Hashing Bcrypt & JWT Generator/Parser
│   └── validator/                   # Custom Validator Gin (@company.co.id)
├── .env.example
├── go.mod
├── go.sum
└── README.md
```

---

## 🛡 Fitur Utama & Keamanan

### 1. Custom Validator Email Domain (`@company.co.id`)
Seluruh pendaftaran karyawan divalidasi oleh **custom validator Gin** (`company_email`). Email wajib diakhiri domain `@company.co.id`.
*Implementasi*: [pkg/validator/custom_validator.go](file:///c:/Users/1/Desktop/code/challenge3/pkg/validator/custom_validator.go)

### 2. Autentikasi JWT & Token Blacklist di Database
Saat pengguna melakukan logout (`POST /api/v1/auth/logout`), token JWT aktif disimpan ke tabel `token_blacklists` di database. Middleware `AuthMiddleware` memeriksa database setiap kali request masuk; jika token ada di blacklist, request ditolak (`401 Unauthorized`).
*Implementasi*: [delivery/http/middleware/auth_middleware.go](file:///c:/Users/1/Desktop/code/challenge3/delivery/http/middleware/auth_middleware.go)

### 3. Otorisasi RBAC (`RequireRole`)
Middleware `RequireRole("HRD")` dan `RequireRole("EMPLOYEE", "HRD")` memisahkan hak akses antara HRD (Akses admin penuh) dan Employee (akses terbatas).
*Implementasi*: [delivery/http/middleware/rbac_middleware.go](file:///c:/Users/1/Desktop/code/challenge3/delivery/http/middleware/rbac_middleware.go)

### 4. Transaksi Database & Row-Level Locking (`FOR UPDATE`)
Proses penggajian bulanan (`POST /api/v1/payroll/process`) dibungkus dalam **`gorm.DB.Transaction`**:
- Mengunci baris anggaran departemen di PostgreSQL menggunakan **`FOR UPDATE`** (`clause.Locking{Strength: "UPDATE"}`).
- Menghitung Total Gaji = Gaji Pokok + Bonus.
- Memotong budget departemen (`department_budgets`).
- Jika sisa anggaran departemen **tidak mencukupi**, transaksi langsung di-**rollback** secara atomik dan mengembalikan pesan error informatif.
*Implementasi*: [usecase/payroll_usecase.go](file:///c:/Users/1/Desktop/code/challenge3/usecase/payroll_usecase.go)

### 5. Pengajuan Cuti & Mitigasi IDOR (Insecure Direct Object Reference)
Karyawan mengakses detail cuti via `GET /api/v1/leaves/:id`. Repository menjalankan **kueri defensif**:
- `EMPLOYEE`: Kueri dipaksa menambahkan `WHERE id = :id AND employee_id = :authenticated_user_id`. Karyawan tidak bisa menebak atau mengintip data cuti milik karyawan lain.
- `HRD`: Kueri mengambil data cuti tanpa memfilter `employee_id` untuk keperluan verifikasi/persetujuan.
*Implementasi*: [repository/leave_repository.go](file:///c:/Users/1/Desktop/code/challenge3/repository/leave_repository.go)

---

## 💻 Prasyarat & Cara Instalasi

1. **Go Compiler**: Versi 1.20 atau lebih baru.
2. **Database PostgreSQL**: Pastikan service PostgreSQL sudah berjalan dan database telah dibuat.

### ⚙️ Konfigurasi Environment (`.env`)

Buat file `.env` di root direktori (atau salin dari `.env.example`):

```env
PORT=8080
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=hris_payroll_db
DB_SSLMODE=disable
JWT_SECRET=super-secret-jwt-key-hris-2026
```

---

## 🚀 Menjalankan Aplikasi

Jalankan perintah berikut pada terminal:

```bash
go run cmd/api/main.go
```

Aplikasi akan secara otomatis:
1. Menjalankan **Auto-Migration** untuk tabel: `departments`, `users`, `payrolls`, `leaves`, `token_blacklists`.
2. Menjalankan **Seeder Data Awal**:
   - Departemen HRD (Budget: Rp100.000.000), IT (Budget: Rp50.000.000), Finance (Budget: Rp30.000.000).
   - Akun Admin HRD bawaan:
     - **Email**: `admin.hrd@company.co.id`
     - **Password**: `hrd123456`
     - **Role**: `HRD`

---

## 📖 Dokumentasi Endpoint API & Contoh Request

Format response sukses dan error konsisten menggunakan JSON:
```json
{
  "success": true,
  "message": "Pesan deskriptif",
  "data": { ... }
}
```

### 1. Autentikasi & Akun

#### A. Login User
- **Endpoint**: `POST /api/v1/auth/login`
- **Akses**: Public
- **Request Body**:
```json
{
  "email": "admin.hrd@company.co.id",
  "password": "hrd123456"
}
```
- **Response Success (200 OK)**:
```json
{
  "success": true,
  "message": "Login berhasil",
  "data": {
    "id": 1,
    "name": "Admin HRD Utama",
    "email": "admin.hrd@company.co.id",
    "role": "HRD",
    "token": "eyJhbGciOiJIUzI1Ni..."
  }
}
```

#### B. Registrasi Karyawan Baru
- **Endpoint**: `POST /api/v1/auth/register`
- **Akses**: Protected (`HRD` Only)
- **Header**: `Authorization: Bearer <JWT_TOKEN_HRD>`
- **Request Body**:
```json
{
  "name": "Budi Santoso",
  "email": "budi@company.co.id",
  "password": "password123",
  "role": "EMPLOYEE",
  "department_id": 2,
  "basic_salary": 8000000.00
}
```
- **Response Success (201 Created)**:
```json
{
  "success": true,
  "message": "Registrasi akun karyawan berhasil",
  "data": {
    "id": 2,
    "name": "Budi Santoso",
    "email": "budi@company.co.id",
    "role": "EMPLOYEE",
    "token": "eyJhbGciOiJIUzI1Ni..."
  }
}
```

#### C. Logout User
- **Endpoint**: `POST /api/v1/auth/logout`
- **Akses**: Protected (`HRD` & `EMPLOYEE`)
- **Header**: `Authorization: Bearer <JWT_TOKEN>`
- **Response Success (200 OK)**:
```json
{
  "success": true,
  "message": "Logout berhasil. Token telah dibatalkan."
}
```

---

### 2. Manajemen Penggajian (Payroll System)

#### Memproses Penggajian Bulanan
- **Endpoint**: `POST /api/v1/payroll/process`
- **Akses**: Protected (`HRD` Only)
- **Header**: `Authorization: Bearer <JWT_TOKEN_HRD>`
- **Request Body**:
```json
{
  "payrolls": [
    {
      "employee_id": 2,
      "bonus": 1000000.00
    }
  ]
}
```
- **Response Success (200 OK)**:
```json
{
  "success": true,
  "message": "Penggajian bulanan berhasil diproses dan anggaran departemen telah dipotong",
  "data": [
    {
      "id": 1,
      "employee_id": 2,
      "basic_salary": 8000000,
      "bonus": 1000000,
      "total_salary": 9000000,
      "period": "2026-08",
      "processed_at": "2026-08-28 19:20:00"
    }
  ]
}
```
- **Response Error (400 Bad Request - Budget Tidak Mencukupi & Automatic Rollback)**:
```json
{
  "success": false,
  "message": "proses penggajian gagal: sisa anggaran departemen tidak mencukupi: Anggaran departemen 'Engineering / IT' tidak mencukupi. Sisa budget: Rp500000.00, Dibutuhkan untuk Karyawan Budi Santoso (ID: 2): Rp9000000.00"
}
```

---

### 3. Pengajuan Cuti (Attendance & Proteksi IDOR)

#### A. Pengajuan Cuti Karyawan
- **Endpoint**: `POST /api/v1/leaves`
- **Akses**: Protected (`EMPLOYEE` & `HRD`)
- **Header**: `Authorization: Bearer <JWT_TOKEN>`
- **Request Body**:
```json
{
  "reason": "Cuti Tahunan Liburan Keluarga",
  "start_date": "2026-09-01",
  "end_date": "2026-09-05"
}
```
- **Response Success (201 Created)**:
```json
{
  "success": true,
  "message": "Pengajuan cuti berhasil dibuat",
  "data": {
    "id": 1,
    "employee_id": 2,
    "reason": "Cuti Tahunan Liburan Keluarga",
    "start_date": "2026-09-01",
    "end_date": "2026-09-05",
    "status": "PENDING",
    "created_at": "2026-08-28T19:20:00Z"
  }
}
```

#### B. Lihat Detail Cuti (Proteksi IDOR)
- **Endpoint**: `GET /api/v1/leaves/:id`
- **Akses**: Protected (`EMPLOYEE` & `HRD`)
- **Header**: `Authorization: Bearer <JWT_TOKEN>`
- **Skenario Mitigasi IDOR**:
  - Jika Employee A (ID 2) mengakses `GET /api/v1/leaves/1` (milik sendiri) -> **200 OK**.
  - Jika Employee A (ID 2) mencoba mengakses `GET /api/v1/leaves/2` (milik Employee B ID 3) -> **404 Not Found** ("Data pengajuan cuti tidak ditemukan atau Anda tidak memiliki akses").
  - Jika HRD mengakses `GET /api/v1/leaves/2` -> **200 OK**.

---

## 🧪 Pengujian Otomatis (Unit Testing)

Untuk menjalankan unit test untuk custom validator dan utility JWT/password:

```bash
go test -v ./...
```
