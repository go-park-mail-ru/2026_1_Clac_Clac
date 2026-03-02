# Архитектура и нормализация базы данных (BuisnesClac)

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
* **`Task`** — базовая неизменяемая сущность задачи.
* **`TaskVersion`** — исторические состояния задачи. Содержит сроки, текстовые данные и текущую привязку к колонке (`SectionID`).
* **`WorkerTask`** — связь многие-ко-многим для назначения исполнителей на задачи.
* **`ListenerTask`** — связь многие-ко-многим для назначения наблюдателей за задачами.
* **`SubTask`** — элементы чек-листа (подзадачи), привязанные к родительской задаче.
* **`TaskDependency`** — связи блокировок: определяет, выполнение какой задачи блокируется другой задачей.
* **`CommentTask`** — комментарии к задачам с поддержкой иерархии (через `ParentID`).

---

## 2. Функциональные зависимости

Ниже представлены все нетривиальные функциональные зависимости для каждого отношения:

**Relation User:**
`{ID} -> Login, Link, DisplayName, Password, Email, Avatar`
`{Login} -> ID, Link, DisplayName, Password, Email, Avatar`
`{Email} -> ID, Login, Link, DisplayName, Password, Avatar`

**Relation Board:**
`{ID} -> Link, CreatedAt`

**Relation BoardVersion:**
`{ID} -> BoardID, BoardName, Description, Background, ValidFrom, ValidTo`

**Relation MemberBoard:**
`{BoardID, UserID} -> IsLike, IsArchive, Level`

**Relation BoardTemplate:**
`{ID} -> AuthorID, TemplateName`

**Relation SectionTemplate:**
`{ID} -> TemplateID, SectionName, Position, IsMandatory, MaxTasks`

**Relation Section:**
`{ID} -> BoardID, Link`

**Relation SectionVersion:**
`{ID} -> SectionID, SectionName, Position, IsMandatory, MaxTasks, ValidFrom, ValidTo`

**Relation Task:**
`{ID} -> AuthorID, Link, CreatedAt`

**Relation TaskVersion:**
`{ID} -> TaskID, SectionID, Title, Description, Position, TaskStartAt, Duedate, ValidFrom, ValidTo`

**Relation WorkerTask:**
`{AssigneeID, TaskID} -> ∅` (Атрибутов нет, ключ составной)

**Relation ListenerTask:**
`{ListenerID, TaskID} -> ∅`

**Relation SubTask:**
`{ID} -> TaskID, Link, Description, IsDone, Position`

**Relation TaskDependency:**
`{BlockingTaskID, BlockedTaskID} -> ∅`

**Relation CommentTask:**
`{ID} -> TaskID, ParentID, Link, Text, CreatedAt`

---

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
**Обоснование:** Исходя из списка функциональных зависимостей (п. 2), в левой части каждого выражения (в роли детерминанта $X$) выступает исключительно первичный ключ или потенциальный ключ (как `{Login}` или `{Email}` в таблице `User`). Ни одна часть составного ключа не зависит от неключевых атрибутов, и нет перекрывающихся потенциальных ключей, вызывающих аномалии. База данных строго соответствует НФБК.

---

## 4. ER-диаграмма (Mermaid)

```mermaid
erDiagram
    User ||--o{ MemberBoard : joins
    User ||--o{ BoardTemplate : creates
    User ||--o{ Task : authors
    User ||--o{ WorkerTask : assigned_to
    User ||--o{ ListenerTask : listens_to
    User {
        ID PK
        Login UK
        Link UK
        DisplayName
        Password
        Email UK
        Avatar
    }

    Board ||--o{ BoardVersion : has_history
    Board ||--o{ MemberBoard : has_members
    Board ||--o{ Section : contains
    Board {
        ID PK
        Link UK
        CreatedAt
    }

    BoardVersion {
        ID PK
        BoardID FK
        BoardName
        Description
        Background
        ValidFrom
        ValidTo
    }

    MemberBoard {
        BoardID PK, FK
        UserID PK, FK
        Level
        IsLike
        IsArchive
    }

    BoardTemplate ||--o{ SectionTemplate : defines_sections
    BoardTemplate {
        ID PK
        AuthorID FK
        TemplateName
    }

    SectionTemplate {
        ID PK
        TemplateID FK
        SectionName
        Position
        IsMandatory
        MaxTasks
    }

    Section ||--o{ SectionVersion : has_history
    Section ||--o{ TaskVersion : holds_tasks
    Section {
        ID PK
        BoardID FK
        Link UK
    }

    SectionVersion {
        ID PK
        SectionID FK
        SectionName
        Position
        IsMandatory
        MaxTasks
        ValidFrom
        ValidTo
    }

    Task ||--o{ TaskVersion : has_history
    Task ||--o{ SubTask : has
    Task ||--o{ WorkerTask : has_workers
    Task ||--o{ ListenerTask : has_listeners
    Task ||--o{ CommentTask : has_comments
    Task ||--o{ TaskDependency : blocks
    Task ||--o{ TaskDependency : is_blocked_by
    Task {
        ID PK
        AuthorID FK
        Link UK
        CreatedAt
    }

    TaskVersion {
        ID PK
        TaskID FK
        SectionID FK
        Title
        Description
        Position
        TaskStartAt
        Duedate
        ValidFrom
        ValidTo
    }

    WorkerTask {
        AssigneeID PK, FK
        TaskID PK, FK
    }

    ListenerTask {
        ListenerID PK, FK
        TaskID PK, FK
    }

    SubTask {
        ID PK
        TaskID FK
        Link UK
        Description
        IsDone
        Position
    }

    TaskDependency {
        BlockingTaskID PK, FK
        BlockedTaskID PK, FK
    }

    CommentTask ||--o{ CommentTask : replies_to
    CommentTask {
        ID PK
        TaskID FK
        ParentID FK
        Link UK
        Text
        CreatedAt
    }
