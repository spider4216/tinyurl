package service

import (
	"context"
	"sync"
)

type BatchJob struct {
	IDs    []string // Тут будет чанк батча
	UserId string
}

type BatchResult struct {
	IDs []string
	Err error
}

func (s Service) GenerateChunk(ids []string, size int, userId string) chan BatchJob {
	out := make(chan BatchJob)

	go func() {
		defer close(out)

		for i := 0; i < len(ids); i += size {
			end := i + size

			if end > len(ids) {
				end = len(ids)
			}

			s.logger.Debug("Generate chunk ", ids[i:end])

			out <- BatchJob{
				IDs:    ids[i:end],
				UserId: userId,
			}
		}
	}()

	return out
}

func (s Service) WorkChunkDelete(ctx context.Context, in <-chan BatchJob) chan BatchResult {
	out := make(chan BatchResult)

	go func() {
		for data := range in {
			s.logger.Debug("Delete batch with chunk ", data.IDs)
			err := s.DeleteBatch(ctx, data.IDs, data.UserId)

			out <- BatchResult{
				IDs: data.IDs,
				Err: err,
			}
		}
	}()

	return out
}

func (s Service) FanOutDeleteBatch(ctx context.Context, in <-chan BatchJob) []chan BatchResult {
	numWorkers := 5

	// Массив каналов с результатами delete батчинга
	channels := make([]chan BatchResult, numWorkers)

	for i := 0; i < numWorkers; i++ {
		s.logger.Debug("Make worker with chunk delete ", i)

		ch := s.WorkChunkDelete(ctx, in)
		channels[i] = ch
	}

	return channels
}

// Объединяет результаты множества каналов в один
func (s Service) FanInDeleteBatch(in []chan BatchResult) chan BatchResult {
	finalCh := make(chan BatchResult)

	var wg sync.WaitGroup

	for _, ch := range in {
		chClosure := ch

		wg.Add(1)

		go func() {
			defer wg.Done()

			for data := range chClosure {
				s.logger.Debug("Collect data to one ch ", data)
				finalCh <- data
			}
		}()
	}

	go func() {
		wg.Wait()
		close(finalCh)
	}()

	return finalCh
}
