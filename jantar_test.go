package jantar

import (
	"crypto/rand"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// modified version of https://github.com/cypriss/golang-mux-benchmark

// Returns a routeset with N *resources per namespace*. so N=1 gives about 15 routes
func resourceSetup(N int) (namespaces []string, resources []string, requests []*http.Request) {
	namespaces = []string{"admin", "api", "site"}
	resources = []string{}

	uniqueID := make([]byte, 32)
	rand.Read(uniqueID)
	cookie := &http.Cookie{Name: "JANTAR_ID", Value: hex.EncodeToString(uniqueID)}

	for i := 0; i < N; i += 1 {
		sha1 := sha1.New()
		io.WriteString(sha1, fmt.Sprintf("%d", i))
		strResource := fmt.Sprintf("%x", sha1.Sum(nil))
		resources = append(resources, strResource)
	}

	for _, ns := range namespaces {
		for _, res := range resources {
			req, _ := http.NewRequest("GET", "/"+ns+"/"+res, nil)
			req.AddCookie(cookie)
			requests = append(requests, req)
			req, _ = http.NewRequest("POST", "/"+ns+"/"+res, nil)
			req.AddCookie(cookie)
			requests = append(requests, req)
			req, _ = http.NewRequest("GET", "/"+ns+"/"+res+"/3937", nil)
			req.AddCookie(cookie)
			requests = append(requests, req)
			req, _ = http.NewRequest("PUT", "/"+ns+"/"+res+"/3937", nil)
			req.AddCookie(cookie)
			requests = append(requests, req)
			req, _ = http.NewRequest("DELETE", "/"+ns+"/"+res+"/3937", nil)
			req.AddCookie(cookie)
			requests = append(requests, req)
		}
	}

	return
}

func routeFor(namespaces []string, resources []string, csrf bool) http.Handler {
	j := setupServer(csrf)

	for _, ns := range namespaces {
		for _, res := range resources {
			j.AddRoute("GET", "/"+ns+"/"+res, helloHandler)
			j.AddRoute("POST", "/"+ns+"/"+res, helloHandler)
			j.AddRoute("GET", "/"+ns+"/"+res+"/:id", helloHandler)
			j.AddRoute("PUT", "/"+ns+"/"+res+"/:id", helloHandler)
			j.AddRoute("DELETE", "/"+ns+"/"+res+"/:id", helloHandler)
		}
	}

	return j
}

func benchmarkSimple(b *testing.B, csrf bool) {
	j := setupServer(csrf)

	j.AddRoute("GET", "/action", helloHandler)

	rw, req := testRequest("GET", "/action")

	uniqueID := make([]byte, 32)
	rand.Read(uniqueID)
	req.AddCookie(&http.Cookie{Name: "JANTAR_ID", Value: hex.EncodeToString(uniqueID)})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		j.ServeHTTP(rw, req)
	}
}

func BenchmarkSimple(b *testing.B) {
	benchmarkSimple(b, false)
}

func BenchmarkRoute15(b *testing.B) {
	benchmarkRoutes(b, 1, false)
}

func BenchmarkRoute75(b *testing.B) {
	benchmarkRoutes(b, 5, false)
}

func BenchmarkRoute150(b *testing.B) {
	benchmarkRoutes(b, 10, false)
}

func BenchmarkRoute300(b *testing.B) {
	benchmarkRoutes(b, 20, false)
}

func BenchmarkRoute3000(b *testing.B) {
	benchmarkRoutes(b, 200, false)
}

func BenchmarkCsrfSimple(b *testing.B) {
	benchmarkSimple(b, true)
}

func BenchmarkCsrfRoute15(b *testing.B) {
	benchmarkRoutes(b, 1, true)
}

func BenchmarkCsrfRoute75(b *testing.B) {
	benchmarkRoutes(b, 5, true)
}

func BenchmarkCsrfRoute150(b *testing.B) {
	benchmarkRoutes(b, 10, true)
}

func BenchmarkCsrfRoute300(b *testing.B) {
	benchmarkRoutes(b, 20, true)
}

func BenchmarkCsrfRoute3000(b *testing.B) {
	benchmarkRoutes(b, 200, true)
}

//
// Helpers:
//
func helloHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(rw, "hello")
}

func testRequest(method, path string) (*httptest.ResponseRecorder, *http.Request) {
	request, _ := http.NewRequest(method, path, nil)
	recorder := httptest.NewRecorder()

	return recorder, request
}

func setupServer(csrf bool) *Jantar {
	j := New(&Config{
		Hostname: "localhost",
		Port:     3000,
	})

	if !csrf {
		j.middleware = nil
	}

	Log.SetMinLevel(LogLevelPanic)

	return j
}

func benchmarkRoutes(b *testing.B, n int, csrf bool) {
	namespaces, resources, requests := resourceSetup(n)
	handler := routeFor(namespaces, resources, csrf)

	recorder := httptest.NewRecorder()
	reqID := 0
	b.ResetTimer()
	for i := 0; i < b.N*10; i++ {
		if reqID >= len(requests) {
			reqID = 0
		}
		req := requests[reqID]
		handler.ServeHTTP(recorder, req)

		if recorder.Code != 200 {
			panic("wat")
		}

		reqID += 1
	}
}
