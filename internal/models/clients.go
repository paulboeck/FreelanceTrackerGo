package models

import (
	"database/sql"
	"errors"
	"time"
)

type Client struct {
	ID      int
	Name    string
	Updated time.Time
	Created time.Time
}

type ClientModel struct {
	DB *sql.DB
}

func (c *ClientModel) Insert(name string) (int, error) {
	stmt := "INSERT INTO client (name) VALUES (?)"

	result, err := c.DB.Exec(stmt, name)
	if err != nil {
		return 0, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (c *ClientModel) Get(id int) (Client, error) {
	stmt := "SELECT id, name, updated_at, created_at FROM client WHERE id=?"

	row := c.DB.QueryRow(stmt, id)

	var client Client
	err := row.Scan(&client.ID, &client.Name, &client.Updated, &client.Created)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Client{}, ErrNoRecord
		} else {
			return Client{}, err
		}
	}
	return client, nil
}

func (c *ClientModel) GetAll() ([]Client, error) {
	stmt := "SELECT id, name, updated_at, created_at FROM client"

	rows, err := c.DB.Query(stmt)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []Client
	for rows.Next() {
		var client Client
		err := rows.Scan(&client.ID, &client.Name, &client.Updated, &client.Created)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	// When the rows.Next() loop has finished we call rows.Err() to retrieve any
	// error that was encountered during the iteration. It's important to
	// call this - don't assume the iteration completed successfully over the
	// entire result set.
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return clients, nil
}
