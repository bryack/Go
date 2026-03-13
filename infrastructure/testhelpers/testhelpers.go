package testhelpers

import (
	"myproject/domain"
)

type SpyTaskService struct {
	LastDescription string
	LastUserID      int
	ResultTask      domain.Task
	ResultErr       error
	TasksTable      []domain.Task
	GetTasksError   error
}

func (ts *SpyTaskService) CreateTask(description string, userID int) (domain.Task, error) {
	ts.LastDescription = description
	ts.LastUserID = userID
	return ts.ResultTask, ts.ResultErr
}

func (ts *SpyTaskService) UpdateTask(taskID, userID int, description *string, done *bool) (domain.Task, error) {
	return domain.Task{}, nil
}

func (ts *SpyTaskService) GetTasks(userID int) ([]domain.Task, error) {
	ts.LastUserID = userID
	return ts.TasksTable, ts.GetTasksError
}

type StubTaskStore struct {
	Tasks            map[int]string
	CreateCall       []int
	TasksTable       []domain.Task
	UpdateTaskCalled int
}

func (s *StubTaskStore) GetTaskByID(id int, userID int) (task domain.Task, err error) {
	t, ok := s.Tasks[id]
	if !ok {
		return domain.Task{}, domain.ErrTaskNotFound
	}
	return domain.Task{ID: id, Description: t}, nil
}

func (s *StubTaskStore) CreateTask(task domain.Task, userID int) (int, error) {
	s.CreateCall = append(s.CreateCall, task.ID)
	return task.ID, nil
}

func (s *StubTaskStore) LoadTasks(userID int) ([]domain.Task, error) {
	return s.TasksTable, nil
}

func (s *StubTaskStore) UpdateTask(task domain.Task, userID int) error {
	s.UpdateTaskCalled++
	s.Tasks[task.ID] = task.Description
	return nil
}

func (s *StubTaskStore) DeleteTask(id int, userID int) error {
	delete(s.Tasks, id)
	return nil
}

func (s *StubTaskStore) Close() error {
	return nil
}

type SpyAuthService struct {
	ResultToken  string
	ResultErr    error
	LastEmail    string
	LastPassword string
}

func (s *SpyAuthService) Register(email, password string) (string, error) {
	s.LastEmail = email
	s.LastPassword = password
	return s.ResultToken, s.ResultErr
}

func (s *SpyAuthService) Login(email, password string) (string, error) {
	s.LastEmail = email
	s.LastPassword = password
	return s.ResultToken, s.ResultErr
}

type StubTokenGenerator struct {
	Token  string
	Claims *domain.Claims
	Err    error
}

func (tg *StubTokenGenerator) GenerateToken(userID int) (string, error) {
	if tg.Err != nil {
		return "", tg.Err
	}
	tg.Claims.UserID = userID
	return tg.Token, nil
}
func (tg *StubTokenGenerator) ValidateToken(tokenString string) (*domain.Claims, error) {
	if tg.Err != nil {
		return nil, tg.Err
	}
	if tokenString == "" {
		return nil, tg.Err
	}
	return tg.Claims, nil
}
