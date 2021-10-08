package repository

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pavel-trbv/go-todo-app/internal/domain"
	"github.com/sirupsen/logrus"
	"reflect"
	"strings"
)

type TodoItemPostgres struct {
	db *sqlx.DB
}

func NewTodoItemPostgres(db *sqlx.DB) *TodoItemPostgres {
	return &TodoItemPostgres{db: db}
}

func (r *TodoItemPostgres) Create(listId int, item domain.TodoItem) (int, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var itemId int
	createItemQuery := fmt.Sprintf("INSERT INTO %s (title, description) VALUES ($1, $2) RETURNING id",
		todoItemsTable)

	row := tx.QueryRow(createItemQuery, item.Title, item.Description)
	if err := row.Scan(&itemId); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, rollbackErr
		}
		return 0, err
	}

	createListsItemsQuery := fmt.Sprintf("INSERT INTO %s (list_id, item_id) VALUES ($1, $2)", listsItemsTable)
	if _, err := tx.Exec(createListsItemsQuery, listId, itemId); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return 0, rollbackErr
		}
		return 0, err
	}

	return itemId, tx.Commit()
}

func (r *TodoItemPostgres) GetAll(userId, listId int) ([]domain.TodoItem, error) {
	var items []domain.TodoItem
	query := fmt.Sprintf(
		`SELECT ti.* FROM %s ti INNER JOIN %s li ON li.item_id = ti.id
				INNER JOIN %s ul ON ul.list_id = li.list_id WHERE li.list_id = $1 AND ul.user_id = $2`,
		todoItemsTable,
		listsItemsTable,
		usersListsTable,
	)

	if err := r.db.Select(&items, query, listId, userId); err != nil {
		return nil, err
	}

	return items, nil
}

func (r *TodoItemPostgres) GetById(userId, itemId int) (domain.TodoItem, error) {
	var item domain.TodoItem
	query := fmt.Sprintf(
		`SELECT ti.* FROM %s ti INNER JOIN %s li ON li.item_id = ti.id
				INNER JOIN %s ul ON ul.list_id = li.list_id WHERE ti.id = $1 AND ul.user_id = $2`,
		todoItemsTable,
		listsItemsTable,
		usersListsTable,
	)

	if err := r.db.Get(&item, query, itemId, userId); err != nil {
		return item, err
	}

	return item, nil
}

func (r *TodoItemPostgres) Delete(userId, itemId int) error {
	query := fmt.Sprintf(
		`DELETE FROM %s ti USING %s li, %s ul 
				WHERE ti.id = li.item_id AND li.list_id = ul.list_id 
				AND ul.user_id = $1 AND ti.id = $2`,
		todoItemsTable,
		listsItemsTable,
		usersListsTable,
	)

	_, err := r.db.Exec(query, userId, itemId)
	return err
}

func (r *TodoItemPostgres) Update(userId, itemId int, input domain.UpdateItemInput) error {
	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	v := reflect.ValueOf(input)
	for i := 0; i < v.NumField(); i++ {
		key := v.Type().Field(i).Name
		if !v.Field(i).IsNil() {
			value := v.Field(i).Elem().Interface()
			setValues = append(setValues, fmt.Sprintf("%s = $%d", key, argId))
			args = append(args, value)
			argId++
		}
	}

	setQuery := strings.Join(setValues, ", ")

	query := fmt.Sprintf(
		`UPDATE %s ti SET %s FROM %s li, %s ul 
				WHERE ti.id = li.item_id AND li.list_id = ul.list_id 
				AND ul.user_id = $%d AND ti.id = $%d`,
		todoItemsTable,
		setQuery,
		listsItemsTable,
		usersListsTable,
		argId,
		argId+1,
	)
	args = append(args, userId, itemId)

	logrus.Debugf("updateQuery: %s", query)
	logrus.Debugf("args: %s", args)

	_, err := r.db.Exec(query, args...)

	return err
}
