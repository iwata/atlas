// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

// Code generated by entc, DO NOT EDIT.

package revision

import (
	"time"

	"ariga.io/atlas/cmd/atlas/internal/migrate/ent/predicate"
	"entgo.io/ent/dialect/sql"
)

// ID filters vertices based on their ID field.
func ID(id string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDEQ applies the EQ predicate on the ID field.
func IDEQ(id string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldID), id))
	})
}

// IDNEQ applies the NEQ predicate on the ID field.
func IDNEQ(id string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldID), id))
	})
}

// IDIn applies the In predicate on the ID field.
func IDIn(ids ...string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.In(s.C(FieldID), v...))
	})
}

// IDNotIn applies the NotIn predicate on the ID field.
func IDNotIn(ids ...string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(ids) == 0 {
			s.Where(sql.False())
			return
		}
		v := make([]interface{}, len(ids))
		for i := range v {
			v[i] = ids[i]
		}
		s.Where(sql.NotIn(s.C(FieldID), v...))
	})
}

// IDGT applies the GT predicate on the ID field.
func IDGT(id string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldID), id))
	})
}

// IDGTE applies the GTE predicate on the ID field.
func IDGTE(id string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldID), id))
	})
}

// IDLT applies the LT predicate on the ID field.
func IDLT(id string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldID), id))
	})
}

// IDLTE applies the LTE predicate on the ID field.
func IDLTE(id string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldID), id))
	})
}

// Description applies equality check predicate on the "description" field. It's identical to DescriptionEQ.
func Description(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldDescription), v))
	})
}

// ExecutedAt applies equality check predicate on the "executed_at" field. It's identical to ExecutedAtEQ.
func ExecutedAt(v time.Time) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldExecutedAt), v))
	})
}

// ExecutionTime applies equality check predicate on the "execution_time" field. It's identical to ExecutionTimeEQ.
func ExecutionTime(v time.Duration) predicate.Revision {
	vc := int64(v)
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldExecutionTime), vc))
	})
}

// Hash applies equality check predicate on the "hash" field. It's identical to HashEQ.
func Hash(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldHash), v))
	})
}

// OperatorVersion applies equality check predicate on the "operator_version" field. It's identical to OperatorVersionEQ.
func OperatorVersion(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldOperatorVersion), v))
	})
}

// DescriptionEQ applies the EQ predicate on the "description" field.
func DescriptionEQ(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldDescription), v))
	})
}

// DescriptionNEQ applies the NEQ predicate on the "description" field.
func DescriptionNEQ(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldDescription), v))
	})
}

// DescriptionIn applies the In predicate on the "description" field.
func DescriptionIn(vs ...string) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldDescription), v...))
	})
}

// DescriptionNotIn applies the NotIn predicate on the "description" field.
func DescriptionNotIn(vs ...string) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldDescription), v...))
	})
}

// DescriptionGT applies the GT predicate on the "description" field.
func DescriptionGT(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldDescription), v))
	})
}

// DescriptionGTE applies the GTE predicate on the "description" field.
func DescriptionGTE(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldDescription), v))
	})
}

// DescriptionLT applies the LT predicate on the "description" field.
func DescriptionLT(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldDescription), v))
	})
}

// DescriptionLTE applies the LTE predicate on the "description" field.
func DescriptionLTE(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldDescription), v))
	})
}

// DescriptionContains applies the Contains predicate on the "description" field.
func DescriptionContains(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldDescription), v))
	})
}

// DescriptionHasPrefix applies the HasPrefix predicate on the "description" field.
func DescriptionHasPrefix(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldDescription), v))
	})
}

// DescriptionHasSuffix applies the HasSuffix predicate on the "description" field.
func DescriptionHasSuffix(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldDescription), v))
	})
}

// DescriptionEqualFold applies the EqualFold predicate on the "description" field.
func DescriptionEqualFold(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldDescription), v))
	})
}

// DescriptionContainsFold applies the ContainsFold predicate on the "description" field.
func DescriptionContainsFold(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldDescription), v))
	})
}

// ExecutionStateEQ applies the EQ predicate on the "execution_state" field.
func ExecutionStateEQ(v ExecutionState) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldExecutionState), v))
	})
}

// ExecutionStateNEQ applies the NEQ predicate on the "execution_state" field.
func ExecutionStateNEQ(v ExecutionState) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldExecutionState), v))
	})
}

// ExecutionStateIn applies the In predicate on the "execution_state" field.
func ExecutionStateIn(vs ...ExecutionState) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldExecutionState), v...))
	})
}

// ExecutionStateNotIn applies the NotIn predicate on the "execution_state" field.
func ExecutionStateNotIn(vs ...ExecutionState) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldExecutionState), v...))
	})
}

// ExecutedAtEQ applies the EQ predicate on the "executed_at" field.
func ExecutedAtEQ(v time.Time) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldExecutedAt), v))
	})
}

// ExecutedAtNEQ applies the NEQ predicate on the "executed_at" field.
func ExecutedAtNEQ(v time.Time) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldExecutedAt), v))
	})
}

// ExecutedAtIn applies the In predicate on the "executed_at" field.
func ExecutedAtIn(vs ...time.Time) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldExecutedAt), v...))
	})
}

// ExecutedAtNotIn applies the NotIn predicate on the "executed_at" field.
func ExecutedAtNotIn(vs ...time.Time) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldExecutedAt), v...))
	})
}

// ExecutedAtGT applies the GT predicate on the "executed_at" field.
func ExecutedAtGT(v time.Time) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldExecutedAt), v))
	})
}

// ExecutedAtGTE applies the GTE predicate on the "executed_at" field.
func ExecutedAtGTE(v time.Time) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldExecutedAt), v))
	})
}

// ExecutedAtLT applies the LT predicate on the "executed_at" field.
func ExecutedAtLT(v time.Time) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldExecutedAt), v))
	})
}

// ExecutedAtLTE applies the LTE predicate on the "executed_at" field.
func ExecutedAtLTE(v time.Time) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldExecutedAt), v))
	})
}

// ExecutionTimeEQ applies the EQ predicate on the "execution_time" field.
func ExecutionTimeEQ(v time.Duration) predicate.Revision {
	vc := int64(v)
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldExecutionTime), vc))
	})
}

// ExecutionTimeNEQ applies the NEQ predicate on the "execution_time" field.
func ExecutionTimeNEQ(v time.Duration) predicate.Revision {
	vc := int64(v)
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldExecutionTime), vc))
	})
}

// ExecutionTimeIn applies the In predicate on the "execution_time" field.
func ExecutionTimeIn(vs ...time.Duration) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = int64(vs[i])
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldExecutionTime), v...))
	})
}

// ExecutionTimeNotIn applies the NotIn predicate on the "execution_time" field.
func ExecutionTimeNotIn(vs ...time.Duration) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = int64(vs[i])
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldExecutionTime), v...))
	})
}

// ExecutionTimeGT applies the GT predicate on the "execution_time" field.
func ExecutionTimeGT(v time.Duration) predicate.Revision {
	vc := int64(v)
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldExecutionTime), vc))
	})
}

// ExecutionTimeGTE applies the GTE predicate on the "execution_time" field.
func ExecutionTimeGTE(v time.Duration) predicate.Revision {
	vc := int64(v)
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldExecutionTime), vc))
	})
}

// ExecutionTimeLT applies the LT predicate on the "execution_time" field.
func ExecutionTimeLT(v time.Duration) predicate.Revision {
	vc := int64(v)
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldExecutionTime), vc))
	})
}

// ExecutionTimeLTE applies the LTE predicate on the "execution_time" field.
func ExecutionTimeLTE(v time.Duration) predicate.Revision {
	vc := int64(v)
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldExecutionTime), vc))
	})
}

// HashEQ applies the EQ predicate on the "hash" field.
func HashEQ(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldHash), v))
	})
}

// HashNEQ applies the NEQ predicate on the "hash" field.
func HashNEQ(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldHash), v))
	})
}

// HashIn applies the In predicate on the "hash" field.
func HashIn(vs ...string) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldHash), v...))
	})
}

// HashNotIn applies the NotIn predicate on the "hash" field.
func HashNotIn(vs ...string) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldHash), v...))
	})
}

// HashGT applies the GT predicate on the "hash" field.
func HashGT(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldHash), v))
	})
}

// HashGTE applies the GTE predicate on the "hash" field.
func HashGTE(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldHash), v))
	})
}

// HashLT applies the LT predicate on the "hash" field.
func HashLT(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldHash), v))
	})
}

// HashLTE applies the LTE predicate on the "hash" field.
func HashLTE(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldHash), v))
	})
}

// HashContains applies the Contains predicate on the "hash" field.
func HashContains(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldHash), v))
	})
}

// HashHasPrefix applies the HasPrefix predicate on the "hash" field.
func HashHasPrefix(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldHash), v))
	})
}

// HashHasSuffix applies the HasSuffix predicate on the "hash" field.
func HashHasSuffix(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldHash), v))
	})
}

// HashEqualFold applies the EqualFold predicate on the "hash" field.
func HashEqualFold(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldHash), v))
	})
}

// HashContainsFold applies the ContainsFold predicate on the "hash" field.
func HashContainsFold(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldHash), v))
	})
}

// OperatorVersionEQ applies the EQ predicate on the "operator_version" field.
func OperatorVersionEQ(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionNEQ applies the NEQ predicate on the "operator_version" field.
func OperatorVersionNEQ(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.NEQ(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionIn applies the In predicate on the "operator_version" field.
func OperatorVersionIn(vs ...string) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.In(s.C(FieldOperatorVersion), v...))
	})
}

// OperatorVersionNotIn applies the NotIn predicate on the "operator_version" field.
func OperatorVersionNotIn(vs ...string) predicate.Revision {
	v := make([]interface{}, len(vs))
	for i := range v {
		v[i] = vs[i]
	}
	return predicate.Revision(func(s *sql.Selector) {
		// if not arguments were provided, append the FALSE constants,
		// since we can't apply "IN ()". This will make this predicate falsy.
		if len(v) == 0 {
			s.Where(sql.False())
			return
		}
		s.Where(sql.NotIn(s.C(FieldOperatorVersion), v...))
	})
}

// OperatorVersionGT applies the GT predicate on the "operator_version" field.
func OperatorVersionGT(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GT(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionGTE applies the GTE predicate on the "operator_version" field.
func OperatorVersionGTE(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.GTE(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionLT applies the LT predicate on the "operator_version" field.
func OperatorVersionLT(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LT(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionLTE applies the LTE predicate on the "operator_version" field.
func OperatorVersionLTE(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.LTE(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionContains applies the Contains predicate on the "operator_version" field.
func OperatorVersionContains(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.Contains(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionHasPrefix applies the HasPrefix predicate on the "operator_version" field.
func OperatorVersionHasPrefix(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.HasPrefix(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionHasSuffix applies the HasSuffix predicate on the "operator_version" field.
func OperatorVersionHasSuffix(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.HasSuffix(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionEqualFold applies the EqualFold predicate on the "operator_version" field.
func OperatorVersionEqualFold(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.EqualFold(s.C(FieldOperatorVersion), v))
	})
}

// OperatorVersionContainsFold applies the ContainsFold predicate on the "operator_version" field.
func OperatorVersionContainsFold(v string) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s.Where(sql.ContainsFold(s.C(FieldOperatorVersion), v))
	})
}

// And groups predicates with the AND operator between them.
func And(predicates ...predicate.Revision) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for _, p := range predicates {
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Or groups predicates with the OR operator between them.
func Or(predicates ...predicate.Revision) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		s1 := s.Clone().SetP(nil)
		for i, p := range predicates {
			if i > 0 {
				s1.Or()
			}
			p(s1)
		}
		s.Where(s1.P())
	})
}

// Not applies the not operator on the given predicate.
func Not(p predicate.Revision) predicate.Revision {
	return predicate.Revision(func(s *sql.Selector) {
		p(s.Not())
	})
}
