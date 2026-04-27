package gstack

import (
	"testing"
)

func TestAllSkills(t *testing.T) {
	skills := AllSkills()
	if len(skills) == 0 {
		t.Fatal("AllSkills should return at least one skill")
	}

	// Verify all skills have required fields
	for _, s := range skills {
		if s.Name == "" {
			t.Error("skill Name should not be empty")
		}
		if s.Dir == "" {
			t.Errorf("skill %s has empty Dir", s.Name)
		}
		if s.Category == "" {
			t.Errorf("skill %s has empty Category", s.Name)
		}
		if s.Description == "" {
			t.Errorf("skill %s has empty Description", s.Name)
		}
	}
}

func TestSkillsByCategory(t *testing.T) {
	byCategory := SkillsByCategory()

	expectedCategories := []string{
		CategoryPlanning, CategoryReview, CategoryQATesting,
		CategorySecurity, CategoryShipping, CategorySafety,
		CategoryDebugging, CategoryRetrospective,
	}

	for _, cat := range expectedCategories {
		skills, ok := byCategory[cat]
		if !ok || len(skills) == 0 {
			t.Errorf("category %s should have at least one skill", cat)
		}
	}
}

func TestFilterSkills(t *testing.T) {
	// Select only review and qa
	selected := []string{"review", "qa"}

	// With runtime available
	withRuntime := FilterSkills(selected, true)
	if len(withRuntime) != 2 {
		t.Errorf("expected 2 skills with runtime, got %d", len(withRuntime))
	}

	// Without runtime — qa requires runtime, should be filtered
	withoutRuntime := FilterSkills(selected, false)
	if len(withoutRuntime) != 1 {
		t.Errorf("expected 1 skill without runtime, got %d", len(withoutRuntime))
	}
	if len(withoutRuntime) > 0 && withoutRuntime[0].Dir != "review" {
		t.Errorf("expected review skill, got %s", withoutRuntime[0].Dir)
	}
}

func TestFilterSkillsEmpty(t *testing.T) {
	result := FilterSkills([]string{}, true)
	if len(result) != 0 {
		t.Errorf("empty selection should return no skills, got %d", len(result))
	}
}

func TestAllCategories(t *testing.T) {
	cats := AllCategories()
	if len(cats) != 10 {
		t.Errorf("expected 10 categories, got %d", len(cats))
	}
}

func TestCategoryLabel(t *testing.T) {
	label := CategoryLabel(CategoryPlanning)
	if label == "" || label == CategoryPlanning {
		t.Errorf("CategoryLabel should return a human-readable label, got %s", label)
	}

	// Unknown category should return the key itself
	unknown := CategoryLabel("unknown-category")
	if unknown != "unknown-category" {
		t.Errorf("unknown category should return key, got %s", unknown)
	}
}
