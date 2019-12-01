package pages

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	tests := []struct {
		name           string
		in             *http.Request
		out            *httptest.ResponseRecorder
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "good",
			in:             httptest.NewRequest("GET", "/", nil),
			out:            httptest.NewRecorder(),
			expectedStatus: http.StatusOK,
			expectedBody:   message,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			h := New(nil, nil)
			h.Home(test.out, test.in)
			if test.out.Code != test.expectedStatus {
				t.Logf("Error! Status expected: %d but got: %d\n", test.out.Code, test.expectedStatus)
				t.Fail()
			}
			body := test.out.Body.String()
			if !strings.Contains(body, message) {
				t.Logf("Error! Not the right body. Expected: %s but got: %s", message, body)
				t.Fail()
			}
		})
	}
}
