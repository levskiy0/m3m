# M3M - План разработки

## Обзор проекта

**M3M** - платформа для создания и управления мини-сервисами/воркерами на JavaScript с возможностью сбора логов, статистики, хранения данных в MongoDB и расширения через плагины.

---

## Архитектура

```
m3m/
├── cmd/
│   └── m3m/                    # CLI точка входа
│       └── main.go
├── internal/
│   ├── app/                    # Uber FX приложение
│   │   └── app.go
│   ├── config/                 # Конфигурация (Viper)
│   │   └── config.go
│   ├── domain/                 # Доменные модели
│   │   ├── user.go
│   │   ├── project.go
│   │   ├── goal.go
│   │   ├── pipeline.go
│   │   ├── model.go
│   │   └── environment.go
│   ├── repository/             # MongoDB репозитории
│   │   ├── user_repository.go
│   │   ├── project_repository.go
│   │   ├── goal_repository.go
│   │   ├── pipeline_repository.go
│   │   └── model_repository.go
│   ├── service/                # Бизнес-логика
│   │   ├── auth_service.go
│   │   ├── user_service.go
│   │   ├── project_service.go
│   │   ├── goal_service.go
│   │   ├── pipeline_service.go
│   │   ├── storage_service.go
│   │   └── runtime_service.go
│   ├── handler/                # HTTP handlers (Gin)
│   │   ├── auth_handler.go
│   │   ├── user_handler.go
│   │   ├── project_handler.go
│   │   ├── goal_handler.go
│   │   ├── pipeline_handler.go
│   │   ├── storage_handler.go
│   │   └── model_handler.go
│   ├── middleware/             # Middleware
│   │   ├── auth.go
│   │   └── cors.go
│   ├── runtime/                # GOJA Runtime
│   │   ├── runtime.go
│   │   ├── modules/
│   │   │   ├── storage.go
│   │   │   ├── image.go
│   │   │   ├── draw.go
│   │   │   ├── database.go
│   │   │   ├── env.go
│   │   │   ├── http.go
│   │   │   ├── smtp.go
│   │   │   ├── logger.go
│   │   │   ├── crypto.go
│   │   │   ├── encoding.go
│   │   │   ├── utils.go
│   │   │   ├── delayed.go
│   │   │   ├── router.go
│   │   │   └── schedule.go
│   │   └── pool.go
│   └── plugin/                 # Система плагинов
│       └── loader.go
├── web/                        # Frontend embedded (React + Vite)
│   ├── static.go               # go:embed для статики
│   ├── dist/                   # Скомпилированный frontend (генерируется)
│   ├── src/
│   │   ├── components/
│   │   │   ├── ui/             # shadcn компоненты
│   │   │   ├── app-sidebar.tsx
│   │   │   ├── nav-main.tsx
│   │   │   ├── nav-projects.tsx
│   │   │   ├── nav-user.tsx
│   │   │   └── project-switcher.tsx
│   │   ├── pages/
│   │   ├── hooks/
│   │   ├── api/
│   │   ├── lib/
│   │   │   ├── utils.ts
│   │   │   └── config.ts       # window.__APP_CONFIG__
│   │   └── App.tsx
│   ├── package.json
│   ├── vite.config.ts
│   └── components.json         # shadcn config
├── pkg/                        # Публичные пакеты
│   └── api/                    # API клиент (если нужен)
├── plugins/                    # SO плагины (примеры)
│   └── example/
├── config.yaml                 # Конфигурация
├── go.mod
├── go.sum
└── Makefile
```

---

## Этапы разработки

### Этап 1: Базовая инфраструктура

#### 1.1 Настройка проекта Go
- [x] Создать структуру директорий
- [ ] Настроить go.mod с зависимостями
- [ ] Создать базовый Makefile
- [ ] Настроить конфигурацию (Viper)

#### 1.2 База данных MongoDB
- [ ] Подключение к MongoDB
- [ ] Базовые модели (User, Project, Goal)
- [ ] Репозитории для CRUD операций

#### 1.3 HTTP сервер (Gin + Uber FX)
- [ ] Инициализация Uber FX приложения
- [ ] Настройка Gin роутера
- [ ] CORS middleware
- [ ] Health check endpoint

---

### Этап 2: Аутентификация и пользователи

#### 2.1 Модель пользователя
- [ ] Структура User (email, password, name, avatar, role, permissions)
- [ ] Хеширование паролей (bcrypt)
- [ ] Права доступа (create_projects, manage_users, project_access)

#### 2.2 CLI команда создания root пользователя
- [ ] `m3m new-admin {email} {password}`
- [ ] Проверка первого пользователя (root)

#### 2.3 JWT аутентификация
- [ ] Генерация токенов
- [ ] Middleware проверки токенов
- [ ] Refresh токены

#### 2.4 API пользователей
- [ ] POST /api/auth/login
- [ ] POST /api/auth/logout
- [ ] GET /api/users/me
- [ ] PUT /api/users/me (обновление профиля)
- [ ] PUT /api/users/me/password
- [ ] PUT /api/users/me/avatar
- [ ] CRUD пользователей (для админов)

---

### Этап 3: Проекты

#### 3.1 Модель проекта
- [ ] Структура Project (name, slug, color, owner_id, members, api_key)
- [ ] Уникальность slug
- [ ] Генерация API ключа

#### 3.2 API проектов
- [ ] POST /api/projects
- [ ] GET /api/projects
- [ ] GET /api/projects/:id
- [ ] PUT /api/projects/:id
- [ ] DELETE /api/projects/:id
- [ ] POST /api/projects/:id/regenerate-key
- [ ] Управление участниками проекта

---

### Этап 4: Цели (Goals)

#### 4.1 Модель целей
- [ ] Структура Goal (name, slug, color, type, description, project_ref)
- [ ] Типы: counter, daily_counter
- [ ] Глобальные цели (project_ref = null)
- [ ] Проектные цели

#### 4.2 Статистика целей
- [ ] Таблица GoalStats (goal_id, date, value)
- [ ] Агрегация статистики за период
- [ ] API для записи и чтения статистики

#### 4.3 API целей
- [ ] CRUD глобальных целей
- [ ] CRUD проектных целей
- [ ] GET /api/goals/stats (с фильтрами)

---

### Этап 5: Окружение проекта

#### 5.1 Модель окружения
- [ ] Структура EnvVar (key, type, value, project_id)
- [ ] Типы: string, text, json, integer, float, boolean

#### 5.2 API окружения
- [ ] GET /api/projects/:id/env
- [ ] POST /api/projects/:id/env
- [ ] PUT /api/projects/:id/env/:key
- [ ] DELETE /api/projects/:id/env/:key

---

### Этап 6: Файловое хранилище (Storage)

#### 6.1 Структура хранилища
- [ ] Базовый путь: `{storage_path}/{project_id}/storage/`
- [ ] Создание директорий при создании проекта

#### 6.2 Файловый менеджер
- [ ] Создание/удаление/переименование каталогов
- [ ] Загрузка/скачивание/удаление файлов
- [ ] Создание/редактирование текстовых файлов
- [ ] Генерация прямых ссылок на файлы
- [ ] Генерация превью для изображений (50x50)

#### 6.3 API хранилища
- [ ] GET /api/projects/:id/storage (листинг)
- [ ] POST /api/projects/:id/storage/mkdir
- [ ] POST /api/projects/:id/storage/upload
- [ ] GET /api/projects/:id/storage/download/:path
- [ ] PUT /api/projects/:id/storage/rename
- [ ] DELETE /api/projects/:id/storage/:path
- [ ] GET /api/storage/public/:token/:filename (прямые ссылки)

---

### Этап 7: Pipeline (Код сервиса)

#### 7.1 Модели версионирования
- [ ] Структура Branch (name, code, project_id, parent_branch, created_from_release)
- [ ] Структура Release (version, code, comment, tag, is_active, project_id)
- [ ] Версионирование: major.minor
- [ ] Теги: stable, hot-fix, night-build, develop

#### 7.2 Управление ветками
- [ ] Создание ветки от релиза или другой ветки
- [ ] Ветка по умолчанию: develop
- [ ] Сохранение/редактирование кода ветки
- [ ] Reset черновика на выбранную версию

#### 7.3 Управление релизами
- [ ] Публикация релиза (minor/major bump)
- [ ] Список релизов
- [ ] Удаление релизов (кроме активного)
- [ ] Активация релиза

#### 7.4 API Pipeline
- [ ] GET /api/projects/:id/pipeline/branches
- [ ] POST /api/projects/:id/pipeline/branches
- [ ] PUT /api/projects/:id/pipeline/branches/:name
- [ ] DELETE /api/projects/:id/pipeline/branches/:name
- [ ] GET /api/projects/:id/pipeline/releases
- [ ] POST /api/projects/:id/pipeline/releases
- [ ] DELETE /api/projects/:id/pipeline/releases/:version
- [ ] POST /api/projects/:id/pipeline/releases/:version/activate

---

### Этап 8: Хранилище данных (Models)

#### 8.1 Динамические модели
- [ ] Структура Model (name, slug, fields, project_id)
- [ ] Типы полей: string, text, number, float, bool, document, file, ref, date, datetime
- [ ] Настройки полей: required, default_value

#### 8.2 Настройки отображения
- [ ] Настройка колонок таблицы
- [ ] Настройка фильтров и сортировки
- [ ] Настройка формы (порядок полей, представление)

#### 8.3 CRUD данных
- [ ] Динамическое создание коллекций
- [ ] Валидация данных по схеме
- [ ] API для работы с данными моделей

#### 8.4 API моделей
- [ ] CRUD моделей (схем)
- [ ] CRUD записей в моделях

---

### Этап 9: GOJA Runtime

#### 9.1 Базовый Runtime
- [x] Инициализация GOJA VM
- [ ] Пул воркеров (настраиваемый размер)
- [x] Изоляция контекстов проектов
- [x] Обработка ошибок и таймауты

#### 9.1.1 Жизненный цикл сервиса (Service Lifecycle)
```javascript
// Модуль service для управления жизненным циклом

// Вызывается при инициализации сервиса (до start)
// Используется для настройки, загрузки конфигов и т.д.
service.boot(function() {
    logger.info('Service booting...');
    // Инициализация переменных, подключений
});

// Вызывается после boot, когда сервис готов к работе
// Здесь регистрируются роуты, планировщики и т.д.
service.start(function() {
    logger.info('Service started');

    router.get('/health', function(ctx) {
        return router.response(200, { status: 'ok' });
    });

    schedule.hourly(function() {
        logger.info('Hourly task');
    });
});

// Вызывается при остановке сервиса
// Используется для корректного завершения (cleanup)
service.shutdown(function() {
    logger.info('Service shutting down...');
    // Закрытие соединений, сохранение состояния
});
```

- [ ] Модуль `service` с lifecycle хуками
- [ ] `service.boot(callback)` - инициализация
- [ ] `service.start(callback)` - запуск
- [ ] `service.shutdown(callback)` - остановка
- [ ] Гарантированный порядок выполнения: boot -> start -> (работа) -> shutdown
- [ ] Graceful shutdown с таймаутом

#### 9.2 Модуль storage
- [ ] Доступ к файлам проекта
- [ ] Чтение/запись файлов
- [ ] Работа с директориями

#### 9.3 Модуль image
- [ ] Ресайз изображений
- [ ] Кроп изображений
- [ ] Конвертация форматов

#### 9.4 Модуль draw
- [ ] Создание холста
- [ ] Рисование примитивов
- [ ] Текст на изображениях

#### 9.5 Модуль database
- [ ] CRUD операции с моделями проекта
- [ ] Поиск и фильтрация
- [ ] Агрегации

#### 9.6 Модуль env
- [ ] `env.get(key)` - получение переменных окружения

#### 9.7 Модуль http
- [ ] GET, POST, PUT, DELETE запросы
- [ ] Заголовки и cookies
- [ ] Таймауты

#### 9.8 Модуль smtp
- [ ] Отправка email
- [ ] Шаблоны писем

#### 9.9 Модуль logger
- [ ] Уровни: debug, info, warn, error
- [ ] Запись в файл лога проекта

#### 9.10 Модули crypto, encoding, utils
- [ ] crypto: md5, sha256, randomBytes
- [ ] encoding: base64, json, url encode/decode
- [ ] utils: sleep, random, uuid, slugify, formatDate, etc.

#### 9.11 Модуль delayed
- [ ] Выполнение в горутине
- [ ] Настраиваемый пул

#### 9.12 Модуль router
- [ ] Регистрация маршрутов
- [ ] Обработка HTTP запросов к проекту
- [ ] `/api/projects/:slug/:route`

#### 9.13 Модуль schedule
- [ ] Планировщик задач (cron)
- [ ] schedule.daily(), schedule.hourly(), schedule.cron()

#### 9.14 Модуль goals
- [ ] Запись в цели из runtime
- [ ] `goals.increment('goal-slug', value)`

---

### Этап 10: Запуск и логирование

#### 10.1 Управление запуском
- [ ] Запуск проекта с выбранным релизом
- [ ] Остановка проекта
- [ ] Перезапуск проекта
- [ ] Статус проекта (running/stopped)

#### 10.1.1 Автостарт рантаймов при запуске сервера
При перезапуске сервера (краш, обновление, ручной рестарт) все проекты со статусом `running` должны быть автоматически запущены.

```go
// При старте приложения:
// 1. Загрузить все проекты со статусом "running"
// 2. Для каждого проекта получить активный релиз
// 3. Запустить runtime с кодом релиза
// 4. Вызвать lifecycle: boot -> start
```

- [ ] Сохранение статуса проекта в БД (`running`/`stopped`)
- [ ] При старте приложения - загрузка проектов со статусом `running`
- [ ] Автоматический запуск всех ранее запущенных проектов
- [ ] Порядок запуска: сначала все `boot()`, затем все `start()`
- [ ] Логирование автостарта
- [ ] Обработка ошибок при автостарте (не блокировать остальные проекты)
- [ ] API для отключения автостарта конкретного проекта

```yaml
# Дополнительная настройка в Project
project:
  auto_start: true  # Автоматически запускать при старте сервера
```

#### 10.2 Логирование
- [ ] Формат: `[{datetime}] [severity] message`
- [ ] Файл лога для каждого запуска
- [ ] Очистка логов при новом запуске
- [ ] API для чтения логов (с пагинацией)
- [ ] Скачивание лога

#### 10.3 API запуска
- [ ] POST /api/projects/:id/start
- [ ] POST /api/projects/:id/stop
- [ ] GET /api/projects/:id/status
- [ ] GET /api/projects/:id/logs
- [ ] GET /api/projects/:id/logs/download

---

### Этап 11: Система плагинов (.so)

Система расширения Runtime API через динамически загружаемые `.so` плагины из директории `./plugins`.

#### 11.1 Интерфейс плагина
```go
// Plugin - интерфейс который должен реализовать каждый плагин
type Plugin interface {
    // Name возвращает имя модуля для регистрации в runtime
    Name() string

    // Version возвращает версию плагина
    Version() string

    // Init инициализирует плагин с конфигурацией
    Init(config map[string]interface{}) error

    // RegisterModule регистрирует функции в GOJA runtime
    RegisterModule(runtime *goja.Runtime) error

    // Shutdown корректно завершает работу плагина
    Shutdown() error

    // TypeDefinitions возвращает TypeScript декларации для Monaco
    TypeDefinitions() string
}
```

#### 11.2 Загрузчик плагинов
- [ ] Сканирование директории `./plugins` при старте
- [ ] Загрузка `.so` файлов через `plugin.Open()`
- [ ] Поиск и вызов экспортированной функции `NewPlugin() Plugin`
- [ ] Инициализация плагинов с конфигурацией
- [ ] Регистрация модулей в пуле runtime
- [ ] Graceful shutdown всех плагинов

#### 11.3 Структура плагина (пример)
```
plugins/
├── telegram/
│   ├── plugin.go
│   ├── go.mod
│   └── Makefile          # go build -buildmode=plugin -o telegram.so
├── telegram.so           # Скомпилированный плагин
├── openai/
│   ├── plugin.go
│   └── ...
└── openai.so
```

#### 11.4 Пример плагина (Telegram)
```go
package main

import "github.com/dop251/goja"

type TelegramPlugin struct {
    botToken string
}

func (p *TelegramPlugin) Name() string { return "telegram" }
func (p *TelegramPlugin) Version() string { return "1.0.0" }

func (p *TelegramPlugin) Init(config map[string]interface{}) error {
    if token, ok := config["bot_token"].(string); ok {
        p.botToken = token
    }
    return nil
}

func (p *TelegramPlugin) RegisterModule(runtime *goja.Runtime) error {
    return runtime.Set("telegram", map[string]interface{}{
        "sendMessage": p.sendMessage,
        "sendPhoto":   p.sendPhoto,
        // ...
    })
}

func (p *TelegramPlugin) TypeDefinitions() string {
    return `
declare const telegram: {
    sendMessage(chatId: number, text: string): Promise<Message>;
    sendPhoto(chatId: number, photo: string, caption?: string): Promise<Message>;
};
`
}

// NewPlugin - экспортируемая функция для загрузчика
func NewPlugin() Plugin {
    return &TelegramPlugin{}
}
```

#### 11.5 Конфигурация плагинов
```yaml
plugins:
  path: "./plugins"
  config:
    telegram:
      bot_token: "your-bot-token"
    openai:
      api_key: "your-openai-key"
      model: "gpt-4"
```

#### 11.6 API плагинов
- [ ] GET /api/plugins - список загруженных плагинов
- [ ] GET /api/plugins/:name - информация о плагине
- [ ] POST /api/plugins/:name/reload - перезагрузка плагина

#### 11.7 Интеграция с Monaco
- [ ] Сбор TypeDefinitions от всех плагинов
- [ ] Объединение с базовыми типами Runtime API
- [ ] Динамическое обновление IntelliSense

---

### Этап 12: Frontend (React + Vite + Embedded)

#### 12.0 Встраивание UI в бинарь

UI компилируется вместе с Go бинарником и доступен по `http://{host}:{port}/`

##### Структура `web/static.go`:
```go
package web

import (
    "bytes"
    "embed"
    "io/fs"
    "net/http"
    "strings"

    "m3m/internal/config"
)

//go:embed dist/*
var staticFS embed.FS

// GetFileSystem возвращает файловую систему для статики
func GetFileSystem() (http.FileSystem, error) {
    subFS, err := fs.Sub(staticFS, "dist")
    if err != nil {
        return nil, err
    }
    return http.FS(subFS), nil
}

// GetIndexHTML возвращает index.html с инжектированным конфигом
func GetIndexHTML(cfg *config.Config) ([]byte, error) {
    indexBytes, err := staticFS.ReadFile("dist/index.html")
    if err != nil {
        return nil, err
    }

    // Инжектируем конфигурацию в HTML
    configScript := `<script>
window.__APP_CONFIG__ = {
    apiURL: "` + cfg.Server.URI + `"
};
</script>`

    htmlContent := string(indexBytes)
    headEndIndex := strings.Index(htmlContent, "</head>")
    if headEndIndex != -1 {
        var buf bytes.Buffer
        buf.WriteString(htmlContent[:headEndIndex])
        buf.WriteString(configScript)
        buf.WriteString(htmlContent[headEndIndex:])
        return buf.Bytes(), nil
    }

    return indexBytes, nil
}
```

##### Регистрация в Gin (app.go):
```go
func RegisterStaticRoutes(r *gin.Engine, cfg *config.Config) {
    // Получаем index.html с инжектированным конфигом
    indexHTML, err := ui.GetIndexHTML(cfg)
    if err != nil {
        panic(err)
    }

    // Статические файлы
    staticFS, err := ui.GetFileSystem()
    if err != nil {
        panic(err)
    }

    // Обслуживание статических файлов
    r.StaticFS("/assets", staticFS)

    // SPA fallback - все не-API роуты отдают index.html
    r.NoRoute(func(c *gin.Context) {
        // Пропускаем API роуты
        if strings.HasPrefix(c.Request.URL.Path, "/api/") {
            c.JSON(404, gin.H{"error": "not found"})
            return
        }
        c.Data(200, "text/html; charset=utf-8", indexHTML)
    })
}
```

##### Использование в React:
```typescript
// src/lib/config.ts
interface AppConfig {
    apiURL: string;
}

declare global {
    interface Window {
        __APP_CONFIG__?: AppConfig;
    }
}

export const config: AppConfig = {
    apiURL: window.__APP_CONFIG__?.apiURL || 'http://localhost:3000',
};
```

##### Makefile команды:
```makefile
# Сборка frontend
ui-build:
	cd ui && npm run build

# Сборка всего (frontend + backend)
build: ui-build
	go build -o bin/m3m ./cmd/m3m

# Dev режим (frontend отдельно)
ui-dev:
	cd ui && npm run dev
```

#### 12.1 Настройка проекта
- [ ] Vite + React + TypeScript
- [ ] Shadcn UI компоненты (см. `ui-style.md`)
- [ ] TailwindCSS
- [ ] React Router
- [ ] Конфигурация через `window.__APP_CONFIG__`

#### 12.2 Аутентификация
- [ ] Страница логина
- [ ] Хранение токена
- [ ] Защищенные маршруты

#### 12.3 Управление пользователями
- [ ] Список пользователей
- [ ] Создание/редактирование пользователя
- [ ] Профиль текущего пользователя

#### 12.4 Управление проектами
- [ ] Список проектов
- [ ] Создание проекта
- [ ] Настройки проекта
- [ ] Dashboard проекта

#### 12.5 Цели
- [ ] Глобальные цели
- [ ] Цели проекта
- [ ] Графики статистики

#### 12.6 Окружение
- [ ] Список переменных
- [ ] Добавление/редактирование переменных

#### 12.7 Файловый менеджер
- [ ] Навигация по папкам
- [ ] Загрузка файлов
- [ ] Превью изображений
- [ ] Редактор текстовых файлов

#### 12.8 Редактор кода (Monaco)
- [ ] Интеграция Monaco Editor
- [ ] Подсветка JavaScript
- [ ] Автодополнение Runtime API
- [ ] Управление ветками
- [ ] Публикация релизов

#### 12.9 Хранилище данных
- [ ] Конструктор моделей
- [ ] Таблица данных с фильтрами
- [ ] Форма создания/редактирования записи
- [ ] TipTap для полей типа Text

#### 12.10 Запуск и логи
- [ ] Кнопки запуска/остановки
- [ ] Выбор релиза для запуска
- [ ] Просмотр логов в реальном времени

---

### Этап 13: Генерация подсказок для Monaco

#### 13.1 Описание Runtime API
- [ ] TypeScript декларации для всех модулей
- [ ] API endpoint для получения деклараций
- [ ] Интеграция с Monaco IntelliSense

---

## Модели данных (MongoDB)

### User
```go
type User struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Email       string             `bson:"email"`
    Password    string             `bson:"password"`
    Name        string             `bson:"name"`
    Avatar      string             `bson:"avatar"`
    IsRoot      bool               `bson:"is_root"`
    Permissions Permissions        `bson:"permissions"`
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}

type Permissions struct {
    CreateProjects bool                 `bson:"create_projects"`
    ManageUsers    bool                 `bson:"manage_users"`
    ProjectAccess  []primitive.ObjectID `bson:"project_access"`
}
```

### Project
```go
type Project struct {
    ID        primitive.ObjectID   `bson:"_id,omitempty"`
    Name      string               `bson:"name"`
    Slug      string               `bson:"slug"`
    Color     string               `bson:"color"`
    OwnerID   primitive.ObjectID   `bson:"owner_id"`
    Members   []primitive.ObjectID `bson:"members"`
    APIKey    string               `bson:"api_key"`
    Status    string               `bson:"status"` // running, stopped
    ActiveRelease string           `bson:"active_release"`
    CreatedAt time.Time            `bson:"created_at"`
    UpdatedAt time.Time            `bson:"updated_at"`
}
```

### Goal
```go
type Goal struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Name        string             `bson:"name"`
    Slug        string             `bson:"slug"`
    Color       string             `bson:"color"`
    Type        string             `bson:"type"` // counter, daily_counter
    Description string             `bson:"description"`
    ProjectRef  *primitive.ObjectID `bson:"project_ref"` // null = global
    AllowedProjects []primitive.ObjectID `bson:"allowed_projects"`
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}
```

### GoalStat
```go
type GoalStat struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    GoalID    primitive.ObjectID `bson:"goal_id"`
    ProjectID primitive.ObjectID `bson:"project_id"`
    Date      string             `bson:"date"` // YYYY-MM-DD for daily
    Value     int64              `bson:"value"`
    UpdatedAt time.Time          `bson:"updated_at"`
}
```

### EnvVar
```go
type EnvVar struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    ProjectID primitive.ObjectID `bson:"project_id"`
    Key       string             `bson:"key"`
    Type      string             `bson:"type"` // string, text, json, integer, float, boolean
    Value     interface{}        `bson:"value"`
    CreatedAt time.Time          `bson:"created_at"`
    UpdatedAt time.Time          `bson:"updated_at"`
}
```

### Branch
```go
type Branch struct {
    ID              primitive.ObjectID  `bson:"_id,omitempty"`
    ProjectID       primitive.ObjectID  `bson:"project_id"`
    Name            string              `bson:"name"`
    Code            string              `bson:"code"`
    ParentBranch    *string             `bson:"parent_branch"`
    CreatedFromRelease *string          `bson:"created_from_release"`
    CreatedAt       time.Time           `bson:"created_at"`
    UpdatedAt       time.Time           `bson:"updated_at"`
}
```

### Release
```go
type Release struct {
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    ProjectID primitive.ObjectID `bson:"project_id"`
    Version   string             `bson:"version"` // X.Y
    Code      string             `bson:"code"`
    Comment   string             `bson:"comment"`
    Tag       string             `bson:"tag"` // stable, hot-fix, night-build, develop
    IsActive  bool               `bson:"is_active"`
    CreatedAt time.Time          `bson:"created_at"`
}
```

### Model (схема)
```go
type Model struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    ProjectID   primitive.ObjectID `bson:"project_id"`
    Name        string             `bson:"name"`
    Slug        string             `bson:"slug"`
    Fields      []ModelField       `bson:"fields"`
    TableConfig TableConfig        `bson:"table_config"`
    FormConfig  FormConfig         `bson:"form_config"`
    CreatedAt   time.Time          `bson:"created_at"`
    UpdatedAt   time.Time          `bson:"updated_at"`
}

type ModelField struct {
    Key          string      `bson:"key"`
    Type         string      `bson:"type"`
    Required     bool        `bson:"required"`
    DefaultValue interface{} `bson:"default_value"`
}

type TableConfig struct {
    Columns     []string `bson:"columns"`
    Filters     []string `bson:"filters"`
    SortColumns []string `bson:"sort_columns"`
}

type FormConfig struct {
    FieldOrder  []string            `bson:"field_order"`
    HiddenFields []string           `bson:"hidden_fields"`
    FieldViews  map[string]string   `bson:"field_views"` // e.g., "description": "tiptap"
}
```

---

## API Endpoints

### Auth
```
POST   /api/auth/login
POST   /api/auth/logout
POST   /api/auth/refresh
```

### Users
```
GET    /api/users
POST   /api/users
GET    /api/users/:id
PUT    /api/users/:id
DELETE /api/users/:id
GET    /api/users/me
PUT    /api/users/me
PUT    /api/users/me/password
PUT    /api/users/me/avatar
```

### Projects
```
GET    /api/projects
POST   /api/projects
GET    /api/projects/:id
PUT    /api/projects/:id
DELETE /api/projects/:id
POST   /api/projects/:id/regenerate-key
POST   /api/projects/:id/members
DELETE /api/projects/:id/members/:userId
```

### Goals
```
GET    /api/goals                    # Global goals
POST   /api/goals
GET    /api/goals/:id
PUT    /api/goals/:id
DELETE /api/goals/:id
GET    /api/goals/stats

GET    /api/projects/:id/goals       # Project goals
POST   /api/projects/:id/goals
PUT    /api/projects/:id/goals/:goalId
DELETE /api/projects/:id/goals/:goalId
```

### Environment
```
GET    /api/projects/:id/env
POST   /api/projects/:id/env
PUT    /api/projects/:id/env/:key
DELETE /api/projects/:id/env/:key
```

### Storage
```
GET    /api/projects/:id/storage
POST   /api/projects/:id/storage/mkdir
POST   /api/projects/:id/storage/upload
GET    /api/projects/:id/storage/download/*path
PUT    /api/projects/:id/storage/rename
DELETE /api/projects/:id/storage/*path
POST   /api/projects/:id/storage/file
PUT    /api/projects/:id/storage/file/*path
GET    /api/storage/public/:token/:filename
```

### Pipeline
```
GET    /api/projects/:id/pipeline/branches
POST   /api/projects/:id/pipeline/branches
GET    /api/projects/:id/pipeline/branches/:name
PUT    /api/projects/:id/pipeline/branches/:name
DELETE /api/projects/:id/pipeline/branches/:name

GET    /api/projects/:id/pipeline/releases
POST   /api/projects/:id/pipeline/releases
DELETE /api/projects/:id/pipeline/releases/:version
POST   /api/projects/:id/pipeline/releases/:version/activate
```

### Models
```
GET    /api/projects/:id/models
POST   /api/projects/:id/models
GET    /api/projects/:id/models/:modelId
PUT    /api/projects/:id/models/:modelId
DELETE /api/projects/:id/models/:modelId

GET    /api/projects/:id/models/:modelId/data
POST   /api/projects/:id/models/:modelId/data
GET    /api/projects/:id/models/:modelId/data/:dataId
PUT    /api/projects/:id/models/:modelId/data/:dataId
DELETE /api/projects/:id/models/:modelId/data/:dataId
```

### Runtime
```
POST   /api/projects/:id/start
POST   /api/projects/:id/stop
POST   /api/projects/:id/restart
GET    /api/projects/:id/status
GET    /api/projects/:id/logs
GET    /api/projects/:id/logs/download

# Project routes (handled by runtime router)
ANY    /api/r/:projectSlug/*route
```

### Runtime API Types (для Monaco)
```
GET    /api/runtime/types
```

---

## Runtime API (JavaScript)

```typescript
// Storage module
declare const storage: {
    read(path: string): string;
    write(path: string, content: string): void;
    exists(path: string): boolean;
    delete(path: string): void;
    list(path: string): string[];
    mkdir(path: string): void;
};

// Image module
declare const image: {
    resize(path: string, width: number, height: number): Buffer;
    crop(path: string, x: number, y: number, width: number, height: number): Buffer;
    convert(path: string, format: string): Buffer;
};

// Draw module
declare const draw: {
    createCanvas(width: number, height: number): Canvas;
};

// Database module
declare const database: {
    collection(name: string): Collection;
};

// Env module
declare const env: {
    get(key: string): any;
};

// HTTP module
declare const http: {
    get(url: string, options?: RequestOptions): Response;
    post(url: string, body: any, options?: RequestOptions): Response;
    put(url: string, body: any, options?: RequestOptions): Response;
    delete(url: string, options?: RequestOptions): Response;
};

// SMTP module
declare const smtp: {
    send(options: EmailOptions): void;
};

// Logger module
declare const logger: {
    debug(...args: any[]): void;
    info(...args: any[]): void;
    warn(...args: any[]): void;
    error(...args: any[]): void;
};

// Crypto module
declare const crypto: {
    md5(data: string): string;
    sha256(data: string): string;
    randomBytes(length: number): string;
};

// Encoding module
declare const encoding: {
    base64Encode(data: string): string;
    base64Decode(data: string): string;
    jsonParse(data: string): any;
    jsonStringify(data: any): string;
    urlEncode(data: string): string;
    urlDecode(data: string): string;
};

// Utils module
declare const utils: {
    sleep(ms: number): void;
    random(): number;
    randomInt(min: number, max: number): number;
    uuid(): string;
    slugify(text: string): string;
    truncate(text: string, length: number): string;
    capitalize(text: string): string;
    regexMatch(text: string, pattern: string): string[];
    regexReplace(text: string, pattern: string, replacement: string): string;
    formatDate(date: Date, format: string): string;
    parseDate(text: string, format: string): Date;
    timestamp(): number;
};

// Delayed module
declare const delayed: {
    run(fn: () => void): void;
};

// Router module
declare const router: {
    get(path: string, handler: (ctx: Context) => Response): void;
    post(path: string, handler: (ctx: Context) => Response): void;
    put(path: string, handler: (ctx: Context) => Response): void;
    delete(path: string, handler: (ctx: Context) => Response): void;
    response(status: number, body: any): Response;
};

// Schedule module
declare const schedule: {
    daily(handler: () => void): void;
    hourly(handler: () => void): void;
    cron(expression: string, handler: () => void): void;
};

// Goals module
declare const goals: {
    increment(slug: string, value?: number): void;
};
```

---

## Конфигурация (config.yaml)

```yaml
server:
  host: "0.0.0.0"
  port: 3000
  uri: "http://127.0.0.1:3000"

mongodb:
  uri: "mongodb://localhost:27017"
  database: "m3m"

jwt:
  secret: "your-jwt-secret-key-change-in-production"
  expiration: 168h

storage:
  path: "./storage"

runtime:
  worker_pool_size: 10
  timeout: 30s

plugins:
  path: "./plugins"

logging:
  level: "info"
  path: "./logs"
```

---

## Порядок реализации

1. **Этап 1-3**: Базовая инфраструктура + Auth + Проекты
2. **Этап 4-6**: Цели + Окружение + Storage
3. **Этап 7-8**: Pipeline + Хранилище данных
4. **Этап 9-10**: Runtime + Логирование
5. **Этап 11**: Плагины
6. **Этап 12-13**: Frontend + Monaco подсказки
