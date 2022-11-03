package monitor

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/baggage"
)

func TestGetBaggageLatency(t *testing.T) {
	type args struct {
		bag baggage.Baggage
		key string
	}
	tests := []struct {
		name string
		args args
		want int64
	}{
		{
			"normal input",
			args{
				createBaggage("latency", "3459"),
				"latency",
			},
			3459,
		},
		{
			"negative input",
			args{
				createBaggage("latency", "-99999"),
				"latency",
			},
			0,
		},
		{
			"empty value",
			args{
				createBaggage("empty", "3459"),
				"latency",
			},
			0,
		},
		{
			"large value",
			args{
				createBaggage("latency", "922337203685477580"),
				"latency",
			},
			922337203685477580,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getBaggageLatency(tt.args.bag, tt.args.key); got != tt.want {
				t.Errorf("Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func createBaggage(key string, value string) baggage.Baggage {
	member, _ := baggage.NewMember(key, value)
	bag, _ := baggage.New(member)
	return bag
}

func TestGetUsageHeaders(t *testing.T) {
	tests := []struct {
		name       string
		arg        *gin.Context
		origin     string
		token      string
		wantOrigin string
		wantToken  string
	}{
		{
			"authed",
			testContext(),
			"https://api.nakji.network",
			"Bearer testtoken123",
			"https://api.nakji.network",
			"testtoken123",
		},
		{
			"no auth",
			testContext(),
			"",
			"",
			"anonymous",
			"none",
		},
		{
			"token no origin",
			testContext(),
			"",
			"Bearer testtoken123",
			"anonymous",
			"none",
		},
		{
			"origin no token",
			testContext(),
			"https://api.nakji.network",
			"",
			"anonymous",
			"none",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.arg
			if tt.origin != "" {
				ctx.Request.Header.Set("Origin", tt.origin)
			}
			if tt.token != "" {
				ctx.Request.Header.Set("Authorization", tt.token)
			}
			if gotO, gotT := getApiAuthHeaders(ctx); gotO != tt.wantOrigin || gotT != tt.wantToken {
				t.Errorf("Get() = (%v, %v), want (%v, %v)", gotO, gotT, tt.wantOrigin, tt.wantToken)
			}
		})
	}
}

func testContext() *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = &http.Request{
		Header: make(http.Header),
	}
	return c
}
