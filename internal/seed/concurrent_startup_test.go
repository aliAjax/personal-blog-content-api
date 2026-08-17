package seed

import (
	"context"
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-sql-driver/mysql"
)

func TestSecondStartupSurvivesConcurrentCategoryInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery("SELECT id FROM categories WHERE slug = \\?").
		WithArgs("golang").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectExec("INSERT INTO categories").
		WithArgs("Go", "golang", "notes").
		WillReturnError(&mysql.MySQLError{Number: 1062, Message: "Duplicate entry 'golang'"})
	mock.ExpectQuery("SELECT id FROM categories WHERE slug = \\?").
		WithArgs("golang").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(int64(17)))

	id, err := ensureCategory(context.Background(), db, "Go", "golang", "notes")
	if err != nil {
		t.Fatalf("second startup failed after the other instance inserted first: %v", err)
	}
	if id != 17 {
		t.Fatalf("category id = %d, want 17", id)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
