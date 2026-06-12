package tddcheck

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rulekit"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/dependency/layer"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/constants"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/context"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/cqrs"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/dto"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/entity"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/errors"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/handler"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/mapper"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/repository"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/schema"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/service"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/utils"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/file/validation"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/name/errorprefix"
	packagename "github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/name/package"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/name/publicapi"
	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/rules/test/database"
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
	Root   string
	Config Config
	// IncludeDatabaseTests enables the database test helper boundary rule.
	IncludeDatabaseTests bool
}

// Assert fails the test when any project architecture boundary is violated.
func (r ProjectRules) Assert(tb testing.TB) {
	tb.Helper()

	result := r.Check()
	if result.Err != nil {
		tb.Fatal(result.Err)
	}
	if !result.Passed {
		tb.Fatal(result.Text())
	}
}

// Check runs the configured architecture rules and returns a normalized result.
func (r ProjectRules) Check() Result {
	start := time.Now()
	root := r.Root
	if root == "" {
		root = "internal"
	}
	config := r.Config

	var violations []Violation
	add := func(rule string, values []Violation, err error) error {
		if err != nil {
			return fmt.Errorf("%s: %w", rule, err)
		}
		violations = append(violations, values...)
		return nil
	}
	run := func(rule string, check func(string, Config) ([]Violation, error)) error {
		values, err := check(root, config)
		return add(rule, values, err)
	}

	if err := run("dependency-layer", checkLayer); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("name-package", checkPackageName); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-constants", checkConstants); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-entity", checkEntity); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-errors", checkErrors); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("name-error-prefix", checkErrorPrefix); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-context", checkContext); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-cqrs", checkCQRS); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-dto", checkDTO); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-mapper", checkMapper); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("name-public-api", checkPublicAPI); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-service", checkService); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-validation", checkValidation); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-handler", checkHandler); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-repository", checkRepository); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-schema", checkSchema); err != nil {
		return resultError(err, violations, start)
	}
	if err := run("file-utils", checkUtils); err != nil {
		return resultError(err, violations, start)
	}
	if r.IncludeDatabaseTests {
		if err := run("test-database", checkDatabaseTests); err != nil {
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

func checkLayer(root string, config Config) ([]Violation, error) {
	values, err := layer.New(root, rulekit.WithConfig(config)).LayerDependencyViolations()
	return mapViolations("dependency-layer", values, err, func(v layer.LayerDependencyViolation) Violation {
		return Violation{Rule: "dependency-layer", File: v.File, Line: v.Line, Message: v.Message + ": " + v.ImportPath}
	})
}

func checkPackageName(root string, config Config) ([]Violation, error) {
	values, err := packagename.New(root, rulekit.WithConfig(config)).PackageNameViolations()
	return mapMessageViolations("name-package", values, err)
}

func checkConstants(root string, config Config) ([]Violation, error) {
	values, err := constants.New(root, rulekit.WithConfig(config)).ConstantsBoundaryViolations()
	return mapMessageViolations("file-constants", values, err)
}

func checkEntity(root string, config Config) ([]Violation, error) {
	values, err := entity.New(root, rulekit.WithConfig(config)).EntityBoundaryViolations()
	return mapMessageViolations("file-entity", values, err)
}

func checkErrors(root string, config Config) ([]Violation, error) {
	values, err := errors.New(root, rulekit.WithConfig(config)).ErrorsBoundaryViolations()
	return mapMessageViolations("file-errors", values, err)
}

func checkErrorPrefix(root string, config Config) ([]Violation, error) {
	values, err := errorprefix.New(root, rulekit.WithConfig(config)).ErrorPrefixViolations()
	return mapViolations("name-error-prefix", values, err, func(v errorprefix.ErrorPrefixViolation) Violation {
		return Violation{Rule: "name-error-prefix", File: v.File, Line: v.Line, Message: fmt.Sprintf("error variable %s must start with Err", v.Name)}
	})
}

func checkContext(root string, config Config) ([]Violation, error) {
	values, err := context.New(root, rulekit.WithConfig(config)).ContextBoundaryViolations()
	return mapMessageViolations("file-context", values, err)
}

func checkCQRS(root string, config Config) ([]Violation, error) {
	rules := cqrs.New(root, rulekit.WithConfig(config))
	var violations []Violation
	structs, err := rules.StructSuffixViolations()
	if err != nil {
		return nil, err
	}
	for _, violation := range structs {
		violations = append(violations, Violation{
			Rule:    "file-cqrs",
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
			Rule:    "file-cqrs",
			File:    violation.File,
			Line:    violation.Line,
			Message: fmt.Sprintf("interface %s %s", violation.Name, violation.Message),
		})
	}
	return violations, nil
}

func checkDTO(root string, config Config) ([]Violation, error) {
	rules := dto.New(root, rulekit.WithConfig(config))
	var violations []Violation
	structs, err := rules.StructSuffixViolations()
	if err != nil {
		return nil, err
	}
	for _, violation := range structs {
		violations = append(violations, Violation{
			Rule:    "file-dto",
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
			Rule:    "file-dto",
			File:    violation.File,
			Line:    violation.Line,
			Message: "dto.go must not declare func " + violation.Name,
		})
	}
	owned, err := rules.FileOwnershipViolations()
	if err != nil {
		return nil, err
	}
	for _, violation := range owned {
		violations = append(violations, Violation{
			Rule:    "file-dto",
			File:    violation.File,
			Line:    violation.Line,
			Message: fmt.Sprintf("DTO struct %s must be declared in dto.go", violation.Name),
		})
	}
	return violations, nil
}

func checkMapper(root string, config Config) ([]Violation, error) {
	values, err := mapper.New(root, rulekit.WithConfig(config)).MapperBoundaryViolations()
	return mapMessageViolations("file-mapper", values, err)
}

func checkPublicAPI(root string, config Config) ([]Violation, error) {
	values, err := publicapi.New(root, rulekit.WithConfig(config)).PublicAPINameViolations()
	return mapMessageViolations("name-public-api", values, err)
}

func checkService(root string, config Config) ([]Violation, error) {
	values, err := service.New(root, rulekit.WithConfig(config)).ServiceBoundaryViolations()
	return mapMessageViolations("file-service", values, err)
}

func checkValidation(root string, config Config) ([]Violation, error) {
	values, err := validation.New(root, rulekit.WithConfig(config)).ValidationBoundaryViolations()
	return mapMessageViolations("file-validation", values, err)
}

func checkHandler(root string, config Config) ([]Violation, error) {
	values, err := handler.New(root, rulekit.WithConfig(config)).HandlerBoundaryViolations()
	return mapMessageViolations("file-handler", values, err)
}

func checkRepository(root string, config Config) ([]Violation, error) {
	values, err := repository.New(root, rulekit.WithConfig(config)).RepositoryBoundaryViolations()
	return mapMessageViolations("file-repository", values, err)
}

func checkSchema(root string, config Config) ([]Violation, error) {
	values, err := schema.New(root, rulekit.WithConfig(config)).SchemaBoundaryViolations()
	return mapMessageViolations("file-schema", values, err)
}

func checkUtils(root string, config Config) ([]Violation, error) {
	values, err := utils.New(root, rulekit.WithConfig(config)).UtilsBoundaryViolations()
	return mapMessageViolations("file-utils", values, err)
}

func checkDatabaseTests(root string, config Config) ([]Violation, error) {
	values, err := database.New(root, rulekit.WithConfig(config)).DatabaseTestBoundaryViolations()
	return mapMessageViolations("test-database", values, err)
}

func mapMessageViolations[T any](rule string, values []T, err error) ([]Violation, error) {
	return mapViolations(rule, values, err, func(v T) Violation {
		return messageViolation(rule, v)
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

func messageViolation(rule string, value any) Violation {
	fields := reflect.ValueOf(value)
	return Violation{
		Rule:    rule,
		File:    fields.FieldByName("File").String(),
		Line:    int(fields.FieldByName("Line").Int()),
		Message: fields.FieldByName("Message").String(),
	}
}
