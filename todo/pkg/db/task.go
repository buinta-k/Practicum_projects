package db

import "fmt"

func (s *Base) GetTask(id string) (*Task, error) {
	var task Task
	err := s.Conn.QueryRow(`SELECT *
		FROM scheduler
		WHERE id = ?`, id,
	).Scan(
		&task.ID,
		&task.Date,
		&task.Title,
		&task.Comment,
		&task.Repeat)

	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Base) PutTask(task Task) error {
	query := `UPDATE scheduler SET title = ?, date = ?, comment = ?, repeat = ? WHERE id = ?`
	res, err := s.Conn.Exec(query, task.Title, task.Date, task.Comment, task.Repeat, task.ID)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func (s *Base) DeleteTask(id string) error {
	query := `DELETE FROM scheduler WHERE id = ?`
	res, err := s.Conn.Exec(query, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}

func (s *Base) UpdateDate(next, id string) error {
	query := `UPDATE scheduler SET date = ? WHERE id = ?`
	res, err := s.Conn.Exec(query, next, id)
	if err != nil {
		return err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return fmt.Errorf(`incorrect id for updating task`)
	}
	return nil
}
