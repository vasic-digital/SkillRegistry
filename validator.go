package agents

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

type SkillValidator struct {
	dependencyResolver *DependencyResolver
}

type DependencyResolver struct {
	visited map[string]bool
	stack   map[string]bool
}

func NewSkillValidator() *SkillValidator {
	return &SkillValidator{
		dependencyResolver: NewDependencyResolver(),
	}
}

func NewDependencyResolver() *DependencyResolver {
	return &DependencyResolver{
		visited: make(map[string]bool),
		stack:   make(map[string]bool),
	}
}

func (sv *SkillValidator) ValidateSkill(skill *Skill) error {
	if skill == nil {
		return fmt.Errorf("%w: skill is nil", ErrSkillInvalid)
	}

	if err := sv.validateRequiredFields(skill); err != nil {
		return err
	}

	if err := sv.validateID(skill.ID); err != nil {
		return err
	}

	if err := sv.validateName(skill.Name); err != nil {
		return err
	}

	if err := sv.validateDescription(skill.Description); err != nil {
		return err
	}

	if err := sv.validateVersion(skill.Version); err != nil {
		return err
	}

	if err := sv.validateCategory(skill.Category); err != nil {
		return err
	}

	if err := sv.validateTriggers(skill.Triggers); err != nil {
		return err
	}

	if err := sv.validateTags(skill.Tags); err != nil {
		return err
	}

	if skill.Definition != nil {
		if err := sv.validateDefinition(skill.Definition); err != nil {
			return err
		}
	}

	return nil
}

func (sv *SkillValidator) validateRequiredFields(skill *Skill) error {
	if strings.TrimSpace(skill.ID) == "" {
		return fmt.Errorf("%w: skill ID is required", ErrSkillInvalid)
	}

	if strings.TrimSpace(skill.Name) == "" {
		return fmt.Errorf("%w: skill name is required", ErrSkillInvalid)
	}

	if strings.TrimSpace(skill.Description) == "" {
		return fmt.Errorf("%w: skill description is required", ErrSkillInvalid)
	}

	return nil
}

func (sv *SkillValidator) validateID(id string) error {
	if len(id) < 1 || len(id) > 100 {
		return fmt.Errorf("%w: skill ID must be between 1 and 100 characters", ErrSkillInvalid)
	}

	validID := regexp.MustCompile(`^[a-z0-9]([a-z0-9-_]*[a-z0-9])?$`)
	if !validID.MatchString(id) {
		return fmt.Errorf("%w: skill ID must contain only lowercase letters, numbers, hyphens, and underscores, and must start and end with alphanumeric characters", ErrSkillInvalid)
	}

	return nil
}

func (sv *SkillValidator) validateName(name string) error {
	if len(name) < 1 || len(name) > 200 {
		return fmt.Errorf("%w: skill name must be between 1 and 200 characters", ErrSkillInvalid)
	}

	return nil
}

func (sv *SkillValidator) validateDescription(desc string) error {
	if len(desc) < 10 || len(desc) > 5000 {
		return fmt.Errorf("%w: skill description must be between 10 and 5000 characters", ErrSkillInvalid)
	}

	return nil
}

func (sv *SkillValidator) validateVersion(version string) error {
	if version == "" {
		return nil
	}

	semverPattern := regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)
	if !semverPattern.MatchString(version) {
		return fmt.Errorf("%w: version must follow semantic versioning (e.g., 1.0.0)", ErrSkillInvalid)
	}

	return nil
}

func (sv *SkillValidator) validateCategory(category SkillCategory) error {
	if category == "" {
		return nil
	}

	validCategories := []SkillCategory{
		SkillCategoryCode,
		SkillCategoryData,
		SkillCategoryDevOps,
		SkillCategoryTesting,
		SkillCategorySecurity,
		SkillCategoryMonitoring,
		SkillCategoryGeneral,
	}

	for _, valid := range validCategories {
		if category == valid {
			return nil
		}
	}

	return fmt.Errorf("%w: invalid category '%s'", ErrSkillInvalid, category)
}

func (sv *SkillValidator) validateTriggers(triggers []string) error {
	for _, trigger := range triggers {
		if strings.TrimSpace(trigger) == "" {
			return fmt.Errorf("%w: trigger cannot be empty", ErrSkillInvalid)
		}
		if len(trigger) > 100 {
			return fmt.Errorf("%w: trigger exceeds maximum length of 100 characters", ErrSkillInvalid)
		}
	}

	return nil
}

func (sv *SkillValidator) validateTags(tags []string) error {
	for _, tag := range tags {
		if strings.TrimSpace(tag) == "" {
			return fmt.Errorf("%w: tag cannot be empty", ErrSkillInvalid)
		}
		if len(tag) > 50 {
			return fmt.Errorf("%w: tag exceeds maximum length of 50 characters", ErrSkillInvalid)
		}
	}

	return nil
}

func (sv *SkillValidator) validateDefinition(def *SkillDefinition) error {
	if def == nil {
		return nil
	}

	if err := sv.validateParameters(def.Parameters); err != nil {
		return err
	}

	if err := sv.validateTimeout(def.Timeout); err != nil {
		return err
	}

	return nil
}

func (sv *SkillValidator) validateParameters(params []SkillParameter) error {
	paramNames := make(map[string]bool)

	for _, param := range params {
		if strings.TrimSpace(param.Name) == "" {
			return fmt.Errorf("%w: parameter name cannot be empty", ErrSkillInvalid)
		}

		if paramNames[param.Name] {
			return fmt.Errorf("%w: duplicate parameter name '%s'", ErrSkillInvalid, param.Name)
		}
		paramNames[param.Name] = true

		validTypes := []string{"string", "number", "boolean", "array", "object", "integer"}
		isValidType := false
		for _, t := range validTypes {
			if param.Type == t {
				isValidType = true
				break
			}
		}
		if !isValidType && param.Type != "" {
			return fmt.Errorf("%w: invalid parameter type '%s' for parameter '%s'", ErrSkillInvalid, param.Type, param.Name)
		}
	}

	return nil
}

func (sv *SkillValidator) validateTimeout(timeout time.Duration) error {
	return nil
}

func (sv *SkillValidator) ValidateSkillDependencies(skill *Skill, availableSkills map[string]*Skill) error {
	if skill.Definition == nil || len(skill.Definition.Dependencies) == 0 {
		return nil
	}

	sv.dependencyResolver.Reset()

	for _, depID := range skill.Definition.Dependencies {
		if _, ok := availableSkills[depID]; !ok {
			return fmt.Errorf("%w: '%s'", ErrDependencyNotFound, depID)
		}
	}

	if err := sv.dependencyResolver.DetectCycle(skill.ID, skill.Definition.Dependencies, availableSkills); err != nil {
		return err
	}

	return nil
}

func (dr *DependencyResolver) Reset() {
	dr.visited = make(map[string]bool)
	dr.stack = make(map[string]bool)
}

func (dr *DependencyResolver) DetectCycle(skillID string, dependencies []string, availableSkills map[string]*Skill) error {
	if dr.stack[skillID] {
		return fmt.Errorf("%w: skill '%s'", ErrCircularDependency, skillID)
	}

	if dr.visited[skillID] {
		return nil
	}

	dr.stack[skillID] = true

	skill, ok := availableSkills[skillID]
	if ok && skill.Definition != nil {
		for _, depID := range skill.Definition.Dependencies {
			depSkill, depOk := availableSkills[depID]
			if !depOk {
				return fmt.Errorf("%w: '%s'", ErrDependencyNotFound, depID)
			}
			
			var depDeps []string
			if depSkill.Definition != nil {
				depDeps = depSkill.Definition.Dependencies
			}
			if err := dr.DetectCycle(depID, depDeps, availableSkills); err != nil {
				return err
			}
		}
	}

	for _, depID := range dependencies {
		depSkill, depOk := availableSkills[depID]
		if !depOk {
			return fmt.Errorf("%w: '%s'", ErrDependencyNotFound, depID)
		}
		
		var depDeps []string
		if depSkill.Definition != nil {
			depDeps = depSkill.Definition.Dependencies
		}
		if err := dr.DetectCycle(depID, depDeps, availableSkills); err != nil {
			return err
		}
	}

	dr.stack[skillID] = false
	dr.visited[skillID] = true

	return nil
}

func (sv *SkillValidator) ValidateBatch(skills []*Skill) error {
	errors := make([]error, 0)

	for _, skill := range skills {
		if err := sv.ValidateSkill(skill); err != nil {
			errors = append(errors, fmt.Errorf("skill '%s': %w", skill.ID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("batch validation failed with %d errors: %v", len(errors), errors)
	}

	return nil
}
