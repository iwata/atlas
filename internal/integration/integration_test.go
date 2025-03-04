// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"text/template"
	"time"

	"ariga.io/atlas/schemahcl"
	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	entsql "entgo.io/ent/dialect/sql"
	entschema "entgo.io/ent/dialect/sql/schema"
	"entgo.io/ent/entc/integration/ent"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	var dialect string
	flag.StringVar(&dialect, "dialect", "", "[mysql56, postgres10, tidb5, ...] what dialect (version) to test")
	flag.Parse()
	var dbs []io.Closer
	dbs = append(dbs, myInit(dialect)...)
	dbs = append(dbs, pgInit(dialect)...)
	dbs = append(dbs, tidbInit(dialect)...)
	dbs = append(dbs, crdbInit(dialect)...)
	defer func() {
		for _, db := range dbs {
			db.Close()
		}
	}()
	os.Exit(m.Run())
}

// T holds the elements common between dialect tests.
type T interface {
	testing.TB
	url() string
	driver() migrate.Driver
	revisionsStorage() migrate.RevisionReadWriter
	realm() *schema.Realm
	loadRealm() *schema.Realm
	users() *schema.Table
	loadUsers() *schema.Table
	posts() *schema.Table
	loadPosts() *schema.Table
	revisions() *schema.Table
	loadTable(string) *schema.Table
	dropTables(...string)
	dropSchemas(...string)
	migrate(...schema.Change)
	diff(*schema.Table, *schema.Table) []schema.Change
	applyHcl(spec string)
	applyRealmHcl(spec string)
}

func testAddDrop(t T) {
	usersT := t.users()
	postsT := t.posts()
	petsT := &schema.Table{
		Name:   "pets",
		Schema: usersT.Schema,
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "bigint"}}},
			{Name: "owner_id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "bigint"}, Null: true}},
		},
	}
	petsT.PrimaryKey = &schema.Index{Parts: []*schema.IndexPart{{C: postsT.Columns[0]}}}
	petsT.ForeignKeys = []*schema.ForeignKey{
		{Symbol: "owner_id", Table: petsT, Columns: petsT.Columns[1:], RefTable: usersT, RefColumns: usersT.Columns[:1]},
	}
	t.dropTables(postsT.Name, usersT.Name, petsT.Name)
	t.migrate(&schema.AddTable{T: petsT}, &schema.AddTable{T: usersT}, &schema.AddTable{T: postsT})
	ensureNoChange(t, usersT, petsT, postsT)
	t.migrate(&schema.DropTable{T: usersT}, &schema.DropTable{T: postsT}, &schema.DropTable{T: petsT})
	// Ensure the realm is empty.
	require.EqualValues(t, t.realm(), t.loadRealm())
}

func testRelation(t T) {
	usersT, postsT := t.users(), t.posts()
	t.dropTables(postsT.Name, usersT.Name)
	t.migrate(
		&schema.AddTable{T: usersT},
		&schema.AddTable{T: postsT},
	)
	ensureNoChange(t, postsT, usersT)
}

func testEntIntegration(t T, dialect string, db *sql.DB, opts ...entschema.MigrateOption) {
	ctx := context.Background()
	drv := entsql.OpenDB(dialect, db)
	client := ent.NewClient(ent.Driver(drv))
	require.NoError(t, client.Schema.Create(ctx, opts...))
	sanity(client)
	realm := t.loadRealm()
	ensureNoChange(t, realm.Schemas[0].Tables...)

	// Drop tables.
	changes := make([]schema.Change, len(realm.Schemas[0].Tables))
	for i, t := range realm.Schemas[0].Tables {
		changes[i] = &schema.DropTable{T: t}
	}
	t.migrate(changes...)

	// Add tables.
	for i, t := range realm.Schemas[0].Tables {
		changes[i] = &schema.AddTable{T: t}
	}
	t.migrate(changes...)
	ensureNoChange(t, realm.Schemas[0].Tables...)
	sanity(client)

	// Drop tables.
	for i, t := range realm.Schemas[0].Tables {
		changes[i] = &schema.DropTable{T: t}
	}
	t.migrate(changes...)
}

func testImplicitIndexes(t T, db *sql.DB) {
	const (
		name = "implicit_indexes"
		ddl  = "create table implicit_indexes(c1 int unique, c2 int unique, unique(c1,c2), unique(c2,c1))"
	)
	t.dropTables(name)
	_, err := db.Exec(ddl)
	require.NoError(t, err)
	current := t.loadTable(name)
	c1, c2 := schema.NewNullIntColumn("c1", "int"), schema.NewNullIntColumn("c2", "int")
	desired := schema.NewTable(name).
		AddColumns(c1, c2).
		AddIndexes(
			schema.NewUniqueIndex("").AddColumns(c1),
			schema.NewUniqueIndex("").AddColumns(c2),
			schema.NewUniqueIndex("").AddColumns(c1, c2),
			schema.NewUniqueIndex("").AddColumns(c2, c1),
		)
	changes := t.diff(current, desired)
	require.Empty(t, changes)
	desired.AddIndexes(
		schema.NewIndex("c1_key").AddColumns(c1),
		schema.NewIndex("c2_key").AddColumns(c2),
	)
	changes = t.diff(current, desired)
	require.NotEmpty(t, changes)
	t.migrate(&schema.ModifyTable{T: desired, Changes: changes})
	ensureNoChange(t, desired)
}

func testHCLIntegration(t T, full string, empty string) {
	t.applyHcl(full)
	users := t.loadUsers()
	posts := t.loadPosts()
	t.dropTables(users.Name, posts.Name)
	column, ok := users.Column("id")
	require.True(t, ok, "expected id column")
	require.Equal(t, "users", users.Name)
	column, ok = posts.Column("author_id")
	require.Equal(t, "author_id", column.Name)
	t.applyHcl(empty)
	require.Empty(t, t.realm().Schemas[0].Tables)
}

func testCLISchemaInspect(t T, h string, dsn string, eval schemahcl.Evaluator, args ...string) {
	err := initCLI()
	require.NoError(t, err)
	t.dropTables("users")
	var expected schema.Schema
	err = evalBytes([]byte(h), &expected, eval)
	require.NoError(t, err)
	t.applyHcl(h)
	runArgs := []string{
		"run", "ariga.io/atlas/cmd/atlas",
		"schema",
		"inspect",
		"-d",
		dsn,
	}
	runArgs = append(runArgs, args...)
	cmd := exec.Command("go", runArgs...)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	require.NoError(t, cmd.Run(), stderr.String())
	var actual schema.Schema
	err = evalBytes(stdout.Bytes(), &actual, eval)
	require.NoError(t, err)
	require.Empty(t, stderr.String())
	require.Equal(t, expected, actual)
}

func testCLISchemaInspectEnv(t T, h string, env string, eval schemahcl.Evaluator) {
	err := initCLI()
	require.NoError(t, err)
	t.dropTables("users")
	var expected schema.Schema
	err = evalBytes([]byte(h), &expected, eval)
	require.NoError(t, err)
	t.applyHcl(h)
	cmd := exec.Command("go", "run", "ariga.io/atlas/cmd/atlas",
		"schema",
		"inspect",
		"--env",
		env,
	)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	require.NoError(t, cmd.Run(), stderr.String())
	var actual schema.Schema
	err = evalBytes(stdout.Bytes(), &actual, eval)
	require.NoError(t, err)
	require.Empty(t, stderr.String())
	require.Equal(t, expected, actual)
}

// initOnce controls that the cli will only be built once.
var initOnce sync.Once

func initCLI() error {
	var err error
	initOnce.Do(func() {
		err = exec.Command("go", "run", "-mod=mod", "ariga.io/atlas/cmd/atlas").Run()
	})
	return err
}

func testCLIMultiSchemaApply(t T, h string, dsn string, schemas []string, eval schemahcl.Evaluator) {
	err := initCLI()
	f := filepath.Join(t.TempDir(), "schema.hcl")
	err = ioutil.WriteFile(f, []byte(h), 0644)
	require.NoError(t, err)
	require.NoError(t, err)
	var expected schema.Realm
	err = evalBytes([]byte(h), &expected, eval)
	require.NoError(t, err)
	cmd := exec.Command("go", "run", "ariga.io/atlas/cmd/atlas",
		"schema",
		"apply",
		"-f",
		f,
		"-d",
		dsn,
		"-s",
		strings.Join(schemas, ","),
	)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	defer stdin.Close()
	_, err = io.WriteString(stdin, "\n")
	require.NoError(t, cmd.Run(), stderr.String())
	require.Contains(t, stdout.String(), `-- Add new schema named "test2"`)
}

func testCLIMultiSchemaInspect(t T, h string, dsn string, schemas []string, eval schemahcl.Evaluator) {
	err := initCLI()
	require.NoError(t, err)
	var expected schema.Realm
	err = evalBytes([]byte(h), &expected, eval)
	require.NoError(t, err)
	t.applyRealmHcl(h)
	cmd := exec.Command("go", "run", "ariga.io/atlas/cmd/atlas",
		"schema",
		"inspect",
		"-d",
		dsn,
		"-s",
		strings.Join(schemas, ","),
	)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	require.NoError(t, cmd.Run(), stderr.String())
	var actual schema.Realm
	err = evalBytes(stdout.Bytes(), &actual, eval)
	require.NoError(t, err)
	require.Empty(t, stderr.String())
	require.Equal(t, expected, actual)
}

func testCLISchemaApply(t T, h string, dsn string, args ...string) {
	err := initCLI()
	require.NoError(t, err)
	t.dropTables("users")
	f := filepath.Join(t.TempDir(), "schema.hcl")
	err = ioutil.WriteFile(f, []byte(h), 0644)
	require.NoError(t, err)
	runArgs := []string{
		"run", "ariga.io/atlas/cmd/atlas",
		"schema",
		"apply",
		"-u",
		dsn,
		"-f",
		f,
		"--dev-url",
		dsn,
	}
	runArgs = append(runArgs, args...)
	cmd := exec.Command("go", runArgs...)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	defer stdin.Close()
	_, err = io.WriteString(stdin, "\n")
	require.NoError(t, err)
	require.NoError(t, cmd.Run(), stderr.String(), stdout.String())
	require.Empty(t, stderr.String(), stderr.String())
	require.Contains(t, stdout.String(), "-- Planned")
	u := t.loadUsers()
	require.NotNil(t, u)
}

func testCLISchemaApplyDry(t T, h string, dsn string) {
	err := initCLI()
	require.NoError(t, err)
	t.dropTables("users")
	f := filepath.Join(t.TempDir(), "schema.hcl")
	err = ioutil.WriteFile(f, []byte(h), 0644)
	require.NoError(t, err)
	cmd := exec.Command("go", "run", "ariga.io/atlas/cmd/atlas",
		"schema",
		"apply",
		"-d",
		dsn,
		"-f",
		f,
		"--dry-run",
	)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)
	defer stdin.Close()
	_, err = io.WriteString(stdin, "\n")
	require.NoError(t, err)
	require.NoError(t, cmd.Run(), stderr.String(), stdout.String())
	require.Empty(t, stderr.String(), stderr.String())
	require.Contains(t, stdout.String(), "-- Planned")
	require.NotContains(t, stdout.String(), "Are you sure?", "dry run should not prompt")
	realm := t.loadRealm()
	_, ok := realm.Schemas[0].Table("users")
	require.False(t, ok, "expected users table not to be created")
}

func testCLISchemaApplyAutoApprove(t T, h string, dsn string, args ...string) {
	err := initCLI()
	require.NoError(t, err)
	t.dropTables("users")
	f := filepath.Join(t.TempDir(), "schema.hcl")
	err = ioutil.WriteFile(f, []byte(h), 0644)
	require.NoError(t, err)
	runArgs := []string{
		"run", "ariga.io/atlas/cmd/atlas",
		"schema",
		"apply",
		"-d",
		dsn,
		"-f",
		f,
		"--auto-approve",
	}
	runArgs = append(runArgs, args...)
	cmd := exec.Command("go", runArgs...)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	require.NoError(t, err)
	require.NoError(t, cmd.Run(), stderr.String(), stdout.String())
	require.Empty(t, stderr.String(), stderr.String())
	require.Contains(t, stdout.String(), "-- Planned")
	u := t.loadUsers()
	require.NotNil(t, u)
}

func testCLISchemaDiff(t T, dsn string) {
	err := initCLI()

	require.NoError(t, err)
	t.dropTables("users")
	cmd := exec.Command("go", "run", "ariga.io/atlas/cmd/atlas",
		"schema",
		"diff",
		"--from",
		dsn,
		"--to",
		dsn,
	)
	stdout, stderr := bytes.NewBuffer(nil), bytes.NewBuffer(nil)
	cmd.Stderr = stderr
	cmd.Stdout = stdout
	require.NoError(t, cmd.Run(), stderr.String(), stdout.String())
	require.Empty(t, stderr.String(), stderr.String())
	require.Contains(t, stdout.String(), "Schemas are synced, no changes to be made.")
}

func ensureNoChange(t T, tables ...*schema.Table) {
	realm := t.loadRealm()
	require.Equal(t, len(realm.Schemas[0].Tables), len(tables))
	for i := range tables {
		tt, ok := realm.Schemas[0].Table(tables[i].Name)
		require.True(t, ok)
		changes := t.diff(tt, tables[i])
		require.Emptyf(t, changes, "changes should be empty for table %s, but instead was %#v", tt.Name, changes)
	}
}

func sanity(c *ent.Client) {
	ctx := context.Background()
	u := c.User.Create().
		SetName("foo").
		SetAge(20).
		AddPets(
			c.Pet.Create().SetName("pedro").SaveX(ctx),
			c.Pet.Create().SetName("xabi").SaveX(ctx),
		).
		AddFiles(
			c.File.Create().SetName("a").SetSize(10).SaveX(ctx),
			c.File.Create().SetName("b").SetSize(20).SaveX(ctx),
		).
		SaveX(ctx)
	c.Group.Create().
		SetName("Github").
		SetExpire(time.Now()).
		AddUsers(u).
		SetInfo(c.GroupInfo.Create().SetDesc("desc").SaveX(ctx)).
		SaveX(ctx)
}

func testAdvisoryLock(t *testing.T, l schema.Locker) {
	t.Run("One", func(t *testing.T) {
		unlock, err := l.Lock(context.Background(), "migrate", 0)
		require.NoError(t, err)
		_, err = l.Lock(context.Background(), "migrate", 0)
		require.Equal(t, schema.ErrLocked, err)
		require.NoError(t, unlock())
	})
	t.Run("Multi", func(t *testing.T) {
		var unlocks []schema.UnlockFunc
		for _, name := range []string{"a", "b", "c"} {
			unlock, err := l.Lock(context.Background(), name, 0)
			require.NoError(t, err)
			unlocks = append(unlocks, unlock)
		}
		for _, unlock := range unlocks {
			require.NoError(t, unlock())
		}
	})
}

func testExecutor(t T) {
	usersT, postsT := t.users(), t.posts()
	petsT := &schema.Table{
		Name:   "pets",
		Schema: usersT.Schema,
		Columns: []*schema.Column{
			{Name: "id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "bigint"}}},
			{Name: "owner_id", Type: &schema.ColumnType{Type: &schema.IntegerType{T: "bigint"}, Null: true}},
		},
	}
	petsT.PrimaryKey = &schema.Index{Parts: []*schema.IndexPart{{C: postsT.Columns[0]}}}
	petsT.ForeignKeys = []*schema.ForeignKey{
		{Symbol: "owner_id", Table: petsT, Columns: petsT.Columns[1:], RefTable: usersT, RefColumns: usersT.Columns[:1]},
	}

	t.dropTables(petsT.Name, postsT.Name, usersT.Name)
	t.Cleanup(func() {
		t.revisionsStorage().(*rrw).clean()
	})

	dir, err := migrate.NewLocalDir(t.TempDir())
	require.NoError(t, err)
	f, err := migrate.NewTemplateFormatter(
		template.Must(template.New("").Parse("{{ .Name }}.sql")),
		template.Must(template.New("").Parse(
			`{{ range .Changes }}{{ with .Comment }}-- {{ println . }}{{ end }}{{ printf "%s;\n" .Cmd }}{{ end }}`,
		)),
	)
	require.NoError(t, err)
	pl := migrate.NewPlanner(t.driver(), dir, migrate.WithFormatter(f))
	require.NoError(t, err)

	require.NoError(t, pl.WritePlan(plan(t, "1_users", &schema.AddTable{T: usersT})))
	require.NoError(t, pl.WritePlan(plan(t, "2_posts", &schema.AddTable{T: postsT})))
	require.NoError(t, pl.WritePlan(plan(t, "3_pets", &schema.AddTable{T: petsT})))

	ex, err := migrate.NewExecutor(t.driver(), dir, t.revisionsStorage())
	require.NoError(t, err)
	require.NoError(t, ex.ExecuteN(context.Background(), 2)) // usersT and postsT
	require.Len(t, *t.revisionsStorage().(*rrw), 2)
	ensureNoChange(t, postsT, usersT)
	require.NoError(t, ex.ExecuteN(context.Background(), 1)) // petsT
	require.Len(t, *t.revisionsStorage().(*rrw), 3)
	ensureNoChange(t, petsT, postsT, usersT)

	require.ErrorIs(t, ex.ExecuteN(context.Background(), 1), migrate.ErrNoPendingFiles)
}

func plan(t T, name string, changes ...schema.Change) *migrate.Plan {
	p, err := t.driver().PlanChanges(context.Background(), name, changes)
	require.NoError(t, err)
	return p
}

type rrw migrate.Revisions

func (r *rrw) WriteRevision(_ context.Context, rev *migrate.Revision) error {
	for i, rev2 := range *r {
		if rev2.Version == rev.Version {
			(*r)[i] = rev
			return nil
		}
	}
	*r = append(*r, rev)
	return nil
}

func (r *rrw) ReadRevisions(context.Context) (migrate.Revisions, error) {
	return migrate.Revisions(*r), nil
}

func (r *rrw) clean() {
	*r = rrw(migrate.Revisions{})
}

func buildCmd(t *testing.T) (string, error) {
	td := t.TempDir()
	if b, err := exec.Command("go", "build", "-o", td, "ariga.io/atlas/cmd/atlas").CombinedOutput(); err != nil {
		return "", fmt.Errorf("%w: %s", err, b)
	}
	return filepath.Join(td, "atlas"), nil
}

func evalBytes(b []byte, v interface{}, ev schemahcl.Evaluator) error {
	p := hclparse.NewParser()
	if _, diag := p.ParseHCL(b, ""); diag.HasErrors() {
		return diag
	}
	return ev.Eval(p, v, nil)
}
