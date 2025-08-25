package models

import (
	"encoding/json"
	"time"
)

// LearnedPattern represents a learned URL pattern with associated tags
type LearnedPattern struct {
	ID             int       `json:"id" db:"id"`
	URLPattern     string    `json:"url_pattern" db:"url_pattern"`
	Domain         string    `json:"domain" db:"domain"`
	ConfirmedTags  []string  `json:"confirmed_tags"`
	RejectedTags   []string  `json:"rejected_tags"`
	Confidence     float64   `json:"confidence" db:"confidence"`
	SampleCount    int       `json:"sample_count" db:"sample_count"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
	
	// Database fields (JSON strings)
	ConfirmedTagsJSON string `db:"confirmed_tags"`
	RejectedTagsJSON  string `db:"rejected_tags"`
}

// TagCorrection represents user feedback on AI-suggested tags
type TagCorrection struct {
	ID             int       `json:"id" db:"id"`
	BookmarkID     int       `json:"bookmark_id" db:"bookmark_id"`
	OriginalTags   []string  `json:"original_tags"`
	FinalTags      []string  `json:"final_tags"`
	CorrectionType string    `json:"correction_type" db:"correction_type"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	
	// Database fields (JSON strings)
	OriginalTagsJSON string `db:"original_tags"`
	FinalTagsJSON    string `db:"final_tags"`
}

// DomainProfile represents user preferences for a specific domain
type DomainProfile struct {
	ID             int            `json:"id" db:"id"`
	Domain         string         `json:"domain" db:"domain"`
	CommonTags     []string       `json:"common_tags"`
	IgnoredTags    []string       `json:"ignored_tags"`
	CustomMappings map[string]string `json:"custom_mappings"`
	BookmarkCount  int            `json:"bookmark_count" db:"bookmark_count"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
	
	// Database fields (JSON strings)
	CommonTagsJSON     string `db:"common_tags"`
	IgnoredTagsJSON    string `db:"ignored_tags"`
	CustomMappingsJSON string `db:"custom_mappings"`
}

// TagFeedback represents user feedback for learning
type TagFeedback struct {
	URL           string   `json:"url"`
	SuggestedTags []string `json:"suggested_tags"`
	KeptTags      []string `json:"kept_tags"`
	AddedTags     []string `json:"added_tags"`
	RemovedTags   []string `json:"removed_tags"`
	Pattern       string   `json:"pattern"`
}

// TagPrediction represents an AI-predicted tag with confidence
type TagPrediction struct {
	Name       string  `json:"name"`
	Confidence float64 `json:"confidence"`
	Source     string  `json:"source"` // "rule", "fasttext", "onnx", "learned"
	IsLearned  bool    `json:"is_learned"`
}

// Helper methods for JSON serialization/deserialization

// BeforeSave prepares the model for database storage
func (lp *LearnedPattern) BeforeSave() error {
	confirmedJSON, err := json.Marshal(lp.ConfirmedTags)
	if err != nil {
		return err
	}
	lp.ConfirmedTagsJSON = string(confirmedJSON)

	rejectedJSON, err := json.Marshal(lp.RejectedTags)
	if err != nil {
		return err
	}
	lp.RejectedTagsJSON = string(rejectedJSON)

	return nil
}

// AfterLoad prepares the model after loading from database
func (lp *LearnedPattern) AfterLoad() error {
	if lp.ConfirmedTagsJSON != "" {
		if err := json.Unmarshal([]byte(lp.ConfirmedTagsJSON), &lp.ConfirmedTags); err != nil {
			return err
		}
	}

	if lp.RejectedTagsJSON != "" {
		if err := json.Unmarshal([]byte(lp.RejectedTagsJSON), &lp.RejectedTags); err != nil {
			return err
		}
	}

	return nil
}

// BeforeSave prepares the model for database storage
func (tc *TagCorrection) BeforeSave() error {
	originalJSON, err := json.Marshal(tc.OriginalTags)
	if err != nil {
		return err
	}
	tc.OriginalTagsJSON = string(originalJSON)

	finalJSON, err := json.Marshal(tc.FinalTags)
	if err != nil {
		return err
	}
	tc.FinalTagsJSON = string(finalJSON)

	return nil
}

// AfterLoad prepares the model after loading from database
func (tc *TagCorrection) AfterLoad() error {
	if tc.OriginalTagsJSON != "" {
		if err := json.Unmarshal([]byte(tc.OriginalTagsJSON), &tc.OriginalTags); err != nil {
			return err
		}
	}

	if tc.FinalTagsJSON != "" {
		if err := json.Unmarshal([]byte(tc.FinalTagsJSON), &tc.FinalTags); err != nil {
			return err
		}
	}

	return nil
}

// BeforeSave prepares the model for database storage
func (dp *DomainProfile) BeforeSave() error {
	commonJSON, err := json.Marshal(dp.CommonTags)
	if err != nil {
		return err
	}
	dp.CommonTagsJSON = string(commonJSON)

	ignoredJSON, err := json.Marshal(dp.IgnoredTags)
	if err != nil {
		return err
	}
	dp.IgnoredTagsJSON = string(ignoredJSON)

	mappingsJSON, err := json.Marshal(dp.CustomMappings)
	if err != nil {
		return err
	}
	dp.CustomMappingsJSON = string(mappingsJSON)

	return nil
}

// AfterLoad prepares the model after loading from database
func (dp *DomainProfile) AfterLoad() error {
	if dp.CommonTagsJSON != "" {
		if err := json.Unmarshal([]byte(dp.CommonTagsJSON), &dp.CommonTags); err != nil {
			return err
		}
	}

	if dp.IgnoredTagsJSON != "" {
		if err := json.Unmarshal([]byte(dp.IgnoredTagsJSON), &dp.IgnoredTags); err != nil {
			return err
		}
	}

	if dp.CustomMappingsJSON != "" {
		if err := json.Unmarshal([]byte(dp.CustomMappingsJSON), &dp.CustomMappings); err != nil {
			return err
		}
	}

	return nil
}