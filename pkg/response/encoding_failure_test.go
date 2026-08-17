package response

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestJSONDoesNotCommitSuccessBeforeEncoding(t *testing.T) {
	recorder := httptest.NewRecorder()
	JSON(recorder, http.StatusOK, map[string]any{"stream": make(chan int)})

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500; body=%q", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "unsupported type") {
		t.Fatalf("encoder details leaked: %q", recorder.Body.String())
	}
}
