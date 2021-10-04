package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pavel-trbv/go-todo-app/internal/domain"
)

type TodoListPostgres struct {
	db *sqlx.DB
}

func NewTodoListPostgres(db *sqlx.DB) *TodoListPostgres {
	return &TodoListPostgres{db: db}
}

func (r *TodoListPostgres) Create(userId int, list domain.TodoList) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var id int
	createListQuery := fmt.Sprintf("INSERT INTO %s (title, description) VALUES ($1, $2) RETURNING id", todoListsTable)
	row := tx.QueryRow(createListQuery, list.Title, list.Description)
	if err := row.Scan(&id); err != nil {
		tx.Rollback()
		return 0, nil
	}

	createUsersListQuery := fmt.Sprintf("INSERT INTO %s (user_id, list_id) VALUES ($1, $2) RETURNING id", usersListsTable)
	if _, err := tx.Exec(createUsersListQuery, userId, id); err != nil {
		tx.Rollback()
		return 0, nil
	}

	return id, tx.Commit()
}

func (r *TodoListPostgres) GetAll(userId int) ([]domain.TodoList, error) {
	var lists []domain.TodoList

	query := fmt.Sprintf("SELECT tl.* FROM %s tl INNER JOIN %s ul ON tl.id = ul.list_id WHERE ul.user_id = $1",
		todoListsTable, usersListsTable)
	err := r.db.Select(&lists, query, userId)

	return lists, err
}

func (r *TodoListPostgres) GetById(userId, listId int) (domain.TodoList, error) {
	var list domain.TodoList

	query := fmt.Sprintf(
		`SELECT tl.* FROM %s tl 
				INNER JOIN %s ul ON tl.id = ul.list_id 
				WHERE ul.user_id = $1 AND ul.list_id = $2
				LIMIT 1`,
		todoListsTable,
		usersListsTable,
	)
	err := r.db.Get(&list, query, userId, listId)

	return list, err
}
