package e12

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestNome(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	mock.ExpectQuery(`SELECT nome FROM usuarios WHERE id = \$1`).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"nome"}).AddRow("Ana"))

	got, err := New(db).Nome(t.Context(), 7)
	if err != nil || got != "Ana" {
		t.Fatalf("got %q, %v", got, err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestNomeNaoEncontrado(t *testing.T) {
	db, mock, _ := sqlmock.New()
	defer db.Close()
	mock.ExpectQuery(`SELECT nome`).WithArgs(1).WillReturnError(sql.ErrNoRows)

	_, err := New(db).Nome(t.Context(), 1)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("esperava ErrNoRows, veio %v", err)
	}
}
