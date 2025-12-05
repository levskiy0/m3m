# m3m

Система которая позволяет создавать мини сервисы/воркеры на JS и запускать их собирая логи, статистику, данные, хранить модели данных в БД.

Система должна быть расширяема `*.so` плагинами которые расширяют RuntimeJS API.

## Стек

- Golang
  - `gingonic`
  - `uber.FX`
  - `slog`
  - `spf13/cast`
  - `github.com/dop251/goja`
  - `jwt`
  - `github.com/spf13/viper`
  - `github.com/golang-jwt/jwt/v5`
- UI
  - `Monaco Editor`
  - `Vite`
  - `React`
  - `Shadcn`
    - https://ui.shadcn.com/docs/registry/mcp (Нужно добавить MCP)
    - https://ui.shadcn.com/docs/components
  - `TipTap`
- MongoDB

Примерный конфиг

``` 
server:
  host: "0.0.0.0"
  port: 3000
  uri: "http://127.0.0.1:3000"

mongodb:
  uri: "mongodb://localhost:27017"
  database: "telegram_bot_farm"

jwt:
  secret: "your-jwt-secret-key-change-in-production"
  expiration: 168h

storage:
  path: "./storage"
```

## Подробное описание

### Администраторы/Пользователи

У системы есть администраторы, первого root пользователя создаем через `cli` -> `>_ m3m new-admin {email} {password}`.
Root пользователь имеет доступ ко всем проектам, и может добавлять администраторов в систему. При добавлении можно назначить права:

- Создание проектов
- Управление пользователями (добавление/удаление/блокировка/редактирование/доступ к проектоам)
- Доступ к проектам
  - Список созданных проектов

Самый первый root пользователь, неудаляемый, им может управлять только он сам.

Пользователи могут менять себе пароль, имя, аватарку.

### Цели

В система можно настроить глобальные цели, которые будут доступны в `runtime` проектов

- Название
- Slug (`{goal-slug}`)
- Цвет (выбор или Нет или из 16 установленных в системе)
- Выбор проектов которым доступно запись в цель (`one to many`)
- Тип
  - Счетчик (просто увеличиваем значение)
  - Счетчик в разрезе дня (значение будет сегментироваться в разрезе дня `2025-01-01 | value = 123`)
- Описание
- `project_ref` - если есть связь с проектом то это цель проекта, если `null` то глобальная 

В Бд должна быть таблица в которой накапливается статистика, и удобный сервис для сбора статистики за период, по выбранным целям и тд;

### Проекты

Когда пользователь создает проект, ему нужно присвоить

- Имя
- Slug (уник)
- Цвет (выбор или Нет или из 16 установленных в системе)

#### Настройки

- Название проекта
- Цвет
- slug (уник)
- Участники проекта (пользователи с доступом к нему, можно дать доступ или отозвать)
- Приватный API ключ (можно обновить)
- Удалить проект (красная зона)


#### Цели

У каждого проекта есть справочник Целей (Добавление/Редактирование/Список/Удаление).
Эти цели будут использоваться для аналитики в проекте.

Ну уровне БД это одна таблица с глобальными целями.

- Название
- Slug (`{project-slug}-{goal-slug}`)
- Цвет (выбор или Нет или из 16 установленных в системе)
- Тип
    - Счетчик
    - Счетчик в разрезе дня
- Описание 

#### Окружение

У проекта можно настроить окружение, это список параметров которые будут доступны в `runtime` (`env.get('...')`).

- Key
- Type
  - String
  - Text
  - Json
  - Integer
  - Float
  - Boolean
- Value


#### Storage

У каждого проекта есть свой storage который будет доступен в `runtime` и функционал `API` файлового менеджера для доступа к нему

`./{base-storage-dir}/{bot-id}/storage/...`

- Каталоги
  - Создание каталога
  - Удаление
  - Переименование
- Файлы
  - Загрузка
  - Скачивание
  - Создание
    - Редактирование json, txt, yaml и прочих текстовых форматов 
  - Получение прямой ссылки (сервер генерирует)
  - Переименование 
  - Удаление
- Просмотр и навигация
  - Для картинок можно генерировать превью `50x50`
  - В остальном фронт будет сам подставлять иконку в соответствии с `mime` типом


#### Service Code (Pipeline)

У каждого проекта есть код сервиса на `JS` + `goja`, который будет запускаться в `runtime`.

`pipeline` должен иметь версии, и черновик. 
Опубликована может быть только одна версия. 
Опубликованную версию нельзя редактировать, что бы ее отредактировать нужно сделать новый черновик от нее. 
Можно удалить устаревшие версии.
Для каждой версии можно назначить ветку.


**Как это выглядит на практике:**

Изначально у нас пустой редактор (с пустым шаблоном сервиса). Мы пишем код и жмем сохранить, теперь есть ветка `develop`. 


Мы можем очистить/удалить версии кроме текущей опубликованной.
Мы всегда можем откатить версию.

Если мы хотим сбросить черновик, то должен быть reset с выбором версии на которую мы хотим сбросить черновик.

- `releases`
  - Список опубликованных версий
  - При публикации выбираем какую версию повысить minor/major (X.X)
  - У текущей запущеной версии должен гореть индикатор что именно она сейчас работает
  - Есть коммент к релизу
  - Когда делаем релиз то выбираем теги - `stable`, `not-fix`, `night-build`, `develop`
  - Код релиза нельзя редактировать
  - Релизы можно удалять (кроме текущего активного / если он запущен то нужно его остановить)
- `branches`
  - Можно делать ветки разработки или от релиза или от другой ветки
  - по умолчанию первая ветка `develop`


#### Запуск и отладка (Pipeline)

Проект может запускать на исполнение одну из своих версий `pipeline`.

При запуске нужно выбрать один из релизов (он будет активный).

Должен быть лог работы сервиса, лог новый для каждого нового запуска. вид лога обычный текстовый

``` 
[{date time}] [severity] ...
```

Лог можно скачать, или посмотреть в админке (как то мб с пагинации при скролле ).

Если запускается новая версия - предыдущие логи чистятся.

---

Тут наверное ты не понял в реализации, должно быть что-то типа pocketbase. На фронте полноценный редактор а на беке валидация, обработка запросов, фильтров, пагинации и все такое.

Изучи текущую реализацию и доработай если нужно

#### Хранилище данных

Для каждого проекта можно настроить своеобразную БД

Настраиваем модели:
- Поля
  - Ключ
  - Тип
    - String
    - Text
    - Number
    - Float
    - Bool
    - Document
    - File
    - Ref
    - Date
    - DateTime
    - и тд ...
  - Required
  - Default Value
    - Nullable
    - Значение должно от типа завить
- Настройка вида таблицы (выбор полей, выбор по каким мы будем фильтровать, сортировать и тд)
- Настройка формы (отображаются поля, можно скрыть те которые по умолчанию и менять их местами), настройка представления, например `Text` может иметь вид или `Textarea` или `TipTap`. Настройка валидации.

Далее мы можем это все просматривать и делать круд.

Так же сделать тестирование создания схем и работы с монго, пускай тестовая БД назвается `m3m_testing_database` и чистится каждый раз при начале тестов. ("mongodb://localhost:27017")

В `runtime` есть доступ к хранилищу данных через `database.mymodel.insert(...)`.


#### Runtime

Собственно говоря это как раз таки `GOJA` запуск.

`Runtiume API` - Это подключаемые модули:
- `storage`
  - Доступ к файлам
- `image`
  - Работа с картинками, ресайз, кроп, и тд
- `draw`
  - Работа с холстом - рисование
- `database`
  - Работа с Хранилищем данных
  - Формируем из существующих подсказки для `Monaco`
- `env`
  - Доступ к переменным окружения
- `http`
  - Полноценная работа с запросами
- `smtp`
  - Отправка писем
- `logger`
- `сrypto`, `encoding`, `strings`
- `delayed`
    - Выполнение задачи в горутине (размер пула настраивается в `config.yaml`)
- `router`
    - Можно сделать свой роутер аля `/bot-{id}/{route-name}`
  ```
    router.get('/health/:test', function(ctx) => {
      return router.response(200, { ok: true })
    })
  ```
- `schedule`
    - Настройка задач на очередь
    - `schedule.daily(() => { ... })`
- `и тд`

Сервер должен формировать в разрезе существующих `Runtiume API` подсказки для `Monaco Editor`. 

Должна быть система `so` плагинов `./plugins` расширяющих `runtime api`

Должен быть автостарт рантаймов, например сервер повис и был перезагружен, то при старте приложения все запущенные сервисы в проекта х должны быть запущены.

Жизненный цикл JS нужно продумать но что то типа такого:

``` 

service.boot(() => {
  ...
})

service.start(() => {
  ...
})

service.shotdown(() => {
  ...
})


```

------------------------

Active Release
No active release

Хотя он есть и запущен!!!

Pipeline в списке релизов вообще не видно какой из них активный.

Для удобства отладки и разработки нужно сделать возможность запуска Debug версии (та что сейчас в редакторе).
То есть прям из пайплайна запускаем и если это не релиз то в пайплайне появляется лог, в Overview посвечиваем что запущен не релиз.

Кнопка Start в Overview должна иметь выпадашку с выбором версии для запуска;

Список релизов - индикатор активный или нет не нужен, активным для проекта становится последний стабильный запущенный.

Удалить релиз если он запущен нельзя.

---

Нужно расширить мониторинг, показывать нагрузку, в том числе CPU, Хиты по роутам и т.д. 

в Overview Monitoring заюзать `sparkline-svg`

Инстанс должен хранить в памяти небольшой срез для графика за 24 часа по 5 минут и возвращать на фронт для формирования данных в мониторинге... 

в Goals заюзать `sparkline-svg` для показа за 7 днней графика целям которое копятся на день.

---

Добавь добавь больше отладочной метрики в проект. и сквозные карточки - CPU и MEM

---


Убери эту карточку 

``` 
<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4"><div data-slot="card" class="bg-card text-card-foreground flex flex-col gap-4 justify-between rounded-xl border py-6 shadow-sm"><div data-slot="card-header" class="@container/card-header auto-rows-min grid-rows-[auto_auto] gap-2 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6 flex flex-row items-center justify-between space-y-0 pb-2"><div data-slot="card-description" class="text-muted-foreground text-sm font-medium">Uptime</div><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-clock size-4 text-muted-foreground" aria-hidden="true"><path d="M12 6v6l4 2"></path><circle cx="12" cy="12" r="10"></circle></svg></div><div data-slot="card-content" class="px-6"><div class="text-2xl font-bold">1m 28s</div><p class="text-xs text-muted-foreground mt-1">Since 12/5/2025, 3:24:50 PM</p></div></div><div data-slot="card" class="bg-card text-card-foreground flex flex-col gap-4 justify-between rounded-xl border py-6 shadow-sm"><div data-slot="card-header" class="@container/card-header auto-rows-min grid-rows-[auto_auto] gap-2 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6 flex flex-row items-center justify-between space-y-0 pb-2"><div data-slot="card-description" class="text-muted-foreground text-sm font-medium">Requests</div><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-zap size-4 text-muted-foreground" aria-hidden="true"><path d="M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z"></path></svg></div><div data-slot="card-content" class="px-6"><div class="flex items-end justify-between gap-4"><div><div class="text-2xl font-bold">0</div><p class="text-xs text-muted-foreground mt-1">1 route</p></div></div></div></div><div data-slot="card" class="bg-card text-card-foreground flex flex-col gap-4 justify-between rounded-xl border py-6 shadow-sm"><div data-slot="card-header" class="@container/card-header auto-rows-min grid-rows-[auto_auto] gap-2 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6 flex flex-row items-center justify-between space-y-0 pb-2"><div data-slot="card-description" class="text-muted-foreground text-sm font-medium">Scheduled Jobs</div><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-clock size-4 text-muted-foreground" aria-hidden="true"><path d="M12 6v6l4 2"></path><circle cx="12" cy="12" r="10"></circle></svg></div><div data-slot="card-content" class="px-6"><div class="text-2xl font-bold">1</div><p class="text-xs text-muted-foreground mt-1">Scheduler inactive</p></div></div><div data-slot="card" class="bg-card text-card-foreground flex flex-col gap-4 justify-between rounded-xl border py-6 shadow-sm"><div data-slot="card-header" class="@container/card-header auto-rows-min grid-rows-[auto_auto] gap-2 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6 flex flex-row items-center justify-between space-y-0 pb-2"><div data-slot="card-description" class="text-muted-foreground text-sm font-medium">Memory</div><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-memory-stick size-4 text-muted-foreground" aria-hidden="true"><path d="M6 19v-3"></path><path d="M10 19v-3"></path><path d="M14 19v-3"></path><path d="M18 19v-3"></path><path d="M8 11V9"></path><path d="M16 11V9"></path><path d="M12 11V9"></path><path d="M2 15h20"></path><path d="M2 7a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v1.1a2 2 0 0 0 0 3.837V17a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2v-5.1a2 2 0 0 0 0-3.837Z"></path></svg></div><div data-slot="card-content" class="px-6"><div class="flex items-end justify-between gap-4"><div><div class="text-2xl font-bold">3.4 MB</div><p class="text-xs text-muted-foreground mt-1">Current usage</p></div></div></div></div></div>
```

И добавь инфу о запущенной ветке тут после Running новый бейдж с сеткой и индикатором стабильности

``` 
<div class="grid gap-4 md:grid-cols-2 lg:grid-cols-4"><div data-slot="card" class="bg-card text-card-foreground flex flex-col gap-4 justify-between rounded-xl border py-6 shadow-sm"><div data-slot="card-header" class="@container/card-header auto-rows-min grid-rows-[auto_auto] gap-2 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6 flex flex-row items-center justify-between space-y-0 pb-2"><div data-slot="card-description" class="text-muted-foreground text-sm font-medium">Uptime</div><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-clock size-4 text-muted-foreground" aria-hidden="true"><path d="M12 6v6l4 2"></path><circle cx="12" cy="12" r="10"></circle></svg></div><div data-slot="card-content" class="px-6"><div class="text-2xl font-bold">1m 28s</div><p class="text-xs text-muted-foreground mt-1">Since 12/5/2025, 3:24:50 PM</p></div></div><div data-slot="card" class="bg-card text-card-foreground flex flex-col gap-4 justify-between rounded-xl border py-6 shadow-sm"><div data-slot="card-header" class="@container/card-header auto-rows-min grid-rows-[auto_auto] gap-2 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6 flex flex-row items-center justify-between space-y-0 pb-2"><div data-slot="card-description" class="text-muted-foreground text-sm font-medium">Requests</div><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-zap size-4 text-muted-foreground" aria-hidden="true"><path d="M4 14a1 1 0 0 1-.78-1.63l9.9-10.2a.5.5 0 0 1 .86.46l-1.92 6.02A1 1 0 0 0 13 10h7a1 1 0 0 1 .78 1.63l-9.9 10.2a.5.5 0 0 1-.86-.46l1.92-6.02A1 1 0 0 0 11 14z"></path></svg></div><div data-slot="card-content" class="px-6"><div class="flex items-end justify-between gap-4"><div><div class="text-2xl font-bold">0</div><p class="text-xs text-muted-foreground mt-1">1 route</p></div></div></div></div><div data-slot="card" class="bg-card text-card-foreground flex flex-col gap-4 justify-between rounded-xl border py-6 shadow-sm"><div data-slot="card-header" class="@container/card-header auto-rows-min grid-rows-[auto_auto] gap-2 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6 flex flex-row items-center justify-between space-y-0 pb-2"><div data-slot="card-description" class="text-muted-foreground text-sm font-medium">Scheduled Jobs</div><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-clock size-4 text-muted-foreground" aria-hidden="true"><path d="M12 6v6l4 2"></path><circle cx="12" cy="12" r="10"></circle></svg></div><div data-slot="card-content" class="px-6"><div class="text-2xl font-bold">1</div><p class="text-xs text-muted-foreground mt-1">Scheduler inactive</p></div></div><div data-slot="card" class="bg-card text-card-foreground flex flex-col gap-4 justify-between rounded-xl border py-6 shadow-sm"><div data-slot="card-header" class="@container/card-header auto-rows-min grid-rows-[auto_auto] gap-2 px-6 has-data-[slot=card-action]:grid-cols-[1fr_auto] [.border-b]:pb-6 flex flex-row items-center justify-between space-y-0 pb-2"><div data-slot="card-description" class="text-muted-foreground text-sm font-medium">Memory</div><svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-memory-stick size-4 text-muted-foreground" aria-hidden="true"><path d="M6 19v-3"></path><path d="M10 19v-3"></path><path d="M14 19v-3"></path><path d="M18 19v-3"></path><path d="M8 11V9"></path><path d="M16 11V9"></path><path d="M12 11V9"></path><path d="M2 15h20"></path><path d="M2 7a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v1.1a2 2 0 0 0 0 3.837V17a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2v-5.1a2 2 0 0 0 0-3.837Z"></path></svg></div><div data-slot="card-content" class="px-6"><div class="flex items-end justify-between gap-4"><div><div class="text-2xl font-bold">3.4 MB</div><p class="text-xs text-muted-foreground mt-1">Current usage</p></div></div></div></div></div>
```























