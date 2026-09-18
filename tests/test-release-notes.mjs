import { readFileSync } from "node:fs";
import { generateNotes } from "@semantic-release/release-notes-generator";

// Validate the shared addon config (.releaserc.json) as-is, so a preset that
// silently ignores presetConfig.types fails here instead of shipping empty notes.
const config = JSON.parse(readFileSync(".releaserc.json", "utf8"));
const notesPlugin = config.plugins.find(
  (p) =>
    Array.isArray(p) && p[0] === "@semantic-release/release-notes-generator",
);
if (!notesPlugin) {
  console.error(
    "❌ release-notes-generator plugin not found in .releaserc.json",
  );
  process.exit(1);
}
const pluginConfig = notesPlugin[1];
const types = pluginConfig.presetConfig.types;

const context = {
  commits: types.map((t, i) => ({
    hash: `abc${i}`,
    message: `${t.type}: test ${t.type} commit for release notes`,
  })),
  lastRelease: { version: "1.2.2", gitTag: "v1.2.2", gitHead: "parent123" },
  nextRelease: { version: "1.3.0", gitTag: "v1.3.0", gitHead: "head456" },
  options: {
    repositoryUrl: "https://github.com/McTalian-WoW-Addons/wow-build-tools",
  },
  cwd: process.cwd(),
};

const notes = await generateNotes(pluginConfig, context);
console.log("Generated notes:");
console.log(notes);
console.log("\n--- Validation ---");

const errors = [];
if (!notes.trim()) errors.push("Notes are empty");

const compareUrlMatch = notes.match(
  /\(https:\/\/github\.com\/[^)]+\/compare\/[^)]+\)/,
);
if (!compareUrlMatch) errors.push("Missing compare URL");

// Every configured section must render with its configured heading.
for (const { type, section } of types) {
  if (!notes.includes(`### ${section}`)) {
    errors.push(`Missing section "${section}" for type "${type}"`);
  }
  if (!notes.includes(`test ${type} commit`)) {
    errors.push(`Missing commit entry for type "${type}"`);
  }
}

console.log("First 100 chars:", JSON.stringify(notes.substring(0, 100)));
if (!notes.match(/^#+ \[\d+\.\d+\.\d+\]/m)) {
  errors.push("Missing version header ([version])");
}

if (errors.length > 0) {
  console.error("\n❌ Validation failed:");
  errors.forEach((e) => console.error(`  - ${e}`));
  process.exit(1);
}

console.log("\n✅ Release notes validation passed");
