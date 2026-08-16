--- CLI entrypoint for boot_sim.lua, invoked as a subprocess by the Go
--- `wow-build-tools boot-sim` command against a real system Lua interpreter
--- (not gopher-lua — see the package doc comment on bootsim.Run for why).
---
---   lua boot_sim_cli.lua <path/to/Addon.toc> <AddonName> [mockPath]
---
--- Prints one line of JSON to stdout: {"ok":bool,"steps":[{"phase","label","ok","err"},...]}
--- Exits 0 if every step passed, 1 if the simulation ran but something
--- failed, 2 on a setup error (bad args, unreadable mocks file, crash before
--- results exist) — diagnostics for setup errors go to stderr, not stdout.

local scriptDir = arg[0]:match("^(.*)/[^/]+$") or "."
package.path = scriptDir .. "/?.lua;" .. package.path

-- Addon code loaded during simulation can print() on its own (alpha
-- slash-command registration, etc.) — redirect that to stderr so stdout
-- stays pure JSON, exactly the one line this script prints itself.
_G.print = function(...)
	local parts = {}
	for i = 1, select("#", ...) do
		parts[i] = tostring(select(i, ...))
	end
	io.stderr:write(table.concat(parts, "\t") .. "\n")
end

local tocPath = arg[1]
local addonName = arg[2]
local mockPath = arg[3]

if not tocPath or not addonName then
	io.stderr:write("usage: lua boot_sim_cli.lua <path/to/Addon.toc> <AddonName> [mockPath]\n")
	os.exit(2, true)
end

if mockPath and mockPath ~= "" then
	local mockFn, mockLoadErr = loadfile(mockPath)
	if not mockFn then
		io.stderr:write("failed to load mocks file " .. mockPath .. ": " .. tostring(mockLoadErr) .. "\n")
		os.exit(2, true)
	end
	local mockOk, mockRunErr = pcall(mockFn)
	if not mockOk then
		io.stderr:write("mocks file " .. mockPath .. " crashed: " .. tostring(mockRunErr) .. "\n")
		os.exit(2, true)
	end
end

local boot_sim = require("boot_sim")

local callOk, simOk, results = pcall(boot_sim.simulateBoot, tocPath, addonName)
if not callOk then
	io.stderr:write("boot simulation crashed before producing results: " .. tostring(simOk) .. "\n")
	os.exit(2, true)
end

--- Minimal JSON string escaper — the only values this ever serializes are
--- file paths, event names, and Lua error messages, so this doesn't need to
--- be a general-purpose encoder.
local function jsonString(value)
	local escaped = tostring(value):gsub('[%c\\"]', function(c)
		if c == '"' then
			return '\\"'
		elseif c == "\\" then
			return "\\\\"
		elseif c == "\n" then
			return "\\n"
		elseif c == "\r" then
			return "\\r"
		elseif c == "\t" then
			return "\\t"
		else
			return string.format("\\u%04x", c:byte())
		end
	end)
	return '"' .. escaped .. '"'
end

local parts = {}
for _, step in ipairs(results) do
	local label = step.phase == "load" and step.file or step.event
	local errField = step.err ~= nil and jsonString(step.err) or "null"
	table.insert(
		parts,
		string.format(
			'{"phase":%s,"label":%s,"ok":%s,"err":%s}',
			jsonString(step.phase),
			jsonString(label),
			tostring(step.ok),
			errField
		)
	)
end

-- Bypasses the print() override above on purpose: this is the one line of
-- real stdout output this script produces.
io.stdout:write(string.format('{"ok":%s,"steps":[%s]}\n', tostring(simOk), table.concat(parts, ",")))

os.exit(simOk and 0 or 1, true)
