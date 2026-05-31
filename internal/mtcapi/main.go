package mtcapi

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/morethancertified/sprintctl/internal/auth"
	"github.com/morethancertified/sprintctl/internal/types"
)

type MtcApiClient struct {
	BaseURL    string
	httpClient *resty.Client
}

func New(baseURL string) *MtcApiClient {
	httpClient := resty.New()
	httpClient.SetBaseURL(baseURL)

	return &MtcApiClient{
		BaseURL:    baseURL,
		httpClient: httpClient,
	}
}

// ensureAuthenticated checks for a valid token and sets it on the request
func (c *MtcApiClient) ensureAuthenticated(req *resty.Request) error {
	token, err := auth.GetToken()
	if err != nil {
		return fmt.Errorf("authentication required: please run 'sprintctl login' first")
	}

	req.SetAuthToken(token)
	return nil
}

func (c *MtcApiClient) GetLesson(lessonToken string) (types.Lesson, error) {
	req := c.httpClient.R().
		SetResult(&types.Lesson{})

	if err := c.ensureAuthenticated(req); err != nil {
		return types.Lesson{}, err
	}

	res, err := req.Get("/lessons/" + lessonToken)
	if err != nil {
		return types.Lesson{}, err
	}

	results := *res.Result().(*types.Lesson)

	if len(results.Tasks) == 0 {
		return types.Lesson{}, fmt.Errorf("token is invalid")
	}

	return results, nil
}

func (c *MtcApiClient) SubmitLesson(lessonToken string, cliCommandResults []types.CLICommandResult) (types.Lesson, error) {
	req := c.httpClient.R().
		SetBody(types.SubmitLessonRequest{
			Type:              types.SubmitLessonRequestTypeCommandResults,
			CliCommandResults: cliCommandResults,
		}).
		SetResult(&types.Lesson{})

	if err := c.ensureAuthenticated(req); err != nil {
		return types.Lesson{}, err
	}

	res, err := req.Post("/lessons/" + lessonToken + "/submit")
	if err != nil {
		return types.Lesson{}, err
	}

	if res.IsError() {
		return types.Lesson{}, fmt.Errorf("%s", res.String())
	}

	return *res.Result().(*types.Lesson), nil
}

func (c *MtcApiClient) ResetLesson(lessonToken string) (types.Lesson, error) {
	req := c.httpClient.R().
		SetResult(&types.Lesson{})

	if err := c.ensureAuthenticated(req); err != nil {
		return types.Lesson{}, err
	}

	res, err := req.Post("/lessons/" + lessonToken + "/reset")
	if err != nil {
		return types.Lesson{}, err
	}
	return *res.Result().(*types.Lesson), nil
}

func (c *MtcApiClient) GetActiveLesson() (types.ActiveLesson, error) {
	req := c.httpClient.R().
		SetResult(&types.ActiveLesson{})

	if err := c.ensureAuthenticated(req); err != nil {
		return types.ActiveLesson{}, err
	}

	res, err := req.Get("/grading/active-lab")
	if err != nil {
		return types.ActiveLesson{}, err
	}

	if res.IsError() {
		return types.ActiveLesson{}, fmt.Errorf("%s", res.String())
	}

	return *res.Result().(*types.ActiveLesson), nil
}

func ValidCUID(cuid string) bool {
	return len(cuid) >= 7 && strings.HasPrefix(cuid, "c")
}

// Admin API methods - use lesson_id directly (not user_lesson_id)

// GetAdminLessonInfo fetches lesson info via admin endpoint
func (c *MtcApiClient) GetAdminLessonInfo(lessonID string) (types.AdminLessonInfo, error) {
	req := c.httpClient.R().
		SetResult(&types.AdminLessonInfo{})

	if err := c.ensureAuthenticated(req); err != nil {
		return types.AdminLessonInfo{}, err
	}

	res, err := req.Get("/admin/labs/" + lessonID)
	if err != nil {
		return types.AdminLessonInfo{}, err
	}

	if res.IsError() {
		return types.AdminLessonInfo{}, fmt.Errorf("API error: %s", res.String())
	}

	return *res.Result().(*types.AdminLessonInfo), nil
}

// GetAdminLabFiles fetches all lab files including solutions via admin endpoint
func (c *MtcApiClient) GetAdminLabFiles(lessonID string) ([]types.LabFile, error) {
	req := c.httpClient.R()

	if err := c.ensureAuthenticated(req); err != nil {
		return nil, err
	}

	res, err := req.Get("/admin/labs/" + lessonID + "/files")
	if err != nil {
		return nil, err
	}

	if res.IsError() {
		return nil, fmt.Errorf("API error: %s", res.String())
	}

	// Parse response - expect { files: [...] }
	var response struct {
		Files []types.LabFile `json:"files"`
	}
	if err := json.Unmarshal(res.Body(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return response.Files, nil
}

// GetAdminProjectLabs fetches all labs for a project via admin endpoint
func (c *MtcApiClient) GetAdminProjectLabs(projectID string) (types.AdminProjectLabs, error) {
	req := c.httpClient.R().
		SetResult(&types.AdminProjectLabs{})

	if err := c.ensureAuthenticated(req); err != nil {
		return types.AdminProjectLabs{}, err
	}

	res, err := req.Get("/admin/projects/" + projectID + "/labs")
	if err != nil {
		return types.AdminProjectLabs{}, err
	}

	if res.IsError() {
		return types.AdminProjectLabs{}, fmt.Errorf("API error: %s", res.String())
	}

	return *res.Result().(*types.AdminProjectLabs), nil
}

// AdminGradeLesson grades a lesson via admin endpoint (no user enrollment required)
func (c *MtcApiClient) AdminGradeLesson(lessonID string, cliCommandResults []types.CLICommandResult) (types.AdminGradeResult, error) {
	req := c.httpClient.R().
		SetBody(map[string]interface{}{
			"cliOutput": cliCommandResults,
		}).
		SetResult(&types.AdminGradeResult{})

	if err := c.ensureAuthenticated(req); err != nil {
		return types.AdminGradeResult{}, err
	}

	res, err := req.Post("/admin/labs/" + lessonID + "/grade")
	if err != nil {
		return types.AdminGradeResult{}, err
	}

	if res.IsError() {
		return types.AdminGradeResult{}, fmt.Errorf("API error: %s", res.String())
	}

	return *res.Result().(*types.AdminGradeResult), nil
}
