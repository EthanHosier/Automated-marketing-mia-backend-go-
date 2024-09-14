package utils

import (
	"fmt"
	"reflect"
	"runtime"
)

type RetryFunc[T any] func() (T, error)

func Retry[T any](attempts int, fn RetryFunc[T]) (T, error) {
	var (
		err    error
		result T
	)

	fnName := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()

	for i := 0; i < attempts; i++ {
		fmt.Printf("Attempt %d: Calling function %s\n", i+1, fnName)
		result, err = fn()
		if err == nil {
			return result, nil
		}
		fmt.Printf("Attempt %d failed with error: %v\n", i+1, err)
	}

	return result, err
}

type Task[T any] struct {
	ch      chan T
	errorCh chan error
}

func DoAsync[T any](fn func() (T, error)) *Task[T] {
	ch := make(chan T)
	errorCh := make(chan error)

	go func() {
		result, err := fn()
		if err != nil {
			errorCh <- err
			return
		}
		ch <- result
	}()

	return &Task[T]{ch, errorCh}
}

func DoAsyncList[T any, U any](items []T, fn func(T) (U, error)) []*Task[U] {
	tasks := make([]*Task[U], len(items))

	for i, item := range items {
		tasks[i] = DoAsync(func() (U, error) {
			return fn(item)
		})
	}

	return tasks
}

func GetAsync[T any](task *Task[T]) (T, error) {
	var zero T // This will initialize `zero` to the zero value for type T
	select {
	case result := <-task.ch:
		return result, nil
	case err := <-task.errorCh:
		return zero, err
	}
}

func GetAsyncList[T any](tasks []*Task[T]) ([]T, error) {
	results := make([]T, len(tasks))

	for i, task := range tasks {
		result, err := GetAsync(task)
		if err != nil {
			return nil, err
		}
		results[i] = result
	}

	return results, nil
}
