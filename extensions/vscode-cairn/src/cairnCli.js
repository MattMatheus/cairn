const fs = require("fs");
const path = require("path");
const { execFile } = require("child_process");

function runCairn(cliPath, args, options = {}) {
  return new Promise((resolve, reject) => {
    execFile(cliPath, args, { cwd: options.cwd, timeout: options.timeout || 30000 }, (error, stdout, stderr) => {
      if (error) {
        if (error.code === "ENOENT") {
          reject(new Error(`Cairn CLI was not found at "${cliPath}". Install Cairn or set cairn.cliPath.`));
          return;
        }
        const detail = String(stderr || error.message || "").trim();
        reject(new Error(detail || `cairn ${args.join(" ")} failed`));
        return;
      }
      resolve(String(stdout || ""));
    });
  });
}

async function resolveWorkspace({ explicitWorkspace, folders, cliPath, runner = runCairn }) {
  if (explicitWorkspace && explicitWorkspace.trim() !== "") {
    return path.resolve(explicitWorkspace.trim());
  }
  for (const folder of folders || []) {
    const folderPath = typeof folder === "string" ? folder : folder.fsPath;
    const configPath = path.join(folderPath, ".cairn", "config.yaml");
    if (fs.existsSync(configPath)) {
      return folderPath;
    }
  }
  for (const folder of folders || []) {
    const folderPath = typeof folder === "string" ? folder : folder.fsPath;
    try {
      const output = await runner(cliPath, ["repo", "discover", "--from", folderPath], { cwd: folderPath });
      const match = output.match(/^Cairn workspace:\s*(.+)$/m);
      if (match && match[1]) {
        return match[1].trim();
      }
    } catch (error) {
      if (isMissingCLIError(error)) {
        throw error;
      }
    }
  }
  throw new Error("No Cairn workspace found. Open a Cairn workspace, attach this repo with `cairn repo attach`, or set cairn.workspacePath.");
}

function isMissingCLIError(error) {
  return error && String(error.message || "").includes("Cairn CLI was not found");
}

function workspaceRelativePath(workspacePath, filePath) {
  const relative = path.relative(workspacePath, filePath);
  if (relative === "" || relative.startsWith("..") || path.isAbsolute(relative)) {
    throw new Error("The active file is not inside the Cairn workspace; open a Cairn document before promoting.");
  }
  return relative.split(path.sep).join("/");
}

module.exports = {
  runCairn,
  resolveWorkspace,
  workspaceRelativePath
};
