// SessionStart hook, wired in .claude/settings.json.
//
// A session that starts on a branch whose pull request was merged or closed,
// or on a main that is behind origin/main, reads an old tree until it
// notices: on 2026-11-26 a session searched a checkout ten commits behind
// and had to search again on origin/main. This hook fetches origin and, only
// when one of those holds, tells the agent in a line or two, before it reads
// anything. It changes nothing in the checkout, since another session or the
// maintainer may be working in it, and on any failure it says nothing.

const { execFileSync } = require("node:child_process");

const checkout = process.env.CLAUDE_PROJECT_DIR || process.cwd();

// run gives what the command printed, or null when it failed or took too long.
function run(command, args, timeout) {
  try {
    return execFileSync(command, args, {
      cwd: checkout,
      timeout,
      encoding: "utf8",
      stdio: ["ignore", "pipe", "ignore"],
      env: { ...process.env, GIT_TERMINAL_PROMPT: "0" },
      windowsHide: true,
    }).trim();
  } catch {
    return null;
  }
}

// behindMain tells how many commits main lacks of origin/main.
function behindMain() {
  const count = Number(run("git", ["rev-list", "--count", "main..origin/main"], 5000));
  return Number.isInteger(count) ? count : 0;
}

// pullRequestOf gives the newest pull request whose head is the branch, or null.
function pullRequestOf(branch) {
  const listed = run("gh", ["pr", "list", "--head", branch, "--state", "all", "--limit", "1", "--json", "number,state"], 15000);
  if (!listed) {
    return null;
  }
  try {
    return JSON.parse(listed)[0] || null;
  } catch {
    return null;
  }
}

function notes() {
  run("git", ["fetch", "--quiet", "--prune", "origin"], 10000);
  const branch = run("git", ["branch", "--show-current"], 5000);
  if (!branch) {
    return [];
  }
  const latest = run("git", ["log", "-1", "--format=%h %s", "origin/main"], 5000) || "origin/main";
  if (branch === "main") {
    const behind = behindMain();
    return behind > 0
      ? [`This checkout is on main, ${behind} commits behind origin/main (${latest}): its files are older than main. Bring it up to date with \`git merge --ff-only -q origin/main\` before reading them, unless the maintainer or another session is working in it.`]
      : [];
  }
  const pullRequest = pullRequestOf(branch);
  if (!pullRequest || pullRequest.state === "OPEN") {
    return [];
  }
  const ended = pullRequest.state === "MERGED" ? "was merged" : "was closed";
  return [
    `This checkout is on the branch ${branch}, whose pull request #${pullRequest.number} ${ended}; origin/main is at ${latest}. Read the files of main with \`git show origin/main:<path>\`, or switch to main or to a branch cut from origin/main, before reading the tree. Ask before switching if the working tree has changes.`,
  ];
}

const lines = notes();
if (lines.length > 0) {
  process.stdout.write(
    JSON.stringify({ hookSpecificOutput: { hookEventName: "SessionStart", additionalContext: lines.join("\n") } }),
  );
}
