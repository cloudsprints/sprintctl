package types

import "time"

// LabFile represents a file in a lab
type LabFile struct {
	Path         string     `json:"path"`
	URL          string     `json:"url"`
	Size         int64      `json:"size"`
	LastModified time.Time  `json:"lastModified"`
	ExpiresAt    time.Time  `json:"expires_at"`
}

// LabFiles represents the response from the lab files API
type LabFiles struct {
	Files struct {
		Public    []LabFile `json:"public"`
		Bootstrap []LabFile `json:"bootstrap"`
		Other     []LabFile `json:"other"`
	} `json:"files"`
}

// LabPublicFiles represents the response from the lab public files API
type LabPublicFiles struct {
	Files []LabFile `json:"files"`
}

// LabFileURL represents the response from the lab file URL API
type LabFileURL struct {
	URL       string    `json:"url"`
	FilePath  string    `json:"file_path"`
	S3Key     string    `json:"s3_key"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LabInfo represents the lab information
type LabInfo struct {
	UserLessonID string    `json:"user_lesson_id"`
	LessonID     string    `json:"lesson_id"`
	Title        string    `json:"title"`
	Course       struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"course"`
	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	S3Paths   struct {
		Base      string `json:"base"`
		Public    string `json:"public"`
		Bootstrap string `json:"bootstrap"`
		Readme    string `json:"readme"`
	} `json:"s3_paths"`
}

// No need for LabMetadata since we're not storing metadata
