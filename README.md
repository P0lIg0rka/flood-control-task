Когда завершите задачу, в этом README опишите свой ход мыслей: как вы пришли к решению, какие были варианты и почему выбрали именно этот. 

Для начала хотелось определиться с начальным решением более простой задачи (той же самой, но без поддержки паралельного вызова Check). В таком случае уже будет готово медленное решение без использования интересных фич языка (во избежании data race-ов надо будет залочить мьютексами обращение к общей памяти в такой реализации).

Перавя же идея реализации - в структуре хранить `map[int64][]int` который будет по `userID` хранить очередь его запросов. Тогда, при каждом вызове `Check(userID)` нкжно циклом проходиться по данному массиву и удалять все значения у которых разница по времени с `time.Now().Second()` > n секунд.

вот примерный код для такой структуры
```go
type FloodControler struct {
	n, k int
	controler map[int64][]int
}

func (fc FloodControler) Check(ctx context.Context, userID int64) (bool, error) {
	now := time.Now().Second()
	fc.controler[userID] = append(fc.controler[userID], now)
	for now - fc.controler[userID][0] > fc.k {
		fc.controler[userID] = fc.controler[userID][1:]
	}
	if len(fc.controler[userID]) >= fc.n {
		return true, nil
	}
	return false, nil
}
```

Как я уже упоминал, далее, чтобы сделать нашу функцию потокобезопасной необходимо убрать все места с гонкой памяти, а именно обращения к словарю, так как функции взаимодействия с ним из разных горутин вызовут undefined behavior. 

Если заменить map на sync.Map (залочить только места с самим обращением), то от проблемы мы не избавимся, так как между удалением элементов и проверкой на их количество в одной функции, вторая может забежать в цикл и удалить ещё парочку элементов всвязи с чем мы можем выдать неверный ответ.

По идее, у нас не получится паралельно обрабатывать 2 запроса для одного и того же пользователя (так как для корректной проверки в ифе в данной горутине нельзя чтобы любая другая заходила в цикл), но вот Check для различных юзеров хотелось бы обрабатывать вместе. Для этого предлагаю хранить в map небольшую структуру из очереди и канала, который будет контролировать для каждого пользователя очерёдность запросов.

# Что нужно сделать

Реализовать интерфейс с методом для проверки правил флуд-контроля. Если за последние N секунд вызовов метода Check будет больше K, значит, проверка на флуд-контроль не пройдена.

- Интерфейс FloodControl располагается в файле main.go.

- Флуд-контроль может быть запущен на нескольких экземплярах приложения одновременно, поэтому нужно предусмотреть общее хранилище данных. Допустимо использовать любое на ваше усмотрение. 

# Необязательно, но было бы круто

Хорошо, если добавите поддержку конфигурации итоговой реализации. Параметры — на ваше усмотрение.
