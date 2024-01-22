package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWebhookHandler(t *testing.T) {
	t.Run("rejects non-POST requests", func(t *testing.T) {
		request, _ := http.NewRequest("GET", "/api/webhook", nil)
		response := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(response)
		context.Request = request

		WebhookHandler(context)

		if status := response.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
		}
	})

	t.Run("rejects invalid post body requests with error", func(t *testing.T) {
		request, _ := http.NewRequest("POST", "/api/webhook", nil)
		response := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(response)
		context.Request = request

		WebhookHandler(context)

		if status := response.Code; status != http.StatusBadRequest {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}
	})
}
