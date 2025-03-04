// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

// Package sqlcheck provides interfaces for analyzing the contents of SQL files
// to generate insights on the safety of many kinds of changes to database
// schemas. With this package developers may define an Analyzer that can be used
// to diagnose the impact of SQL statements on the target database. For instance,
// The `destructive` package exposes an Analyzer that detects destructive changes
// to the database schema, such as the dropping of tables or columns.
package sqlcheck

import (
	"context"

	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlclient"
)

type (
	// An Analyzer describes a migration file analyzer.
	Analyzer interface {
		// Analyze executes the analysis function.
		Analyze(context.Context, *Pass) error
	}

	// A Pass provides information to the Run function that
	// applies a specific analyzer to an SQL file.
	Pass struct {
		// A migration file and the changes it describes.
		File *File

		// Dev is a driver-specific environment used to execute analysis work.
		Dev *sqlclient.Client

		// Report reports analysis reports.
		Reporter ReportWriter
	}

	// File represents a parsed version of a migration file.
	File struct {
		migrate.File

		// Changes represents the list of changes this file represents.
		Changes []*Change

		// Sum represents a summary of changes this file represents. For example,
		// in case of a file that contains exactly two statements, and the first
		// statement is reverted by the one after it, the Sum is nil.
		Sum schema.Changes

		// schema spans. lazily initialized.
		spans map[string]*schemaSpan
	}

	// A Change in a migration file.
	Change struct {
		schema.Changes        // The actual changes.
		Stmt           string // The SQL statement generated this change.
		Pos            int    // The position of the statement in the file.
	}

	// A Report describes an analysis report with an optional specific diagnostic.
	Report struct {
		Text        string       // Report text.
		Diagnostics []Diagnostic // Report diagnostics.
	}

	// A Diagnostic is a text associated with a specific position of a statement in a file.
	Diagnostic struct {
		Pos  int    // Diagnostic position.
		Text string // Diagnostic text.
	}

	// ReportWriter represents a writer for analysis reports.
	ReportWriter interface {
		WriteReport(Report)
	}
)

// Analyzers implements Analyzer.
type Analyzers []Analyzer

// Analyze implements Analyzer.
func (a Analyzers) Analyze(ctx context.Context, p *Pass) error {
	for _, a := range a {
		if err := a.Analyze(ctx, p); err != nil {
			return err
		}
	}
	return nil
}

// ReportWriterFunc is a function that implements Reporter.
type ReportWriterFunc func(Report)

// WriteReport calls f(r).
func (f ReportWriterFunc) WriteReport(r Report) {
	f(r)
}

// ResourceSpan describes the lifespan of a resource
// in perspective to the migration file.
type ResourceSpan uint

const (
	// SpanUnknown describes unknown lifespan.
	// e.g. resource may exist before this file.
	SpanUnknown ResourceSpan = iota

	// SpanAdded describes that a span of
	// a resource was started in this file.
	SpanAdded

	// SpanDropped describes that a span of
	// a resource was ended in this file.
	SpanDropped

	// SpanTemporary indicates that a resource lifetime
	// was started and ended in this file (CREATE and DROP).
	SpanTemporary = SpanAdded | SpanDropped
)

// SchemaSpan returns the span information for the schema.
func (f *File) SchemaSpan(s *schema.Schema) ResourceSpan {
	return f.schemaSpan(s).state
}

// TableSpan returns the span information for the table.
func (f *File) TableSpan(t *schema.Table) ResourceSpan {
	return f.tableSpan(t).state
}

// ColumnSpan returns the span information for the column.
func (f *File) ColumnSpan(t *schema.Table, c *schema.Column) ResourceSpan {
	return f.tableSpan(t).columns[c.Name]
}

// IndexSpan returns the span information for the span.
func (f *File) IndexSpan(t *schema.Table, i *schema.Index) ResourceSpan {
	return f.tableSpan(t).indexes[i.Name]
}

type (
	// schemaSpan holds the span structure of a schema.
	schemaSpan struct {
		state  ResourceSpan
		tables map[string]*tableSpan
	}
	// schemaSpan holds the span structure of a table.
	tableSpan struct {
		state   ResourceSpan
		columns map[string]ResourceSpan
		indexes map[string]ResourceSpan
	}
)

func (f *File) loadSpans() {
	f.spans = make(map[string]*schemaSpan)
	for _, sc := range f.Changes {
		for _, c := range sc.Changes {
			switch c := c.(type) {
			case *schema.AddSchema:
				f.schemaSpan(c.S).state = SpanAdded
			case *schema.DropSchema:
				f.schemaSpan(c.S).state |= SpanDropped
			case *schema.AddTable:
				span := f.tableSpan(c.T)
				span.state = SpanAdded
				for _, column := range c.T.Columns {
					span.columns[column.Name] = SpanAdded
				}
				for _, idx := range c.T.Indexes {
					span.indexes[idx.Name] = SpanAdded
				}
			case *schema.DropTable:
				f.tableSpan(c.T).state |= SpanDropped
			case *schema.ModifyTable:
				span := f.tableSpan(c.T)
				for _, c1 := range c.Changes {
					switch c1 := c1.(type) {
					case *schema.AddColumn:
						span.columns[c1.C.Name] = SpanAdded
					case *schema.DropColumn:
						span.columns[c1.C.Name] |= SpanDropped
					case *schema.AddIndex:
						span.indexes[c1.I.Name] = SpanAdded
					case *schema.DropIndex:
						span.indexes[c1.I.Name] |= SpanDropped
					}
				}
			}
		}
	}
}

func (f *File) schemaSpan(s *schema.Schema) *schemaSpan {
	if f.spans == nil {
		f.loadSpans()
	}
	if f.spans[s.Name] == nil {
		f.spans[s.Name] = &schemaSpan{tables: make(map[string]*tableSpan)}
	}
	return f.spans[s.Name]
}

func (f *File) tableSpan(t *schema.Table) *tableSpan {
	span := f.schemaSpan(t.Schema)
	if span.tables[t.Name] == nil {
		span.tables[t.Name] = &tableSpan{
			columns: make(map[string]ResourceSpan),
			indexes: make(map[string]ResourceSpan),
		}
	}
	return f.spans[t.Schema.Name].tables[t.Name]
}
