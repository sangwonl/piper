package piper_test

import (
	"fmt"
	"testing"

	"github.com/sangwonl/piper"
)

func TestSingleWorker(t *testing.T) {
	fanoutSize := 3
	fruits := []string{"berry", "banana", "apple", "grape"}

	workerHandler := func(w *piper.Worker, job piper.Job) piper.Result {
		j := job.(string)
		if len(j) == 0 {
			t.Errorf("Expected job value should be > 0, but 0")
		}

		foundIdx := -1
		for idx, f := range fruits {
			if f == j {
				foundIdx = idx
				break
			}
		}
		return foundIdx
	}

	workerCallback := func(w *piper.Worker, job piper.Job, result piper.Result) {
		f := job.(string)
		idx := result.(int)

		if idx == -1 {
			if f != "pumpkin" {
				t.Errorf("Expected pumpkin if idx is -1, but %s", f)
			}
		} else {
			if f != fruits[idx] {
				t.Errorf("Expected %s, but %s", fruits[idx], f)
			}
		}
	}

	worker := piper.NewWorker(fanoutSize, workerHandler).Get(workerCallback)

	worker.Go()
	worker.Queue("berry")
	worker.Queue("apple")
	worker.Queue("pumpkin")
	worker.Done()
	worker.Wait()
}

func TestChainedWorker(t *testing.T) {
	worker1 := piper.NewWorker(1, func(w *piper.Worker, job piper.Job) piper.Result {
		sum := 0
		numList := job.([]int)
		for _, v := range numList {
			sum += v
		}
		return sum
	})

	worker2 := piper.NewWorker(1, func(w *piper.Worker, job piper.Job) piper.Result {
		sum := job.(int)
		return float32(sum) / 10
	}).Get(func(w *piper.Worker, job piper.Job, result piper.Result) {
		avg := result.(float32)
		if avg != 5.5 {
			t.Errorf("Expected avg is %.1f, but %.1f", 5.5, avg)
		}
	})

	worker3 := piper.NewWorker(1, func(w *piper.Worker, job piper.Job) piper.Result {
		avg := job.(float32)
		return fmt.Sprintf("average is %.1f", avg)
	}).Get(func(w *piper.Worker, job piper.Job, result piper.Result) {
		msg := result.(string)
		expected := "average is 5.5"
		if msg != expected {
			t.Errorf("Expected msg is '%s', but '%s'", expected, msg)
		}
	})

	worker1.
		Chain(worker2).
		Chain(worker3)

	worker1.Go()
	worker1.Queue([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	worker1.Done()
	worker1.Wait()
}
