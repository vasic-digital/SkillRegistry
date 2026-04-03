package agents

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

type Loader struct {
	supportedFormats []string
}

func NewLoader() *Loader {
	return &Loader{
		supportedFormats: []string{".yaml", ".yml", ".json"},
	}
}

func (l *Loader) LoadSkillFromFile(path string) (*Skill, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("cannot access skill file: %w", err)
	}

	if info.IsDir() {
		skillFile := filepath.Join(path, "SKILL.md")
		if _, err := os.Stat(skillFile); err == nil {
			return l.loadSkillFromMarkdown(skillFile)
		}
		return nil, fmt.Errorf("directory does not contain SKILL.md: %s", path)
	}

	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md":
		return l.loadSkillFromMarkdown(path)
	case ".yaml", ".yml":
		return l.loadSkillFromYAML(path)
	case ".json":
		return l.loadSkillFromJSON(path)
	default:
		return nil, fmt.Errorf("unsupported skill format: %s", ext)
	}
}

func (l *Loader) LoadSkillsFromDirectory(dir string) ([]*Skill, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read directory: %w", err)
	}

	var skills []*Skill
	var errs []error

	for _, entry := range entries {
		if entry.IsDir() {
			skillDir := filepath.Join(dir, entry.Name())
			skill, err := l.LoadSkillFromFile(skillDir)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to load skill from %s: %w", skillDir, err))
				continue
			}
			skills = append(skills, skill)
		} else {
			ext := strings.ToLower(filepath.Ext(entry.Name()))
			if !l.isSupportedFormat(ext) {
				continue
			}

			skillFile := filepath.Join(dir, entry.Name())
			skill, err := l.LoadSkillFromFile(skillFile)
			if err != nil {
				errs = append(errs, fmt.Errorf("failed to load skill from %s: %w", skillFile, err))
				continue
			}
			skills = append(skills, skill)
		}
	}

	if len(skills) == 0 && len(errs) > 0 {
		return nil, fmt.Errorf("no skills loaded, errors: %v", errs)
	}

	return skills, nil
}

func (l *Loader) loadSkillFromYAML(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read YAML file: %w", err)
	}

	skill := &Skill{}
	if err := yaml.Unmarshal(data, skill); err != nil {
		return nil, fmt.Errorf("cannot parse YAML: %w", err)
	}

	skill.ContentPath = path
	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = skill.UpdatedAt
	}

	if skill.ID == "" {
		skill.ID = generateIDFromName(skill.Name)
	}

	return skill, nil
}

func (l *Loader) loadSkillFromJSON(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read JSON file: %w", err)
	}

	skill := &Skill{}
	if err := json.Unmarshal(data, skill); err != nil {
		return nil, fmt.Errorf("cannot parse JSON: %w", err)
	}

	skill.ContentPath = path
	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = skill.UpdatedAt
	}

	if skill.ID == "" {
		skill.ID = generateIDFromName(skill.Name)
	}

	return skill, nil
}

func (l *Loader) loadSkillFromMarkdown(path string) (*Skill, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read markdown file: %w", err)
	}

	content := string(data)
	
	frontmatter, body, err := extractFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("cannot extract frontmatter: %w", err)
	}

	skill := &Skill{}
	if err := yaml.Unmarshal([]byte(frontmatter), skill); err != nil {
		return nil, fmt.Errorf("cannot parse frontmatter YAML: %w", err)
	}

	skill.ContentPath = path
	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = skill.UpdatedAt
	}

	if skill.ID == "" {
		skill.ID = generateIDFromName(skill.Name)
	}

	if skill.Metadata == nil {
		skill.Metadata = make(map[string]interface{})
	}
	skill.Metadata["markdown_content"] = body

	return skill, nil
}

func (l *Loader) isSupportedFormat(ext string) bool {
	for _, format := range l.supportedFormats {
		if ext == format {
			return true
		}
	}
	return ext == ".md"
}

func extractFrontmatter(content string) (string, string, error) {
	if !strings.HasPrefix(content, "---") {
		return "", content, nil
	}

	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return "", "", fmt.Errorf("invalid frontmatter format")
	}

	frontmatter := strings.TrimSpace(parts[1])
	body := strings.TrimSpace(parts[2])

	return frontmatter, body, nil
}

func generateIDFromName(name string) string {
	id := strings.ToLower(name)
	id = strings.ReplaceAll(id, " ", "-")
	id = strings.ReplaceAll(id, "_", "-")
	id = strings.ReplaceAll(id, ".", "-")
	
	var result strings.Builder
	for _, r := range id {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}

func (l *Loader) ParseSkillYAML(data []byte) (*Skill, error) {
	skill := &Skill{}
	if err := yaml.Unmarshal(data, skill); err != nil {
		return nil, fmt.Errorf("cannot parse YAML: %w", err)
	}

	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = skill.UpdatedAt
	}

	if skill.ID == "" {
		skill.ID = generateIDFromName(skill.Name)
	}

	return skill, nil
}

func (l *Loader) ParseSkillJSON(data []byte) (*Skill, error) {
	skill := &Skill{}
	if err := json.Unmarshal(data, skill); err != nil {
		return nil, fmt.Errorf("cannot parse JSON: %w", err)
	}

	skill.UpdatedAt = time.Now()
	if skill.CreatedAt.IsZero() {
		skill.CreatedAt = skill.UpdatedAt
	}

	if skill.ID == "" {
		skill.ID = generateIDFromName(skill.Name)
	}

	return skill, nil
}

func (l *Loader) LoadSkillsRecursive(rootDir string) ([]*Skill, error) {
	var skills []*Skill

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			if info.Name() == ".git" || info.Name() == "node_modules" || info.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}

		if info.Name() == "SKILL.md" {
			skill, err := l.loadSkillFromMarkdown(path)
			if err != nil {
				return fmt.Errorf("failed to load skill from %s: %w", path, err)
			}
			skills = append(skills, skill)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return skills, nil
}
