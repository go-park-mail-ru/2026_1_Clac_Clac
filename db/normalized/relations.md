# Архитектура и нормализация базы данных (NeXus)

Данный документ описывает структуру реляционной базы данных для проекта NeXus, функциональные зависимости отношений и доказательство соответствия схемы Нормальной форме Бойса-Кодда (НФБК).

---

## 1. Описание отношений (таблиц)

* **`User`** — уникальные учетные данные и профили пользователей системы.
* **`Board`** — базовая неизменяемая сущность доски (идентификаторы и дата создания).
* **`BoardVersion`** — исторические состояния доски. Хранит название, описание и фон в определенный период времени (SCD Type 2).
* **`MemberBoard`** — отношение многие-ко-многим. Определяет права доступа пользователей к доскам (`Level`) и их персональные настройки (`IsLike`, `IsArchive`).
* **`BoardTemplate`** — реестр шаблонов для быстрого создания досок.
* **`SectionTemplate`** — колонки, привязанные к конкретному шаблону доски (нормализованная замена массиву секций).
* **`Section`** — базовая неизменяемая сущность колонки (секции) на конкретной доске.
* **`SectionVersion`** — исторические состояния колонки (название, позиция на доске, лимиты задач) с привязкой ко времени.
* **`Task`** — базовая неизменяемая сущность задачи. Хранит только вечные данные: автора, уникальную ссылку и дату создания.
* **`TaskVersion`** — исторические состояния задачи. Содержит сроки, текстовые данные и текущую привязку к колонке (`SectionID`), что позволяет отслеживать историю перемещения задачи по доске.
* **`WorkerTask`** — связь многие-ко-многим для назначения исполнителей (Assignee) на задачи.
* **`ListenerTask`** — связь многие-ко-многим для назначения наблюдателей (Listener) за задачами.
* **`SubTask`** — элементы чек-листа (подзадачи), привязанные к родительской задаче.
* **`TaskDependency`** — связи блокировок: определяет, выполнение какой задачи блокируется другой задачей.
* **`CommentTask`** — комментарии к задачам с поддержкой иерархии (через `ParentID` для реализации ответов на комментарии).

---

## 2. Функциональные зависимости

Ниже представлены все нетривиальные функциональные зависимости для каждого отношения с учетом системных полей аудита:

**Relation User:**
`{ID} -> Login, Link, DisplayName, Password, Email, Avatar, CreatedAt, UpdatedAt`
`{Email} -> ID, Login, Link, DisplayName, Password, Avatar, CreatedAt, UpdatedAt`
`{Link} -> ID, Login, DisplayName, Password, Email, Avatar, CreatedAt, UpdatedAt`

**Relation Board:**
`{ID} -> Link, CreatedAt`

**Relation BoardVersion:**
`{ID} -> BoardID, BoardName, Description, Background, ValidFrom, ValidTo`

**Relation MemberBoard:**
`{BoardID, UserID} -> IsLike, IsArchive, Level, CreatedAt, UpdatedAt`

**Relation BoardTemplate:**
`{ID} -> AuthorID, TemplateName, CreatedAt, UpdatedAt`

**Relation SectionTemplate:**
`{ID} -> TemplateID, SectionName, Position, IsMandatory, MaxTasks, CreatedAt, UpdatedAt`

**Relation Section:**
`{ID} -> BoardID, Link`

**Relation SectionVersion:**
`{ID} -> SectionID, SectionName, Position, IsMandatory, MaxTasks, ValidFrom, ValidTo`

**Relation Task:**
`{ID} -> AuthorID, Link, CreatedAt`

**Relation TaskVersion:**
`{ID} -> TaskID, SectionID, Title, Description, Position, TaskStartAt, Duedate, ValidFrom, ValidTo`

**Relation WorkerTask:**
`{AssigneeID, TaskID} -> CreatedAt`

**Relation ListenerTask:**
`{ListenerID, TaskID} -> CreatedAt`

**Relation SubTask:**
`{ID} -> TaskID, Link, Description, IsDone, Position, CreatedAt, UpdatedAt`

**Relation TaskDependency:**
`{BlockingTaskID, BlockedTaskID} -> CreatedAt`

**Relation CommentTask:**
`{ID} -> TaskID, ParentID, Link, Text, CreatedAt, UpdatedAt`

## 3. Доказательство нормализации

Схема данных спроектирована с учетом требований строгой нормализации и полностью соответствует Нормальной форме Бойса-Кодда (НФБК).

### Первая нормальная форма (1НФ)
**Требование:** Отсутствие повторяющихся групп и составных атрибутов; все атрибуты атомарны.
**Обоснование:** В схеме нет массивов или JSON-полей для хранения множественных данных. Например, потенциальный массив секций в шаблоне доски вынесен в отдельное отношение `SectionTemplate`. У каждого отношения определен первичный ключ (включая составные ключи в связующих таблицах).

### Вторая нормальная форма (2НФ)
**Требование:** Выполнение 1НФ и отсутствие частичных зависимостей от составного первичного ключа.
**Обоснование:** Отношения с одиночным первичным ключом (`ID`) автоматически находятся во 2НФ. В таблицах с составным ключом (`MemberBoard`, `WorkerTask`, `ListenerTask`, `TaskDependency`) все неключевые атрибуты зависят строго от всего ключа целиком. Например, в `MemberBoard` атрибуты `Level`, `IsLike` и `IsArchive` зависят от комбинации `{BoardID, UserID}`, а не отдельно от пользователя или доски.

### Третья нормальная форма (3НФ)
**Требование:** Выполнение 2НФ и отсутствие транзитивных зависимостей (когда неключевой атрибут зависит от другого неключевого атрибута).
**Обоснование:** В схеме принципиально отсутствуют вычисляемые атрибуты (например, количество подзадач или прогресс выполнения), которые создавали бы транзитивную зависимость от записей в других таблицах. Все неключевые атрибуты в каждом отношении зависят **только** от первичного ключа этого отношения.

### Нормальная форма Бойса-Кодда (НФБК)
**Требование:** Для любой нетривиальной функциональной зависимости $X \rightarrow Y$, детерминант $X$ обязан быть суперключом.
**Обоснование:** Исходя из списка функциональных зависимостей (п. 2), в левой части каждого выражения (в роли детерминанта $X$) выступает исключительно первичный ключ или потенциальный ключ (`{Email}` в таблице `User`). Ни одна часть составного ключа не зависит от неключевых атрибутов, и нет перекрывающихся потенциальных ключей, вызывающих аномалии. База данных строго соответствует НФБК.

---

## 4. Дополнительные СУБД

Для обеспечения высокой производительности, масштабируемости и снижения нагрузки на основную реляционную базу данных, архитектура BuisnesClac использует гибридный подход к хранению различных типов данных:

* **S3-совместимое объектное хранилище (Object Storage):** Используется для хранения статического и бинарного медиаконтента (пользовательские аватары, фоновые изображения досок).
    **Обоснование:** Хранение BLOB-объектов в реляционной БД приводит к фрагментации файлов данных и деградации производительности. Вынесение статики в S3 позволяет эффективно управлять большими объемами медиафайлов, снижает стоимость хранения и открывает возможность легкой интеграции с CDN (Content Delivery Network) для ускорения загрузки контента на клиенте. В реляционной БД (поля `Avatar`, `Background`) хранятся исключительно легковесные URL-ссылки или идентификаторы объектов S3.
* **Redis (In-memory Data Structure Store):** Применяется в качестве высокоскоростного хранилища (Key-Value) для управления пользовательскими сессиями (Session Management).
    **Обоснование:** Аутентификация и валидация токенов/сессий происходит при каждом запросе к API. Использование оперативной памяти (Redis) для этих целей гарантирует минимальную задержку (low latency) при чтении. Кроме того, Redis предоставляет нативные механизмы TTL (Time-To-Live) для автоматической экспирации и инвалидации устаревших сессий, полностью снимая эту нагрузку (высокочастотные операции чтения/записи) с транзакционной базы данных.

---

## 5. ER-диаграмма (Mermaid)

```mermaid
erDiagram
    %% Внешние хранилища
    S3 {
        string Storage "Для хранения медиа"
    }
    Redis {
        string Session "Для хранения сессий"
    }

    %% Сущности
    User {
        int ID PK
        uuid Link
        string DisplayName
        string Password
        string Email
        string Avatar
        timestamp CreatedAt
        timestamp UpdatedAt
    }

    BoardTemplate {
        int ID PK
        int AuthorID FK
        string TemplateName
        timestamp CreatedAt
        timestamp UpdatedAt
    }

    SectionTemplate {
        int ID PK
        int TemplateID FK
        int Position
        boolean IsMandatory
        int MaxTasks
        string SectionName
        timestamp CreatedAt
        timestamp UpdatedAt
    }

    MemberBoard {
        int BoardID FK
        int UserID FK
        boolean IsLike
        boolean IsArchive
        int Level
        timestamp CreatedAt
    }

    Board {
        int ID PK
        uuid Link
    }

    BoardVersion {
        int ID PK
        int BoardID FK
        string BoardName
        string Description
        string BackGround
        timestamp ValidFrom
        timestamp ValidTo
    }

    Section {
        int ID PK
        int BoardID FK
        uuid Link
    }

    SectionVersion {
        int ID PK
        int SectionID FK
        string SectionName
        int Position
        boolean IsMandatory
        int MaxTasks
        timestamp ValidFrom
        timestamp ValidTo
    }

    WorkerTask {
        int AssigneeID PK
        int TaskID PK
        timestamp CreatedAt
    }

    ListenerTask {
        int ListenerID PK
        int TaskID PK
        timestamp CreatedAt
    }

    TaskDependency {
        int BlockingTaskID PK
        int BlockedTaskID PK
        timestamp CreatedAt
    }

    Task {
        int ID PK
        int AuthorID FK
        int SectionID FK
        uuid Link
    }

    TaskVersion {
        int ID PK
        int TaskID FK
        int SectionID FK
        string Title
        string Description
        int Position
        timestamp DueDate
        timestamp ValidFrom
        timestamp ValidTo
    }

    SubTask {
        int ID PK
        int TaskID FK
        uuid Link
        string Description
        boolean IsDone
        int Position
        timestamp CreatedAt
        timestamp UpdatedAt
    }

    CommentTask {
        int ID PK
        int TaskID FK
        int ParentID FK
        uuid Link
        string Text
        timestamp CreatedAt
        timestamp UpdatedAt
    }

    %% Связи
    S3 ||--o| User : "Avatar"
    S3 ||--o| BoardVersion : "BackGround"

    User ||--o{ BoardTemplate : "AuthorID"
    User ||--o{ MemberBoard : "UserID"
    User ||--o{ WorkerTask : "AssigneeID"
    User ||--o{ ListenerTask : "ListenerID"
    User ||--o{ Task : "AuthorID"

    BoardTemplate ||--|{ SectionTemplate : "TemplateID"

    Board ||--|{ MemberBoard : "BoardID"
    Board ||--|{ BoardVersion : "BoardID"
    Board ||--o{ Section : "BoardID"

    Section ||--|{ SectionVersion : "SectionID"
    Section ||--|{ TaskVersion : "SectionID"

    Task ||--|{ WorkerTask : "TaskID"
    Task ||--|{ ListenerTask : "TaskID"
    Task ||--|{ TaskVersion : "TaskID"
    Task ||--o{ SubTask : "TaskID"
    Task ||--o{ CommentTask : "TaskID"
    Task ||--o{ TaskDependency : "BlockingTaskID"
    Task ||--o{ TaskDependency : "BlockedTaskID"

    CommentTask ||--o{ CommentTask : "ParentID"
```
