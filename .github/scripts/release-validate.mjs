#!/usr/bin/env node
import { analyzeCommits } from "@semantic-release/commit-analyzer";
import { execa } from "execa";
import { readFileSync } from "fs";
import { resolve } from "path";

const ROOT = resolve(process.cwd());

async function getCommits(base, head) {
  const { stdout } = await execa(
    "git",
    ["log", "--oneline", "--format=%H %s", `${base}..${head}`],
    { cwd: ROOT },
  );
  return stdout
    .trim()
    .split("\n")
    .filter(Boolean)
    .map((line) => {
      const [hash, ...msgParts] = line.split(" ");
      return { hash, message: msgParts.join(" ") };
    });
}

async function getAnalyzerConfig() {
  const config = JSON.parse(
    readFileSync(resolve(ROOT, ".releaserc.json"), "utf8"),
  );
  const analyzer = config.plugins.find(
    (p) => Array.isArray(p) && p[0] === "@semantic-release/commit-analyzer",
  );
  return analyzer?.[1] || {};
}

async function getReleaseType(commits, analyzerConfig) {
  const releaseType = await analyzeCommits(
    {
      preset: analyzerConfig.preset || "angular",
      releaseRules: analyzerConfig.releaseRules,
      config: analyzerConfig.config,
      parserOpts: analyzerConfig.parserOpts,
    },
    {
      commits,
      cwd: ROOT,
      logger: { log: () => {} },
    },
  );
  return releaseType || "no-release";
}

async function validateTitle(title, analyzerConfig) {
  try {
    const commits = [{ hash: "title", message: title }];
    const releaseType = await getReleaseType(commits, analyzerConfig);
    const type = title.match(/^([^:!( ]+)/)?.[1] || "unknown";
    const valid = type !== "unknown";
    return { valid, type, releaseType: releaseType || "no-release" };
  } catch {
    return { valid: false, type: "unknown", releaseType: "no-release" };
  }
}

async function main() {
  const args = process.argv.slice(2);
  const command = args[0];

  if (command === "validate-pr") {
    const { baseSha, headSha, title, squashAllowed, rebaseAllowed } = {
      baseSha: args[1],
      headSha: args[2],
      title: args[3],
      squashAllowed: args[4] !== "false",
      rebaseAllowed: args[5] !== "false",
    };

    const analyzerConfig = await getAnalyzerConfig();
    const results = {};

    // Validate squash (title)
    if (squashAllowed) {
      const titleResult = await validateTitle(title, analyzerConfig);
      results.squashValid = titleResult.valid;
      results.titleType = titleResult.type;
      results.titleReleaseType = titleResult.releaseType;
    } else {
      results.squashValid = false;
    }

    // Validate rebase (all commits)
    if (rebaseAllowed) {
      const commits = await getCommits(baseSha, headSha);
      const commitReleaseType = await getReleaseType(commits, analyzerConfig);
      results.rebaseValid = commitReleaseType !== "no-release";
      results.commitReleaseType = commitReleaseType;
    } else {
      results.rebaseValid = false;
    }

    // Determine overall validity
    results.valid = results.squashValid || results.rebaseValid;

    // Check consistency if both valid
    if (results.squashValid && results.rebaseValid) {
      results.consistent =
        results.titleReleaseType === results.commitReleaseType;
      if (!results.consistent) {
        results.error = `PR title implies ${results.titleReleaseType} but commits imply ${results.commitReleaseType}`;
      }
    }

    // Determine labels
    if (results.squashValid && !results.rebaseValid) {
      results.labels = ["squash valid", results.titleReleaseType];
    } else if (!results.squashValid && results.rebaseValid) {
      results.labels = ["rebase valid", results.commitReleaseType];
    } else if (results.squashValid && results.rebaseValid) {
      if (results.consistent) {
        results.labels = [
          "squash valid",
          "rebase valid",
          results.titleReleaseType,
        ];
      }
    }

    console.log(JSON.stringify(results, null, 2));
    process.exit(results.valid ? 0 : 1);
  }

  console.error(
    "Usage: node release-validate.mjs validate-pr <baseSha> <headSha> <title> [squashAllowed] [rebaseAllowed]",
  );
  process.exit(1);
}

main().catch((err) => {
  console.error(
    JSON.stringify({ error: err.message, stack: err.stack }, null, 2),
  );
  process.exit(1);
});
