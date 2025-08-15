package response

import (
	"net/http"

	"multi-tenant-service/package/structs"

	"github.com/labstack/echo/v4"
)


func JSONSuccess(c echo.Context, result interface{}, msg string) error {
	return c.JSON(http.StatusOK, structs.Response{
		Result:     result,
		Message:    msg,
		Status:     true,
		StatusCode: http.StatusOK,
	})
 }
 

 func JSONResponse(c echo.Context, code int, status bool, message string, data interface{}) error {
    return c.JSON(code, map[string]interface{}{
        "status":  status,
        "message": message,
        "data":    data,
    })
}