package storage

import (
	"encoding/json"
	"errors"
	"myproject/task"
	"os"
)

type Storage interface {
	LoadTasks() ([]task.Task, error)
	SaveTasks(tasks []task.Task) error
}

type JsonStorage struct{}

var (
	ErrFileNotFound    = errors.New("file not found, tasks not downloaded")
	ErrParseJson       = errors.New("error parsing JSON")
	ErrConversionTask  = errors.New("tasks conversion error")
	ErrFailedWriteFile = errors.New("failed to write tasks.json")
)

// LoadTasks загружает задачи из файла tasks.json
func (j JsonStorage) LoadTasks() ([]task.Task, error) {
	// Попытка прочитать весь файл tasks.json
	data, err := os.ReadFile("tasks.json")
	if err != nil {
		// Если файл не существует или другая ошибка, возвращаем пустой список
		return []task.Task{}, ErrFileNotFound
	}
	// Декодируем JSON из []byte в срез Task
	var tasks []task.Task
	if err := json.Unmarshal(data, &tasks); err != nil {
		return []task.Task{}, ErrParseJson
	}
	return tasks, nil
}

// SaveTasks сохраняет задачи в файл tasks.json
func (j JsonStorage) SaveTasks(tasks []task.Task) error {
	// Преобразуем срез задач в JSON-формат ([]byte)
	data, err := json.Marshal(tasks)
	if err != nil {
		return ErrConversionTask
	}
	if err = os.WriteFile("tasks.json", data, 0644); err != nil {
		return ErrFailedWriteFile
	}
	return nil
}
