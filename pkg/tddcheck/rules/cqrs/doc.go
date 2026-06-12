// Package cqrs checks naming rules for command/query responsibility
// segregation types.
//
// It scans cqrs.go files and reports structs that do not end with Query,
// Result, or Command. It also reports interfaces that do not express accepted
// use case contracts such as UseCase, CommandHandler, QueryHandler, Access,
// Policy, or Authorizer.
package cqrs
