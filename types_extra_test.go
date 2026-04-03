package agents

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSkill_WithMetadata(t *testing.T) {
	skill := &Skill{
		ID:          "metadata-skill",
		Name:        "Metadata Skill",
		Description: "A skill with metadata",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	assert.Equal(t, "value1", skill.Metadata["key1"])
	assert.Equal(t, 42, skill.Metadata["key2"])
}

func TestSkill_HasTrigger_NotFound(t *testing.T) {
	skill := &Skill{
		Triggers: []string{"trigger1", "trigger2"},
	}

	assert.False(t, skill.HasTrigger("nonexistent"))
}

func TestSkill_HasTag_NotFound(t *testing.T) {
	skill := &Skill{
		Tags: []string{"tag1", "tag2"},
	}

	assert.False(t, skill.HasTag("nonexistent"))
}

func TestSkill_HasTrigger_EmptyTriggers(t *testing.T) {
	skill := &Skill{
		Triggers: []string{},
	}

	assert.False(t, skill.HasTrigger("any"))
}

func TestSkill_HasTag_EmptyTags(t *testing.T) {
	skill := &Skill{
		Tags: []string{},
	}

	assert.False(t, skill.HasTag("any"))
}

func TestSkillFilter_Matches_MultipleTags(t *testing.T) {
	skill := &Skill{
		ID:       "multi-tag-skill",
		Name:     "Multi Tag Skill",
		Category: SkillCategoryCode,
		Tags:     []string{"tag1", "tag2", "tag3"},
	}

	filter := &SkillFilter{
		Tags: []string{"tag1", "tag3"},
	}

	assert.True(t, filter.Matches(skill))
}

func TestSkillFilter_Matches_PartialTagMatch(t *testing.T) {
	skill := &Skill{
		ID:   "partial-tag-skill",
		Name: "Partial Tag Skill",
		Tags: []string{"tag1"},
	}

	filter := &SkillFilter{
		Tags: []string{"tag1", "tag2"},
	}

	// Should pass because skill has tag1 (OR logic - any match is enough)
	assert.True(t, filter.Matches(skill))
}

func TestSkillFilter_Matches_SearchQueryName(t *testing.T) {
	skill := &Skill{
		ID:          "searchable-skill",
		Name:        "Searchable Skill Name",
		Description: "Some description",
	}

	filter := &SkillFilter{
		SearchQuery: "Skill Name",
	}

	assert.True(t, filter.Matches(skill))
}

func TestSkillFilter_Matches_SearchQueryDescription(t *testing.T) {
	skill := &Skill{
		ID:          "desc-searchable",
		Name:        "My Skill",
		Description: "This skill has a unique description",
	}

	filter := &SkillFilter{
		SearchQuery: "unique",
	}

	assert.True(t, filter.Matches(skill))
}

func TestSkillResult_MetadataOperations(t *testing.T) {
	result := NewSkillResult("exec-123", "skill-456")

	// Add some metadata
	result.Metadata["key1"] = "value1"
	result.Metadata["key2"] = 42

	assert.Equal(t, "value1", result.Metadata["key1"])
	assert.Equal(t, 42, result.Metadata["key2"])
}

func TestSkillExecutionContext_WithTimeout(t *testing.T) {
	ctx := NewSkillExecutionContext("test-skill")
	ctx.Timeout = 5 * time.Minute

	assert.Equal(t, "test-skill", ctx.SkillID)
	assert.Equal(t, 5*time.Minute, ctx.Timeout)
	assert.NotEmpty(t, ctx.ExecutionID)
}

func TestSkillExecutionContext_WithEnvironment(t *testing.T) {
	ctx := NewSkillExecutionContext("test-skill")
	ctx.Environment["VAR1"] = "value1"
	ctx.Environment["VAR2"] = "value2"

	assert.Equal(t, "value1", ctx.Environment["VAR1"])
	assert.Equal(t, "value2", ctx.Environment["VAR2"])
}

func TestSkillDefinition_WithExamples(t *testing.T) {
	def := &SkillDefinition{
		Parameters: []SkillParameter{
			{Name: "input", Type: "string", Required: true},
		},
		Examples: []SkillExample{
			{
				Name:        "basic",
				Description: "Basic example",
				Input:       map[string]interface{}{"input": "hello"},
				Output:      "result",
			},
			{
				Name:        "advanced",
				Description: "Advanced example",
				Input:       map[string]interface{}{"input": "world"},
				Output:      map[string]string{"result": "done"},
			},
		},
	}

	assert.Len(t, def.Examples, 2)
	assert.Equal(t, "basic", def.Examples[0].Name)
	assert.Equal(t, "advanced", def.Examples[1].Name)
}

func TestSkillDefinition_WithConfig(t *testing.T) {
	def := &SkillDefinition{
		Config: map[string]interface{}{
			"timeout":     30,
			"retries":     3,
			"debug_mode":  true,
		},
	}

	assert.Equal(t, 30, def.Config["timeout"])
	assert.Equal(t, 3, def.Config["retries"])
	assert.Equal(t, true, def.Config["debug_mode"])
}

func TestSkillParameter_WithDefault(t *testing.T) {
	param := SkillParameter{
		Name:        "optional_param",
		Type:        "string",
		Description: "An optional parameter",
		Required:    false,
		Default:     "default_value",
	}

	assert.Equal(t, "default_value", param.Default)
	assert.False(t, param.Required)
}

func TestSkillParameter_WithValidation(t *testing.T) {
	param := SkillParameter{
		Name:        "validated_param",
		Type:        "string",
		Description: "A validated parameter",
		Required:    true,
		Validation:  "^email@example.com$",
	}

	assert.Equal(t, "^email@example.com$", param.Validation)
}

func TestSkillMetrics_Update(t *testing.T) {
	now := time.Now()
	metrics := &SkillMetrics{
		SkillID:          "test-skill",
		TotalExecutions:  10,
		SuccessfulRuns:   8,
		FailedRuns:       2,
		AverageDuration:  100 * time.Millisecond,
		LastExecutedAt:   &now,
		LastError:        "some error",
		UsageCount30Days: 5,
	}

	assert.Equal(t, int64(10), metrics.TotalExecutions)
	assert.Equal(t, int64(8), metrics.SuccessfulRuns)
	assert.Equal(t, int64(2), metrics.FailedRuns)
	assert.Equal(t, "some error", metrics.LastError)
}

func TestContains_EmptySubstring(t *testing.T) {
	// Empty substring should be found in any string
	assert.True(t, contains("hello", ""))
	assert.True(t, contains("", ""))
}

func TestContains_Match(t *testing.T) {
	assert.True(t, contains("hello world", "world"))
	assert.True(t, contains("hello world", "hello"))
	assert.True(t, contains("hello world", "lo wo"))
}

func TestContains_NoMatch(t *testing.T) {
	assert.False(t, contains("hello world", "foo"))
	assert.False(t, contains("hello", "hello world"))
	assert.False(t, contains("", "hello"))
}

func TestRandomString(t *testing.T) {
	str1 := randomString(8)
	str2 := randomString(8)

	assert.Len(t, str1, 8)
	assert.Len(t, str2, 8)
	// Two random strings should be different (with very high probability)
	assert.NotEqual(t, str1, str2)
}

func TestGenerateExecutionID(t *testing.T) {
	id1 := generateExecutionID()
	id2 := generateExecutionID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	// Should contain a timestamp prefix and a random suffix
	assert.Contains(t, id1, "-")
}
