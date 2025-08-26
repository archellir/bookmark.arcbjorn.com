package db

import (
	"database/sql"
	"encoding/json/v2"
	"time"

	"torimemo/internal/models"
)

// LearningRepository handles database operations for the AI learning system
type LearningRepository struct {
	db *DB
}

// NewLearningRepository creates a new learning repository
func NewLearningRepository(db *DB) *LearningRepository {
	return &LearningRepository{db: db}
}

// SavePattern saves a learned pattern to the database
func (r *LearningRepository) SavePattern(pattern *models.LearnedPattern) error {
	confirmedTagsJSON, err := json.Marshal(pattern.ConfirmedTags)
	if err != nil {
		return err
	}

	rejectedTagsJSON, err := json.Marshal(pattern.RejectedTags)
	if err != nil {
		return err
	}

	query := `
		INSERT OR REPLACE INTO learned_patterns 
		(url_pattern, domain, confirmed_tags, rejected_tags, confidence, sample_count, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	
	_, err = r.db.Exec(query, 
		pattern.URLPattern, 
		pattern.Domain,
		string(confirmedTagsJSON), 
		string(rejectedTagsJSON),
		pattern.Confidence, 
		pattern.SampleCount,
		now, 
		now)
	
	return err
}

// GetPatternByURL gets a learned pattern by URL
func (r *LearningRepository) GetPatternByURL(url string) (*models.LearnedPattern, error) {
	query := `SELECT id, url_pattern, domain, confirmed_tags, rejected_tags, confidence, sample_count, created_at, updated_at 
			  FROM learned_patterns WHERE url_pattern = ?`
	
	var pattern models.LearnedPattern
	var confirmedTagsJSON, rejectedTagsJSON string
	var createdAt, updatedAt time.Time
	
	err := r.db.QueryRow(query, url).Scan(
		&pattern.ID,
		&pattern.URLPattern,
		&pattern.Domain,
		&confirmedTagsJSON,
		&rejectedTagsJSON,
		&pattern.Confidence,
		&pattern.SampleCount,
		&createdAt,
		&updatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	err = json.Unmarshal([]byte(confirmedTagsJSON), &pattern.ConfirmedTags)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal([]byte(rejectedTagsJSON), &pattern.RejectedTags)
	if err != nil {
		return nil, err
	}
	
	pattern.CreatedAt = createdAt
	pattern.UpdatedAt = updatedAt
	
	return &pattern, nil
}

// SaveTagCorrection saves user tag corrections for learning
func (r *LearningRepository) SaveTagCorrection(correction *models.TagCorrection) error {
	query := `
		INSERT INTO tag_corrections 
		(bookmark_id, original_tags, final_tags, correction_type, created_at) 
		VALUES (?, ?, ?, ?, ?)`
	
	originalJSON, err := json.Marshal(correction.OriginalTags)
	if err != nil {
		return err
	}
	
	finalJSON, err := json.Marshal(correction.FinalTags)
	if err != nil {
		return err
	}
	
	_, err = r.db.Exec(query, 
		correction.BookmarkID,
		string(originalJSON), 
		string(finalJSON),
		correction.CorrectionType,
		time.Now())
	
	return err
}

// GetTagCorrections gets all tag corrections for analysis
func (r *LearningRepository) GetTagCorrections(limit int) ([]models.TagCorrection, error) {
	query := `SELECT id, bookmark_id, original_tags, final_tags, correction_type, created_at 
			  FROM tag_corrections 
			  ORDER BY created_at DESC 
			  LIMIT ?`
	
	rows, err := r.db.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var corrections []models.TagCorrection
	
	for rows.Next() {
		var correction models.TagCorrection
		var originalJSON, finalJSON string
		
		err := rows.Scan(
			&correction.ID,
			&correction.BookmarkID,
			&originalJSON,
			&finalJSON,
			&correction.CorrectionType,
			&correction.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		
		err = json.Unmarshal([]byte(originalJSON), &correction.OriginalTags)
		if err != nil {
			return nil, err
		}
		
		err = json.Unmarshal([]byte(finalJSON), &correction.FinalTags)
		if err != nil {
			return nil, err
		}
		
		corrections = append(corrections, correction)
	}
	
	return corrections, nil
}

// SaveDomainProfile saves domain-specific patterns
func (r *LearningRepository) SaveDomainProfile(profile *models.DomainProfile) error {
	commonTagsJSON, err := json.Marshal(profile.CommonTags)
	if err != nil {
		return err
	}

	ignoredTagsJSON, err := json.Marshal(profile.IgnoredTags)
	if err != nil {
		return err
	}

	customMappingsJSON, err := json.Marshal(profile.CustomMappings)
	if err != nil {
		return err
	}

	query := `
		INSERT OR REPLACE INTO domain_profiles 
		(domain, common_tags, ignored_tags, custom_mappings, bookmark_count, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?)`
	
	now := time.Now()
	
	_, err = r.db.Exec(query,
		profile.Domain,
		string(commonTagsJSON),
		string(ignoredTagsJSON),
		string(customMappingsJSON),
		profile.BookmarkCount,
		now,
		now)
	
	return err
}

// GetDomainProfile gets domain profile by domain name
func (r *LearningRepository) GetDomainProfile(domain string) (*models.DomainProfile, error) {
	query := `SELECT id, domain, common_tags, ignored_tags, custom_mappings, bookmark_count, created_at, updated_at 
			  FROM domain_profiles WHERE domain = ?`
	
	var profile models.DomainProfile
	var commonTagsJSON, ignoredTagsJSON, customMappingsJSON string
	var createdAt, updatedAt time.Time
	
	err := r.db.QueryRow(query, domain).Scan(
		&profile.ID,
		&profile.Domain,
		&commonTagsJSON,
		&ignoredTagsJSON,
		&customMappingsJSON,
		&profile.BookmarkCount,
		&createdAt,
		&updatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	err = json.Unmarshal([]byte(commonTagsJSON), &profile.CommonTags)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal([]byte(ignoredTagsJSON), &profile.IgnoredTags)
	if err != nil {
		return nil, err
	}
	
	err = json.Unmarshal([]byte(customMappingsJSON), &profile.CustomMappings)
	if err != nil {
		return nil, err
	}
	
	profile.CreatedAt = createdAt
	profile.UpdatedAt = updatedAt
	
	return &profile, nil
}

// UpdateDomainProfile updates existing domain profile statistics
func (r *LearningRepository) UpdateDomainProfile(domain string, newTags []string) error {
	// First, get existing profile
	existingProfile, err := r.GetDomainProfile(domain)
	if err != nil {
		return err
	}
	
	if existingProfile == nil {
		// Create new profile
		profile := &models.DomainProfile{
			Domain:         domain,
			CommonTags:     newTags,
			IgnoredTags:    []string{},
			CustomMappings: make(map[string]string),
			BookmarkCount:  1,
		}
		return r.SaveDomainProfile(profile)
	}
	
	// Update existing profile
	existingProfile.BookmarkCount++
	
	// Merge tags (simple approach - could be more sophisticated)
	tagMap := make(map[string]bool)
	for _, tag := range existingProfile.CommonTags {
		tagMap[tag] = true
	}
	for _, tag := range newTags {
		tagMap[tag] = true
	}
	
	// Convert back to slice
	var mergedTags []string
	for tag := range tagMap {
		mergedTags = append(mergedTags, tag)
	}
	existingProfile.CommonTags = mergedTags
	
	return r.SaveDomainProfile(existingProfile)
}