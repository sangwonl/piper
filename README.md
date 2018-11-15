# Piper
is Go Chained Worker

## Why
In job/worker perspective, you might need series of workers chained. For example, assume that too many of jobs should be processed by fanout and in order accross many workers(`goroutines`), so it should be able to pass the results of over some results of a worker into another next worker and then do so... `Piper` provides chaining worker in simple way.

## Get
```bash
$ go get github.com/sangwonl/piper
```

## Example
```go
import "github.com/sangwonl/piper"

func seriesJobs() {
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
	})

	worker3 := piper.NewWorker(1, func(w *piper.Worker, job piper.Job) piper.Result {
		avg := job.(float32)
		return fmt.Sprintf("average is %.1f", avg)
	}).Get(func(w *piper.Worker, job piper.Job, result piper.Result) {
		msg := result.(string)
        fmt.Println(msg)    // "average is 5.5"
	})

	worker1.
		Chain(worker2).
		Chain(worker3)

	worker1.Go()
	worker1.Queue([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
	worker1.Done()
	worker1.Wait()    
}
```

## Feedback
If you have any idea or suggestions, always welcome to your PR and issues. Thank you.
