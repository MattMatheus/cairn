const assert = require("assert");
const fs = require("fs");
const os = require("os");
const path = require("path");
const { resolveWorkspace, workspaceRelativePath } = require("../src/cairnCli");

async function testExplicitWorkspaceWins() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "cairn-vscode-"));
  const resolved = await resolveWorkspace({
    explicitWorkspace: root,
    folders: [],
    cliPath: "cairn",
    runner: async () => {
      throw new Error("runner should not be called");
    }
  });
  assert.strictEqual(resolved, root);
}

async function testOpenCairnWorkspace() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "cairn-vscode-"));
  fs.mkdirSync(path.join(root, ".cairn"), { recursive: true });
  fs.writeFileSync(path.join(root, ".cairn", "config.yaml"), "schema_version: 1\n");
  const resolved = await resolveWorkspace({ folders: [root], cliPath: "cairn" });
  assert.strictEqual(resolved, root);
}

async function testPointerDiscovery() {
  const repo = fs.mkdtempSync(path.join(os.tmpdir(), "cairn-vscode-repo-"));
  const kb = fs.mkdtempSync(path.join(os.tmpdir(), "cairn-vscode-kb-"));
  const resolved = await resolveWorkspace({
    folders: [repo],
    cliPath: "cairn",
    runner: async (_cliPath, args) => {
      assert.deepStrictEqual(args, ["repo", "discover", "--from", repo]);
      return `Cairn workspace: ${kb}\nDiscovered via: ${path.join(repo, ".cairn-workspace")}\n`;
    }
  });
  assert.strictEqual(resolved, kb);
}

async function testMissingCLIIsActionable() {
  const repo = fs.mkdtempSync(path.join(os.tmpdir(), "cairn-vscode-repo-"));
  await assert.rejects(
    () =>
      resolveWorkspace({
        folders: [repo],
        cliPath: "missing-cairn",
        runner: async () => {
          throw new Error('Cairn CLI was not found at "missing-cairn". Install Cairn or set cairn.cliPath.');
        }
      }),
    /Install Cairn or set cairn\.cliPath/
  );
}

function testWorkspaceRelativePath() {
  const root = fs.mkdtempSync(path.join(os.tmpdir(), "cairn-vscode-"));
  assert.strictEqual(workspaceRelativePath(root, path.join(root, "agents", "ado", "note.md")), "agents/ado/note.md");
  assert.throws(() => workspaceRelativePath(root, path.join(os.tmpdir(), "other.md")), /not inside the Cairn workspace/);
}

async function main() {
  await testExplicitWorkspaceWins();
  await testOpenCairnWorkspace();
  await testPointerDiscovery();
  await testMissingCLIIsActionable();
  testWorkspaceRelativePath();
}

main().catch((error) => {
  console.error(error);
  process.exit(1);
});
