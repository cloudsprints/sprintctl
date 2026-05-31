package mtcapi

import (
	"encoding/json"
	"fmt"

	"github.com/morethancertified/sprintctl/internal/types"
)

// GetLabInfo fetches information about a lab
func (c *MtcApiClient) GetLabInfo(userLessonID string) (types.LabInfo, error) {
	res, err := c.httpClient.R().
		SetResult(&types.LabInfo{}).
		Get("/labs/" + userLessonID)
	if err != nil {
		return types.LabInfo{}, err
	}

	if res.IsError() {
		return types.LabInfo{}, fmt.Errorf("API error: %s", res.String())
	}

	return *res.Result().(*types.LabInfo), nil
}

// GetLabFiles fetches all files for a lab (public, bootstrap, and other)
func (c *MtcApiClient) GetLabFiles(userLessonID string) ([]types.LabFile, error) {
	// Make a single request and print the raw response for debugging
	rawRes, err := c.httpClient.R().
		Get("/labs/" + userLessonID + "/files")
	if err != nil {
		return nil, err
	}

	if rawRes.IsError() {
		return nil, fmt.Errorf("API error: %s", rawRes.String())
	}

	// No need for debug printing in production code

	// Check for empty files response: {"files":[]}
	if string(rawRes.Body()) == "{\"files\":[]}" {
		return []types.LabFile{}, nil
	}

	// Try to unmarshal as a direct array of LabFile
	var directFiles []types.LabFile
	err = json.Unmarshal(rawRes.Body(), &directFiles)
	if err == nil {
		// If this succeeds, return the files directly
		return directFiles, nil
	}

	// Try to unmarshal as a simple wrapper with files array
	var simpleFiles struct {
		Files []types.LabFile `json:"files"`
	}
	err = json.Unmarshal(rawRes.Body(), &simpleFiles)
	if err == nil && simpleFiles.Files != nil {
		return simpleFiles.Files, nil
	}

	// If simple unmarshaling fails, try the full structured approach
	var labFiles types.LabFiles
	err = json.Unmarshal(rawRes.Body(), &labFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// Combine all file types
	var files []types.LabFile
	files = append(files, labFiles.Files.Public...)
	files = append(files, labFiles.Files.Bootstrap...)
	files = append(files, labFiles.Files.Other...)

	return files, nil
}

// GetLabPublicFiles fetches only public files for a lab
func (c *MtcApiClient) GetLabPublicFiles(userLessonID string) ([]types.LabFile, error) {
	// Make a single request and print the raw response for debugging
	rawRes, err := c.httpClient.R().
		Get("/labs/" + userLessonID + "/files/public")
	if err != nil {
		return nil, err
	}

	if rawRes.IsError() {
		return nil, fmt.Errorf("API error: %s", rawRes.String())
	}

	// No need for debug printing in production code

	// Check for empty files response: {"files":[]}
	if string(rawRes.Body()) == "{\"files\":[]}" {
		return []types.LabFile{}, nil
	}

	// Try to unmarshal as a direct array of LabFile
	var directFiles []types.LabFile
	err = json.Unmarshal(rawRes.Body(), &directFiles)
	if err == nil {
		// If this succeeds, return the files directly
		return directFiles, nil
	}

	// Try to unmarshal as a simple wrapper with files array
	var simpleFiles struct {
		Files []types.LabFile `json:"files"`
	}
	err = json.Unmarshal(rawRes.Body(), &simpleFiles)
	if err == nil && simpleFiles.Files != nil {
		return simpleFiles.Files, nil
	}

	// If simple unmarshaling fails, try the structured approach
	var labFiles types.LabPublicFiles
	err = json.Unmarshal(rawRes.Body(), &labFiles)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public files response: %v", err)
	}

	return labFiles.Files, nil
}

// GetLabFileURL fetches a pre-signed URL for a specific file
func (c *MtcApiClient) GetLabFileURL(userLessonID string, filePath string) (types.LabFileURL, error) {
	res, err := c.httpClient.R().
		SetResult(&types.LabFileURL{}).
		Get("/labs/" + userLessonID + "/files/" + filePath)
	if err != nil {
		return types.LabFileURL{}, err
	}

	if res.IsError() {
		return types.LabFileURL{}, fmt.Errorf("API error: %s", res.String())
	}

	return *res.Result().(*types.LabFileURL), nil
}
