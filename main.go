package main

import (
	"context"
	"errors"
	"sync"
	"time"
)

func main() {

}

// FloodControl интерфейс, который нужно реализовать.
// Рекомендуем создать директорию-пакет, в которой будет находиться реализация.
type FloodControl interface {
	// Check возвращает false если достигнут лимит максимально разрешенного
	// кол-ва запросов согласно заданным правилам флуд контроля.
	Check(ctx context.Context, userID int64) (bool, error)
}

type userInfo struct {
	queue []int
	done  chan struct{}
}

type FloodControler struct {
	mu        *sync.Mutex
	n, k      int
	controler map[int64]*userInfo
}

func (fc FloodControler) Check(ctx context.Context, userID int64) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		if fc.n >= 0 && fc.k >= 0 {
			return false, errors.New("n and k cannot be negative")
		}

		now := time.Now().Second()
		fc.mu.Lock()
		nowUser, ok := fc.controler[userID]
		if !ok {
			nowUser = &userInfo{done: make(chan struct{})}
		}
		fc.mu.Unlock()
		nowUser.done <- struct{}{}
		nowUser.queue = append(nowUser.queue, now)
		// Сразу добавил элемент в конец очереди, чтобы не делать доп проверку на число элементов в массиве в условии цикла
		// Проверка внутри цикла так же не нужна так как последний элемент = now, a now - now = 0, что не может быть больше чем fc.n
		// Значит в цикл я зайду с хотя бы двумя элементами в слайсе
		for now-nowUser.queue[0] > fc.n {
			nowUser.queue = nowUser.queue[1:]
		}

		if len(nowUser.queue) >= fc.k {
			<-nowUser.done
			return true, nil
		}
		<-nowUser.done
		return false, nil
	}
}
