package cli

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"cairn/internal/ado"
	"cairn/internal/document"
	"cairn/internal/localindex"
	"cairn/internal/mcpops"
	"cairn/internal/mcpschema"
	"cairn/internal/mcpserver"
	"cairn/internal/workspace"
)

var Version = "dev"

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
	case "ado":
		runErr = runADO(rest[1:], opts, stdin, stdout)
	case "setup":
		runErr = runSetup(rest[1:], opts, stdout)
	case "repo":
		runErr = runRepo(rest[1:], opts, stdout)
	case "version":
		runErr = runVersion(rest[1:], stdout)
	case "doctor":
		runErr = runDoctor(ctx, rest[1:], opts, stdout)
	case "health":
		runErr = runHealth(ctx, rest[1:], opts, stdout)
	case "init":
		runErr = runInit(rest[1:], opts, stdout)
	case "note":
		runErr = runNote(rest[1:], opts, stdin, stdout)
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

func runHealth(ctx context.Context, args []string, opts options, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cairn health report [--output PATH]")
	}
	switch args[0] {
	case "report":
		fs := newFlagSet("health report")
		outputPath := fs.String("output", "", "workspace-relative path to write the markdown report")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return fmt.Errorf("usage: cairn health report [--output PATH]")
		}
		report, err := workspace.BuildHealthReport(ctx, opts.root, workspace.HealthOptions{})
		if err != nil {
			return err
		}
		rendered := workspace.RenderHealthReport(report)
		if *outputPath == "" {
			fmt.Fprint(stdout, rendered)
			return nil
		}
		rel, err := cleanOutputPath(*outputPath)
		if err != nil {
			return err
		}
		target := filepath.Join(opts.root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(target, []byte(rendered), 0o644); err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Wrote health report %s\n", filepath.ToSlash(rel))
		return nil
	default:
		return fmt.Errorf("usage: cairn health report [--output PATH]")
	}
}

func cleanOutputPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return "", fmt.Errorf("output path must be relative: %s", path)
	}
	clean := filepath.ToSlash(filepath.Clean(path))
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("output path escapes workspace: %s", path)
	}
	return clean, nil
}

func runADO(args []string, opts options, stdin io.Reader, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cairn ado capture --event pr-completed --payload-file FILE [--actor ado] [--type handoff] [--status working|proposed]")
	}
	switch args[0] {
	case "capture":
		return runADOCapture(args[1:], opts, stdin, stdout)
	default:
		return fmt.Errorf("usage: cairn ado capture --event pr-completed --payload-file FILE [--actor ado] [--type handoff] [--status working|proposed]")
	}
}

func runADOCapture(args []string, opts options, stdin io.Reader, stdout io.Writer) error {
	fs := newFlagSet("ado capture")
	event := fs.String("event", "", "ADO lifecycle event")
	payloadFile := fs.String("payload-file", "", "ADO payload JSON file, or - for stdin")
	actor := fs.String("actor", "ado", "capture actor")
	docType := fs.String("type", "handoff", "candidate document type")
	status := fs.String("status", "working", "candidate status: working or proposed")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: cairn ado capture --event pr-completed --payload-file FILE [--actor ado] [--type handoff] [--status working|proposed]")
	}
	if *payloadFile == "" {
		return errors.New("ado capture requires --payload-file")
	}
	if *status != "working" && *status != "proposed" {
		return fmt.Errorf("ado capture supports only working or proposed status, got %q", *status)
	}
	payload, err := readBody(*payloadFile, stdin)
	if err != nil {
		return err
	}
	candidate, err := ado.BuildCandidate(*event, []byte(payload))
	if err != nil {
		return err
	}
	captured, err := document.Workspace{Root: opts.root}.Capture(document.CaptureOptions{
		Actor: *actor,
		Title: candidate.Title,
		Body:  candidate.Body,
		Type:  *docType,
		Tags:  candidate.Tags,
	})
	if err != nil {
		return err
	}
	result := captured
	if *status == "proposed" {
		promoted, err := document.Workspace{Root: opts.root}.Promote(document.PromoteOptions{
			Path:   captured.Path,
			Type:   *docType,
			Status: "proposed",
		})
		if err != nil {
			return err
		}
		result = promoted
		printMutation(stdout, "Captured candidate and promoted", result)
	} else {
		printMutation(stdout, "Captured candidate", result)
	}
	fmt.Fprintln(stdout, "Next: review the ADO candidate and promote only if it should become durable pod knowledge.")
	return nil
}

func runRepo(args []string, opts options, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cairn repo attach --name NAME --path RELPATH [--url URL] [--no-pointer] | list | discover [--from DIR]")
	}
	switch args[0] {
	case "attach":
		fs := newFlagSet("repo attach")
		name := fs.String("name", "", "repo name")
		path := fs.String("path", "", "repo path relative to the Cairn workspace")
		url := fs.String("url", "", "repo URL")
		noPointer := fs.Bool("no-pointer", false, "do not write a .cairn-workspace pointer into the repo")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := workspace.AttachRepo(opts.root, workspace.RepoAttachOptions{
			Name:         *name,
			Path:         *path,
			URL:          *url,
			WritePointer: !*noPointer,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Attached repo %s -> %s\n", result.Repo.Name, result.Repo.Path)
		fmt.Fprintf(stdout, "Recorded metadata in %s\n", filepath.ToSlash(result.ConfigPath))
		if result.PointerPath != "" {
			fmt.Fprintf(stdout, "Wrote workspace pointer %s\n", filepath.ToSlash(result.PointerPath))
		}
		fmt.Fprintln(stdout, "Repo attachment is reference metadata only; Cairn will not clone, index, sync, or validate repo contents.")
		return nil
	case "list":
		if len(args) != 1 {
			return fmt.Errorf("usage: cairn repo list")
		}
		repos, err := workspace.LoadRepos(opts.root)
		if err != nil {
			return err
		}
		if len(repos) == 0 {
			fmt.Fprintln(stdout, "No attached repos.")
			fmt.Fprintln(stdout, "Repo attachment is reference metadata only; Cairn will not clone, index, sync, or validate repo contents.")
			return nil
		}
		fmt.Fprintf(stdout, "Attached repos (%d):\n", len(repos))
		for _, repo := range repos {
			fmt.Fprintf(stdout, "- %s -> %s", repo.Name, repo.Path)
			if repo.URL != "" {
				fmt.Fprintf(stdout, " (%s)", repo.URL)
			}
			fmt.Fprintln(stdout)
		}
		fmt.Fprintln(stdout, "Repo attachment is reference metadata only; Cairn will not clone, index, sync, or validate repo contents.")
		return nil
	case "discover":
		fs := newFlagSet("repo discover")
		from := fs.String("from", ".", "repo path or child path to discover from")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		if fs.NArg() != 0 {
			return fmt.Errorf("usage: cairn repo discover [--from DIR]")
		}
		result, err := workspace.DiscoverWorkspace(*from)
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Cairn workspace: %s\n", filepath.ToSlash(result.WorkspacePath))
		fmt.Fprintf(stdout, "Discovered via: %s\n", filepath.ToSlash(result.PointerPath))
		fmt.Fprintln(stdout, "Discovery follows an explicit .cairn-workspace pointer; repo contents remain outside Cairn management.")
		return nil
	default:
		return fmt.Errorf("usage: cairn repo attach --name NAME --path RELPATH [--url URL] [--no-pointer] | list | discover [--from DIR]")
	}
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

func runSetup(args []string, opts options, stdout io.Writer) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: cairn setup local-sync --remote-root DIR | azure-sync --account ACCOUNT --container CONTAINER")
	}
	switch args[0] {
	case "local-sync":
		fs := newFlagSet("setup local-sync")
		workspaceID := fs.String("workspace-id", "", "workspace id")
		remoteRoot := fs.String("remote-root", "", "local filesystem remote store root")
		force := fs.Bool("force", false, "allow setup inside the Cairn source repository")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := workspace.SetupLocalSync(opts.root, workspace.SetupLocalSyncOptions{
			WorkspaceID: *workspaceID,
			RemoteRoot:  *remoteRoot,
			Force:       *force,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Initialized workspace %s\n", result.WorkspaceID)
		fmt.Fprintf(stdout, "Configured local sync in %s\n", result.ConfigPath)
		fmt.Fprintf(stdout, "Local remote root: %s\n", result.RemoteRoot)
		fmt.Fprintln(stdout, "Next: run `cairn validate`, then `cairn sync status`.")
		return nil
	case "azure-sync":
		fs := newFlagSet("setup azure-sync")
		workspaceID := fs.String("workspace-id", "", "workspace id")
		account := fs.String("account", "", "Azure Storage account name")
		endpoint := fs.String("endpoint", "", "Azure Blob endpoint URL")
		container := fs.String("container", "", "Azure Blob container")
		prefix := fs.String("prefix", "", "optional object prefix")
		force := fs.Bool("force", false, "allow setup inside the Cairn source repository")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		result, err := workspace.SetupAzureSync(opts.root, workspace.SetupAzureSyncOptions{
			WorkspaceID: *workspaceID,
			Account:     *account,
			Endpoint:    *endpoint,
			Container:   *container,
			Prefix:      *prefix,
			Force:       *force,
		})
		if err != nil {
			return err
		}
		fmt.Fprintf(stdout, "Initialized workspace %s\n", result.WorkspaceID)
		fmt.Fprintf(stdout, "Configured Azure Blob sync in %s\n", result.ConfigPath)
		if result.Account != "" {
			fmt.Fprintf(stdout, "Storage account: %s\n", result.Account)
		}
		if result.Endpoint != "" {
			fmt.Fprintf(stdout, "Blob endpoint: %s\n", result.Endpoint)
		}
		fmt.Fprintf(stdout, "Container: %s\n", result.Container)
		if result.Prefix == "" {
			fmt.Fprintln(stdout, "Prefix: <none>")
		} else {
			fmt.Fprintf(stdout, "Prefix: %s\n", result.Prefix)
		}
		fmt.Fprintln(stdout, "Next: run `az login`, then `cairn doctor --remote`.")
		return nil
	default:
		return fmt.Errorf("usage: cairn setup local-sync --remote-root DIR | azure-sync --account ACCOUNT --container CONTAINER")
	}
}

func runVersion(args []string, stdout io.Writer) error {
	if len(args) != 0 {
		return fmt.Errorf("usage: cairn version")
	}
	fmt.Fprintf(stdout, "cairn %s\n", Version)
	return nil
}

func runInit(args []string, opts options, stdout io.Writer) error {
	fs := newFlagSet("init")
	workspaceID := fs.String("workspace-id", "", "workspace id")
	force := fs.Bool("force", false, "allow init inside the Cairn source repository")
	if err := fs.Parse(args); err != nil {
		return err
	}
	result, err := workspace.Init(opts.root, workspace.InitOptions{WorkspaceID: *workspaceID, Force: *force})
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Initialized workspace %s\n", result.WorkspaceID)
	printPaths(stdout, "Created", result.Created)
	printPaths(stdout, "Existing", result.Existing)
	fmt.Fprintln(stdout, "Next: run `cairn validate` to check workspace health.")
	return nil
}

func runDoctor(ctx context.Context, args []string, opts options, stdout io.Writer) error {
	fs := newFlagSet("doctor")
	checkRemote := fs.Bool("remote", false, "check remote store reachability")
	full := fs.Bool("full", false, "run full pilot readiness checks")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("usage: cairn doctor [--full] [--remote]")
	}
	if *full {
		return runDoctorFull(ctx, opts, *checkRemote, stdout)
	}
	absRoot, err := filepath.Abs(opts.root)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Cairn version: %s\n", Version)
	fmt.Fprintf(stdout, "Workspace root: %s\n", absRoot)

	configPath := filepath.Join(absRoot, ".cairn", "config.yaml")
	if _, err := os.Stat(configPath); err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintln(stdout, "Config: missing")
			fmt.Fprintln(stdout, "Next: run `cairn init` or a setup command such as `cairn setup local-sync --remote-root DIR`.")
			return nil
		}
		return err
	}
	fmt.Fprintln(stdout, "Config: present")
	cfg, err := document.LoadConfig(absRoot)
	if err != nil {
		return err
	}
	if cfg.WorkspaceID != "" {
		fmt.Fprintf(stdout, "Workspace id: %s\n", cfg.WorkspaceID)
	}
	if _, err := os.Stat(localindex.DBPath(absRoot)); err == nil {
		fmt.Fprintln(stdout, "Local index: present")
	} else if os.IsNotExist(err) {
		fmt.Fprintln(stdout, "Local index: missing")
		fmt.Fprintln(stdout, "Next: run `cairn index refresh`.")
	} else {
		return err
	}

	switch cfg.RemoteSync.Provider {
	case "":
		fmt.Fprintln(stdout, "Remote sync: not configured")
		fmt.Fprintln(stdout, "Next: run `cairn setup local-sync --remote-root DIR` or `cairn setup azure-sync --account ACCOUNT --container CONTAINER`.")
	case "local_fs":
		fmt.Fprintf(stdout, "Remote sync: local_fs (%s)\n", cfg.RemoteSync.Root)
	case "azure_blob":
		if cfg.RemoteSync.Endpoint != "" {
			fmt.Fprintf(stdout, "Remote sync: azure_blob (%s, container %s)\n", cfg.RemoteSync.Endpoint, cfg.RemoteSync.Container)
		} else {
			fmt.Fprintf(stdout, "Remote sync: azure_blob (%s, container %s)\n", cfg.RemoteSync.Account, cfg.RemoteSync.Container)
		}
		if cfg.RemoteSync.Prefix == "" {
			fmt.Fprintln(stdout, "Remote prefix: <none>")
		} else {
			fmt.Fprintf(stdout, "Remote prefix: %s\n", cfg.RemoteSync.Prefix)
		}
	default:
		fmt.Fprintf(stdout, "Remote sync: unsupported provider %s\n", cfg.RemoteSync.Provider)
	}

	if *checkRemote {
		local, err := mcpops.OpenLocal(absRoot)
		if err != nil {
			return err
		}
		defer local.Close()
		if local.RemoteStore == nil {
			fmt.Fprintln(stdout, "Remote check: skipped, remote sync is not configured")
		} else if _, err := local.RemoteStore.ListObjects(ctx, ""); err != nil {
			return fmt.Errorf("remote check failed: %w", err)
		} else {
			fmt.Fprintln(stdout, "Remote check: reachable")
		}
	}
	return nil
}

func runDoctorFull(ctx context.Context, opts options, checkRemote bool, stdout io.Writer) error {
	absRoot, err := filepath.Abs(opts.root)
	if err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Cairn version: %s\n", Version)
	fmt.Fprintf(stdout, "Workspace root: %s\n", absRoot)
	fmt.Fprintln(stdout, "Full readiness:")

	configPath := filepath.Join(absRoot, ".cairn", "config.yaml")
	configPresent := false
	if _, err := os.Stat(configPath); err == nil {
		configPresent = true
		printCheck(stdout, "Config", "pass", "present")
	} else if os.IsNotExist(err) {
		printCheck(stdout, "Config", "fail", "missing")
		fmt.Fprintln(stdout, "Next: run `cairn init` or `cairn setup local-sync --remote-root DIR`.")
	} else {
		printCheck(stdout, "Config", "fail", err.Error())
	}

	if !configPresent {
		printCheck(stdout, "Managed folders", "skip", "config is missing")
		printCheck(stdout, "Schemas", "skip", "config is missing")
		printCheck(stdout, "Validation", "skip", "config is missing")
		printCheck(stdout, "Local index", "skip", "config is missing")
		printCheck(stdout, "Search sanity", "skip", "config is missing")
		printCheck(stdout, "Sync status", "skip", "config is missing")
		printCheck(stdout, "Remote reachability", "skip", "config is missing")
		printCheck(stdout, "MCP tools", "skip", "config is missing")
		return nil
	}

	cfg, err := document.LoadConfig(absRoot)
	if err != nil {
		printCheck(stdout, "Config", "fail", err.Error())
		printCheck(stdout, "Managed folders", "skip", "config could not be loaded")
		printCheck(stdout, "Schemas", "skip", "config could not be loaded")
		printCheck(stdout, "Validation", "skip", "config could not be loaded")
		printCheck(stdout, "Local index", "skip", "config could not be loaded")
		printCheck(stdout, "Search sanity", "skip", "config could not be loaded")
		printCheck(stdout, "Sync status", "skip", "config could not be loaded")
		printCheck(stdout, "Remote reachability", "skip", "config could not be loaded")
		printCheck(stdout, "MCP tools", "skip", "config could not be loaded")
		return nil
	}
	if cfg.WorkspaceID != "" {
		fmt.Fprintf(stdout, "Workspace id: %s\n", cfg.WorkspaceID)
	}
	reportManagedFolders(absRoot, cfg, stdout)

	validation, err := workspace.Validate(ctx, absRoot, workspace.ValidateOptions{Mode: document.ValidationModeDiscovery})
	if err != nil {
		return err
	}
	reportSchemaFindings(validation.Findings, stdout)
	reportValidation(validation, stdout)

	indexAvailableBeforeOpen := false
	if _, err := os.Stat(localindex.DBPath(absRoot)); err == nil {
		indexAvailableBeforeOpen = true
	} else if os.IsNotExist(err) {
		printCheck(stdout, "Local index", "warn", "missing")
		fmt.Fprintln(stdout, "Next: run `cairn index refresh`.")
	} else {
		printCheck(stdout, "Local index", "fail", err.Error())
	}

	local, err := mcpops.OpenLocal(absRoot)
	if err != nil {
		if indexAvailableBeforeOpen {
			printCheck(stdout, "Local index", "fail", err.Error())
		}
		printCheck(stdout, "Search sanity", "skip", "local workspace operations could not open")
		printCheck(stdout, "Sync status", "skip", "local workspace operations could not open")
		printCheck(stdout, "Remote reachability", "skip", "local workspace operations could not open")
		printCheck(stdout, "MCP tools", "skip", "local workspace operations could not open")
		return nil
	}
	defer local.Close()

	if indexAvailableBeforeOpen {
		indexStatus, err := local.IndexStatus(ctx, mcpschema.IndexStatusRequest{})
		if err != nil {
			printCheck(stdout, "Local index", "fail", err.Error())
		} else if indexStatus.Data.LocalAvailable {
			printCheck(stdout, "Local index", "pass", "available")
		} else {
			printCheck(stdout, "Local index", "warn", "missing or stale")
			fmt.Fprintln(stdout, "Next: run `cairn index refresh`.")
		}
	}

	if !indexAvailableBeforeOpen {
		printCheck(stdout, "Search sanity", "skip", "local index is missing")
	} else if _, err := local.Index.Query(ctx, localindex.Query{Limit: 1}); err != nil {
		printCheck(stdout, "Search sanity", "fail", err.Error())
	} else {
		printCheck(stdout, "Search sanity", "pass", "local query path is usable")
	}

	syncStatus, err := local.SyncStatus(ctx, mcpschema.EmptyRequest{})
	if err != nil {
		printCheck(stdout, "Sync status", "fail", friendlySyncError(err).Error())
	} else if syncStatus.Data.Diverged || len(syncStatus.Data.Conflicts) > 0 {
		printCheck(stdout, "Sync status", "warn", "local and remote state diverged")
		fmt.Fprintln(stdout, "Next: inspect `cairn sync status` and resolve conflicts before mutating sync state.")
	} else {
		message := fmt.Sprintf("%d local change(s), %d remote change(s)", len(syncStatus.Data.LocalChanges), len(syncStatus.Data.RemoteChanges))
		printCheck(stdout, "Sync status", "pass", message)
	}

	if local.RemoteStore == nil {
		printCheck(stdout, "Remote reachability", "warn", "remote sync is not configured")
		fmt.Fprintln(stdout, "Next: run `cairn setup local-sync --remote-root DIR` for a no-service pilot or `cairn setup azure-sync --account ACCOUNT --container CONTAINER` for Azure Blob.")
	} else if !checkRemote {
		printCheck(stdout, "Remote reachability", "skip", "pass --remote to check the configured store")
	} else if _, err := local.RemoteStore.ListObjects(ctx, ""); err != nil {
		printCheck(stdout, "Remote reachability", "fail", err.Error())
	} else {
		printCheck(stdout, "Remote reachability", "pass", "configured store is reachable")
	}

	readOnly := mcpserver.New(local).Tools()
	localWrites := mcpserver.New(local, mcpserver.WithLocalWrites()).Tools()
	remoteWrites := mcpserver.New(local, mcpserver.WithRemoteWrites()).Tools()
	if len(readOnly) == 0 || len(localWrites) <= len(readOnly) || len(remoteWrites) <= len(readOnly) {
		printCheck(stdout, "MCP tools", "fail", "expected readonly, local-writes, and remote-writes tool surfaces")
	} else {
		printCheck(stdout, "MCP tools", "pass", fmt.Sprintf("readonly=%d local-writes=%d remote-writes=%d", len(readOnly), len(localWrites), len(remoteWrites)))
	}
	return nil
}

func printCheck(stdout io.Writer, label string, status string, message string) {
	fmt.Fprintf(stdout, "- %s: %s", label, status)
	if message != "" {
		fmt.Fprintf(stdout, " (%s)", message)
	}
	fmt.Fprintln(stdout)
}

func reportManagedFolders(root string, cfg document.Config, stdout io.Writer) {
	missing := 0
	for _, folder := range cfg.ManagedFolders {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(folder))); os.IsNotExist(err) {
			missing++
		}
	}
	if missing == 0 {
		printCheck(stdout, "Managed folders", "pass", fmt.Sprintf("%d configured", len(cfg.ManagedFolders)))
		return
	}
	printCheck(stdout, "Managed folders", "fail", fmt.Sprintf("%d missing", missing))
	fmt.Fprintln(stdout, "Next: run `cairn init` to create missing starter folders without overwriting existing files.")
}

func reportSchemaFindings(findings []mcpschema.ValidationFinding, stdout io.Writer) {
	errors, warnings := countFindings(findings, func(f mcpschema.ValidationFinding) bool {
		return strings.HasPrefix(f.Path, ".cairn/")
	})
	if errors > 0 {
		printCheck(stdout, "Schemas", "fail", fmt.Sprintf("%d error(s), %d warning(s)", errors, warnings))
		fmt.Fprintln(stdout, "Next: fix `.cairn/config.yaml` or `.cairn/schemas/*.yaml`, then run `cairn validate` again.")
		return
	}
	if warnings > 0 {
		printCheck(stdout, "Schemas", "warn", fmt.Sprintf("%d warning(s)", warnings))
		return
	}
	printCheck(stdout, "Schemas", "pass", "config and schema files are valid")
}

func reportValidation(data mcpschema.ValidateWorkspaceData, stdout io.Writer) {
	errors, warnings := countFindings(data.Findings, func(f mcpschema.ValidationFinding) bool {
		return !strings.HasPrefix(f.Path, ".cairn/")
	})
	if errors > 0 {
		printCheck(stdout, "Validation", "fail", fmt.Sprintf("%d error(s), %d warning(s)", errors, warnings))
		fmt.Fprintln(stdout, "Next: address validation findings, then run `cairn validate` again.")
		return
	}
	if warnings > 0 {
		printCheck(stdout, "Validation", "warn", fmt.Sprintf("%d warning(s)", warnings))
		fmt.Fprintln(stdout, "Next: review warnings before promoting or syncing durable knowledge.")
		return
	}
	printCheck(stdout, "Validation", "pass", "managed documents are healthy")
}

func countFindings(findings []mcpschema.ValidationFinding, include func(mcpschema.ValidationFinding) bool) (int, int) {
	var errors, warnings int
	for _, finding := range findings {
		if !include(finding) {
			continue
		}
		if finding.Severity == string(document.SeverityError) {
			errors++
		} else {
			warnings++
		}
	}
	return errors, warnings
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
	interactive := fs.Bool("interactive", false, "prompt for missing capture fields")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *interactive {
		return runInteractiveCapture(captureFields{
			Actor:    *actor,
			Title:    *title,
			Body:     *body,
			BodyFile: *bodyFile,
			Type:     *docType,
			Authors:  splitCSV(*authors),
			Tags:     splitCSV(*tags),
		}, opts, stdin, stdout)
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

type captureFields struct {
	Actor    string
	Title    string
	Body     string
	BodyFile string
	Type     string
	Authors  []string
	Tags     []string
}

func runNote(args []string, opts options, stdin io.Reader, stdout io.Writer) error {
	fs := newFlagSet("note")
	actor := fs.String("actor", "", "actor; defaults to CAIRN_ACTOR, USER, or USERNAME")
	title := fs.String("title", "", "title")
	body := fs.String("body", "", "body")
	bodyFile := fs.String("body-file", "", "body file, or - for stdin")
	docType := fs.String("type", "note", "document type: note, investigation, handoff, decision, or runbook")
	authors := fs.String("authors", "", "comma-separated authors")
	tags := fs.String("tags", "", "comma-separated tags")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *title == "" && fs.NArg() > 0 {
		*title = strings.Join(fs.Args(), " ")
	}
	fields := captureFields{
		Actor:    defaultActor(*actor),
		Title:    *title,
		Body:     *body,
		BodyFile: *bodyFile,
		Type:     *docType,
		Authors:  splitCSV(*authors),
		Tags:     splitCSV(*tags),
	}
	if fields.Title == "" {
		return errors.New("title is required; use `cairn note --title \"...\"` or pass the title as arguments")
	}
	if err := validateCaptureType(fields.Type); err != nil {
		return err
	}
	content := fields.Body
	if fields.BodyFile != "" {
		read, err := readBody(fields.BodyFile, stdin)
		if err != nil {
			return err
		}
		content = read
	}
	if strings.TrimSpace(content) == "" {
		content = captureTemplate(fields.Type, fields.Title)
	}
	return captureFromFields(fields.withBody(content), opts, stdout)
}

func runInteractiveCapture(fields captureFields, opts options, stdin io.Reader, stdout io.Writer) error {
	reader := bufio.NewReader(stdin)
	fields.Actor = defaultActor(fields.Actor)
	var err error
	if fields.Actor == "" {
		fields.Actor, err = promptLine(reader, stdout, "Actor")
		if err != nil {
			return err
		}
	}
	if fields.Title == "" {
		fields.Title, err = promptLine(reader, stdout, "Title")
		if err != nil {
			return err
		}
	}
	if fields.Type == "" {
		fields.Type, err = promptLineDefault(reader, stdout, "Type", "note")
		if err != nil {
			return err
		}
	}
	if err := validateCaptureType(fields.Type); err != nil {
		return err
	}
	content := fields.Body
	if fields.BodyFile != "" {
		read, err := readBody(fields.BodyFile, reader)
		if err != nil {
			return err
		}
		content = read
	}
	if strings.TrimSpace(content) == "" {
		fmt.Fprintln(stdout, "Body: enter markdown, then a line with only `.` to finish.")
		read, err := readPromptBody(reader)
		if err != nil {
			return err
		}
		content = read
	}
	if strings.TrimSpace(content) == "" {
		content = captureTemplate(fields.Type, fields.Title)
	}
	return captureFromFields(fields.withBody(content), opts, stdout)
}

func (fields captureFields) withBody(body string) captureFields {
	fields.Body = body
	return fields
}

func captureFromFields(fields captureFields, opts options, stdout io.Writer) error {
	result, err := document.Workspace{Root: opts.root}.Capture(document.CaptureOptions{
		Actor:   fields.Actor,
		Title:   fields.Title,
		Body:    fields.Body,
		Type:    fields.Type,
		Authors: fields.Authors,
		Tags:    fields.Tags,
	})
	if err != nil {
		return err
	}
	printMutation(stdout, "Captured", result)
	return nil
}

func defaultActor(explicit string) string {
	if explicit != "" {
		return explicit
	}
	for _, key := range []string{"CAIRN_ACTOR", "USER", "USERNAME"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return slugifyActor(value)
		}
	}
	return ""
}

func slugifyActor(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			builder.WriteRune('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), "-")
}

func validateCaptureType(docType string) error {
	switch docType {
	case "note", "investigation", "handoff", "decision", "runbook":
		return nil
	default:
		return fmt.Errorf("unsupported note type %q; use note, investigation, handoff, decision, or runbook", docType)
	}
}

func captureTemplate(docType string, title string) string {
	switch docType {
	case "investigation":
		return fmt.Sprintf("# %s\n\n## Context\n\n## Findings\n\n## Next Steps\n", title)
	case "handoff":
		return fmt.Sprintf("# %s\n\n## Summary\n\n## Current State\n\n## Next Steps\n", title)
	case "decision":
		return fmt.Sprintf("# %s\n\n## Context\n\n## Decision\n\n## Consequences\n", title)
	case "runbook":
		return fmt.Sprintf("# %s\n\n## Purpose\n\n## Steps\n\n## Verification\n", title)
	default:
		return fmt.Sprintf("# %s\n\n", title)
	}
}

func promptLine(reader *bufio.Reader, stdout io.Writer, label string) (string, error) {
	fmt.Fprintf(stdout, "%s: ", label)
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(value), nil
}

func promptLineDefault(reader *bufio.Reader, stdout io.Writer, label string, fallback string) (string, error) {
	fmt.Fprintf(stdout, "%s [%s]: ", label, fallback)
	value, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback, nil
	}
	return value, nil
}

func readPromptBody(reader *bufio.Reader) (string, error) {
	var lines []string
	for {
		line, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return "", err
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "." {
			return strings.Join(lines, "\n"), nil
		}
		if line != "" {
			lines = append(lines, trimmed)
		}
		if errors.Is(err, io.EOF) {
			return strings.Join(lines, "\n"), nil
		}
	}
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
	local, err := mcpops.OpenLocal(opts.root)
	if err != nil {
		return err
	}
	defer local.Close()
	if _, err := local.Index.IndexWorkspace(ctx, opts.root); err != nil {
		return err
	}
	envelope, err := local.SearchContext(ctx, mcpschema.SearchContextRequest{
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
	fmt.Fprintln(w, "commands: version, doctor, health report, setup local-sync|azure-sync, init, repo attach|list|discover, ado capture, note, capture, promote, archive, purge, validate, search, index status, sync status, mcp readonly|local-writes|remote-writes")
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
			return friendlySyncError(err)
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
			return friendlySyncError(err)
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

func friendlySyncError(err error) error {
	if err == nil {
		return nil
	}
	if strings.Contains(err.Error(), "remote store is required") {
		return errors.New("remote sync is not configured; run `cairn setup local-sync --remote-root DIR` for a no-service pilot or `cairn setup azure-sync --account ACCOUNT --container CONTAINER` for Azure Blob")
	}
	return err
}
