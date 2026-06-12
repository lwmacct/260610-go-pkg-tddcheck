package tddcheck

func (v ConstantsBoundaryViolation) GetFile() string { return v.File }
func (v ConstantsBoundaryViolation) GetLine() int    { return v.Line }
func (v ConstantsBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v EntityBoundaryViolation) GetFile() string { return v.File }
func (v EntityBoundaryViolation) GetLine() int    { return v.Line }
func (v EntityBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v PackageNameViolation) GetFile() string { return v.File }
func (v PackageNameViolation) GetLine() int    { return v.Line }
func (v PackageNameViolation) GetMessage() string {
	return v.Message
}

func (v DatabaseTestViolation) GetFile() string { return v.File }
func (v DatabaseTestViolation) GetLine() int    { return v.Line }
func (v DatabaseTestViolation) GetMessage() string {
	return v.Message
}

func (v ContextBoundaryViolation) GetFile() string { return v.File }
func (v ContextBoundaryViolation) GetLine() int    { return v.Line }
func (v ContextBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v MapperBoundaryViolation) GetFile() string { return v.File }
func (v MapperBoundaryViolation) GetLine() int    { return v.Line }
func (v MapperBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v HandlerBoundaryViolation) GetFile() string { return v.File }
func (v HandlerBoundaryViolation) GetLine() int    { return v.Line }
func (v HandlerBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v ErrorsBoundaryViolation) GetFile() string { return v.File }
func (v ErrorsBoundaryViolation) GetLine() int    { return v.Line }
func (v ErrorsBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v ServiceBoundaryViolation) GetFile() string { return v.File }
func (v ServiceBoundaryViolation) GetLine() int    { return v.Line }
func (v ServiceBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v SchemaBoundaryViolation) GetFile() string { return v.File }
func (v SchemaBoundaryViolation) GetLine() int    { return v.Line }
func (v SchemaBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v ValidationBoundaryViolation) GetFile() string { return v.File }
func (v ValidationBoundaryViolation) GetLine() int    { return v.Line }
func (v ValidationBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v PublicAPINameViolation) GetFile() string { return v.File }
func (v PublicAPINameViolation) GetLine() int    { return v.Line }
func (v PublicAPINameViolation) GetMessage() string {
	return v.Message
}

func (v UtilsBoundaryViolation) GetFile() string { return v.File }
func (v UtilsBoundaryViolation) GetLine() int    { return v.Line }
func (v UtilsBoundaryViolation) GetMessage() string {
	return v.Message
}

func (v RepositoryBoundaryViolation) GetFile() string { return v.File }
func (v RepositoryBoundaryViolation) GetLine() int    { return v.Line }
func (v RepositoryBoundaryViolation) GetMessage() string {
	return v.Message
}
