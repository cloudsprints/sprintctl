package types

import (
	"time"
)

type CLICommandResult struct {
	ExitCode int    `json:"exit_code"`
	Command  string `json:"command"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
}

type Lesson struct {
	ID          string    `json:"id"`
	CliCommands []string  `json:"cli_commands"`
	Tasks       []Task    `json:"tasks"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	// Enrichment served by newer servers for IDE/ticket surfaces.
	// Zero-valued when talking to older servers — callers must not require them.
	Title            string           `json:"title,omitempty"`
	LabContent       string           `json:"lab_content,omitempty"`
	Course           *LessonCourse    `json:"course,omitempty"`
	HelpdeskPriority int              `json:"helpdesk_priority,omitempty"`
	ReporterPersona  *ReporterPersona `json:"reporter_persona,omitempty"`
	// Workspace time-limit surface for the IDE countdown. Pointers so a null
	// deadline (uncapped) is distinguishable from an older server that omits
	// them; both mean "no countdown". SecondsRemaining is server-computed
	// (never a local clock).
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	SecondsRemaining *int       `json:"seconds_remaining,omitempty"`
	TimeLimitMinutes *int       `json:"time_limit_minutes,omitempty"`
}

type LessonCourse struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	CourseType string `json:"course_type"`
}

type ReporterPersona struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	AvatarURL   string `json:"avatar_url"`
}

type Task struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Status        string    `json:"status"`
	AiExplanation string    `json:"ai_explanation"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SubmitLessonRequestType string

const (
	SubmitLessonRequestTypeCommandResults SubmitLessonRequestType = "COMMAND_RESULTS"
)

type SubmitLessonRequest struct {
	Type              SubmitLessonRequestType `json:"type"`
	CliCommandResults []CLICommandResult      `json:"cli_command_results"`
}

type ActiveLesson struct {
	LessonID       string `json:"lessonId"`
	UserLessonID   string `json:"userLessonId"`
	LessonToken    string `json:"lessonToken"`
	Title          string `json:"title"`
	CourseTitle    string `json:"courseTitle"`
	LastAccessedAt string `json:"lastAccessedAt"`
}

// Admin types - for direct lesson access without user enrollment

type AdminLessonInfo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	CliCommands []string `json:"cliCommands"`
	Tasks       []Task   `json:"tasks"`
}

type AdminGradeResult struct {
	Success     bool              `json:"success"`
	LessonID    string            `json:"lesson_id"`
	LessonTitle string            `json:"lesson_title"`
	Results     []AdminTaskResult `json:"results"`
	Summary     AdminGradeSummary `json:"summary"`
}

type AdminTaskResult struct {
	TaskNumber  int    `json:"taskNumber"`
	Title       string `json:"title"`
	Passed      bool   `json:"passed"`
	Explanation string `json:"explanation"`
}

type AdminGradeSummary struct {
	Total  int `json:"total"`
	Passed int `json:"passed"`
	Failed int `json:"failed"`
}

// AdminProjectLabs - response from project labs endpoint
type AdminProjectLabs struct {
	ProjectID    string           `json:"project_id"`
	ProjectTitle string           `json:"project_title"`
	Lessons      []AdminLabLesson `json:"lessons"`
}

type AdminLabLesson struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	SequenceOrder int       `json:"sequenceOrder"`
	CliCommands   []string  `json:"cliCommands"`
	Files         []LabFile `json:"files"`
}
