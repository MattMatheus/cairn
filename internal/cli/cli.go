package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cairn/internal/document"
	"cairn/internal/localindex"
	"cairn/internal/mcpops"
	"cairn/internal/mcpschema"
	"cairn/internal/mcpserver"
	"cairn/internal/workspace"
)

type options struct {
	root string
}

func Run(ctx context.Context, args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) int {
	opts, rest, err := parseGlobal(args)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	if len(rest) == 0 {
		usage(stderr)
		return 2
	}

	var runErr error
	switch rest[0] {
	case "init":
		runErr = runInit(rest[1:], opts, stdout)
	case "capture":
		runErr = runCapture(rest[1:], opts, stdin, stdout)
	case "promote":
		runErr = runPromote(rest[1:], opts, stdout)
	case "archive":
		runErr = runArchive(rest[1:], opts, stdout)
	case "purge":
		runErr = runPurge(rest[1:], opts, stdout)
	case "validate":
		runErr = runValidate(ctx, rest[1:], opts, stdout)
	case "search":
		runErr = runSearch(ctx, rest[1:], opts, stdout)
	case "index":
		runErr = runIndex(ctx, rest[1:], opts, stdout)
	case "sync":
		runErr = runSync(ctx, rest[1:], opts, stdout)
	case "mcp":
		runErr = runMCP(ctx, rest[1:], opts, stdin, stdout)
	case "help", "-h", "--help":
		usage(stdout)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n", rest[0])
		usage(stderr)
		return 2
	}
	if runErr != nil {
		fmt.Fprintln(stderr, runErr)
		return 1
	}
	return 0
}

func parseGlobal(args []string) (options, []string, error) {
	opts := options{root: "."}
	for len(args) > 0 {
		arg := args[0]
		if arg == "--" {
			return opts, args[1:], nil
		}
		if arg == "--root" || arg == "-root" {
			if len(args) < 2 {
				return opts, nil, fmt.Errorf("%s requires a value", arg)
			}
			opts.root = args[1]
			args = args[2:]
			continue
		}
		if strings.HasPrefix(arg, "--root=") {
			opts.root = strings.TrimPrefix(arg, "--root=")
			args = args[1:]
			continue
		}
		break
	}
	if opts.root == "" {
		opts.root = "."
	}
	return opts, args, nil
}

func runInit(args []string, opts options, stdout io.Writer) error {
	fs := newFlagSet("init")
	workspaceID := fs.String("workspace-id", "", "workspace id")
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := workspace.Init(opts.root, workspace.InitOptions{WorkspaceID: *workspaceID})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Initialized workspace %s\n", result.WorkspaceID)
	printPaths(stdout, "Created", result.Created)
	printPaths(stdout, "Existing", result.Existing)
	fmt.Fprintln(stdout, "Next: run `cairn validate` to check workspace health.")
	return nil
}

func runCapture(args []string, opts options, stdin io.Reader, stdout io.Writer) error {
	fs := newFlagSet("capture")
	actor := fs.String("actor", "", "actor")
	title := fs.String("title", "", "title")
	body := fs.String("body", "", "body")
	bodyFile := fs.String("body-file", "", "body file, or - for stdin")
	docType := fs.String("type", "", "document type")
	authors := fs.String("authors", "", "comma-separated authors")
	tags := fs.String("tags", "", "comma-separated tags")
	if err := fs.Parse(args); err != nil {
		return err
	}
	content := *body
	if *bodyFile != "" {
		read, err := readBody(*bodyFile, stdin)
		if err != nil {
			return err
		}
		content = read
	}
	result, err := document.Workspace{Root: opts.root}.Capture(document.CaptureOptions{
		Actor:   *actor,
		Title:   *title,
		Body:    content,
		Type:    *docType,
		Authors: splitCSV(*authors),
		Tags:    splitCSV(*tags),
	})
	if err != nil {
		return err
	}
	printMutation(stdout, "Captured", result)
	return nil
}

func runPromote(args []string, opts options, stdout io.Writer) error {
	fs := newFlagSet("promote")
	path := fs.String("path", "", "workspace path")
	docType := fs.String("type", "", "document type")
	status := fs.String("status", "", "target status")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" && fs.NArg() > 0 {
		*path = fs.Arg(0)
	}
	result, err := document.Workspace{Root: opts.root}.Promote(document.PromoteOptions{
		Path:   *path,
		Type:   *docType,
		Status: *status,
	})
	if err != nil {
		return err
	}
	printMutation(stdout, "Promoted", result)
	return nil
}

func runArchive(args []string, opts options, stdout io.Writer) error {
	fs := newFlagSet("archive")
	path := fs.String("path", "", "workspace path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" && fs.NArg() > 0 {
		*path = fs.Arg(0)
	}
	result, err := document.Workspace{Root: opts.root}.Archive(document.ArchiveOptions{Path: *path})
	if err != nil {
		return err
	}
	printMutation(stdout, "Archived", result)
	return nil
}

func runPurge(args []string, opts options, stdout io.Writer) error {
	fs := newFlagSet("purge")
	path := fs.String("path", "", "workspace path")
	confirm := fs.Bool("confirm-purge", false, "confirm hard deletion of an archived document")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *path == "" && fs.NArg() > 0 {
		*path = fs.Arg(0)
	}
	if !*confirm {
		return errors.New("purge requires --confirm-purge")
	}
	result, err := document.Workspace{Root: opts.root}.Purge(document.PurgeOptions{Path: *path})
	if err != nil {
		return err
	}
	printMutation(stdout, "Purged", result)
	return nil
}

func runValidate(ctx context.Context, args []string, opts options, stdout io.Writer) error {
	fs := newFlagSet("validate")
	mode := fs.String("mode", string(document.ValidationModeDiscovery), "validation mode")
	if err := fs.Parse(args); err != nil {
		return err
	}
	data, err := workspace.Validate(ctx, opts.root, workspace.ValidateOptions{
		Mode:  validationMode(*mode),
		Paths: fs.Args(),
	})
	if err != nil {
		return err
	}
	if data.Healthy && len(data.Findings) == 0 {
		fmt.Fprintln(stdout, "Workspace validation passed.")
	} else if data.Healthy {
		fmt.Fprintln(stdout, "Workspace validation passed with warnings.")
	} else {
		fmt.Fprintln(stdout, "Workspace validation failed.")
	}
	for _, finding := range data.Findings {
		fmt.Fprintf(stdout, "- %s %s %s: %s", finding.Severity, finding.Code, finding.Path, finding.Message)
		if finding.DocumentID != "" {
			fmt.Fprintf(stdout, " [%s]", finding.DocumentID)
		}
		fmt.Fprintln(stdout)
	}
	if len(data.Findings) > 0 {
		fmt.Fprintln(stdout, "Next: address findings, then run `cairn validate` again.")
	}
	return nil
}

func runSearch(ctx context.Context, args []string, opts options, stdout io.Writer) error {
	fs := newFlagSet("search")
	query := fs.String("query", "", "search query")
	mode := fs.String("mode", string(mcpschema.SearchModeAuto), "search mode")
	limit := fs.Int("limit", 10, "result limit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *query == "" {
		*query = strings.Join(fs.Args(), " ")
	}
	index, err := localindex.Open(opts.root)
	if err != nil {
		return err
	}
	defer index.Close()
	if _, err := index.IndexWorkspace(ctx, opts.root); err != nil {
		return err
	}
	envelope, err := index.Search(ctx, opts.root, localindex.SearchOptions{
		Query: *query,
		Mode:  mcpschema.SearchMode(*mode),
		Limit: *limit,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Found %d result(s).\n", len(envelope.Data.Results))
	for _, result := range envelope.Data.Results {
		fmt.Fprintf(stdout, "- %s", result.Path)
		if result.Title != "" {
			fmt.Fprintf(stdout, " - %s", result.Title)
		}
		if result.Snippet != "" {
			fmt.Fprintf(stdout, "\n  %s", result.Snippet)
		}
		fmt.Fprintln(stdout)
	}
	printWarnings(stdout, envelope.Warnings)
	fmt.Fprintln(stdout, "Next: run `cairn index status` to inspect local index health.")
	return nil
}

func runIndex(ctx context.Context, args []string, opts options, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cairn index status|refresh")
	}
	local, err := mcpops.OpenLocal(opts.root)
	if err != nil {
		return err
	}
	defer local.Close()
	if args[0] == "refresh" {
		envelope, err := local.IndexRefresh(ctx, mcpschema.IndexRefreshRequest{})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Local index refreshed: %t\n", envelope.Data.LocalRefreshed)
		fmt.Fprintf(stdout, "Remote index refresh accepted: %t\n", envelope.Data.Accepted)
		fmt.Fprintf(stdout, "Remote index refreshed: %t\n", envelope.Data.RemoteRefreshed)
		if envelope.Data.JobID != "" {
			fmt.Fprintf(stdout, "Job id: %s\n", envelope.Data.JobID)
		}
		if envelope.Data.Message != "" {
			fmt.Fprintf(stdout, "Message: %s\n", envelope.Data.Message)
		}
		printWarnings(stdout, envelope.Warnings)
		for _, step := range envelope.NextSteps {
			fmt.Fprintf(stdout, "Next: %s\n", step.Label)
		}
		return nil
	}
	if args[0] != "status" {
		return fmt.Errorf("usage: cairn index status|refresh")
	}
	envelope, err := local.IndexStatus(ctx, mcpschema.IndexStatusRequest{})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Local index available: %t\n", envelope.Data.LocalAvailable)
	fmt.Fprintf(stdout, "Remote index available: %t\n", envelope.Data.RemoteAvailable)
	fmt.Fprintf(stdout, "Fresh: %t\n", envelope.Data.Fresh)
	printWarnings(stdout, envelope.Warnings)
	for _, step := range envelope.NextSteps {
		fmt.Fprintf(stdout, "Next: %s\n", step.Label)
	}
	return nil
}

func validationMode(mode string) document.ValidationMode {
	if mode == string(document.ValidationModeDurableBoundary) {
		return document.ValidationModeDurableBoundary
	}
	return document.ValidationModeDiscovery
}

func printMutation(stdout io.Writer, label string, result document.OperationResult) {
	fmt.Fprintf(stdout, "%s %s\n", label, filepath.ToSlash(result.Path))
	if result.OriginalPath != "" {
		fmt.Fprintf(stdout, "Previous path: %s\n", filepath.ToSlash(result.OriginalPath))
	}
	if result.DocumentID != "" {
		fmt.Fprintf(stdout, "Document id: %s\n", result.DocumentID)
	}
	for _, step := range result.NextSteps {
		fmt.Fprintf(stdout, "Next: %s\n", step)
	}
}

func printPaths(stdout io.Writer, label string, paths []string) {
	if len(paths) == 0 {
		return
	}
	fmt.Fprintf(stdout, "%s:\n", label)
	for _, path := range paths {
		fmt.Fprintf(stdout, "- %s\n", filepath.ToSlash(path))
	}
}

func printWarnings(stdout io.Writer, warnings []mcpschema.Warning) {
	for _, warning := range warnings {
		fmt.Fprintf(stdout, "Warning: %s", warning.Message)
		if warning.Path != "" {
			fmt.Fprintf(stdout, " (%s)", warning.Path)
		}
		fmt.Fprintln(stdout)
	}
}

func readBody(path string, stdin io.Reader) (string, error) {
	var content []byte
	var err error
	if path == "-" {
		content, err = io.ReadAll(stdin)
	} else {
		content, err = os.ReadFile(path)
	}
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func splitCSV(value string) []string {
	var out []string
	for _, part := range strings.Split(value, ",") {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func newFlagSet(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "usage: cairn [--root DIR] <command> [options]")
	fmt.Fprintln(w, "commands: init, capture, promote, archive, purge, validate, search, index status, sync status, mcp readonly|local-writes|remote-writes")
}

func runMCP(ctx context.Context, args []string, opts options, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cairn mcp readonly|local-writes|remote-writes")
	}
	local, err := mcpops.OpenLocal(opts.root)
	if err != nil {
		return err
	}
	defer local.Close()
	switch args[0] {
	case "readonly":
		return mcpserver.New(local).Serve(ctx, stdin, stdout)
	case "local-writes":
		return mcpserver.New(local, mcpserver.WithLocalWrites()).Serve(ctx, stdin, stdout)
	case "remote-writes":
		return mcpserver.New(local, mcpserver.WithRemoteWrites()).Serve(ctx, stdin, stdout)
	default:
		return fmt.Errorf("usage: cairn mcp readonly|local-writes|remote-writes")
	}
}

func runSync(ctx context.Context, args []string, opts options, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cairn sync status|dry-run|pull|push")
	}
	local, err := mcpops.OpenLocal(opts.root)
	if err != nil {
		return err
	}
	defer local.Close()
	if args[0] == "dry-run" {
		envelope, err := local.SyncDryRun(ctx, mcpschema.SyncRequest{DryRun: true})
		if err != nil {
			return err
		}
		plan := envelope.Data.Plan
		if plan == nil {
			return fmt.Errorf("sync dry-run did not return a plan")
		}
		fmt.Fprintf(stdout, "Sync dry-run direction: %s\n", plan.Direction)
		fmt.Fprintf(stdout, "Safe: %t\n", plan.Safe)
		for _, change := range plan.PlannedChanges {
			fmt.Fprintf(stdout, "- %s %s\n", change.Type, change.Path)
		}
		for _, conflict := range plan.Conflicts {
			fmt.Fprintf(stdout, "Conflict: local %s %s / remote %s %s\n", conflict.Local.Type, conflict.Local.Path, conflict.Remote.Type, conflict.Remote.Path)
		}
		printWarnings(stdout, envelope.Warnings)
		for _, step := range envelope.NextSteps {
			fmt.Fprintf(stdout, "Next: %s\n", step.Label)
		}
		return nil
	}
	if args[0] == "pull" {
		envelope, err := local.SyncPull(ctx, mcpschema.SyncRequest{})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Sync pull applied: %t\n", envelope.Data.Applied)
		for _, changed := range envelope.Data.ChangedPaths {
			fmt.Fprintf(stdout, "- %s %s\n", changed.Kind, changed.Path)
		}
		for _, step := range envelope.NextSteps {
			fmt.Fprintf(stdout, "Next: %s\n", step.Label)
		}
		return nil
	}
	if args[0] == "push" {
		envelope, err := local.SyncPush(ctx, mcpschema.SyncRequest{})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Sync push applied: %t\n", envelope.Data.Applied)
		for _, changed := range envelope.Data.ChangedPaths {
			fmt.Fprintf(stdout, "- %s %s\n", changed.Kind, changed.Path)
		}
		for _, step := range envelope.NextSteps {
			fmt.Fprintf(stdout, "Next: %s\n", step.Label)
		}
		return nil
	}
	if args[0] != "status" {
		return fmt.Errorf("usage: cairn sync status|dry-run|pull|push")
	}
	envelope, err := local.SyncStatus(ctx, mcpschema.EmptyRequest{})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Sync diverged: %t\n", envelope.Data.Diverged)
	fmt.Fprintf(stdout, "Local changes: %d\n", len(envelope.Data.LocalChanges))
	for _, change := range envelope.Data.LocalChanges {
		fmt.Fprintf(stdout, "- local %s %s\n", change.Type, change.Path)
	}
	fmt.Fprintf(stdout, "Remote changes: %d\n", len(envelope.Data.RemoteChanges))
	for _, change := range envelope.Data.RemoteChanges {
		fmt.Fprintf(stdout, "- remote %s %s\n", change.Type, change.Path)
	}
	for _, conflict := range envelope.Data.Conflicts {
		fmt.Fprintf(stdout, "Conflict: local %s %s / remote %s %s\n", conflict.Local.Type, conflict.Local.Path, conflict.Remote.Type, conflict.Remote.Path)
	}
	printWarnings(stdout, envelope.Warnings)
	for _, step := range envelope.NextSteps {
		fmt.Fprintf(stdout, "Next: %s\n", step.Label)
	}
	return nil
}
