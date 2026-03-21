package postgres_repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/vekshinnikita/pulse_watch/internal/dbs/postgres"
	"github.com/vekshinnikita/pulse_watch/internal/entities"
	"github.com/vekshinnikita/pulse_watch/internal/errs"
	"github.com/vekshinnikita/pulse_watch/internal/models"
	"github.com/vekshinnikita/pulse_watch/internal/testutils"
)

func TestMain(m *testing.M) {
	// подготовка перед тестами
	testutils.RepositorySetup()

	// запуск всех тестов
	code := m.Run()

	os.Exit(code)
}

func TestAuthPostgres_CreateUser(t *testing.T) {
	user := testutils.NewTestUser("test")

	userInput := &entities.SignUpUser{
		Name:     user.Name,
		Username: user.Username,
		Password: "Test1234",
		Email:    user.Email,
		TgId:     user.TgId,
	}

	testCases := []struct {
		name         string
		input        *entities.SignUpUser
		expected     int
		mockBehavior mockBehavior
		wantSomeErr  bool
		wantErr      error
	}{
		{
			name:     "OK",
			input:    userInput,
			expected: user.Id,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id"}).AddRow(user.Id)
				mock.ExpectQuery("INSERT INTO app_user").
					WithArgs(userInput.Name, userInput.Username, userInput.Password, userInput.Email, userInput.TgId).
					WillReturnRows(rows)
			},
		},

		{
			name:        "DB error",
			input:       userInput,
			wantSomeErr: true,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO app_user").
					WithArgs(userInput.Name, userInput.Username, userInput.Password, userInput.Email, userInput.TgId).
					WillReturnError(
						fmt.Errorf("some error"),
					)
			},
		},

		{
			name:    "Not unique username",
			input:   userInput,
			wantErr: &errs.UniqueFieldError{Field: "username"},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO app_user").
					WithArgs(userInput.Name, userInput.Username, userInput.Password, userInput.Email, userInput.TgId).
					WillReturnError(
						&pq.Error{
							Code:       postgres.UniqueViolationCode,
							Constraint: "app_user_username_key",
						},
					)
			},
		},

		{
			name:    "Not unique email",
			input:   userInput,
			wantErr: &errs.UniqueFieldError{Field: "email"},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO app_user").
					WithArgs(userInput.Name, userInput.Username, userInput.Password, userInput.Email, userInput.TgId).
					WillReturnError(
						&pq.Error{
							Code:       postgres.UniqueViolationCode,
							Constraint: "app_user_email_key",
						},
					)
			},
		},

		{
			name:    "Not unique tg id",
			input:   userInput,
			wantErr: &errs.UniqueFieldError{Field: "tg_id"},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("INSERT INTO app_user").
					WithArgs(userInput.Name, userInput.Username, userInput.Password, userInput.Email, userInput.TgId).
					WillReturnError(
						&pq.Error{
							Code:       postgres.UniqueViolationCode,
							Constraint: "app_user_tg_id_key",
						},
					)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mock, repo := NewMockAuthPostgres(t)
			testCase.mockBehavior(mock)

			id, err := repo.CreateUser(context.Background(), testCase.input)
			if testCase.wantSomeErr {
				assert.Error(t, err)
				return
			}

			if testCase.wantErr != nil {
				assert.Equal(t, err, testCase.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, id)
		})
	}
}

func TestAuthPostgres_GetUserById(t *testing.T) {
	user := testutils.NewTestUser("test")
	userFields, userValues := testutils.ExtractDBFields(user)

	testCases := []struct {
		name         string
		input        int
		expected     *models.User
		mockBehavior mockBehavior
		wantSomeErr  bool
		wantErr      error
	}{
		{
			name:     "OK",
			input:    user.Id,
			expected: user,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows(userFields).AddRow(userValues...)
				mock.ExpectQuery("SELECT (.+) FROM app_user").
					WithArgs(user.Id).
					WillReturnRows(rows)
			},
		},

		{
			name:        "DB error",
			wantSomeErr: true,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM app_user").
					WithArgs(user.Id).
					WillReturnError(fmt.Errorf("some error"))
			},
		},

		{
			name:    "User not found",
			input:   user.Id,
			wantErr: &errs.UserNotFoundError{Message: errs.UserNotFoundErrorMessage},
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM app_user").
					WithArgs(user.Id).
					WillReturnError(sql.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mock, repo := NewMockAuthPostgres(t)
			testCase.mockBehavior(mock)

			user, err := repo.GetUserById(context.Background(), testCase.input)
			if testCase.wantSomeErr {
				assert.Error(t, err)
				return
			}

			if testCase.wantErr != nil {
				assert.Equal(t, err, testCase.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, user)
		})
	}
}

func TestAuthPostgres_GetUserByUsernameAndPassword(t *testing.T) {
	user := testutils.NewTestUser("test")
	userFields, userValues := testutils.ExtractDBFields(user)

	input := &entities.SignInUser{
		Username: user.Username,
		Password: "Test1234",
	}

	testCases := []struct {
		name         string
		input        *entities.SignInUser
		expected     *models.User
		mockBehavior mockBehavior
		wantSomeErr  bool
		wantErr      error
	}{
		{
			name:     "OK",
			input:    input,
			expected: user,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows(userFields).AddRow(userValues...)
				mock.ExpectQuery("SELECT (.+) FROM app_user").
					WithArgs(input.Username, input.Password).
					WillReturnRows(rows)
			},
		},

		{
			name:        "DB error",
			wantSomeErr: true,
			input:       input,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM app_user").
					WithArgs(input.Username, input.Password).
					WillReturnError(fmt.Errorf("some error"))
			},
		},

		{
			name:    "User not found",
			wantErr: &errs.UserNotFoundError{Message: errs.UserNotFoundErrorMessage},
			input:   input,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery("SELECT (.+) FROM app_user").
					WithArgs(input.Username, input.Password).
					WillReturnError(sql.ErrNoRows)
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mock, repo := NewMockAuthPostgres(t)
			testCase.mockBehavior(mock)

			user, err := repo.GetUserByUsernameAndPassword(
				context.Background(),
				testCase.input.Username,
				testCase.input.Password,
			)
			if testCase.wantSomeErr {
				assert.Error(t, err)
				return
			}

			if testCase.wantErr != nil {
				assert.Equal(t, err, testCase.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, user)
		})
	}
}

func TestAuthPostgres_IsRefreshTokenValid(t *testing.T) {
	type args struct {
		userId int
		jti    string
	}
	input := &args{
		userId: 1,
		jti:    "jti",
	}

	testCases := []struct {
		name         string
		input        *args
		expected     bool
		mockBehavior mockBehavior
		wantSomeErr  bool
		wantErr      error
	}{
		{
			name:     "OK",
			input:    input,
			expected: true,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"exists"}).AddRow(true)
				mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM refresh_token (.+)\)`).
					WithArgs(input.jti, false, input.userId).
					WillReturnRows(rows)
			},
		},

		{
			name:     "Not valid token",
			input:    input,
			expected: false,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"exists"}).AddRow(false)
				mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM refresh_token (.+)\)`).
					WithArgs(input.jti, false, input.userId).
					WillReturnRows(rows)
			},
		},

		{
			name:        "DB error",
			wantSomeErr: true,
			input:       input,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`SELECT EXISTS \(SELECT 1 FROM refresh_token (.+)\)`).
					WithArgs(input.jti, false, input.userId).
					WillReturnError(fmt.Errorf("some error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mock, repo := NewMockAuthPostgres(t)
			testCase.mockBehavior(mock)

			ok, err := repo.IsRefreshTokenValid(
				context.Background(),
				testCase.input.userId,
				testCase.input.jti,
			)
			if testCase.wantSomeErr {
				assert.Error(t, err)
				return
			}

			if testCase.wantErr != nil {
				assert.Equal(t, err, testCase.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, ok)
		})
	}
}

func TestAuthPostgres_SaveRefreshToken(t *testing.T) {
	type args struct {
		userId    int
		jti       string
		expiresAt *time.Time
	}
	input := &args{
		userId:    1,
		jti:       "jti",
		expiresAt: &testutils.FixedTime,
	}

	testCases := []struct {
		name         string
		input        *args
		expected     int
		mockBehavior mockBehavior
		wantSomeErr  bool
		wantErr      error
	}{
		{
			name:     "OK",
			input:    input,
			expected: 1,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				rows := mock.NewRows([]string{"id"}).AddRow(1)
				mock.ExpectQuery(`INSERT INTO refresh_token (.+) RETURNING id`).
					WithArgs(input.jti, input.userId, input.expiresAt).
					WillReturnRows(rows)
			},
		},

		{
			name:        "DB error",
			wantSomeErr: true,
			input:       input,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`INSERT INTO refresh_token (.+) RETURNING id`).
					WithArgs(input.jti, input.userId, input.expiresAt).
					WillReturnError(fmt.Errorf("some error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mock, repo := NewMockAuthPostgres(t)
			testCase.mockBehavior(mock)

			ok, err := repo.SaveRefreshToken(
				context.Background(),
				testCase.input.userId,
				testCase.input.jti,
				testCase.input.expiresAt,
			)
			if testCase.wantSomeErr {
				assert.Error(t, err)
				return
			}

			if testCase.wantErr != nil {
				assert.Equal(t, err, testCase.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, testCase.expected, ok)
		})
	}
}

func TestAuthPostgres_RevokeRefreshToken(t *testing.T) {
	type args struct {
		jti string
	}
	input := &args{
		jti: "jti",
	}

	testCases := []struct {
		name         string
		input        *args
		mockBehavior mockBehavior
		wantSomeErr  bool
	}{
		{
			name:  "OK",
			input: input,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectExec(`UPDATE refresh_token`).
					WithArgs(true, input.jti).
					WillReturnResult(sqlmock.NewResult(0, 1))
			},
		},

		{
			name:        "DB error",
			wantSomeErr: true,
			input:       input,
			mockBehavior: func(mock sqlmock.Sqlmock) {
				mock.ExpectQuery(`UPDATE refresh_token`).
					WithArgs(true, input.jti).
					WillReturnError(fmt.Errorf("some error"))
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			mock, repo := NewMockAuthPostgres(t)
			testCase.mockBehavior(mock)

			err := repo.RevokeRefreshToken(
				context.Background(),
				testCase.input.jti,
			)
			if testCase.wantSomeErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
		})
	}
}
