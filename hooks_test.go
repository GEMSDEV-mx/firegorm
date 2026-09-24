package firegorm

import (
	"context"
	"testing"
	"time"
)

func TestHookCanRegisterAnotherHookWithoutDeadlock(t *testing.T) {
	registry := NewHookRegistry()
	done := make(chan error, 1)
	registry.RegisterHook("items", PreCreate, func(context.Context, interface{}) error {
		registry.RegisterHook("items", PostCreate, func(context.Context, interface{}) error { return nil })
		return nil
	})
	go func() {
		done <- registry.RunHooks(context.Background(), "items", PreCreate, nil)
	}()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("hook registry deadlocked while a hook registered another hook")
	}
}

func TestLogIsSafeBeforeInit(t *testing.T) {
	Log(INFO, "logger is initialized at package load")
}

func TestCloseBeforeInitIsSafe(t *testing.T) {
	original := Client
	Client = nil
	t.Cleanup(func() { Client = original })
	if err := Close(); err != nil {
		t.Fatal(err)
	}
}
