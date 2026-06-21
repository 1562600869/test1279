package main

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

const (
	MovieGenreLove     = "爱情"
	MovieGenreAction   = "动作"
	MovieGenreComedy   = "喜剧"
	MovieGenreSciFi    = "科幻"
	MovieGenreMystery  = "悬疑"
	MovieGenreAnime    = "动画"

	SessionStatusOnSale = "售票中"
	SessionStatusEnded  = "已结束"
	SessionStatusCancel = "已取消"
)

var ValidGenres = map[string]bool{
	MovieGenreLove:    true,
	MovieGenreAction:  true,
	MovieGenreComedy:  true,
	MovieGenreSciFi:   true,
	MovieGenreMystery: true,
	MovieGenreAnime:   true,
}

var ValidStatuses = map[string]bool{
	SessionStatusOnSale: true,
	SessionStatusEnded:  true,
	SessionStatusCancel: true,
}

type Movie struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Genre    string `json:"genre"`
	Duration int    `json:"duration"`
}

type Session struct {
	ID         int64  `json:"id"`
	MovieID    int64  `json:"movie_id"`
	MovieTitle string `json:"movie_title,omitempty"`
	ShowTime   string `json:"show_time"`
	TotalSlots int    `json:"total_slots"`
	PriceFen   int    `json:"price_fen"`
	Status     string `json:"status"`
	SoldSlots  int    `json:"sold_slots,omitempty"`
	LeftSlots  int    `json:"left_slots,omitempty"`
}

type Booking struct {
	ID          int64  `json:"id"`
	SessionID   int64  `json:"session_id"`
	MovieTitle  string `json:"movie_title,omitempty"`
	ShowTime    string `json:"show_time,omitempty"`
	Customer    string `json:"customer"`
	Phone       string `json:"phone"`
	Slots       int    `json:"slots"`
	TotalFen    int    `json:"total_fen"`
	CreatedAt   string `json:"created_at"`
	IsCancelled bool   `json:"is_cancelled"`
}

type GenreStat struct {
	Genre        string `json:"genre"`
	SessionCount int    `json:"session_count"`
	BoxOfficeFen int    `json:"box_office_fen"`
}

func InitDB(path string) error {
	var err error
	DB, err = sql.Open("sqlite3", path+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		return err
	}
	DB.SetMaxOpenConns(1)

	schema := `
	CREATE TABLE IF NOT EXISTS movies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		genre TEXT NOT NULL,
		duration INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		movie_id INTEGER NOT NULL,
		show_time TEXT NOT NULL,
		total_slots INTEGER NOT NULL,
		price_fen INTEGER NOT NULL,
		status TEXT NOT NULL DEFAULT '售票中',
		FOREIGN KEY (movie_id) REFERENCES movies(id)
	);

	CREATE TABLE IF NOT EXISTS bookings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id INTEGER NOT NULL,
		customer TEXT NOT NULL,
		phone TEXT NOT NULL,
		slots INTEGER NOT NULL,
		total_fen INTEGER NOT NULL,
		created_at TEXT NOT NULL,
		is_cancelled INTEGER NOT NULL DEFAULT 0,
		FOREIGN KEY (session_id) REFERENCES sessions(id)
	);

	CREATE INDEX IF NOT EXISTS idx_sessions_movie ON sessions(movie_id);
	CREATE INDEX IF NOT EXISTS idx_bookings_session ON bookings(session_id);
	`
	_, err = DB.Exec(schema)
	return err
}

func CreateMovie(m *Movie) error {
	res, err := DB.Exec("INSERT INTO movies (title, genre, duration) VALUES (?, ?, ?)",
		m.Title, m.Genre, m.Duration)
	if err != nil {
		return err
	}
	m.ID, err = res.LastInsertId()
	return err
}

func ListMovies() ([]Movie, error) {
	rows, err := DB.Query("SELECT id, title, genre, duration FROM movies ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Movie
	for rows.Next() {
		var m Movie
		if err := rows.Scan(&m.ID, &m.Title, &m.Genre, &m.Duration); err != nil {
			return nil, err
		}
		list = append(list, m)
	}
	return list, rows.Err()
}

func GetMovie(id int64) (*Movie, error) {
	var m Movie
	err := DB.QueryRow("SELECT id, title, genre, duration FROM movies WHERE id=?", id).
		Scan(&m.ID, &m.Title, &m.Genre, &m.Duration)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func UpdateMovie(m *Movie) error {
	_, err := DB.Exec("UPDATE movies SET title=?, genre=?, duration=? WHERE id=?",
		m.Title, m.Genre, m.Duration, m.ID)
	return err
}

func DeleteMovie(id int64) error {
	_, err := DB.Exec("DELETE FROM movies WHERE id=?", id)
	return err
}

func CreateSession(s *Session) error {
	res, err := DB.Exec("INSERT INTO sessions (movie_id, show_time, total_slots, price_fen, status) VALUES (?, ?, ?, ?, ?)",
		s.MovieID, s.ShowTime, s.TotalSlots, s.PriceFen, s.Status)
	if err != nil {
		return err
	}
	s.ID, err = res.LastInsertId()
	return err
}

func ListSessions() ([]Session, error) {
	rows, err := DB.Query(`
		SELECT s.id, s.movie_id, s.show_time, s.total_slots, s.price_fen, s.status, m.title
		FROM sessions s JOIN movies m ON s.movie_id = m.id
		ORDER BY s.show_time DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Session
	for rows.Next() {
		var s Session
		if err := rows.Scan(&s.ID, &s.MovieID, &s.ShowTime, &s.TotalSlots, &s.PriceFen, &s.Status, &s.MovieTitle); err != nil {
			return nil, err
		}
		s.SoldSlots, s.LeftSlots = calcSessionSlots(s.ID)
		list = append(list, s)
	}
	return list, rows.Err()
}

func GetSession(id int64) (*Session, error) {
	var s Session
	err := DB.QueryRow(`
		SELECT s.id, s.movie_id, s.show_time, s.total_slots, s.price_fen, s.status, m.title
		FROM sessions s JOIN movies m ON s.movie_id = m.id WHERE s.id=?
	`, id).Scan(&s.ID, &s.MovieID, &s.ShowTime, &s.TotalSlots, &s.PriceFen, &s.Status, &s.MovieTitle)
	if err != nil {
		return nil, err
	}
	s.SoldSlots, s.LeftSlots = calcSessionSlots(id)
	return &s, nil
}

func UpdateSession(s *Session) error {
	_, err := DB.Exec("UPDATE sessions SET movie_id=?, show_time=?, total_slots=?, price_fen=?, status=? WHERE id=?",
		s.MovieID, s.ShowTime, s.TotalSlots, s.PriceFen, s.Status, s.ID)
	return err
}

func DeleteSession(id int64) error {
	_, err := DB.Exec("DELETE FROM sessions WHERE id=?", id)
	return err
}

func calcSessionSlots(sessionID int64) (sold, left int) {
	s, err := GetSessionTotal(sessionID)
	if err != nil {
		return 0, 0
	}
	row := DB.QueryRow(`
		SELECT COALESCE(SUM(slots), 0) FROM bookings
		WHERE session_id=? AND is_cancelled=0
	`, sessionID)
	var used int
	_ = row.Scan(&used)
	return used, s.TotalSlots - used
}

func GetSessionTotal(sessionID int64) (*Session, error) {
	var s Session
	err := DB.QueryRow("SELECT id, total_slots, status FROM sessions WHERE id=?", sessionID).
		Scan(&s.ID, &s.TotalSlots, &s.Status)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func CreateBooking(b *Booking) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var status string
	var totalSlots int
	err = tx.QueryRow("SELECT status, total_slots FROM sessions WHERE id=?", b.SessionID).
		Scan(&status, &totalSlots)
	if err != nil {
		return fmt.Errorf("场次不存在: %w", err)
	}
	if status != SessionStatusOnSale {
		return fmt.Errorf("场次状态为 %s，无法订票", status)
	}

	var used int
	err = tx.QueryRow(`
		SELECT COALESCE(SUM(slots), 0) FROM bookings
		WHERE session_id=? AND is_cancelled=0
	`, b.SessionID).Scan(&used)
	if err != nil {
		return err
	}
	if used+b.Slots > totalSlots {
		return fmt.Errorf("剩余车位不足，仅剩 %d 个", totalSlots-used)
	}

	b.CreatedAt = time.Now().Format("2006-01-02 15:04:05")
	b.IsCancelled = false
	res, err := tx.Exec(`
		INSERT INTO bookings (session_id, customer, phone, slots, total_fen, created_at, is_cancelled)
		VALUES (?, ?, ?, ?, ?, ?, 0)
	`, b.SessionID, b.Customer, b.Phone, b.Slots, b.TotalFen, b.CreatedAt)
	if err != nil {
		return err
	}
	b.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	return tx.Commit()
}

func ListBookings() ([]Booking, error) {
	rows, err := DB.Query(`
		SELECT b.id, b.session_id, b.customer, b.phone, b.slots, b.total_fen, b.created_at, b.is_cancelled,
		       s.show_time, m.title
		FROM bookings b
		JOIN sessions s ON b.session_id = s.id
		JOIN movies m ON s.movie_id = m.id
		ORDER BY b.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []Booking
	for rows.Next() {
		var b Booking
		var cancelledInt int
		if err := rows.Scan(&b.ID, &b.SessionID, &b.Customer, &b.Phone, &b.Slots, &b.TotalFen,
			&b.CreatedAt, &cancelledInt, &b.ShowTime, &b.MovieTitle); err != nil {
			return nil, err
		}
		b.IsCancelled = cancelledInt == 1
		list = append(list, b)
	}
	return list, rows.Err()
}

func CancelBooking(id int64) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var cancelled int
	var sessionID int64
	var slots int
	err = tx.QueryRow("SELECT is_cancelled, session_id, slots FROM bookings WHERE id=?", id).
		Scan(&cancelled, &sessionID, &slots)
	if err != nil {
		return fmt.Errorf("订票记录不存在: %w", err)
	}
	if cancelled == 1 {
		return fmt.Errorf("该订票已退票")
	}

	_, err = tx.Exec("UPDATE bookings SET is_cancelled=1 WHERE id=?", id)
	if err != nil {
		return err
	}

	return tx.Commit()
}

func MonthlyGenreStats() ([]GenreStat, error) {
	now := time.Now()
	year, month, _ := now.Date()
	start := fmt.Sprintf("%04d-%02d-01", year, int(month))
	end := fmt.Sprintf("%04d-%02d-01", year, int(month)+1)
	if month == 12 {
		end = fmt.Sprintf("%04d-01-01", year+1)
	}

	rows, err := DB.Query(`
		SELECT m.genre,
		       COUNT(DISTINCT s.id) AS session_count,
		       COALESCE(SUM(b.slots * s.price_fen), 0) AS box_office
		FROM movies m
		LEFT JOIN sessions s ON s.movie_id = m.id
		LEFT JOIN bookings b ON b.session_id = s.id AND b.is_cancelled = 0
		WHERE s.show_time >= ? AND s.show_time < ?
		GROUP BY m.genre
		ORDER BY box_office DESC
	`, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []GenreStat
	for rows.Next() {
		var g GenreStat
		if err := rows.Scan(&g.Genre, &g.SessionCount, &g.BoxOfficeFen); err != nil {
			return nil, err
		}
		list = append(list, g)
	}
	return list, rows.Err()
}
