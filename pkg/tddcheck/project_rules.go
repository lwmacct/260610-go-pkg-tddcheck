package tddcheck

import (
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"
)

// Violation is a normalized architecture rule violation.
type Violation struct {
	Rule    string
	File    string
	Line    int
	Message string
}

func (v Violation) String() string {
	location := v.File
	if v.Line > 0 {
		location = fmt.Sprintf("%s:%d", v.File, v.Line)
	}
	if location == "" {
		location = "-"
	}
	return fmt.Sprintf("%s [%s] %s", location, v.Rule, v.Message)
}

// Result is the normalized output of a project architecture check.
type Result struct {
	Passed     bool
	Err        error
	Violations []Violation
	Duration   time.Duration
}

// Text renders the result in a stable, line-oriented format.
func (r Result) Text() string {
	if r.Err != nil {
		return "tddcheck: " + r.Err.Error()
	}
	if len(r.Violations) == 0 {
		return "tddcheck: passed"
	}

	lines := make([]string, 0, len(r.Violations)+1)
	lines = append(lines, "tddcheck: failed")
	for _, violation := range r.Violations {
		lines = append(lines, violation.String())
	}
	return strings.Join(lines, "\n")
}

// ProjectRules runs the default project architecture boundary rules.
type ProjectRules struct {
	// Root is the module subtree to scan. Relative paths are resolved from go.mod.
	Root string
	// IncludeDatabaseTests enables the database test helper boundary rule.
	IncludeDatabaseTests bool
}

// Assert fails the test when any project architecture boundary is violated.
func (r ProjectRules) Assert(t testing.TB) {
	t.Helper()

	result := r.Check()
	if result.Err != nil {
		t.Fatal(result.Err)
	}
	if !result.Passed {
		t.Fatal(result.Text())
	}
}

// Check runs the configured architecture rules and returns a normalized result.
func (r ProjectRules) Check() Result {
	start := time.Now()
	root := r.Root
	if root == "" {
		root = "internal"
	}

	var violations []Violation
	add := func(rule string, values []Violation, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %w", rule, err)
		}
		violations = append(violations, values...)
		return nil
	}
	run := func(rule string, check func(string) ([]Violation, error)) error {
		values, err := check(root)
		return add(rule, values, err)
	}

	if err := run("layer", checkLayer); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("package-name", checkPackageName); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("constants", checkConstants); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("entity", checkEntity); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("errors", checkErrors); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("error-prefix", checkErrorPrefix); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("context", checkContext); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("cqrs", checkCQRS); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("dto", checkDTO); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("mapper", checkMapper); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("public-api", checkPublicAPI); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("service", checkService); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("validation", checkValidation); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("handler", checkHandler); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("repository", checkRepository); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("schema", checkSchema); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("utils", checkUtils); err != nil {
		return resultError(err, violations, start)
	}
	if r.IncludeDatabaseTests {
		if err := run("database-test", checkDatabaseTests); err != nil {
			return resultError(err, violations, start)
		}
	}

	slices.SortFunc(violations, func(a, b Violation) int {
		return strings.Compare(a.String(), b.String())
	})
	return Result{
		Passed:     len(violations) == 0,
		Violations: violations,
		Duration:   time.Since(start),
	}
}

func resultError(err error, violations []Violation, start time.Time) Result {
	return Result{
		Passed:     false,
		Err:        err,
		Violations: violations,
		Duration:   time.Since(start),
	}
}

func checkLayer(root string) ([]Violation, error) {
	values, err := (ModuleLayerRules{Root: root}).LayerDependencyViolations()
	return mapViolations("layer", values, err, func(v LayerDependencyViolation) Violation {
		return Violation{Rule: "layer", File: v.File, Line: v.Line, Message: v.Message + ": " + v.ImportPath}
	})
}

func checkPackageName(root string) ([]Violation, error) {
	values, err := (ModulePackageNameRules{Root: root}).PackageNameViolations()
	return mapMessageViolations("package-name", values, err)
}

func checkConstants(root string) ([]Violation, error) {
	values, err := (ModuleConstantsRules{Root: root}).ConstantsBoundaryViolations()
	return mapMessageViolations("constants", values, err)
}

func checkEntity(root string) ([]Violation, error) {
	values, err := (ModuleEntityRules{Root: root}).EntityBoundaryViolations()
	return mapMessageViolations("entity", values, err)
}

func checkErrors(root string) ([]Violation, error) {
	values, err := (ModuleErrorRules{Root: root}).ErrorsBoundaryViolations()
	return mapMessageViolations("errors", values, err)
}

func checkErrorPrefix(root string) ([]Violation, error) {
	values, err := (ModuleErrorRules{Root: root}).ErrorPrefixViolations()
	return mapViolations("error-prefix", values, err, func(v ErrorPrefixViolation) Violation {
		return Violation{Rule: "error-prefix", File: v.File, Line: v.Line, Message: fmt.Sprintf("error variable %s must start with Err", v.Name)}
	})
}

func checkContext(root string) ([]Violation, error) {
	values, err := (ModuleContextRules{Root: root}).ContextBoundaryViolations()
	return mapMessageViolations("context", values, err)
}

func checkCQRS(root string) ([]Violation, error) {
	rules := ModuleCQRSRules{Root: root}
	var violations []Violation
	structs, err := rules.StructSuffixViolations()
	if err != nil {
		return nil, err
	}
	for _, violation := range structs {
		violations = append(violations, Violation{
			Rule:    "cqrs",
			File:    violation.File,
			Line:    violation.Line,
			Message: fmt.Sprintf("struct %s must end with Query, Result, or Command", violation.Name),
		})
	}
	interfaces, err := rules.InterfaceNameViolations()
	if err != nil {
		return nil, err
	}
	for _, violation := range interfaces {
		violations = append(violations, Violation{
			Rule:    "cqrs",
			File:    violation.File,
			Line:    violation.Line,
			Message: fmt.Sprintf("interface %s %s", violation.Name, violation.Message),
		})
	}
	return violations, nil
}

func checkDTO(root string) ([]Violation, error) {
	rules := ModuleDTORules{Root: root}
	var violations []Violation
	structs, err := rules.StructSuffixViolations()
	if err != nil {
		return nil, err
	}
	for _, violation := range structs {
		violations = append(violations, Violation{
			Rule:    "dto",
			File:    violation.File,
			Line:    violation.Line,
			Message: fmt.Sprintf("struct %s must end with DTO or DTOs", violation.Name),
		})
	}
	funcs, err := rules.FuncViolations()
	if err != nil {
		return nil, err
	}
	for _, violation := range funcs {
		violations = append(violations, Violation{
			Rule:    "dto",
			File:    violation.File,
			Line:    violation.Line,
			Message: fmt.Sprintf("dto.go must not declare func %s", violation.Name),
		})
	}
	owned, err := rules.FileOwnershipViolations()
	if err != nil {
		return nil, err
	}
	for _, violation := range owned {
		violations = append(violations, Violation{
			Rule:    "dto",
			File:    violation.File,
			Line:    violation.Line,
			Message: fmt.Sprintf("DTO struct %s must be declared in dto.go", violation.Name),
		})
	}
	return violations, nil
}

func checkMapper(root string) ([]Violation, error) {
	values, err := (ModuleMapperRules{Root: root}).MapperBoundaryViolations()
	return mapMessageViolations("mapper", values, err)
}

func checkPublicAPI(root string) ([]Violation, error) {
	values, err := (ModulePublicAPIRules{Root: root}).PublicAPINameViolations()
	return mapMessageViolations("public-api", values, err)
}

func checkService(root string) ([]Violation, error) {
	values, err := (ModuleServiceRules{Root: root}).ServiceBoundaryViolations()
	return mapMessageViolations("service", values, err)
}

func checkValidation(root string) ([]Violation, error) {
	values, err := (ModuleValidationRules{Root: root}).ValidationBoundaryViolations()
	return mapMessageViolations("validation", values, err)
}

func checkHandler(root string) ([]Violation, error) {
	values, err := (ModuleHandlerRules{Root: root}).HandlerBoundaryViolations()
	return mapMessageViolations("handler", values, err)
}

func checkRepository(root string) ([]Violation, error) {
	values, err := (ModuleRepositoryRules{Root: root}).RepositoryBoundaryViolations()
	return mapMessageViolations("repository", values, err)
}

func checkSchema(root string) ([]Violation, error) {
	values, err := (ModuleSchemaRules{Root: root}).SchemaBoundaryViolations()
	return mapMessageViolations("schema", values, err)
}

func checkUtils(root string) ([]Violation, error) {
	values, err := (ModuleUtilsRules{Root: root}).UtilsBoundaryViolations()
	return mapMessageViolations("utils", values, err)
}

func checkDatabaseTests(root string) ([]Violation, error) {
	values, err := (DatabaseTestRules{Root: root}).DatabaseTestBoundaryViolations()
	return mapMessageViolations("database-test", values, err)
}

type messageViolation interface {
	GetFile() string
	GetLine() int
	GetMessage() string
}

func mapMessageViolations[T messageViolation](rule string, values []T, err error) ([]Violation, error) {
	return mapViolations(rule, values, err, func(v T) Violation {
		return Violation{Rule: rule, File: v.GetFile(), Line: v.GetLine(), Message: v.GetMessage()}
	})
}

func mapViolations[T any](rule string, values []T, err error, convert func(T) Violation) ([]Violation, error) {
	if err != nil {
		return nil, err
	}
	violations := make([]Violation, 0, len(values))
	for _, value := range values {
		violation := convert(value)
		if violation.Rule == "" {
			violation.Rule = rule
		}
		violations = append(violations, violation)
	}
	return violations, nil
}
