package application

import (
	"myproject/infrastructure/testhelpers"
	"testing"

	"github.com/stretchr/testify/assert"
)

type updateTask struct {
	taskID, userID int
	description    *string
	done           *bool
}

func TestUpdateTask(t *testing.T) {
	tests := []struct {
		name                string
		up                  updateTask
		setupStore          *testhelpers.StubTaskStore
		expectedDescription string
		expectedDone        bool
		expectedUpdateCalls int
		wantErr             bool
	}{
		{
			name: "update both fields successfully",
			up: updateTask{
				taskID:      1,
				userID:      1,
				description: stringPtr("new task 1"),
				done:        boolPtr(true),
			},
			setupStore: &testhelpers.StubTaskStore{
				Tasks: map[int]string{
					1: "task 1",
					2: "task 2",
				},
				UpdateTaskCalled: 0,
			},
			expectedDescription: "new task 1",
			expectedDone:        true,
			expectedUpdateCalls: 1,
			wantErr:             false,
		},
		{
			name: "error when both fields are nil",
			up: updateTask{
				taskID:      1,
				userID:      1,
				description: nil,
				done:        nil,
			},
			setupStore: &testhelpers.StubTaskStore{
				Tasks: map[int]string{
					1: "task 1",
					2: "task 2",
				},
				UpdateTaskCalled: 0,
			},
			expectedDescription: "",
			expectedDone:        false,
			expectedUpdateCalls: 0,
			wantErr:             true,
		},
		{
			name: "error when task not found",
			up: updateTask{
				taskID:      200,
				userID:      1,
				description: stringPtr("new task 1"),
				done:        nil,
			},
			setupStore: &testhelpers.StubTaskStore{
				Tasks: map[int]string{
					1: "task 1",
					2: "task 2",
				},
				UpdateTaskCalled: 0,
			},
			expectedDescription: "",
			expectedDone:        false,
			expectedUpdateCalls: 0,
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := tt.setupStore
			service := NewService(store)

			task, err := service.UpdateTask(tt.up.taskID, tt.up.userID, tt.up.description, tt.up.done)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedUpdateCalls, store.UpdateTaskCalled)
			assert.Equal(t, tt.expectedDescription, task.Description)
			assert.Equal(t, tt.expectedDone, task.Done)
		})
	}
}

func stringPtr(s string) *string { return &s }
func boolPtr(b bool) *bool       { return &b }

func TestCreateTask(t *testing.T) {
	tests := []struct {
		name                string
		description         string
		expectedCreateCall  int
		expectedDescription string
		wantErr             bool
	}{
		{
			name:                "successfully created task",
			description:         "task 1",
			expectedCreateCall:  1,
			expectedDescription: "task 1",
			wantErr:             false,
		},
		{
			name:                "empty description",
			description:         "",
			expectedCreateCall:  0,
			expectedDescription: "",
			wantErr:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &testhelpers.StubTaskStore{}
			service := NewService(store)

			task, err := service.CreateTask(tt.description, 1)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedCreateCall, len(store.CreateCall))
			assert.Equal(t, tt.expectedDescription, task.Description)
			assert.False(t, task.Done)
		})
	}
}
