package main

import (
	"errors"
	"fmt"
	"log"

	"go.uber.org/dig"
)

// 4. Config: Thành phần gốc
type Config struct {
	ConnectionString string
}

func NewConfig() *Config {
	fmt.Println("... Đang tạo Config")
	return &Config{
		ConnectionString: "postgres://user:pass@localhost:5432/mydb",
	}
}

// ---

// 3. Database: Phụ thuộc vào Config
type Database struct {
	conn string
}

func NewDatabase(cfg *Config) (*Database, error) {
	fmt.Println("... Đang tạo Database")
	if cfg.ConnectionString == "" {
		return nil, errors.New("chuỗi kết nối rỗng")
	}
	return &Database{conn: cfg.ConnectionString}, nil
}

// ---

// 2. UserRepository: Phụ thuộc vào Database
type UserRepository struct {
	db *Database
}

func NewUserRepository(db *Database) *UserRepository {
	fmt.Println("... Đang tạo UserRepository")
	return &UserRepository{db: db}
}

func (ur *UserRepository) GetUser(id int) string {
	return fmt.Sprintf("User %d (từ DB: %s)", id, ur.db.conn)
}

// ---

// 1. Server: Phụ thuộc vào UserRepository
type Server struct {
	repo *UserRepository
}

func NewServer(repo *UserRepository) *Server {
	fmt.Println("... Đang tạo Server")
	return &Server{repo: repo}
}

func (s *Server) Start() {
	fmt.Println(">>> Server đang chạy...")
	user := s.repo.GetUser(123)
	fmt.Printf(">>> Xử lý yêu cầu: %s\n", user)
}

// ---
// HÀM MAIN
// ---
func main() {
	container := dig.New()

	// 2. Cung cấp (Provide)
	// Lưu ý: Thứ tự Provide không quan trọng
	if err := container.Provide(NewServer); err != nil {
		log.Fatal(err)
	}
	if err := container.Provide(NewUserRepository); err != nil {
		log.Fatal(err)
	}
	if err := container.Provide(NewDatabase); err != nil {
		log.Fatal(err)
	}
	if err := container.Provide(NewConfig); err != nil {
		log.Fatal(err)
	}

	// 3. Gọi (Invoke)
	fmt.Println("--- Bắt đầu Invoke ---")
	err := container.Invoke(func(server *Server) {
		server.Start()
	})
	fmt.Println("--- Kết thúc Invoke ---")

	if err != nil {
		log.Fatal(err)
	}
}
