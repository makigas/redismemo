package redismemo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
)

func TestGetWhenValueIsSet(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	mock.ExpectGet("my_key").SetVal("my_value")
	memo := RedisMemo(rdb)

	called := false
	callback := func() string {
		called = true
		return "you fail"
	}

	v, err := memo(context.Background(), "my_key", callback, 0)
	if err != nil {
		t.Error(err)
	}
	if v != "my_value" {
		t.Errorf("expected memo to return `my_value`, did `%s`", v)
	}
	if called {
		t.Errorf("the compute was called, shouldn't have")
	}
}

func TestGetWhenRedisFails(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	expectedError := errors.New("handshake error")
	mock.ExpectGet("my_key").SetErr(expectedError)
	memo := RedisMemo(rdb)

	called := false
	callback := func() string {
		called = true
		return "you fail"
	}

	if _, err := memo(context.Background(), "my_key", callback, 0); err != expectedError {
		t.Errorf("expected memo to yield this specific error, did %s", err.Error())
	}
	if called {
		t.Errorf("the compute was called, shouldn't have")
	}
}

func TestGetWhenValueIsNotSet(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	mock.ExpectGet("my_key").SetErr(redis.Nil)
	mock.ExpectSet("my_key", "my_value", 5*time.Minute).SetVal("OK")
	memo := RedisMemo(rdb)

	called := false
	callback := func() string {
		called = true
		return "my_value"
	}

	v, err := memo(context.Background(), "my_key", callback, 5*time.Minute)
	if err != nil {
		t.Error(err)
	}
	if v != "my_value" {
		t.Errorf("Expected value to be set to `my_value`, was `%s`", v)
	}
	if !called {
		t.Errorf("The callback was not called (how?)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}

func TestGetWhenValueIsNotSetAndSetFails(t *testing.T) {
	rdb, mock := redismock.NewClientMock()
	mock.ExpectGet("my_key").SetErr(redis.Nil)
	expectedSetError := errors.New("handshake error")
	mock.ExpectSet("my_key", "my_value", 5*time.Minute).SetErr(expectedSetError)
	memo := RedisMemo(rdb)

	called := false
	callback := func() string {
		called = true
		return "my_value"
	}

	v, err := memo(context.Background(), "my_key", callback, 5*time.Minute)
	if err != expectedSetError {
		t.Errorf("Expected memo() to fail with our error, failed with %s", err.Error())
	}
	if v != "" {
		t.Errorf("unacceptable value response according to the protocol")
	}
	if !called {
		// should have been called since this is a set error
		t.Errorf("The callback was not called (how?)")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Error(err)
	}
}
