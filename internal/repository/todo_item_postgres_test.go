package repository

import (
	"errors"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/pavel-trbv/go-todo-app/internal/domain"
	"github.com/stretchr/testify/assert"
	"log"
	"testing"
)

func TestTodoItemPostgres_Create(t *testing.T) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		log.Fatal(err)
	}
	defer mockDB.Close()

	db := sqlx.NewDb(mockDB, "sqlmock")
	r := NewTodoItemPostgres(db)

	type args struct {
		listId int
		item   domain.TodoItem
	}
	type mockBehavior func(args args, id int)

	testTable := []struct {
		name         string
		mockBehavior mockBehavior
		args         args
		id           int
		wantErr      bool
	}{
		{
			name: "OK",
			args: args{
				listId: 1,
				item: domain.TodoItem{
					Title:       "test title",
					Description: "test description",
				},
			},
			id: 2,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("INSERT INTO todo_item").
					WithArgs(args.item.Title, args.item.Description).
					WillReturnRows(rows)

				mock.ExpectExec("INSERT INTO lists_items").
					WithArgs(args.listId, id).
					WillReturnResult(sqlmock.NewResult(1, 1))

				mock.ExpectCommit()
			},
		},
		{
			name: "Empty Fields",
			args: args{
				listId: 1,
				item: domain.TodoItem{
					Title:       "",
					Description: "test description",
				},
			},
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id).
					RowError(1, errors.New("some error"))
				mock.ExpectQuery("INSERT INTO todo_item").
					WithArgs(args.item.Title, args.item.Description).
					WillReturnRows(rows)

				mock.ExpectRollback()
			},
			wantErr: true,
		},
		{
			name: "2nd Insert Error",
			args: args{
				listId: 1,
				item: domain.TodoItem{
					Title:       "test title",
					Description: "test description",
				},
			},
			id: 2,
			mockBehavior: func(args args, id int) {
				mock.ExpectBegin()

				rows := sqlmock.NewRows([]string{"id"}).AddRow(id)
				mock.ExpectQuery("INSERT INTO todo_item").
					WithArgs(args.item.Title, args.item.Description).
					WillReturnRows(rows)

				mock.ExpectExec("INSERT INTO lists_items").
					WithArgs(args.listId, id).
					WillReturnError(errors.New("some error"))

				mock.ExpectRollback()
			},
			wantErr: true,
		},
	}

	for _, testCase := range testTable {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.mockBehavior(testCase.args, testCase.id)

			got, err := r.Create(testCase.args.listId, testCase.args.item)
			if testCase.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.id, got)
			}
		})
	}
}
