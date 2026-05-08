const vscode = require("vscode");
const path = require("path");
const { runCairn, resolveWorkspace, workspaceRelativePath } = require("./cairnCli");

function activate(context) {
  const output = vscode.window.createOutputChannel("Cairn");
  context.subscriptions.push(output);

  context.subscriptions.push(vscode.commands.registerCommand("cairn.captureNote", () => captureNote(output)));
  context.subscriptions.push(vscode.commands.registerCommand("cairn.search", () => search(output)));
  context.subscriptions.push(vscode.commands.registerCommand("cairn.promoteCurrentFile", () => promoteCurrentFile(output)));
  context.subscriptions.push(vscode.commands.registerCommand("cairn.validateWorkspace", () => runWorkspaceCommand(output, ["validate"], "Validate Workspace")));
  context.subscriptions.push(vscode.commands.registerCommand("cairn.doctorFull", () => runWorkspaceCommand(output, ["doctor", "--full"], "Show Doctor")));
}

function deactivate() {}

async function captureNote(output) {
  try {
    const workspacePath = await currentWorkspace();
    const title = await vscode.window.showInputBox({ title: "Cairn note title", prompt: "Title for the captured note" });
    if (!title) {
      return;
    }
    const type = await vscode.window.showQuickPick(["note", "investigation", "handoff", "decision", "runbook"], {
      title: "Cairn note type"
    });
    if (!type) {
      return;
    }
    const body = await vscode.window.showInputBox({ title: "Cairn note body", prompt: "Optional one-line body. Leave empty for a starter template." });
    const cliPath = currentCliPath();
    const args = ["--root", workspacePath, "note", "--title", title, "--type", type];
    if (body) {
      args.push("--body", body);
    }
    showOutput(output, "Capture Note", await runCairn(cliPath, args, { cwd: workspacePath }));
  } catch (error) {
    showError(error);
  }
}

async function search(output) {
  try {
    const workspacePath = await currentWorkspace();
    const query = await vscode.window.showInputBox({ title: "Cairn search", prompt: "Search query" });
    if (!query) {
      return;
    }
    const cliPath = currentCliPath();
    showOutput(output, "Search", await runCairn(cliPath, ["--root", workspacePath, "search", "--query", query], { cwd: workspacePath }));
  } catch (error) {
    showError(error);
  }
}

async function promoteCurrentFile(output) {
  try {
    const editor = vscode.window.activeTextEditor;
    if (!editor) {
      throw new Error("Open a Cairn markdown document before running Cairn: Promote Current File.");
    }
    const workspacePath = await currentWorkspace();
    const relPath = workspaceRelativePath(workspacePath, editor.document.uri.fsPath);
    const cliPath = currentCliPath();
    showOutput(output, "Promote Current File", await runCairn(cliPath, ["--root", workspacePath, "promote", relPath], { cwd: workspacePath }));
  } catch (error) {
    showError(error);
  }
}

async function runWorkspaceCommand(output, args, label) {
  try {
    const workspacePath = await currentWorkspace();
    const cliPath = currentCliPath();
    showOutput(output, label, await runCairn(cliPath, ["--root", workspacePath, ...args], { cwd: workspacePath }));
  } catch (error) {
    showError(error);
  }
}

async function currentWorkspace() {
  const config = vscode.workspace.getConfiguration("cairn");
  const explicitWorkspace = config.get("workspacePath");
  const folders = (vscode.workspace.workspaceFolders || []).map((folder) => folder.uri.fsPath);
  if (folders.length === 0 && (!explicitWorkspace || explicitWorkspace.trim() === "")) {
    throw new Error("Open a folder or set cairn.workspacePath before using Cairn commands.");
  }
  return resolveWorkspace({ explicitWorkspace, folders, cliPath: currentCliPath() });
}

function currentCliPath() {
  return vscode.workspace.getConfiguration("cairn").get("cliPath") || "cairn";
}

function showOutput(output, label, text) {
  output.clear();
  output.appendLine(`# ${label}`);
  output.appendLine("");
  output.append(text);
  output.show(true);
}

function showError(error) {
  vscode.window.showErrorMessage(error && error.message ? error.message : String(error));
}

module.exports = {
  activate,
  deactivate
};
