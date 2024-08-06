// Code generated by mockery v2.44.1. DO NOT EDIT.

package mocks

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	models "github.com/a-bondar/gophermart/internal/models"

	time "time"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// AuthenticateUser provides a mock function with given fields: ctx, login, password
func (_m *Service) AuthenticateUser(ctx context.Context, login string, password string) (string, error) {
	ret := _m.Called(ctx, login, password)

	if len(ret) == 0 {
		panic("no return value specified for AuthenticateUser")
	}

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (string, error)); ok {
		return rf(ctx, login, password)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, login, password)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, login, password)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateOrder provides a mock function with given fields: ctx, userID, orderNumber
func (_m *Service) CreateOrder(ctx context.Context, userID int, orderNumber string) (*models.Order, bool, error) {
	ret := _m.Called(ctx, userID, orderNumber)

	if len(ret) == 0 {
		panic("no return value specified for CreateOrder")
	}

	var r0 *models.Order
	var r1 bool
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, int, string) (*models.Order, bool, error)); ok {
		return rf(ctx, userID, orderNumber)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int, string) *models.Order); ok {
		r0 = rf(ctx, userID, orderNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Order)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int, string) bool); ok {
		r1 = rf(ctx, userID, orderNumber)
	} else {
		r1 = ret.Get(1).(bool)
	}

	if rf, ok := ret.Get(2).(func(context.Context, int, string) error); ok {
		r2 = rf(ctx, userID, orderNumber)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// CreateUser provides a mock function with given fields: ctx, login, password
func (_m *Service) CreateUser(ctx context.Context, login string, password string) error {
	ret := _m.Called(ctx, login, password)

	if len(ret) == 0 {
		panic("no return value specified for CreateUser")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, login, password)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetUserBalance provides a mock function with given fields: ctx, userID
func (_m *Service) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetUserBalance")
	}

	var r0 *models.Balance
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int) (*models.Balance, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int) *models.Balance); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Balance)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserOrders provides a mock function with given fields: ctx, userID
func (_m *Service) GetUserOrders(ctx context.Context, userID int) ([]struct {
	UploadedAt  time.Time          `json:"uploaded_at"`
	Status      models.OrderStatus `json:"status"`
	OrderNumber string             `json:"number"`
	Accrual     float64            `json:"accrual"`
}, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetUserOrders")
	}

	var r0 []struct {
		UploadedAt  time.Time          `json:"uploaded_at"`
		Status      models.OrderStatus `json:"status"`
		OrderNumber string             `json:"number"`
		Accrual     float64            `json:"accrual"`
	}
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int) ([]struct {
		UploadedAt  time.Time          `json:"uploaded_at"`
		Status      models.OrderStatus `json:"status"`
		OrderNumber string             `json:"number"`
		Accrual     float64            `json:"accrual"`
	}, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int) []struct {
		UploadedAt  time.Time          `json:"uploaded_at"`
		Status      models.OrderStatus `json:"status"`
		OrderNumber string             `json:"number"`
		Accrual     float64            `json:"accrual"`
	}); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]struct {
				UploadedAt  time.Time          `json:"uploaded_at"`
				Status      models.OrderStatus `json:"status"`
				OrderNumber string             `json:"number"`
				Accrual     float64            `json:"accrual"`
			})
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetUserWithdrawals provides a mock function with given fields: ctx, userID
func (_m *Service) GetUserWithdrawals(ctx context.Context, userID int) ([]struct {
	ProcessedAt time.Time `json:"processed_at"`
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
}, error) {
	ret := _m.Called(ctx, userID)

	if len(ret) == 0 {
		panic("no return value specified for GetUserWithdrawals")
	}

	var r0 []struct {
		ProcessedAt time.Time `json:"processed_at"`
		Order       string    `json:"order"`
		Sum         float64   `json:"sum"`
	}
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int) ([]struct {
		ProcessedAt time.Time `json:"processed_at"`
		Order       string    `json:"order"`
		Sum         float64   `json:"sum"`
	}, error)); ok {
		return rf(ctx, userID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int) []struct {
		ProcessedAt time.Time `json:"processed_at"`
		Order       string    `json:"order"`
		Sum         float64   `json:"sum"`
	}); ok {
		r0 = rf(ctx, userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]struct {
				ProcessedAt time.Time `json:"processed_at"`
				Order       string    `json:"order"`
				Sum         float64   `json:"sum"`
			})
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int) error); ok {
		r1 = rf(ctx, userID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Ping provides a mock function with given fields: ctx
func (_m *Service) Ping(ctx context.Context) error {
	ret := _m.Called(ctx)

	if len(ret) == 0 {
		panic("no return value specified for Ping")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context) error); ok {
		r0 = rf(ctx)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UserWithdrawBonuses provides a mock function with given fields: ctx, userID, orderNumber, sum
func (_m *Service) UserWithdrawBonuses(ctx context.Context, userID int, orderNumber string, sum float64) error {
	ret := _m.Called(ctx, userID, orderNumber, sum)

	if len(ret) == 0 {
		panic("no return value specified for UserWithdrawBonuses")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, int, string, float64) error); ok {
		r0 = rf(ctx, userID, orderNumber, sum)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService(t interface {
	mock.TestingT
	Cleanup(func())
}) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
