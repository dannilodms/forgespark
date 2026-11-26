// PreToolUse hook (Bash and PowerShell), wired in .claude/settings.json.
//
// The rule .claude/rules/staging.md asks the agent, before it opens a pull
// request of this repository, one question: open it with staging, without
// staging, or not at all. A rule is only context, so this hook stops a
// `gh pr create` of this repository until the command declares the answer:
//   STAGING_QUESTION=with gh pr create --label staging ...      (Bash)
//   STAGING_QUESTION=without gh pr create ...                   (Bash)
//   $env:STAGING_QUESTION='with'; gh pr create --label staging  (PowerShell)
// "with" must carry the label staging and "without" must not.
// Any other command, or a pull request of another repository, passes untouched.

// The repository was renamed from software-factory-manager to forgespark on
// 2026-11-25; GitHub still answers to the old name, so both are matched.
const repositoryNames = ["dannilodms/forgespark", "dannilodms/software-factory-manager"];

function normalized(text) {
  return String(text || "").replace(/\\/g, "/").replace(/\/+$/, "").toLowerCase();
}

// The checkout this session runs in (the main one or a worktree), spelled the
// way Windows writes it (c:/...) and the way Git Bash writes it (/c/...).
function repositoryPaths() {
  const project = normalized(process.env.CLAUDE_PROJECT_DIR);
  if (project === "") {
    return [];
  }
  const drive = /^([a-z]):\//.exec(project);
  return drive ? [project, `/${drive[1]}/${project.slice(3)}`] : [project];
}

function opensPullRequest(command) {
  return /\bgh(\.exe)?["']?\s+pr\s+create\b/i.test(command);
}

function inRepository(command, cwd) {
  const where = normalized(cwd);
  const what = normalized(command);
  return (
    repositoryPaths().some((path) => where.startsWith(path) || what.includes(path)) ||
    repositoryNames.some((name) => what.includes(name))
  );
}

function declaredAnswer(command) {
  const match = /STAGING_QUESTION\s*=\s*["']?(with|without)\b/i.exec(command);
  return match ? match[1].toLowerCase() : "";
}

function carriesStagingLabel(command) {
  return /(--label|-l)(\s+|=)["']?([^"'\s]*,)?staging\b/i.test(command);
}

function deny(reason) {
  process.stdout.write(
    JSON.stringify({
      hookSpecificOutput: {
        hookEventName: "PreToolUse",
        permissionDecision: "deny",
        permissionDecisionReason: reason,
      },
    }),
  );
}

const rule = ".claude/rules/staging.md";

function judge(input) {
  const command = (input.tool_input && input.tool_input.command) || "";
  if (!opensPullRequest(command) || !inRepository(command, input.cwd)) {
    return;
  }
  const answer = declaredAnswer(command);
  if (answer === "") {
    deny(
      `Before \`gh pr create\` in this repository, ask the maintainer the staging question of the rule ${rule}: ` +
        "AskUserQuestion, in Portuguese, with three answers (open it with staging, open it without staging, do not open it). " +
        "With staging, run `task staging:on` first and stop if it fails. Then run the same command again, declaring the answer at its start: " +
        "`STAGING_QUESTION=with` with `--label staging`, or `STAGING_QUESTION=without` without it (in PowerShell, `$env:STAGING_QUESTION='with'; ` before the command).",
    );
    return;
  }
  const labelled = carriesStagingLabel(command);
  if (answer === "with" && !labelled) {
    deny(
      `The maintainer chose staging, so the pull request is opened with \`--label staging\` in this same command (rule ${rule}). ` +
        "Turn staging on first with `task staging:on` if it is not on.",
    );
    return;
  }
  if (answer === "without" && labelled) {
    deny(
      `The maintainer chose no staging, so the pull request is opened without the label \`staging\` (rule ${rule}).`,
    );
  }
}

let raw = "";
process.stdin.setEncoding("utf8");
process.stdin.on("data", (chunk) => {
  raw += chunk;
});
process.stdin.on("end", () => {
  try {
    judge(JSON.parse(raw || "{}"));
  } catch {
    // A payload this hook cannot read is left to the other checks.
  }
});
