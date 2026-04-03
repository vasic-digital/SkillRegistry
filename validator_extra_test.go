package agents

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSkillValidator_validateTriggers_WithInvalid(t *testing.T) {
	validator := NewSkillValidator()

	// Test with valid triggers
	err := validator.validateTriggers([]string{"trigger1", "trigger2"})
	assert.NoError(t, err)

	// Test with empty trigger
	err = validator.validateTriggers([]string{"trigger1", ""})
	assert.Error(t, err)
}

func TestSkillValidator_validateTriggers_LongTrigger(t *testing.T) {
	validator := NewSkillValidator()

	// Test with trigger longer than 100 characters
	longTrigger := string(make([]byte, 101))
	err := validator.validateTriggers([]string{longTrigger})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestSkillValidator_validateTags_LongTag(t *testing.T) {
	validator := NewSkillValidator()

	// Test with tag longer than 50 characters
	longTag := string(make([]byte, 51))
	err := validator.validateTags([]string{longTag})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum length")
}

func TestSkillValidator_validateDefinition_Nil(t *testing.T) {
	validator := NewSkillValidator()

	// Test with nil definition
	err := validator.validateDefinition(nil)
	assert.NoError(t, err)
}

func TestSkillValidator_validateDefinition_WithTimeout(t *testing.T) {
	validator := NewSkillValidator()

	def := &SkillDefinition{
		Timeout: 30 * time.Second,
	}

	err := validator.validateDefinition(def)
	assert.NoError(t, err)
}

func TestSkillValidator_validateDescription_Short(t *testing.T) {
	validator := NewSkillValidator()

	skill := &Skill{
		ID:          "test-id",
		Name:        "Test Name",
		Description: "Short",
	}

	err := validator.ValidateSkill(skill)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description must be between")
}

func TestSkillValidator_validateDescription_Long(t *testing.T) {
	validator := NewSkillValidator()

	// Create a description longer than 5000 characters
	longDesc := string(make([]byte, 5001))

	skill := &Skill{
		ID:          "test-id",
		Name:        "Test Name",
		Description: longDesc,
	}

	err := validator.ValidateSkill(skill)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "description must be between")
}

func TestSkillValidator_validateID_TooLong(t *testing.T) {
	validator := NewSkillValidator()

	// Create an ID longer than 100 characters
	longID := string(make([]byte, 101))

	skill := &Skill{
		ID:          longID,
		Name:        "Test Name",
		Description: "A valid description that is long enough",
	}

	err := validator.ValidateSkill(skill)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID must be between")
}

func TestSkillValidator_validateID_InvalidCharacters(t *testing.T) {
	validator := NewSkillValidator()

	skill := &Skill{
		ID:          "test@id!",
		Name:        "Test Name",
		Description: "A valid description that is long enough",
	}

	err := validator.ValidateSkill(skill)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "ID must contain only")
}

func TestSkillValidator_validateID_StartsWithHyphen(t *testing.T) {
	validator := NewSkillValidator()

	skill := &Skill{
		ID:          "-test-id",
		Name:        "Test Name",
		Description: "A valid description that is long enough",
	}

	err := validator.ValidateSkill(skill)
	assert.Error(t, err)
}

func TestSkillValidator_validateID_EndsWithHyphen(t *testing.T) {
	validator := NewSkillValidator()

	skill := &Skill{
		ID:          "test-id-",
		Name:        "Test Name",
		Description: "A valid description that is long enough",
	}

	err := validator.ValidateSkill(skill)
	assert.Error(t, err)
}

func TestSkillValidator_validateName_TooLong(t *testing.T) {
	validator := NewSkillValidator()

	// Create a name longer than 200 characters
	longName := string(make([]byte, 201))

	skill := &Skill{
		ID:          "test-id",
		Name:        longName,
		Description: "A valid description that is long enough",
	}

	err := validator.ValidateSkill(skill)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name must be between")
}

func TestSkillValidator_validateCategory_Valid(t *testing.T) {
	validator := NewSkillValidator()

	validCategories := []SkillCategory{
		SkillCategoryCode,
		SkillCategoryData,
		SkillCategoryDevOps,
		SkillCategoryTesting,
		SkillCategorySecurity,
		SkillCategoryMonitoring,
		SkillCategoryGeneral,
	}

	for _, category := range validCategories {
		skill := &Skill{
			ID:          "test-id",
			Name:        "Test Name",
			Description: "A valid description that is long enough",
			Category:    category,
		}

		err := validator.ValidateSkill(skill)
		assert.NoError(t, err)
	}
}

func TestDependencyResolver_DetectCycle_WithNestedDependencies(t *testing.T) {
	dr := NewDependencyResolver()

	skills := map[string]*Skill{
		"skill-a": {
			ID: "skill-a",
			Definition: &SkillDefinition{
				Dependencies: []string{"skill-b"},
			},
		},
		"skill-b": {
			ID: "skill-b",
			Definition: &SkillDefinition{
				Dependencies: []string{"skill-c"},
			},
		},
		"skill-c": {
			ID:         "skill-c",
			Definition: &SkillDefinition{
				Dependencies: []string{},
			},
		},
	}

	err := dr.DetectCycle("skill-a", []string{"skill-b"}, skills)
	assert.NoError(t, err)
}
