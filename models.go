package entry

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Password string    `json:"password"`
	Email    string    `json:"email"`
	Boards   []Board   `json:"boards"`
}

type MemberBoard struct {
	BoardID uuid.UUID `json:"board_id"`
	UserID  uuid.UUID `json:"user_id"`
	Level   int       `json:"level"`
}

type Board struct {
	ID             uuid.UUID `json:"id"`
	BoardName      string    `json:"board_name"`
	Description    string    `json:"description"`
	IsPublic       bool      `json:"is_public"`
	NumberTemplate int       `json:"number_template"`
	CreatedAt      time.Time `json:"created_at"`
	Contributers   []User    `json:"contributers"` // под вопросом, так как возможно лучше
	// делать отдельный запрос, чтобы не было перегруженного запрса
}

type Section struct {
	ID          uuid.UUID `json:"id"`
	BoardID     uuid.UUID `json:"board_id"` // чтобы быстро дсотавать инфорамацию
	SectionName string    `json:"section_name"`
	Position    int       `json:"position"`
	MaxTasks    *int      `json:"max_tasks"`
	Tasks       []Task    `json:"tasks"`
}

type Task struct {
	ID          uuid.UUID  `json:"id"`
	SectionID   uuid.UUID  `json:"section_id"`
	BoardID     uuid.UUID  `json:"board_id"`
	AuthorID    uuid.UUID  `json:"author_id"`
	Title       string     `json:"title"`
	Description *string    `json:"description"`
	StoryPoints int        `json:"rating"`
	Position    int        `json:"position"`
	StartAt     *time.Time `json:"start_at"`
	Duedate     *time.Time `json:"due_date"`
	CreatedAt   *time.Time `json:"created_at"`
}

type WorkerTask struct {
	AssigneeID uuid.UUID `json:"assignee_id"`
	TaskID     uuid.UUID `json:"task_id"`
}

type ListenerTask struct {
	ListenerID uuid.UUID `json:"listener_id"`
	TaskID     uuid.UUID `json:"task_id"`
}
