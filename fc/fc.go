package fc

import (
	"context"
	"errors"
	"time"
)

type userInfo struct {
	queue []int
	done  chan struct{}
}

type FloodControler struct {
	mu        chan struct{}
	n, k      int
	controler map[int64]*userInfo
}

func (fc FloodControler) Check(ctx context.Context, userID int64) (bool, error) {
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	default:
		if fc.n < 0 || fc.k < 0 {
			return false, errors.New("n and k cannot be negative")
		}

		now := time.Now().Second()

		// Получение элемента из мапы залочил, так как канал нужно создать при первом запросе от пользователя
		// А канала будет возникать data race и так как в каналах поддерживается принцип FIFO, то те горутины, которые пришли раньше - пройдут дальше тоже раньше
		fc.mu <- struct{}{}
		nowUser, ok := fc.controler[userID]
		if !ok {
			nowUser = &userInfo{done: make(chan struct{}, 1)}
		}

		nowUser.done <- struct{}{}
		defer func() {
			<-nowUser.done
		}()

		<-fc.mu
		// После этого в map не идёт новых записей, так что она не сможет реалоцировать, и для чтения можно дальше обращаться без опаски
		// А вот запись для каждого userID будет происходить в той очерёдности, в которой они пришли

		// Сразу добавил элемент в конец очереди, чтобы не делать доп проверку на число элементов в массиве в условии цикла
		// Проверка внутри цикла так же не нужна так как последний элемент = now, a now - now = 0, что не может быть больше чем fc.n
		// Значит в цикл я зайду с хотя бы двумя элементами в слайсе
		nowUser.queue = append(nowUser.queue, now)
		for now-nowUser.queue[0] > fc.n {
			nowUser.queue = nowUser.queue[1:]
		}
		if len(nowUser.queue) >= fc.k {
			return true, nil
		}
		return false, nil
	}
}

func CreateFC(n int, k int) *FloodControler {
	return &FloodControler{
		n:         n,
		k:         k,
		mu:        make(chan struct{}, 1),
		controler: make(map[int64]*userInfo),
	}
}
