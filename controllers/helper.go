package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"go-jc-challange2/utils"
)

// ============================================================================
// HELPER INTERNAL CONTROLLER
// ============================================================================

// parseIDParam membaca parameter path bertipe angka (misalnya :id) dan
// mengubahnya menjadi uint. Bila nilainya bukan angka positif, fungsi ini
// langsung mengirim response 400 dan mengembalikan ok = false, sehingga
// controller pemanggil cukup melakukan `return`.
func parseIDParam(c *gin.Context, name string) (uint, bool) {
	raw := c.Param(name)

	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || value == 0 {
		utils.Error(c, http.StatusBadRequest,
			"Parameter '"+name+"' harus berupa angka bulat positif")
		return 0, false
	}
	return uint(value), true
}

// parseUintQuery membaca query string bertipe angka (misalnya ?department_id=2).
// Mengembalikan ok = false hanya bila nilainya ada tetapi tidak valid;
// query yang memang tidak dikirim dianggap sah dan menghasilkan found = false.
func parseUintQuery(c *gin.Context, name string) (value uint, found bool, ok bool) {
	raw := c.Query(name)
	if raw == "" {
		return 0, false, true
	}

	parsed, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || parsed == 0 {
		utils.Error(c, http.StatusBadRequest,
			"Query '"+name+"' harus berupa angka bulat positif")
		return 0, false, false
	}
	return uint(parsed), true, true
}
