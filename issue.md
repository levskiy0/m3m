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

## UI

должен компилироваться в бинарь. Доступен по `http://{port:host}/`

В фронт нужно прокидывать адрес фронта из `server.uri`

`./web/static.go`

`./web/ui/...`

FrontEnd

используй schadcn и стиль [ui-style.md](ui-style.md) 

- Проанализируй требования логики в [issue.md](issue.md)
- сделай план `ui-plan.md` а далее следуя плану
- реализуй фронт

---

регистрируй `runtime api` с `$` префиксом, например - `$env.get(...)`

обнови код, схему и тесты

---

В `Monacoeditor` на фронте не попадается схема??? Почему он не знает:

``` 
Cannot find name 'schedule'.
```

---

```
logs-page.tsx:42 Uncaught TypeError: logs.filter is not a function
at LogsPage (logs-page.tsx:42:29) 
```

``` 
breadcrumb.tsx:71 In HTML, <li> cannot be a descendant of <li>.
This will cause a hydration error.

  ...
    <TooltipProvider delayDuration={0}>
      <TooltipProvider data-slot="tooltip-pr..." delayDuration={0}>
        <TooltipProviderProvider scope={undefined} isOpenDelayedRef={{current:true}} delayDuration={0} onOpen={function} ...>
          <div data-slot="sidebar-wr..." style={{...}} className="group/side...">
            <AppSidebar>
            <SidebarInset>
              <main data-slot="sidebar-inset" className="bg-backgro...">
                <header className="flex h-16 ...">
                  <div className="flex items...">
                    <SidebarTrigger>
                    <Separator>
                    <Breadcrumb>
                      <nav aria-label="breadcrumb" data-slot="breadcrumb">
                        <BreadcrumbList>
                          <ol data-slot="breadcrumb..." className="text-muted...">
                            <BreadcrumbItem>
                            <BreadcrumbItem>
>                             <li data-slot="breadcrumb-item" className="inline-flex items-center gap-1.5">
                                <BreadcrumbSeparator>
>                                 <li
>                                   data-slot="breadcrumb-separator"
>                                   role="presentation"
>                                   aria-hidden="true"
>                                   className={"[&>svg]:size-3.5"}
>                                 >
                                ...
                            ...
                ...
            ...
```

```
??? какой блять тег при создании папки ??? 
{
    "error": "Key: 'Path' Error:Field validation for 'Path' failed on the 'required' tag"
}
```

При создании релиза

```
@radix-ui_react-select.js?v=ecbf922e:1062 Uncaught Error: A <Select.Item /> must have a value prop that is not an empty string. This is because the Select value can be set to an empty string to clear the selection and show the placeholder.
    at SelectItem (@radix-ui_react-select.js?v=ecbf922e:1062:13)
    at Object.react_stack_bottom_frame (react-dom_client.js?v=ecbf922e:18509:20)
    at renderWithHooks (react-dom_client.js?v=ecbf922e:5654:24)
    at updateForwardRef (react-dom_client.js?v=ecbf922e:7198:21)
    at beginWork (react-dom_client.js?v=ecbf922e:8735:20)
    at runWithFiberInDEV (react-dom_client.js?v=ecbf922e:997:72)
    at performUnitOfWork (react-dom_client.js?v=ecbf922e:12561:98)
    at workLoopSync (react-dom_client.js?v=ecbf922e:12424:43)
    at renderRootSync (react-dom_client.js?v=ecbf922e:12408:13)
    at performWorkOnRoot (react-dom_client.js?v=ecbf922e:11827:37)
```

