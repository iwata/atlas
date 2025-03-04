// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package cmdapi

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"ariga.io/atlas/cmd/atlas/internal/ci"
	entmigrate "ariga.io/atlas/cmd/atlas/internal/migrate"
	"ariga.io/atlas/sql/migrate"
	"ariga.io/atlas/sql/schema"
	"ariga.io/atlas/sql/sqlcheck"
	"ariga.io/atlas/sql/sqlcheck/datadepend"
	"ariga.io/atlas/sql/sqlcheck/destructive"
	"ariga.io/atlas/sql/sqlclient"
	"ariga.io/atlas/sql/sqltool"
	"github.com/fatih/color"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	migrateFlagURL             = "url"
	migrateFlagDevURL          = "dev-url"
	migrateFlagDir             = "dir"
	migrateFlagForce           = "force"
	migrateFlagFormat          = "format"
	migrateFlagLog             = "log"
	migrateFlagRevisionsSchema = "revisions-schema"
	migrateFlagDryRun          = "dry-run"
	migrateFlagTo              = "to"
	migrateFlagSchema          = "schema"
	migrateDiffFlagVerbose     = "verbose"
	migrateLintLatest          = "latest"
	migrateLintGitDir          = "git-dir"
	migrateLintGitBase         = "git-base"
)

var (
	// MigrateFlags are the flags used in MigrateCmd (and sub-commands).
	MigrateFlags struct {
		URL            string
		DirURL         string
		DevURL         string
		ToURL          string
		Schemas        []string
		Format         string
		LogFormat      string
		RevisionSchema string
		DryRun         bool
		Force          bool
		Verbose        bool
		Lint           struct {
			Format  string // log formatting
			Latest  uint   // latest N migration files
			GitDir  string // repository working dir
			GitBase string // branch name to compare with
		}
	}
	// MigrateCmd represents the migrate command. It wraps several other sub-commands.
	MigrateCmd = &cobra.Command{
		Use:   "migrate",
		Short: "Manage versioned migration files",
		Long:  "'atlas migrate' wraps several sub-commands for migration management.",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := migrateFlagsFromEnv(cmd, nil); err != nil {
				return err
			}
			// Migrate commands will not run on a broken migration directory, unless the force flag is given.
			if !MigrateFlags.Force {
				dir, err := dir()
				if err != nil {
					return err
				}
				if err := migrate.Validate(dir); err != nil {
					fmt.Fprintf(cmd.OutOrStderr(), `You have a checksum error in your migration directory.
This happens if you manually create or edit a migration file.
Please check your migration files and run

'atlas migrate hash --force'

to re-hash the contents and resolve the error

`)
					cmd.SilenceUsage = true
					return err
				}
			}
			return nil
		},
	}
	// MigrateApplyCmd represents the 'atlas migrate apply' subcommand.
	MigrateApplyCmd = &cobra.Command{
		Use:   "apply [flags] [count]",
		Short: "Applies pending migration files on the connected database.",
		Long: `'atlas migrate apply' reads the migration state of the connected database and computes what migrations are pending.
It then attempts to apply the pending migration files in the correct order onto the database. 
The first argument denotes the maximum number of migration files to apply.
As a safety measure 'atlas migrate apply' will abort with an error, if:
  - the migration directory is not in sync with the 'atlas.sum' file
  - the migration and database history do not match each other

If run with the "--dry-run" flag, atlas will not execute any SQL.`,
		Example: `  atlas migrate apply -u mysql://user:pass@localhost:3306/dbname
  atlas migrate apply --dir file:///path/to/migration/directory --url mysql://user:pass@localhost:3306/dbname 1
  atlas migrate apply --env dev 1
  atlas migrate apply --dry-run --env dev 1`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: migrateFlagsFromEnv,
		RunE:    CmdMigrateApplyRun,
	}
	// MigrateDiffCmd represents the 'atlas migrate diff' subcommand.
	MigrateDiffCmd = &cobra.Command{
		Use:   "diff [flags] [name]",
		Short: "Compute the diff between the migration directory and a desired state and create a new migration file.",
		Long: `'atlas migrate diff' uses the dev-database to re-run all migration files in the migration directory, compares
it to a given desired state and create a new migration file containing SQL statements to migrate the migration
directory state to the desired schema. The desired state can be another connected database or an HCL file.`,
		Example: `  atlas migrate diff --dev-url mysql://user:pass@localhost:3306/dev --to file://atlas.hcl
  atlas migrate diff --dev-url mysql://user:pass@localhost:3306/dev --to file://atlas.hcl add_users_table
  atlas migrate diff --dev-url mysql://user:pass@localhost:3306/dev --to mysql://user:pass@localhost:3306/dbname
  atlas migrate diff --env dev`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: migrateFlagsFromEnv,
		RunE:    CmdMigrateDiffRun,
	}
	// MigrateHashCmd represents the 'atlas migrate hash' command.
	MigrateHashCmd = &cobra.Command{
		Use:   "hash",
		Short: "Hash (re-)creates an integrity hash file for the migration directory.",
		Long: `'atlas migrate hash' computes the integrity hash sum of the migration directory and stores it in the atlas.sum file.
This command should be used whenever a manual change in the migration directory was made.`,
		Example: `  atlas migrate hash --force`,
		PreRunE: migrateFlagsFromEnv,
		RunE:    CmdMigrateHashRun,
	}
	// MigrateNewCmd represents the 'atlas migrate new' command.
	MigrateNewCmd = &cobra.Command{
		Use:     "new [name]",
		Short:   "Creates a new empty migration file in the migration directory.",
		Long:    `'atlas migrate new' creates a new migration according to the configured formatter without any statements in it.`,
		Example: `  atlas migrate new my-new-migration`,
		Args:    cobra.MaximumNArgs(1),
		PreRunE: migrateFlagsFromEnv,
		RunE:    CmdMigrateNewRun,
	}
	// MigrateValidateCmd represents the 'atlas migrate validate' command.
	MigrateValidateCmd = &cobra.Command{
		Use:   "validate [flags]",
		Short: "Validates the migration directories checksum and SQL statements.",
		Long: `'atlas migrate validate' computes the integrity hash sum of the migration directory and compares it to the
atlas.sum file. If there is a mismatch it will be reported. If the --dev-url flag is given, the migration
files are executed on the connected database in order to validate SQL semantics.`,
		Example: `  atlas migrate validate
  atlas migrate validate --dir file:///path/to/migration/directory
  atlas migrate validate --dir file:///path/to/migration/directory --dev-url mysql://user:pass@localhost:3306/dev
  atlas migrate validate --env dev`,
		PreRunE: migrateFlagsFromEnv,
		RunE:    CmdMigrateValidateRun,
	}
	// MigrateLintCmd represents the 'atlas migrate Lint' command.
	MigrateLintCmd = &cobra.Command{
		Use:   "lint",
		Short: "Run analysis on the migration directory",
		Example: `  atlas migrate lint --env dev
  atlas migrate lint --dir file:///path/to/migration/directory --dev-url mysql://root:pass@localhost:3306 --latest 1
  atlas migrate lint --dir file:///path/to/migration/directory --dev-url mysql://root:pass@localhost:3306 --git-base master
  atlas migrate lint --dir file:///path/to/migration/directory --dev-url mysql://root:pass@localhost:3306 --log '{{ json .Files }}'`,
		// Override the parent 'migrate' pre-run function to allow executing
		// 'migrate lint' on directories that are not maintained by Atlas.
		PersistentPreRunE: migrateFlagsFromEnv,
		RunE:              CmdMigrateLintRun,
	}
)

func init() {
	// Add sub-commands.
	Root.AddCommand(MigrateCmd)
	MigrateCmd.AddCommand(MigrateApplyCmd)
	MigrateCmd.AddCommand(MigrateDiffCmd)
	MigrateCmd.AddCommand(MigrateHashCmd)
	MigrateCmd.AddCommand(MigrateNewCmd)
	MigrateCmd.AddCommand(MigrateValidateCmd)
	MigrateCmd.AddCommand(MigrateLintCmd)
	// Reusable flags.
	urlFlag := func(f *string, name, short string, set *pflag.FlagSet) {
		set.StringVarP(f, name, short, "", "[driver://username:password@address/dbname?param=value] select a database using the URL format")
	}
	// Global flags.
	MigrateCmd.PersistentFlags().StringVarP(&MigrateFlags.DirURL, migrateFlagDir, "", "file://migrations", "select migration directory using URL format")
	MigrateCmd.PersistentFlags().StringSliceVarP(&MigrateFlags.Schemas, migrateFlagSchema, "", nil, "set schema names")
	MigrateCmd.PersistentFlags().StringVarP(&MigrateFlags.Format, migrateFlagFormat, "", formatAtlas, "set migration file format")
	MigrateCmd.PersistentFlags().BoolVarP(&MigrateFlags.Force, migrateFlagForce, "", false, "force a command to run on a broken migration directory state")
	MigrateCmd.PersistentFlags().SortFlags = false
	// Apply flags.
	MigrateApplyCmd.Flags().StringVarP(&MigrateFlags.LogFormat, migrateFlagLog, "", logFormatTTY, "log format to use")
	MigrateApplyCmd.Flags().StringVarP(&MigrateFlags.RevisionSchema, migrateFlagRevisionsSchema, "", entmigrate.DefaultRevisionSchema, "schema name where the revisions table resides")
	MigrateApplyCmd.Flags().BoolVarP(&MigrateFlags.DryRun, migrateFlagDryRun, "", false, "do not actually execute any SQL but show it on screen")
	urlFlag(&MigrateFlags.URL, migrateFlagURL, "u", MigrateApplyCmd.Flags())
	cobra.CheckErr(MigrateApplyCmd.MarkFlagRequired(migrateFlagURL))
	// Diff flags.
	urlFlag(&MigrateFlags.DevURL, migrateFlagDevURL, "", MigrateDiffCmd.Flags())
	urlFlag(&MigrateFlags.ToURL, migrateFlagTo, "", MigrateDiffCmd.Flags())
	MigrateDiffCmd.Flags().BoolVarP(&MigrateFlags.Verbose, migrateDiffFlagVerbose, "", false, "enable verbose logging")
	MigrateDiffCmd.Flags().SortFlags = false
	cobra.CheckErr(MigrateDiffCmd.MarkFlagRequired(migrateFlagDevURL))
	cobra.CheckErr(MigrateDiffCmd.MarkFlagRequired(migrateFlagTo))
	// Validate flags.
	urlFlag(&MigrateFlags.DevURL, migrateFlagDevURL, "", MigrateValidateCmd.Flags())
	// Lint flags.
	urlFlag(&MigrateFlags.DevURL, migrateFlagDevURL, "", MigrateLintCmd.Flags())
	MigrateLintCmd.PersistentFlags().StringVarP(&MigrateFlags.Lint.Format, migrateFlagLog, "", "", "custom logging using a Go template")
	MigrateLintCmd.PersistentFlags().UintVarP(&MigrateFlags.Lint.Latest, migrateLintLatest, "", 0, "run analysis on the latest N migration files")
	MigrateLintCmd.PersistentFlags().StringVarP(&MigrateFlags.Lint.GitBase, migrateLintGitBase, "", "", "run analysis against the base Git branch")
	MigrateLintCmd.PersistentFlags().StringVarP(&MigrateFlags.Lint.GitDir, migrateLintGitDir, "", ".", "path to the repository working directory")
	cobra.CheckErr(MigrateLintCmd.MarkFlagRequired(migrateFlagDevURL))
	receivesEnv(MigrateCmd)
}

// CmdMigrateApplyRun is the command executed when running the CLI with 'migrate apply' args.
func CmdMigrateApplyRun(cmd *cobra.Command, args []string) error {
	var (
		n   int
		err error
	)
	if len(args) > 0 {
		n, err = strconv.Atoi(args[0])
		if err != nil {
			return err
		}
	}
	// Open the migration directory.
	dir, err := dir()
	if err != nil {
		return err
	}
	// Open a client to the database.
	c, err := sqlclient.Open(cmd.Context(), MigrateFlags.URL)
	if err != nil {
		return err
	}
	// Get the correct log format and destination. Currently, only os.Stdout is supported.
	l, err := logFormat(cmd.OutOrStdout())
	if err != nil {
		return err
	}
	// Currently, only in DB revisions are supported.
	var rrw migrate.RevisionReadWriter
	rrw, err = entmigrate.NewEntRevisions(c, []entmigrate.Option{entmigrate.WithSchema(MigrateFlags.RevisionSchema)}...)
	if err != nil {
		return err
	}
	if err := rrw.(*entmigrate.EntRevisions).Init(cmd.Context()); err != nil {
		return err
	}
	defer func(rrw *entmigrate.EntRevisions, ctx context.Context) {
		if err2 := rrw.Flush(ctx); err2 != nil {
			if err != nil {
				err = fmt.Errorf("%v: %w", err2, err)
			} else {
				err = err2
			}
		}
	}(rrw.(*entmigrate.EntRevisions), cmd.Context())
	var (
		drv migrate.Driver
		tx  *sqlclient.TxClient
	)
	if MigrateFlags.DryRun {
		drv = &dryRunDriver{c.Driver}
		rrw = &dryRunRevisions{rrw}
	} else {
		// Wrap the whole execution in one transaction. This behaviour will change once
		// there are insights about migration files available.
		tx, err = c.Tx(cmd.Context(), nil)
		if err != nil {
			return err
		}
		drv = tx.Driver
	}
	// Get the executor.
	ex, err := migrate.NewExecutor(drv, dir, rrw, migrate.WithLogger(l))
	if err != nil {
		return err
	}
	if err = ex.ExecuteN(cmd.Context(), n); errors.Is(err, migrate.ErrNoPendingFiles) {
		cmd.Println("The migration directory is synced with the database, no migration files to execute")
	} else if err != nil {
		if !MigrateFlags.DryRun {
			if err2 := tx.Rollback(); err2 != nil {
				err = fmt.Errorf("%v: %w", err2, err)
			}
		}
		return err
	}
	if !MigrateFlags.DryRun {
		return tx.Commit()
	}
	return nil
}

// CmdMigrateDiffRun is the command executed when running the CLI with 'migrate diff' args.
func CmdMigrateDiffRun(cmd *cobra.Command, args []string) error {
	// Open a dev driver.
	dev, err := sqlclient.Open(cmd.Context(), MigrateFlags.DevURL)
	if err != nil {
		return err
	}
	defer dev.Close()
	// Acquire a lock.
	if l, ok := dev.Driver.(schema.Locker); ok {
		unlock, err := l.Lock(cmd.Context(), "atlas_migrate_diff", 0)
		if err != nil {
			return err
		}
		// If unlocking fails notify the user about it.
		defer cobra.CheckErr(unlock())
	}
	// Open the migration directory.
	dir, err := dir()
	if err != nil {
		return err
	}
	// Get a state reader for the desired state.
	desired, err := to(cmd.Context(), dev)
	if src, ok := desired.(io.Closer); ok {
		defer src.Close()
	}
	if err != nil {
		return err
	}
	f, err := formatter()
	if err != nil {
		return err
	}
	// Plan the changes and create a new migration file.
	pl := migrate.NewPlanner(dev.Driver, dir, migrate.WithFormatter(f))
	var name string
	if len(args) > 0 {
		name = args[0]
	}
	switch plan, err := pl.Plan(cmd.Context(), name, desired); {
	case errors.Is(err, migrate.ErrNoPlan):
		cmd.Println("The migration directory is synced with the desired state, no changes to be made")
		return nil
	case err != nil:
		return err
	default:
		// Write the plan to a new file.
		return pl.WritePlan(plan)
	}
}

// CmdMigrateHashRun is the command executed when running the CLI with 'migrate hash' args.
func CmdMigrateHashRun(*cobra.Command, []string) error {
	dir, err := dir()
	if err != nil {
		return err
	}
	sum, err := migrate.HashSum(dir)
	if err != nil {
		return err
	}
	return migrate.WriteSumFile(dir, sum)
}

// CmdMigrateNewRun is the command executed when running the CLI with 'migrate new' args.
func CmdMigrateNewRun(_ *cobra.Command, args []string) error {
	dir, err := dir()
	if err != nil {
		return err
	}
	f, err := formatter()
	if err != nil {
		return err
	}
	var name string
	if len(args) > 0 {
		name = args[0]
	}
	return migrate.NewPlanner(nil, dir, migrate.WithFormatter(f)).WritePlan(&migrate.Plan{Name: name})
}

// CmdMigrateValidateRun is the command executed when running the CLI with 'migrate validate' args.
func CmdMigrateValidateRun(cmd *cobra.Command, _ []string) error {
	// Validating the integrity is done by the PersistentPreRun already.
	if MigrateFlags.DevURL == "" {
		// If there is no --dev-url given do not attempt to replay the migration directory.
		return nil
	}
	// Open a client for the dev-db.
	dev, err := sqlclient.Open(cmd.Context(), MigrateFlags.DevURL)
	if err != nil {
		return err
	}
	defer dev.Close()
	// Currently, only our own migration file format is supported.
	dir, err := dir()
	if err != nil {
		return err
	}
	ex, err := migrate.NewExecutor(dev.Driver, dir, migrate.NopRevisionReadWriter{})
	if err != nil {
		return err
	}
	if _, err := ex.ReadState(cmd.Context()); err != nil && !errors.Is(err, migrate.ErrNoPendingFiles) {
		return fmt.Errorf("replaying the migration directory: %w", err)
	}
	return nil
}

// CmdMigrateLintRun is the command executed when running the CLI with 'migrate lint' args.
func CmdMigrateLintRun(cmd *cobra.Command, _ []string) error {
	dev, err := sqlclient.Open(cmd.Context(), MigrateFlags.DevURL)
	if err != nil {
		return err
	}
	defer dev.Close()
	dir, err := dir()
	if err != nil {
		return err
	}
	var (
		detect ci.ChangeDetector
		local  = dir.(*migrate.LocalDir)
	)
	switch {
	case MigrateFlags.Lint.Latest == 0 && MigrateFlags.Lint.GitBase == "":
		return fmt.Errorf("--%s or --%s is required", migrateLintLatest, migrateLintGitBase)
	case MigrateFlags.Lint.Latest > 0 && MigrateFlags.Lint.GitBase != "":
		return fmt.Errorf("--%s and --%s are mutually exclusive", migrateLintLatest, migrateLintGitBase)
	case MigrateFlags.Lint.Latest > 0:
		detect = ci.LatestChanges(local, int(MigrateFlags.Lint.Latest))
	case MigrateFlags.Lint.GitBase != "":
		detect, err = ci.NewGitChangeDetector(
			local,
			ci.WithWorkDir(MigrateFlags.Lint.GitDir),
			ci.WithBase(MigrateFlags.Lint.GitBase),
			ci.WithMigrationsPath(local.Path()),
		)
		if err != nil {
			return err
		}
	}
	format := ci.DefaultTemplate
	if f := MigrateFlags.Lint.Format; f != "" {
		format, err = template.New("format").Funcs(ci.TemplateFuncs).Parse(f)
		if err != nil {
			return fmt.Errorf("parse log format: %w", err)
		}
	}
	r := &ci.Runner{
		Dev:            dev,
		Dir:            local,
		ChangeDetector: detect,
		ReportWriter: &ci.TemplateWriter{
			T: format,
			W: cmd.OutOrStdout(),
		},
		Analyzer: sqlcheck.Analyzers{
			datadepend.New(datadepend.Options{}),
			destructive.New(destructive.Options{}),
		},
	}
	return r.Run(cmd.Context())
}

// dir returns a migrate.Dir to use as migration directory. For now only local directories are supported.
func dir() (migrate.Dir, error) {
	parts := strings.SplitN(MigrateFlags.DirURL, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid dir url %q", MigrateFlags.DirURL)
	}
	if parts[0] != "file" {
		return nil, fmt.Errorf("unsupported driver %q", parts[0])
	}
	return migrate.NewLocalDir(parts[1])
}

// to returns a migrate.StateReader for the given to flag.
func to(ctx context.Context, client *sqlclient.Client) (migrate.StateReader, error) {
	parts := strings.SplitN(MigrateFlags.ToURL, "://", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid driver url %q", MigrateFlags.ToURL)
	}
	schemas := MigrateFlags.Schemas
	switch parts[0] {
	case "file": // hcl file
		realm := &schema.Realm{}
		parsed, err := parseHCLPaths(parts[1])
		if err != nil {
			return nil, err
		}
		if err := client.Eval(parsed, realm, nil); err != nil {
			return nil, err
		}
		if len(schemas) > 0 {
			// Validate all schemas in file were selected by user.
			sm := make(map[string]bool, len(schemas))
			for _, s := range schemas {
				sm[s] = true
			}
			for _, s := range realm.Schemas {
				if !sm[s.Name] {
					return nil, fmt.Errorf("schema %q from file %q is not requested (all schemas in HCL must be requested)", s.Name, parts[1])
				}
			}
		}
		if norm, ok := client.Driver.(schema.Normalizer); ok {
			realm, err = norm.NormalizeRealm(ctx, realm)
			if err != nil {
				return nil, err
			}
		}
		return migrate.Realm(realm), nil
	default: // database connection
		client, err := sqlclient.Open(ctx, MigrateFlags.ToURL)
		if err != nil {
			return nil, err
		}
		if client.URL.Schema != "" {
			schemas = append(schemas, client.URL.Schema)
		}
		return struct {
			migrate.StateReader
			io.Closer
		}{
			Closer:      client,
			StateReader: migrate.Conn(client, &schema.InspectRealmOption{Schemas: schemas}),
		}, nil
	}
}

// parseHCL paths parses the HCL files in the given paths. If a path represents a directory,
// its direct descendants will be considered, skipping any subdirectories. If a project file
// is present in the input paths, an error is returned.
func parseHCLPaths(paths ...string) (*hclparse.Parser, error) {
	p := hclparse.NewParser()
	for _, path := range paths {
		switch stat, err := os.Stat(path); {
		case err != nil:
			return nil, err
		case stat.IsDir():
			dir, err := os.ReadDir(path)
			if err != nil {
				return nil, err
			}
			for _, f := range dir {
				// Skip nested dirs.
				if f.IsDir() {
					continue
				}
				if err := mayParse(p, filepath.Join(path, f.Name())); err != nil {
					return nil, err
				}
			}
		default:
			if err := mayParse(p, path); err != nil {
				return nil, err
			}
		}
	}
	if len(p.Files()) == 0 {
		return nil, fmt.Errorf("no schema files found in: %s", paths)
	}
	return p, nil
}

// mayParse will parse the file in path if it is an HCL file. If the file is an Atlas
// project file an error is returned.
func mayParse(p *hclparse.Parser, path string) error {
	if n := filepath.Base(path); filepath.Ext(n) != ".hcl" {
		return nil
	}
	switch f, diag := p.ParseHCLFile(path); {
	case diag.HasErrors():
		return diag
	case isProjectFile(f):
		return fmt.Errorf("cannot parse project file %q as a schema file", path)
	default:
		return nil
	}
}

func isProjectFile(f *hcl.File) bool {
	for _, blk := range f.Body.(*hclsyntax.Body).Blocks {
		if blk.Type == "env" {
			return true
		}
	}
	return false
}

const (
	formatAtlas         = "atlas"
	formatGolangMigrate = "golang-migrate"
	formatGoose         = "goose"
	formatFlyway        = "flyway"
	formatLiquibase     = "liquibase"
	formatDbmate        = "dbmate"
)

func formatter() (migrate.Formatter, error) {
	switch MigrateFlags.Format {
	case formatAtlas:
		return migrate.DefaultFormatter, nil
	case formatGolangMigrate:
		return sqltool.GolangMigrateFormatter, nil
	case formatGoose:
		return sqltool.GooseFormatter, nil
	case formatFlyway:
		return sqltool.FlywayFormatter, nil
	case formatLiquibase:
		return sqltool.LiquibaseFormatter, nil
	case formatDbmate:
		return sqltool.DbmateFormatter, nil
	default:
		return nil, fmt.Errorf("unknown format %q", MigrateFlags.Format)
	}
}

const (
	logFormatTTY = "tty"
)

// LogTTY is a migrate.Logger that pretty prints execution progress.
// If the connected out is not a tty, it will fall back to a non-colorful output.
type LogTTY struct {
	out         io.Writer
	start       time.Time
	fileStart   time.Time
	fileCounter int
	stmtCounter int
}

var (
	cyan    = color.CyanString
	yellow  = color.YellowString
	dash    = yellow("--")
	arr     = cyan("->")
	indent2 = strings.Repeat(" ", 2)
	indent4 = strings.Repeat(indent2, 2)
)

// Log implements the migrate.Logger interface.
func (l *LogTTY) Log(e migrate.LogEntry) {
	switch e := e.(type) {
	case migrate.LogExecution:
		l.start = time.Now()
		fmt.Fprintf(l.out, "Migrating to version %v", cyan(e.To))
		if e.From != "" {
			fmt.Fprintf(l.out, " from %v", cyan(e.From))
		}
		fmt.Fprintf(l.out, " (%d migrations in total):\n", len(e.Files))
	case migrate.LogFile:
		l.fileCounter++
		if !l.fileStart.IsZero() {
			l.reportFileEnd()
		}
		l.fileStart = time.Now()
		fmt.Fprintf(l.out, "\n%s%v migrating version %v\n", indent2, dash, cyan(e.Version))
	case migrate.LogStmt:
		l.stmtCounter++
		fmt.Fprintf(l.out, "%s%v %s\n", indent4, arr, e.SQL)
	case migrate.LogDone:
		l.reportFileEnd()
		fmt.Fprintf(l.out, "\n%s%v\n", indent2, cyan(strings.Repeat("-", 25)))
		fmt.Fprintf(l.out, "%s%v %v\n", indent2, dash, time.Since(l.start))
		fmt.Fprintf(l.out, "%s%v %v migrations\n", indent2, dash, l.fileCounter)
		fmt.Fprintf(l.out, "%s%v %v sql statements\n", indent2, dash, l.stmtCounter)
	default:
		fmt.Fprintf(l.out, "%v", e)
	}
}

func (l *LogTTY) reportFileEnd() {
	fmt.Fprintf(l.out, "%s%v ok (%v)\n", indent2, dash, yellow("%s", time.Since(l.fileStart)))
}

func logFormat(out io.Writer) (migrate.Logger, error) {
	switch MigrateFlags.LogFormat {
	case logFormatTTY:
		return &LogTTY{out: out}, nil
	default:
		return nil, fmt.Errorf("unknown log-format %q", MigrateFlags.LogFormat)
	}
}

func migrateFlagsFromEnv(cmd *cobra.Command, _ []string) error {
	activeEnv, err := selectEnv(GlobalFlags.SelectedEnv)
	if err != nil {
		return err
	}
	if err := inputValsFromEnv(cmd); err != nil {
		return err
	}
	if err := maySetFlag(cmd, migrateFlagURL, activeEnv.URL); err != nil {
		return err
	}
	if err := maySetFlag(cmd, migrateFlagDevURL, activeEnv.DevURL); err != nil {
		return err
	}
	if err := maySetFlag(cmd, migrateFlagFormat, activeEnv.MigrationDir.Format); err != nil {
		return err
	}
	// Transform "src" to a URL.
	toURL := activeEnv.Source
	if toURL != "" {
		if toURL, err = filepath.Abs(activeEnv.Source); err != nil {
			return fmt.Errorf("finding abs path to source: %q: %w", activeEnv.Source, err)
		}
		toURL = "file://" + toURL
	}
	if err := maySetFlag(cmd, migrateFlagTo, toURL); err != nil {
		return err
	}
	if s := "[" + strings.Join(activeEnv.Schemas, "") + "]"; len(activeEnv.Schemas) > 0 {
		if err := maySetFlag(cmd, migrateFlagSchema, s); err != nil {
			return err
		}
	}
	return nil
}

type (
	// dryRunDriver wraps a migrate.Driver without executing any SQL statements.
	dryRunDriver struct {
		migrate.Driver
	}

	// dryRunRevisions wraps a migrate.RevisionReadWriter without executing any SQL statements.
	dryRunRevisions struct {
		migrate.RevisionReadWriter
	}
)

// QueryContext overrides the wrapped schema.ExecQuerier to not execute any SQL.
func (dryRunDriver) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, nil
}

// ExecContext overrides the wrapped schema.ExecQuerier to not execute any SQL.
func (dryRunDriver) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return nil, nil
}

// Lock implements the schema.Locker interface.
func (dryRunDriver) Lock(context.Context, string, time.Duration) (schema.UnlockFunc, error) {
	// We dry-run, we don't execute anything. Locking is not required.
	return func() error { return nil }, nil
}

// WriteRevision overrides the wrapped migrate.RevisionReadWriter to not saved any changes to revisions.
func (dryRunRevisions) WriteRevision(context.Context, *migrate.Revision) error {
	return nil
}
