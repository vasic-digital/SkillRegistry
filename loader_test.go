package agents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLoader(t *testing.T) {
	loader := NewLoader()
	assert.NotNil(t, loader)
	assert.Equal(t, []string{".yaml", ".yml", ".json"}, loader.supportedFormats)
}

func TestLoader_LoadSkillFromFile_YAML(t *testing.T) {
	tmpDir := t.TempDir()
	yamlContent := `
name: test-skill
description: A test skill for loading
version: 1.0.0
category: code
tags:
  - test
  - yaml
author: Test Author
triggers:
  - /test
`
	yamlPath := filepath.Join(tmpDir, "test-skill.yaml")
	err := os.WriteFile(yamlPath, []byte(yamlContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	skill, err := loader.LoadSkillFromFile(yamlPath)

	require.NoError(t, err)
	assert.Equal(t, "test-skill", skill.ID)
	assert.Equal(t, "test-skill", skill.Name)
	assert.Equal(t, "A test skill for loading", skill.Description)
	assert.Equal(t, "1.0.0", skill.Version)
	assert.Equal(t, SkillCategoryCode, skill.Category)
	assert.Equal(t, []string{"test", "yaml"}, skill.Tags)
	assert.Equal(t, "Test Author", skill.Author)
	assert.Equal(t, []string{"/test"}, skill.Triggers)
}

func TestLoader_LoadSkillFromFile_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	jsonContent := `{
  "name": "json-skill",
  "description": "A JSON skill",
  "version": "2.0.0",
  "category": "data",
  "tags": ["json", "test"]
}`

	jsonPath := filepath.Join(tmpDir, "json-skill.json")
	err := os.WriteFile(jsonPath, []byte(jsonContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	skill, err := loader.LoadSkillFromFile(jsonPath)

	require.NoError(t, err)
	assert.Equal(t, "json-skill", skill.ID)
	assert.Equal(t, "json-skill", skill.Name)
	assert.Equal(t, "A JSON skill", skill.Description)
	assert.Equal(t, "2.0.0", skill.Version)
	assert.Equal(t, SkillCategoryData, skill.Category)
}

func TestLoader_LoadSkillFromFile_Markdown(t *testing.T) {
	tmpDir := t.TempDir()
	mdContent := `---
name: markdown-skill
description: A markdown skill with frontmatter
triggers:
  - /md-test
tags:
  - markdown
---

# Markdown Skill

This is the skill documentation.

## Usage

Some usage instructions here.
`

	mdPath := filepath.Join(tmpDir, "SKILL.md")
	err := os.WriteFile(mdPath, []byte(mdContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	skill, err := loader.LoadSkillFromFile(mdPath)

	require.NoError(t, err)
	assert.Equal(t, "markdown-skill", skill.ID)
	assert.Equal(t, "markdown-skill", skill.Name)
	assert.Equal(t, "A markdown skill with frontmatter", skill.Description)
	assert.Equal(t, []string{"/md-test"}, skill.Triggers)
	assert.Equal(t, []string{"markdown"}, skill.Tags)
	assert.NotNil(t, skill.Metadata)
	assert.Contains(t, skill.Metadata, "markdown_content")
}

func TestLoader_LoadSkillFromFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	skillDir := filepath.Join(tmpDir, "my-skill")
	err := os.MkdirAll(skillDir, 0755)
	require.NoError(t, err)

	mdContent := `---
name: dir-skill
description: A skill in a directory
---

# Directory Skill
`

	mdPath := filepath.Join(skillDir, "SKILL.md")
	err = os.WriteFile(mdPath, []byte(mdContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	skill, err := loader.LoadSkillFromFile(skillDir)

	require.NoError(t, err)
	assert.Equal(t, "dir-skill", skill.ID)
	assert.Equal(t, "dir-skill", skill.Name)
}

func TestLoader_LoadSkillFromFile_NotFound(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadSkillFromFile("/nonexistent/file.yaml")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot access skill file")
}

func TestLoader_LoadSkillFromFile_UnsupportedFormat(t *testing.T) {
	tmpDir := t.TempDir()
	txtPath := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(txtPath, []byte("content"), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	_, err = loader.LoadSkillFromFile(txtPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported skill format")
}

func TestLoader_LoadSkillsFromDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	yaml1 := `name: skill-one
description: First skill`
	yaml2 := `name: skill-two
description: Second skill`

	err := os.WriteFile(filepath.Join(tmpDir, "skill1.yaml"), []byte(yaml1), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(tmpDir, "skill2.yaml"), []byte(yaml2), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	skills, err := loader.LoadSkillsFromDirectory(tmpDir)

	require.NoError(t, err)
	assert.Len(t, skills, 2)

	names := make([]string, len(skills))
	for i, s := range skills {
		names[i] = s.Name
	}
	assert.Contains(t, names, "skill-one")
	assert.Contains(t, names, "skill-two")
}

func TestLoader_LoadSkillsFromDirectory_Subdirectories(t *testing.T) {
	tmpDir := t.TempDir()

	skillDir := filepath.Join(tmpDir, "skill-in-dir")
	err := os.MkdirAll(skillDir, 0755)
	require.NoError(t, err)

	mdContent := `---
name: nested-skill
description: A nested skill
---
`
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(mdContent), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	skills, err := loader.LoadSkillsFromDirectory(tmpDir)

	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "nested-skill", skills[0].Name)
}

func TestLoader_LoadSkillsFromDirectory_NotFound(t *testing.T) {
	loader := NewLoader()
	_, err := loader.LoadSkillsFromDirectory("/nonexistent/directory")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read directory")
}

func TestLoader_ParseSkillYAML(t *testing.T) {
	yamlContent := []byte(`
name: parsed-skill
description: A parsed skill
category: testing
version: 1.0.0
`)

	loader := NewLoader()
	skill, err := loader.ParseSkillYAML(yamlContent)

	require.NoError(t, err)
	assert.Equal(t, "parsed-skill", skill.ID)
	assert.Equal(t, "A parsed skill", skill.Description)
	assert.Equal(t, SkillCategoryTesting, skill.Category)
}

func TestLoader_ParseSkillJSON(t *testing.T) {
	jsonContent := []byte(`{
  "name": "parsed-json-skill",
  "description": "A parsed JSON skill",
  "category": "security"
}`)

	loader := NewLoader()
	skill, err := loader.ParseSkillJSON(jsonContent)

	require.NoError(t, err)
	assert.Equal(t, "parsed-json-skill", skill.ID)
	assert.Equal(t, "A parsed JSON skill", skill.Description)
	assert.Equal(t, SkillCategorySecurity, skill.Category)
}

func TestLoader_LoadSkillsRecursive(t *testing.T) {
	tmpDir := t.TempDir()

	subDir1 := filepath.Join(tmpDir, "subdir1")
	subDir2 := filepath.Join(tmpDir, "subdir2")
	err := os.MkdirAll(subDir1, 0755)
	require.NoError(t, err)
	err = os.MkdirAll(subDir2, 0755)
	require.NoError(t, err)

	md1 := `---
name: recursive-skill-1
description: First recursive skill
---
`
	md2 := `---
name: recursive-skill-2
description: Second recursive skill
---
`

	err = os.WriteFile(filepath.Join(subDir1, "SKILL.md"), []byte(md1), 0644)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(subDir2, "SKILL.md"), []byte(md2), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	skills, err := loader.LoadSkillsRecursive(tmpDir)

	require.NoError(t, err)
	assert.Len(t, skills, 2)
}

func TestLoader_LoadSkillsRecursive_SkipsIgnoredDirs(t *testing.T) {
	tmpDir := t.TempDir()

	gitDir := filepath.Join(tmpDir, ".git")
	err := os.MkdirAll(gitDir, 0755)
	require.NoError(t, err)

	md := `---
name: git-skill
description: Should not be loaded
---
`
	err = os.WriteFile(filepath.Join(gitDir, "SKILL.md"), []byte(md), 0644)
	require.NoError(t, err)

	loader := NewLoader()
	skills, err := loader.LoadSkillsRecursive(tmpDir)

	require.NoError(t, err)
	assert.Len(t, skills, 0)
}

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name              string
		content           string
		wantFrontmatter   string
		wantBody          string
		wantErr           bool
	}{
		{
			name: "valid frontmatter",
			content: `---
name: test
---

# Body`,
			wantFrontmatter: "name: test",
			wantBody:        "# Body",
			wantErr:         false,
		},
		{
			name:            "no frontmatter",
			content:         "# Just body",
			wantFrontmatter: "",
			wantBody:        "# Just body",
			wantErr:         false,
		},
		{
			name: "empty content",
			content: `---
---`,
			wantFrontmatter: "",
			wantBody:        "",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frontmatter, body, err := extractFrontmatter(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantFrontmatter, frontmatter)
				assert.Equal(t, tt.wantBody, body)
			}
		})
	}
}

func TestGenerateIDFromName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"simple", "My Skill", "my-skill"},
		{"with underscores", "my_skill_name", "my-skill-name"},
		{"lowercase", "alreadylowercase", "alreadylowercase"},
		{"with dots", "my.skill.name", "my-skill-name"},
		{"mixed case", "MyAwesomeSkill", "myawesomeskill"},
		{"with numbers", "Skill123Test", "skill123test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateIDFromName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
