import { generateNotes } from "@semantic-release/release-notes-generator";

const pluginConfig = {
  preset: "angular",
  presetConfig: {
    types: [
      { type: "feat", section: "Features" },
      { type: "fix", section: "Bug Fixes" },
      { type: "perf", section: "Performance Improvements" },
      { type: "locale", section: "i18n Translations" },
      { type: "toc", section: "TOC Version Changes" },
    ],
  },
};

const context = {
  commits: [
    { hash: "abc123", message: "feat: test feature for release notes" },
    { hash: "def456", message: "fix: test bug fix for release notes" },
    { hash: "ghi789", message: "perf: test performance improvement" },
    { hash: "jkl012", message: "locale: update translations" },
    { hash: "mno345", message: "toc: bump interface version" },
  ],
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
if (!notes.includes("https://github.com/")) errors.push("Missing compare URL");

const expectedSections = [
  "Features",
  "Bug Fixes",
  "Performance Improvements",
  "i18n Translations",
  "TOC Version Changes",
];

const foundSections = expectedSections.filter((s) =>
  notes.includes(`### ${s}`),
);
console.log(`Found sections: ${foundSections.join(", ")}`);

if (foundSections.length === 0) {
  errors.push("No expected sections found");
}

// Check version header format (angular uses #)
console.log("First 100 chars:", JSON.stringify(notes.substring(0, 100)));
if (!notes.match(/# \[\d+\.\d+\.\d+\]/)) {
  errors.push("Missing version header (# [version])");
}

if (errors.length > 0) {
  console.error("\n❌ Validation failed:");
  errors.forEach((e) => console.error(`  - ${e}`));
  process.exit(1);
}

console.log("\n✅ Release notes validation passed");
