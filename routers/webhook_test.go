package routers

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestWebhookHandler(t *testing.T) {
	t.Run("accepts POST requests", func(t *testing.T) {
		body, err := os.ReadFile("../test_data/solana_swap_example.json")
		if err != nil {
			log.Fatalf("unable to read file: %v", err)
		}
		request, _ := http.NewRequest("POST", "/api/webhook", strings.NewReader(string(body)))
		response := httptest.NewRecorder()
		context, _ := gin.CreateTestContext(response)
		context.Request = request

		WebhookHandler(context)

		if status := response.Code; status != http.StatusOK {
			t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

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
