# Практическая работа №13
## Профилирование Go-приложения (pprof). Измерение времени работы функций
### Саттаров Булат Рамилевич, ЭФМО-01-25

---

> В этом репозитории дополнительно выполнено профилирование и оптимизация производительности HTTP-обработчика с помощью **pprof** (CPU/heap), таймеров и нагрузочного тестирования **hey**. Ниже — отчёт в формате README с результатами “до/после”.

## 1) Цель работы и используемый стек

**Цель:** найти узкое место по производительности в обработчике запроса и ускорить работу сервиса на основе результатов профилирования.

**Стек/окружение (по выводу `go test`):**
- OS/Arch: `darwin/arm64`
- CPU: Apple M3 Pro


---

## 2) Описание “проблемного запроса” (до оптимизации)

Проблемный участок — вычисление числа Фибоначчи **рекурсией**:

### CPU-профиль (до)
По `pprof top` видно, что почти всё CPU-время уходит в `work.Fib`:

![pprof top (до)](docs/screenshots/top.png)

Граф вызовов показывает, что “горячая точка” — именно `work.Fib`:

![pprof graph (до)](docs/screenshots/graph.png)

После все хорошо
![top_fast.png](docs/screenshots/top_fast.png)
![graph_fast.png](docs/screenshots/graph_fast.png)
---

## 3) Применённая оптимизация

Оптимизация выполнена переписыванием алгоритма:

- `internal/work/fast.go` — `work.FibFast(n)` (итеративный вариант, сложность **O(n)**).
- В обработчике `/work` заменён вызов `Fib` на `FibFast`.

---

## 4) Профили памяти (heap)

### alloc_space (до/после)
Снятие heap-профиля показало, что существенных аллокаций со стороны “вычисления” почти нет — основная память уходит на инфраструктуру HTTP/pprof.

**До:**
![heap alloc_space (до)](docs/screenshots/heap_alloc_space.png)

**После:**
![heap alloc_space (после)](docs/screenshots/heap_alloc_fast.png)

### inuse_space (до/после)

**До:**
![heap inuse_space (до)](docs/screenshots/heap_inuse_space.png)

**После:**
![heap inuse_space (после)](docs/screenshots/heap_inuse_fast.png)

---

## 6) Нагрузочное тестирование (до/после)

Использовалась утилита **hey**:
```bash
hey -n 5000 -c 20 http://localhost:8080/work
```

### До оптимизации
- RPS: **65.2575**
- p95: **0.3739s**
- p99: **0.4203s**
- Ошибки: **0%** (все ответы `200`)

![hey (до)](docs/screenshots/load.png)

### После оптимизации
- RPS: **64405.1859**
- p95: **0.0006s**
- p99: **0.0010s**
- Ошибки: **0%** (все ответы `200`)

![hey (после)](docs/screenshots/load_fast.png)

---

## 7) Бенчмарки (до/после)

Команда:
```bash
go test ./internal/work -bench=. -benchmem
```

### До оптимизации
`BenchmarkFib-11`: **2_378_417 ns/op**, `0 B/op`, `0 allocs/op`

![benchmark (до)](docs/screenshots/benchmark.png)

### После оптимизации
`BenchmarkFibFast-11`: **9.217 ns/op**, `0 B/op`, `0 allocs/op`

![benchmark (после)](docs/screenshots/benchmark_fast.png)

---

## 8) Сводная таблица “до/после” (ключевой эндпоинт `/work`)

| Метрика | До | После |
|---|---:|---:|
| Нагрузочный тест (RPS) | 65.2575 | 64405.1859 |
| p95 (сек) | 0.3739 | 0.0006 |
| p99 (сек) | 0.4203 | 0.0010 |
| Error rate | 0% | 0% |
| Benchmark (ns/op) | 2_378_417 | 9.217 |

---

## 9) Таймеры выполнения (логирование)

Логи “до” показывают время порядка **~150 мс** на `Fib(38)`:

![timer (до)](docs/screenshots/timer.png)

После замены на `FibFast(38)` — время порядка **микросекунд/наносекунд**:

![timer (после)](docs/screenshots/timer_after.png)

---

## 10) Как запустить

### Запуск сервера
(у тебя запуск был так)

```bash
go run ./cmd/api
```

Если `main.go` лежит в корне проекта:
```bash
go run .
```

После старта:
- `http://localhost:8080/work` — выполнить работу
- `http://localhost:8080/debug/pprof/` — pprof

![pprof index](/docs/screenshots/index.png)

---

## 11) Как снять профили pprof

### CPU
```bash
go tool pprof -http=:0 http://localhost:8080/debug/pprof/profile?seconds=30
```

### Heap
```bash
go tool pprof -http=:0 http://localhost:8080/debug/pprof/heap
```

---

## 12) Структура проекта (пример)

```
.
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   └── work/
│       ├── slow.go
│       ├── fast.go
│       ├── timer.go
│       └── slow_test.go
└── docs/screenshots/
    ├── benchmark.png
    ├── benchmark_fast.png
    ├── graph.png
    ├── graph_fast.png
    ├── heap_alloc_fast.png
    ├── heap_alloc_space.png
    ├── heap_inuse_fast.png
    ├── heap_inuse_space.png
    ├── index.png
    ├── load.png
    ├── load_fast.png
    ├── timer.png
    ├── timer_after.png
    ├── top.png
    └── top_fast.png
```

---

## 13) Выводы

- Наивная рекурсия `Fib(n)` создаёт экспоненциальное число вызовов, из-за чего в CPU-профиле она занимает ~96% времени.
- Замена на итеративный вариант `FibFast(n)` устранила bottleneck и дала резкий рост производительности:
  - RPS вырос примерно **в ~987 раз** (65 → 64k),
  - время на запрос упало до **миллисекунд/микросекунд**,
  - аллокации не выросли (`0 B/op`, `0 allocs/op`).

**Дальнейшие шаги:** добавить параметр `n` в запрос, кеширование/мемоизацию при повторяющихся значениях, и (если по заданию нужна БД) — подключить Postgres и сделать оптимизацию уже на SQL-эндпоинтах с настройкой connection pool.
