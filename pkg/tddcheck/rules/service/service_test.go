package service

import (
	"reflect"
	"testing"

	"github.com/lwmacct/260610-go-pkg-tddcheck/pkg/tddcheck/testkit"
)

func TestModuleServiceRuleViolations(t *testing.T) {
	root := t.TempDir()

	testkit.WriteFile(t, root, "good/service.go", `package good

type Config struct {}
type Service struct {}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Run() {}
`)
	testkit.WriteFile(t, root, "good/service.auth.go", `package good

func (s *Service) Login() {}
func (s Service) Logout() {}
`)
	testkit.WriteFile(t, root, "bad/service.extra.go", `package bad

type Extra struct {}

func helper() {}

func (s *Service) Run() {}
`)
	testkit.WriteFile(t, root, "bad/handler.go", `package bad

type Service struct {}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Handle() {}
`)
	testkit.WriteFile(t, root, "bad/service.go", `package bad

const mode = "bad"
var enabled = true

type Service struct {}
type Config struct {}
type Dependencies struct {}
type Loader interface {}
type LoaderFunc func()
type Entity struct {}
type Alias = string

func NewService() *Service {
	return &Service{}
}

func New() *Service {
	return &Service{}
}

func helper() {}

func (s *Service) Run() {}
`)
	testkit.WriteFile(t, root, "bad/handler_test.go", `package bad

func (s *Service) HandleTest() {}
`)

	violations, err := New(root).ServiceBoundaryViolations()
	if err != nil {
		t.Fatal(err)
	}

	got := serviceViolationMessages(violations)
	want := []string{
		"type Service must be declared in service.go",
		"NewService must be declared in service.go",
		"Service receiver method Handle must be declared in service.go or service.*.go",
		"service.*.go must only declare Service receiver methods",
		"service.*.go must only declare Service receiver methods",
		"service.go must not declare const or var",
		"service.go must not declare const or var",
		"service.go type Entity must be Service, Config, Dependencies, interface, or *Func dependency",
		"service.go type Alias must be Service, Config, Dependencies, interface, or *Func dependency",
		"service.go must only declare NewService and Service receiver methods",
		"service.go must only declare NewService and Service receiver methods",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("violations = %v, want %v", got, want)
	}
}

func TestRulesRequiresRoot(t *testing.T) {
	_, err := New("").ServiceBoundaryViolations()
	if err == nil {
		t.Fatal("expected error for empty root")
	}
}

func serviceViolationMessages(violations []ServiceBoundaryViolation) []string {
	messages := make([]string, 0, len(violations))
	for _, violation := range violations {
		messages = append(messages, violation.Message)
	}
	return messages
}
