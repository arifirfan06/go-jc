package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go-jc-challange2/controllers"
	"go-jc-challange2/utils"
)

// ============================================================================
// REGISTRASI PERUTEAN (ROUTING) GIN
// ============================================================================

// SetupRouter membuat instance Gin dan mendaftarkan seluruh endpoint API.
// Semua rute dikelompokkan di bawah prefix /api sesuai ketentuan soal.
func SetupRouter() *gin.Engine {
	router := gin.Default()

	// Endpoint kesehatan aplikasi, berguna untuk memastikan server sudah hidup
	// sebelum mulai menguji endpoint lain.
	router.GET("/health", func(c *gin.Context) {
		utils.Success(c, http.StatusOK, "Mini HRIS API berjalan normal", gin.H{
			"service": "Mini HRIS",
			"status":  "UP",
		})
	})

	// Halaman ringkasan daftar endpoint, memudahkan penguji menelusuri API.
	router.GET("/", handleIndex)

	api := router.Group("/api")
	{
		// --------------------------------------------------------------
		// A. DATA MASTER - DEPARTEMEN (CRUD LENGKAP)
		// --------------------------------------------------------------
		departments := api.Group("/departments")
		{
			departments.POST("", controllers.CreateDepartment)
			departments.GET("", controllers.GetDepartments)
			departments.GET("/:id", controllers.GetDepartmentByID)
			departments.PUT("/:id", controllers.UpdateDepartment)
			departments.DELETE("/:id", controllers.DeleteDepartment)
		}

		// --------------------------------------------------------------
		// A. DATA MASTER - JABATAN (CRUD LENGKAP)
		// --------------------------------------------------------------
		positions := api.Group("/positions")
		{
			positions.POST("", controllers.CreatePosition)
			positions.GET("", controllers.GetPositions)
			positions.GET("/:id", controllers.GetPositionByID)
			positions.PUT("/:id", controllers.UpdatePosition)
			positions.DELETE("/:id", controllers.DeletePosition)
		}

		// --------------------------------------------------------------
		// B. MANAJEMEN KARYAWAN (RELASIONAL)
		// --------------------------------------------------------------
		employees := api.Group("/employees")
		{
			employees.POST("", controllers.CreateEmployee)
			// GET mendukung ?search=, ?department_id=, ?position_id=, ?status=
			employees.GET("", controllers.GetEmployees)
			employees.GET("/:id", controllers.GetEmployeeByID)
			employees.PUT("/:id", controllers.UpdateEmployee)
			employees.DELETE("/:id", controllers.DeleteEmployee)
		}

		// --------------------------------------------------------------
		// C. PENCATATAN KEHADIRAN
		// --------------------------------------------------------------
		attendances := api.Group("/attendances")
		{
			attendances.POST("", controllers.CreateAttendance)
			attendances.GET("", controllers.GetAttendances)
			attendances.PATCH("/:id/checkout", controllers.CheckOutAttendance)
		}

		// --------------------------------------------------------------
		// C. PENGAJUAN CUTI
		// --------------------------------------------------------------
		leaves := api.Group("/leaves")
		{
			leaves.POST("", controllers.CreateLeave)
			leaves.GET("", controllers.GetLeaves)
			leaves.PATCH("/:id/approve", controllers.ApproveLeave)
		}

		// --------------------------------------------------------------
		// D. PROSES PAYROLL BULANAN
		// --------------------------------------------------------------
		salaries := api.Group("/salaries")
		{
			salaries.POST("/calculate", controllers.CalculateSalaries)
			salaries.GET("/period/:period", controllers.GetSalariesByPeriod)
		}
	}

	// Penangan rute tidak dikenal, supaya klien menerima JSON yang konsisten
	// alih-alih halaman 404 bawaan Gin berformat teks biasa.
	router.NoRoute(func(c *gin.Context) {
		utils.Error(c, http.StatusNotFound,
			"Endpoint '"+c.Request.Method+" "+c.Request.URL.Path+"' tidak tersedia")
	})

	return router
}

// handleIndex menampilkan ringkasan seluruh endpoint yang tersedia.
func handleIndex(c *gin.Context) {
	utils.Success(c, http.StatusOK, "Selamat datang di Mini HRIS API", gin.H{
		"data_master": gin.H{
			"departments": []string{
				"POST   /api/departments",
				"GET    /api/departments",
				"GET    /api/departments/:id",
				"PUT    /api/departments/:id",
				"DELETE /api/departments/:id",
			},
			"positions": []string{
				"POST   /api/positions",
				"GET    /api/positions",
				"GET    /api/positions/:id",
				"PUT    /api/positions/:id",
				"DELETE /api/positions/:id",
			},
		},
		"karyawan": []string{
			"POST   /api/employees",
			"GET    /api/employees?search=&department_id=&position_id=&status=",
			"GET    /api/employees/:id",
			"PUT    /api/employees/:id",
			"DELETE /api/employees/:id",
		},
		"kehadiran": []string{
			"POST   /api/attendances",
			"GET    /api/attendances?employee_id=&period=&date=&status=",
			"PATCH  /api/attendances/:id/checkout",
		},
		"cuti": []string{
			"POST   /api/leaves",
			"GET    /api/leaves?employee_id=&status=",
			"PATCH  /api/leaves/:id/approve",
		},
		"payroll": []string{
			"POST   /api/salaries/calculate",
			"GET    /api/salaries/period/:period",
		},
	})
}
