package domain

func (t *Tree) DependencyOutcome(ref string) DependencyOutcome {
	if _, _, cross := SplitRef(ref); !cross {
		if t.Issues[ref] == nil {
			return DependencyUnknown
		}
		if t.State(ref) == StateDone {
			return DependencySatisfied
		}
		return DependencyUnsatisfied
	}
	x := t.XIssueOf(ref)
	if x.Outcome != "" {
		return x.Outcome
	}
	if x.Status != XFound {
		return DependencyUnknown
	}
	if x.State == StateDone {
		return DependencySatisfied
	}
	return DependencyUnsatisfied
}

// ExternalDependency returns what is known about a cross-project dependency
// of an issue.
func (t *Tree) ExternalDependency(issueID, ref string) XIssue {
	return t.XIssueOf(ref)
}
