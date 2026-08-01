package rules

// Default returns all built-in rule packs.
func Default() []Rule {
	var result []Rule
	result = append(result, DefaultQuality()...)
	result = append(result, DefaultSecurity()...)
	result = append(result, DefaultHardening()...)
	result = append(result, DefaultReliability()...)
	return result
}
