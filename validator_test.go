package agents

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSkillValidator(t *testing.T) {
	validator := NewSkillValidator()
	assert.NotNil(t, validator)
	assert.NotNil(t, validator.dependencyResolver)
}

func TestNewDependencyResolver(t *testing.T) {
	dr := NewDependencyResolver()
	assert.NotNil(t, dr)
	assert.NotNil(t, dr.visited)
	assert.NotNil(t, dr.stack)
}

func TestSkillValidator_ValidateSkill(t *testing.T) {
	validator := NewSkillValidator()

	tests := []struct {
		name    string
		skill   *Skill
		wantErr bool
		errType error
	}{
		{
			name: "valid skill",
			skill: &Skill{
				ID:          "valid-skill",
				Name:        "Valid Skill",
				Description: "A valid skill for testing purposes",
				Version:     "1.0.0",
				Category:    SkillCategoryCode,
			},
			wantErr: false,
		},
		{
			name:    "nil skill",
			skill:   nil,
			wantErr: true,
			errType: ErrSkillInvalid,
		},
		{
			name: "missing ID",
			skill: &Skill{
				ID:          "",
				Name:        "Skill Name",
				Description: "Description here",
			},
			wantErr: true,
			errType: ErrSkillInvalid,
		},
		{
			name: "missing name",
			skill: &Skill{
				ID:          "skill-id",
				Name:        "",
				Description: "Description here",
			},
			wantErr: true,
			errType: ErrSkillInvalid,
		},
		{
			name: "missing description",
			skill: &Skill{
				ID:          "skill-id",
				Name:        "Skill Name",
				Description: "",
			},
			wantErr: true,
			errType: ErrSkillInvalid,
		},
		{
			name: "invalid ID with spaces",
			skill: &Skill{
				ID:          "invalid id",
				Name:        "Skill Name",
				Description: "Description here that is long enough",
			},
			wantErr: true,
			errType: ErrSkillInvalid,
		},
		{
			name: "invalid ID uppercase",
			skill: &Skill{
				ID:          "Invalid-ID",
				Name:        "Skill Name",
				Description: "Description here that is long enough",
			},
			wantErr: true,
			errType: ErrSkillInvalid,
		},
		{
			name: "invalid version",
			skill: &Skill{
				ID:          "skill-id",
				Name:        "Skill Name",
				Description: "Description here that is long enough",
				Version:     "not-semver",
			},
			wantErr: true,
			errType: ErrSkillInvalid,
		},
		{
			name: "invalid category",
			skill: &Skill{
				ID:          "skill-id",
				Name:        "Skill Name",
				Description: "Description here that is long enough",
				Category:    "invalid-category",
			},
			wantErr: true,
			errType: ErrSkillInvalid,
		},
		{
			name: "empty trigger",
			skill: &Skill{
				ID:          "skill-id",
				Name:        "Skill Name",
				Description: "Description here that is long enough",
				Triggers:    []string{""},
			},
			wantErr: true,
			errType: ErrSkillInvalid,
		},
		{
			name: "empty tag",
			skill: &Skill{
				ID:          "skill-id",
				Name:        "Skill Name",
				Description: "Description here that is long enough",
				Tags:        []string{""},
			},
			wantErr: true,
			errType: ErrSkillInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSkill(tt.skill)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSkillValidator_ValidateSkillDependencies(t *testing.T) {
	validator := NewSkillValidator()

	availableSkills := map[string]*Skill{
		"skill-a": {
			ID:   "skill-a",
			Name: "Skill A",
			Definition: &SkillDefinition{
				Dependencies: []string{},
			},
		},
		"skill-b": {
			ID:   "skill-b",
			Name: "Skill B",
			Definition: &SkillDefinition{
				Dependencies: []string{"skill-a"},
			},
		},
		"skill-c": {
			ID:   "skill-c",
			Name: "Skill C",
			Definition: &SkillDefinition{
				Dependencies: []string{"skill-b"},
			},
		},
	}

	tests := []struct {
		name    string
		skill   *Skill
		wantErr bool
		errType error
	}{
		{
			name: "no dependencies",
			skill: &Skill{
				ID:   "no-deps",
				Name: "No Dependencies",
			},
			wantErr: false,
		},
		{
			name: "valid dependency",
			skill: &Skill{
				ID:   "with-dep",
				Name: "With Dependency",
				Definition: &SkillDefinition{
					Dependencies: []string{"skill-a"},
				},
			},
			wantErr: false,
		},
		{
			name: "missing dependency",
			skill: &Skill{
				ID:   "missing-dep",
				Name: "Missing Dependency",
				Definition: &SkillDefinition{
					Dependencies: []string{"nonexistent"},
				},
			},
			wantErr: true,
			errType: ErrDependencyNotFound,
		},
		{
			name: "indirect dependency",
			skill: &Skill{
				ID:   "indirect",
				Name: "Indirect Dependency",
				Definition: &SkillDefinition{
					Dependencies: []string{"skill-c"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSkillDependencies(tt.skill, availableSkills)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDependencyResolver_DetectCycle(t *testing.T) {
	dr := NewDependencyResolver()

	tests := []struct {
		name         string
		skillID      string
		dependencies []string
		skills       map[string]*Skill
		wantErr      bool
		errType      error
	}{
		{
			name:         "no cycle - self",
			skillID:      "skill-a",
			dependencies: []string{},
			skills: map[string]*Skill{
				"skill-a": {ID: "skill-a"},
			},
			wantErr: false,
		},
		{
			name:         "simple cycle",
			skillID:      "skill-a",
			dependencies: []string{"skill-b"},
			skills: map[string]*Skill{
				"skill-a": {
					ID: "skill-a",
					Definition: &SkillDefinition{
						Dependencies: []string{"skill-b"},
					},
				},
				"skill-b": {
					ID: "skill-b",
					Definition: &SkillDefinition{
						Dependencies: []string{"skill-a"},
					},
				},
			},
			wantErr: true,
			errType: ErrCircularDependency,
		},
		{
			name:         "indirect cycle",
			skillID:      "skill-a",
			dependencies: []string{"skill-b"},
			skills: map[string]*Skill{
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
					ID: "skill-c",
					Definition: &SkillDefinition{
						Dependencies: []string{"skill-a"},
					},
				},
			},
			wantErr: true,
			errType: ErrCircularDependency,
		},
		{
			name:         "no cycle - linear",
			skillID:      "skill-a",
			dependencies: []string{"skill-b"},
			skills: map[string]*Skill{
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
					ID: "skill-c",
					Definition: &SkillDefinition{
						Dependencies: []string{},
					},
				},
			},
			wantErr: false,
		},
		{
			name:         "missing dependency",
			skillID:      "skill-a",
			dependencies: []string{"nonexistent"},
			skills: map[string]*Skill{
				"skill-a": {
					ID: "skill-a",
					Definition: &SkillDefinition{
						Dependencies: []string{"nonexistent"},
					},
				},
			},
			wantErr: true,
			errType: ErrDependencyNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dr.Reset()
			err := dr.DetectCycle(tt.skillID, tt.dependencies, tt.skills)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDependencyResolver_Reset(t *testing.T) {
	dr := NewDependencyResolver()
	dr.visited["skill-a"] = true
	dr.stack["skill-a"] = true

	dr.Reset()

	assert.Empty(t, dr.visited)
	assert.Empty(t, dr.stack)
}

func TestSkillValidator_ValidateBatch(t *testing.T) {
	validator := NewSkillValidator()

	tests := []struct {
		name    string
		skills  []*Skill
		wantErr bool
	}{
		{
			name: "all valid",
			skills: []*Skill{
				{ID: "skill-1", Name: "Skill 1", Description: "Description here that is long enough"},
				{ID: "skill-2", Name: "Skill 2", Description: "Description here that is long enough"},
			},
			wantErr: false,
		},
		{
			name: "one invalid",
			skills: []*Skill{
				{ID: "skill-1", Name: "Skill 1", Description: "Description here that is long enough"},
				{ID: "", Name: "Skill 2", Description: "Description here that is long enough"},
			},
			wantErr: true,
		},
		{
			name:    "empty batch",
			skills:  []*Skill{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBatch(tt.skills)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSkillValidator_validateVersion(t *testing.T) {
	validator := NewSkillValidator()

	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"valid semver", "1.0.0", false},
		{"valid with prerelease", "1.0.0-alpha", false},
		{"valid with build", "1.0.0+build123", false},
		{"valid complex", "1.2.3-alpha.1+build.123", false},
		{"empty version", "", false},
		{"invalid format", "v1.0", true},
		{"no version", "version", true},
		{"just number", "1", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateVersion(tt.version)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSkillValidator_validateParameters(t *testing.T) {
	validator := NewSkillValidator()

	tests := []struct {
		name    string
		params  []SkillParameter
		wantErr bool
	}{
		{
			name:    "empty params",
			params:  []SkillParameter{},
			wantErr: false,
		},
		{
			name: "valid param",
			params: []SkillParameter{
				{Name: "param1", Type: "string", Description: "A parameter"},
			},
			wantErr: false,
		},
		{
			name: "empty name",
			params: []SkillParameter{
				{Name: "", Type: "string"},
			},
			wantErr: true,
		},
		{
			name: "duplicate name",
			params: []SkillParameter{
				{Name: "param1", Type: "string"},
				{Name: "param1", Type: "number"},
			},
			wantErr: true,
		},
		{
			name: "invalid type",
			params: []SkillParameter{
				{Name: "param1", Type: "invalid-type"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateParameters(tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSkillValidator_validateTimeout(t *testing.T) {
	validator := NewSkillValidator()

	err := validator.validateTimeout(0)
	assert.NoError(t, err)

	err = validator.validateTimeout(100)
	assert.NoError(t, err)
}
