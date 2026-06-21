package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
)

type ApiResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, resp ApiResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(resp)
}

func writeSuccess(w http.ResponseWriter, data interface{}) {
	writeJSON(w, http.StatusOK, ApiResponse{Code: 0, Message: "ok", Data: data})
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ApiResponse{Code: status, Message: msg})
}

func parseIDFromPath(path string, prefix string) (int64, error) {
	trimmed := strings.TrimPrefix(path, prefix)
	trimmed = strings.Trim(trimmed, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return 0, nil
	}
	return strconv.ParseInt(parts[0], 10, 64)
}

func handleMovies(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list, err := ListMovies()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, list)

	case http.MethodPost:
		var m Movie
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			writeError(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}
		if m.Title == "" {
			writeError(w, http.StatusBadRequest, "片名不能为空")
			return
		}
		if !ValidGenres[m.Genre] {
			writeError(w, http.StatusBadRequest, "无效的影片类型")
			return
		}
		if m.Duration <= 0 {
			writeError(w, http.StatusBadRequest, "时长必须是正整数")
			return
		}
		if err := CreateMovie(&m); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, m)

	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func handleMovie(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r.URL.Path, "/api/movies/")
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "无效的ID")
		return
	}
	switch r.Method {
	case http.MethodGet:
		m, err := GetMovie(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "影片不存在")
			return
		}
		writeSuccess(w, m)

	case http.MethodPut:
		var m Movie
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			writeError(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}
		m.ID = id
		if m.Title == "" {
			writeError(w, http.StatusBadRequest, "片名不能为空")
			return
		}
		if !ValidGenres[m.Genre] {
			writeError(w, http.StatusBadRequest, "无效的影片类型")
			return
		}
		if m.Duration <= 0 {
			writeError(w, http.StatusBadRequest, "时长必须是正整数")
			return
		}
		if err := UpdateMovie(&m); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, m)

	case http.MethodDelete:
		if err := DeleteMovie(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, nil)

	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func handleSessions(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list, err := ListSessions()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, list)

	case http.MethodPost:
		var s Session
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			writeError(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}
		if s.MovieID <= 0 {
			writeError(w, http.StatusBadRequest, "请选择影片")
			return
		}
		if s.ShowTime == "" {
			writeError(w, http.StatusBadRequest, "请选择放映时间")
			return
		}
		if s.TotalSlots <= 0 {
			writeError(w, http.StatusBadRequest, "停车位总数必须是正整数")
			return
		}
		if s.PriceFen <= 0 {
			writeError(w, http.StatusBadRequest, "票价必须是正整数分")
			return
		}
		if !ValidStatuses[s.Status] {
			s.Status = SessionStatusOnSale
		}
		if err := CreateSession(&s); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, s)

	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func handleSession(w http.ResponseWriter, r *http.Request) {
	id, err := parseIDFromPath(r.URL.Path, "/api/sessions/")
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "无效的ID")
		return
	}
	switch r.Method {
	case http.MethodGet:
		s, err := GetSession(id)
		if err != nil {
			writeError(w, http.StatusNotFound, "场次不存在")
			return
		}
		writeSuccess(w, s)

	case http.MethodPut:
		var s Session
		if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
			writeError(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}
		s.ID = id
		if s.MovieID <= 0 {
			writeError(w, http.StatusBadRequest, "请选择影片")
			return
		}
		if s.ShowTime == "" {
			writeError(w, http.StatusBadRequest, "请选择放映时间")
			return
		}
		if s.TotalSlots <= 0 {
			writeError(w, http.StatusBadRequest, "停车位总数必须是正整数")
			return
		}
		if s.PriceFen <= 0 {
			writeError(w, http.StatusBadRequest, "票价必须是正整数分")
			return
		}
		if !ValidStatuses[s.Status] {
			writeError(w, http.StatusBadRequest, "无效的场次状态")
			return
		}
		if err := UpdateSession(&s); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, s)

	case http.MethodDelete:
		if err := DeleteSession(id); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, nil)

	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func handleSessionSlots(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}
	id, err := parseIDFromPath(r.URL.Path, "/api/sessions/")
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "无效的ID")
		return
	}
	s, err := GetSession(id)
	if err != nil {
		writeError(w, http.StatusNotFound, "场次不存在")
		return
	}
	writeSuccess(w, map[string]int{
		"sold_slots": s.SoldSlots,
		"left_slots": s.LeftSlots,
	})
}

func handleBookings(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		list, err := ListBookings()
		if err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
		writeSuccess(w, list)

	case http.MethodPost:
		var b Booking
		if err := json.NewDecoder(r.Body).Decode(&b); err != nil {
			writeError(w, http.StatusBadRequest, "请求参数错误: "+err.Error())
			return
		}
		if b.Customer == "" {
			writeError(w, http.StatusBadRequest, "客户姓名不能为空")
			return
		}
		if b.Phone == "" {
			writeError(w, http.StatusBadRequest, "手机号不能为空")
			return
		}
		if b.SessionID <= 0 {
			writeError(w, http.StatusBadRequest, "请选择场次")
			return
		}
		if b.Slots <= 0 {
			writeError(w, http.StatusBadRequest, "购买车位数必须是正整数")
			return
		}
		s, err := GetSession(b.SessionID)
		if err != nil {
			writeError(w, http.StatusBadRequest, "场次不存在")
			return
		}
		b.TotalFen = b.Slots * s.PriceFen
		if err := CreateBooking(&b); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeSuccess(w, b)

	default:
		writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
	}
}

func handleBookingCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}
	id, err := parseIDFromPath(r.URL.Path, "/api/bookings/")
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "无效的ID")
		return
	}
	if err := CancelBooking(id); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeSuccess(w, nil)
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "不支持的方法")
		return
	}
	stats, err := MonthlyGenreStats()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeSuccess(w, stats)
}
