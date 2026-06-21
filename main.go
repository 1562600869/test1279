package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func main() {
	port := flag.Int("port", 8374, "服务端口")
	flag.Parse()

	if err := InitDB("./drivein.db"); err != nil {
		log.Fatal("数据库初始化失败:", err)
	}
	defer DB.Close()

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/api/movies", routeMovies)
	http.HandleFunc("/api/movies/", routeMovie)
	http.HandleFunc("/api/sessions", routeSessions)
	http.HandleFunc("/api/sessions/", routeSession)
	http.HandleFunc("/api/bookings", routeBookings)
	http.HandleFunc("/api/bookings/", routeBooking)
	http.HandleFunc("/api/stats/monthly-genre", handleStats)

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("汽车影院管理系统启动成功，访问地址: http://localhost%s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(IndexHTML))
}

func routeMovies(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	handleMovies(w, r)
}

func routeMovie(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/slots") {
		handleSessionSlots(w, r)
		return
	}
	handleMovie(w, r)
}

func routeSessions(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	handleSessions(w, r)
}

func routeSession(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.Contains(r.URL.Path, "/slots") {
		handleSessionSlots(w, r)
		return
	}
	handleSession(w, r)
}

func routeBookings(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	handleBookings(w, r)
}

func routeBooking(w http.ResponseWriter, r *http.Request) {
	setCORS(w)
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	if strings.HasSuffix(r.URL.Path, "/cancel") {
		handleBookingCancel(w, r)
		return
	}
	writeError(w, http.StatusNotFound, "未找到接口")
}

func setCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

var IndexHTML = func() string {
	data, err := os.ReadFile("index.html")
	if err != nil {
		return fallbackHTML
	}
	return string(data)
}()

var fallbackHTML = `<html><body><h1>Drive-in Cinema</h1><p>Loading...</p></body></html>`
