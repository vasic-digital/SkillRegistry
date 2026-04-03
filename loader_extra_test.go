package agents

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoader_isSupportedFormat(t *testing.T) {
	loader := NewLoader()

	tests := []struct {
		format   string
		expected bool
	}{
		{".yaml", true},
		{".yml", true},
		{".json", true},
		{".md", true},
		{".txt", false},
		{".go", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := loader.isSupportedFormat(tt.format)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestLoader_LoadSkillsFromDirectory_NoSupportedFiles(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	// Create only unsupported files
	err := os.WriteFile(filepath.Join(tmpDir, "test.txt"), []byte("content"), 0644)
	require.NoError(t, err)

	skills, err := loader.LoadSkillsFromDirectory(tmpDir)

	// Should return empty without error
	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestLoader_LoadSkillsFromDirectory_WithError(t *testing.T) {
	loader := NewLoader()

	// Try to load from non-existent directory
	_, err := loader.LoadSkillsFromDirectory("/nonexistent/directory/path")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot read directory")
}

func TestLoader_LoadSkillFromFile_InvalidYAML(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	invalidYAML := `name: [invalid
  yaml: content`
	yamlPath := filepath.Join(tmpDir, "invalid.yaml")
	err := os.WriteFile(yamlPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	_, err = loader.LoadSkillFromFile(yamlPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse YAML")
}

func TestLoader_LoadSkillFromFile_InvalidJSON(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	invalidJSON := `{"name": "test", "invalid": json}`
	jsonPath := filepath.Join(tmpDir, "invalid.json")
	err := os.WriteFile(jsonPath, []byte(invalidJSON), 0644)
	require.NoError(t, err)

	_, err = loader.LoadSkillFromFile(jsonPath)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse JSON")
}

func TestLoader_LoadSkillFromFile_DirectoryNoSkillMd(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	// Create a directory without SKILL.md
	subDir := filepath.Join(tmpDir, "empty-skill")
	err := os.MkdirAll(subDir, 0755)
	require.NoError(t, err)

	_, err = loader.LoadSkillFromFile(subDir)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "does not contain SKILL.md")
}

func TestExtractFrontmatter_Invalid(t *testing.T) {
	// Test with invalid frontmatter (only one ---)
	content := `---
name: test`

	_, _, err := extractFrontmatter(content)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid frontmatter format")
}

func TestExtractFrontmatter_MultipleDelimiters(t *testing.T) {
	content := `---
name: test
---

# Header
---
Some other content
---`

	frontmatter, body, err := extractFrontmatter(content)

	require.NoError(t, err)
	assert.Equal(t, "name: test", frontmatter)
	assert.Contains(t, body, "# Header")
}

func TestLoader_ParseSkillYAML_Invalid(t *testing.T) {
	loader := NewLoader()

	invalidYAML := []byte(`{invalid yaml`)

	_, err := loader.ParseSkillYAML(invalidYAML)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse YAML")
}

func TestLoader_ParseSkillJSON_Invalid(t *testing.T) {
	loader := NewLoader()

	invalidJSON := []byte(`{invalid json`)

	_, err := loader.ParseSkillJSON(invalidJSON)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot parse JSON")
}

func TestLoader_LoadSkillsRecursive_WithNodeModules(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	// Create node_modules with a skill file that should be ignored
	nodeModules := filepath.Join(tmpDir, "node_modules", "some-package")
	err := os.MkdirAll(nodeModules, 0755)
	require.NoError(t, err)

	mdContent := `---
name: ignored-skill
description: Should be ignored
---
`
	err = os.WriteFile(filepath.Join(nodeModules, "SKILL.md"), []byte(mdContent), 0644)
	require.NoError(t, err)

	// Create a valid skill outside node_modules
	skillDir := filepath.Join(tmpDir, "valid-skill")
	err = os.MkdirAll(skillDir, 0755)
	require.NoError(t, err)
	err = os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(`---
name: valid-skill
description: A valid skill
---
`), 0644)
	require.NoError(t, err)

	skills, err := loader.LoadSkillsRecursive(tmpDir)

	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "valid-skill", skills[0].Name)
}

func TestLoader_LoadSkillsRecursive_WithVendor(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	// Create vendor directory with a skill file that should be ignored
	vendorDir := filepath.Join(tmpDir, "vendor", "some-package")
	err := os.MkdirAll(vendorDir, 0755)
	require.NoError(t, err)

	mdContent := `---
name: vendor-skill
description: Should be ignored
---
`
	err = os.WriteFile(filepath.Join(vendorDir, "SKILL.md"), []byte(mdContent), 0644)
	require.NoError(t, err)

	skills, err := loader.LoadSkillsRecursive(tmpDir)

	require.NoError(t, err)
	assert.Empty(t, skills)
}

func TestLoader_LoadSkillsRecursive_ErrorWalking(t *testing.T) {
	loader := NewLoader()

	// Try to walk a non-existent directory
	_, err := loader.LoadSkillsRecursive("/nonexistent/path/that/does/not/exist")

	assert.Error(t, err)
}

func TestLoader_LoadSkillFromFile_ReadError(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	// Create a file, then remove read permissions
	yamlPath := filepath.Join(tmpDir, "unreadable.yaml")
	err := os.WriteFile(yamlPath, []byte("name: test"), 0000)
	require.NoError(t, err)
	defer os.Chmod(yamlPath, 0644) // Restore permissions for cleanup

	_, err = loader.LoadSkillFromFile(yamlPath)
	assert.Error(t, err)
}

func TestLoader_LoadSkillsFromDirectory_PartialSuccess(t *testing.T) {
	loader := NewLoader()
	tmpDir := t.TempDir()

	// Create one valid skill - the function will load valid ones and skip errors
	validYAML := `name: valid-skill
description: A valid skill that has a long enough description`

	err := os.WriteFile(filepath.Join(tmpDir, "valid.yaml"), []byte(validYAML), 0644)
	require.NoError(t, err)

	// The function should succeed and return the valid skill
	skills, err := loader.LoadSkillsFromDirectory(tmpDir)
	require.NoError(t, err)
	assert.Len(t, skills, 1)
	assert.Equal(t, "valid-skill", skills[0].Name)
}
