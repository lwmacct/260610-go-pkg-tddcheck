package tddcheck

// DefaultRules returns rules intended for normal go test usage.
func DefaultRules() []Rule {
	return []Rule{
		PublicAPIsHaveTests(),
		TestsAreNotEmpty(),
	}
}
