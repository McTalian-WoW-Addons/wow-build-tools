--- Boot-order smoke test.
---
--- Resolves an addon's real .toc/XML load order, loadfile()s every source
--- file against an auto-mocked WoW API, then fires ADDON_LOADED and
--- PLAYER_ENTERING_WORLD on whatever frames the addon registered — the same
--- two events that run during an actual client login. Nil-reference and
--- load-order bugs that would otherwise only show up as a Lua error on
--- login surface here instead, with no client required.
---
--- This is deliberately low-fidelity: a curated list of known WoW API
--- globals (KNOWN_WOW_APIS below) resolves to a silent chainable stub
--- rather than erroring; everything else is real nil. That list is
--- deliberate, not "any unknown global" — treating every miss as present
--- breaks the extremely common `local X = _G.Foo; if not X then ... end`
--- init-guard idiom (LibStub, loaded first by nearly every Ace3-based
--- addon, is exactly this pattern), since a stub reads as truthy. So the
--- only failures that surface are bugs in the addon's own code (nil
--- ns.Foo, wrong file order, syntax errors) or gaps in KNOWN_WOW_APIS —
--- not general WoW API coverage. That's the tradeoff that keeps this
--- cheap; full API fidelity is a separate, much bigger problem (see
--- wowless) that this does not attempt to solve.

local M = {}

--- Get just the filename from a path (no directory component).
local function basename(path)
	return path:match("([^/]+)$") or path
end

--- Get the directory name from a path.
local function dirname(path)
	return path:match("^(.*)/[^/]+$")
end

-- ==========================================================================
-- .toc / XML parsing
-- ==========================================================================

--- Parse a WoW `.toc` file into an ordered list of {type="script"|"include", file=...}
--- entries, honoring `#@name@ ... #@end-name@` build-directive blocks.
--- @param tocPath string Path to the .toc file
--- @param opts table|nil opts.exclude = {alpha = true, ...} — directive names whose
---   blocks should be skipped (default: none excluded, i.e. the most permissive
---   superset of what could ever load)
--- @return table entries Ordered {type, file} entries
function M.parseToc(tocPath, opts)
	opts = opts or {}
	local exclude = opts.exclude or {}

	local f = io.open(tocPath, "r")
	if not f then
		error("Could not open TOC file: " .. tocPath)
	end

	local entries = {}
	local stack = {}

	local function isExcluded()
		for _, name in ipairs(stack) do
			if exclude[name] then
				return true
			end
		end
		return false
	end

	for rawLine in f:lines() do
		local line = rawLine:match("^%s*(.-)%s*$")
		if line == "" then
			-- blank, skip
		elseif line:match("^##") then
			-- TOC metadata, skip
		else
			local endName = line:match("^#@end%-([%w%-]+)@$")
			local startName = (not endName) and line:match("^#@([%w%-]+)@$")
			if endName then
				for i = #stack, 1, -1 do
					if stack[i] == endName then
						table.remove(stack, i)
						break
					end
				end
			elseif startName then
				table.insert(stack, startName)
			elseif line:match("^#") then
				-- plain comment, skip
			elseif not isExcluded() then
				local normalized = line:gsub("\\", "/")
				local entryType = normalized:match("%.xml$") and "include" or "script"
				table.insert(entries, { type = entryType, file = normalized })
			end
		end
	end
	f:close()

	return entries
end

--- Parse a simple addon XML file for <Script file="..."/> and <Include file="..."/> entries.
--- @param xmlPath string Path to the XML file
--- Strip `<!-- ... -->` XML comments out of a line, tracking comment state
--- across lines since a comment can span more than one. Real-world addon
--- XML (including third-party libs like AceConfig-3.0) comments out
--- <Include>/<Script> entries this way to disable them without deleting
--- them, and those must not be treated as real load-order entries.
--- @return string strippedLine, boolean stillInComment
local function stripXmlComments(line, inComment)
	local result = {}
	local pos = 1
	while true do
		if inComment then
			local closeEnd = select(2, line:find("%-%->", pos))
			if closeEnd then
				inComment = false
				pos = closeEnd + 1
			else
				return table.concat(result), true
			end
		else
			local openStart, openEnd = line:find("<!%-%-", pos)
			if openStart then
				table.insert(result, line:sub(pos, openStart - 1))
				inComment = true
				pos = openEnd + 1
			else
				table.insert(result, line:sub(pos))
				return table.concat(result), false
			end
		end
	end
end

--- @return table entries Ordered {type, file} entries
function M.parseXml(xmlPath)
	local f = io.open(xmlPath, "r")
	if not f then
		error("Could not open XML file: " .. xmlPath)
	end

	local entries = {}
	local inComment = false
	for line in f:lines() do
		local stripped
		stripped, inComment = stripXmlComments(line, inComment)

		local scriptFile = stripped:match('<Script%s+file="([^"]+)"')
		if scriptFile then
			table.insert(entries, { type = "script", file = scriptFile:gsub("\\", "/") })
		end
		local includeFile = stripped:match('<Include%s+file="([^"]+)"')
		if includeFile then
			table.insert(entries, { type = "include", file = includeFile:gsub("\\", "/") })
		end
	end
	f:close()

	return entries
end

--- Recursively resolve a list of {type, file} entries (relative to basePath) into
--- an ordered list of absolute .lua file paths, recursing into .xml includes.
local function resolveEntries(entries, basePath)
	local luaFiles = {}
	for _, entry in ipairs(entries) do
		local fullPath = basePath .. "/" .. entry.file
		if entry.file:match("%.lua$") then
			table.insert(luaFiles, fullPath)
		elseif entry.type == "include" and entry.file:match("%.xml$") then
			local includeDir = dirname(fullPath) or basePath
			local nested = resolveEntries(M.parseXml(fullPath), includeDir)
			for _, nestedFile in ipairs(nested) do
				table.insert(luaFiles, nestedFile)
			end
		end
	end
	return luaFiles
end

--- Resolve the full, real load order for an addon starting from its .toc.
--- @param tocPath string Path to the addon's .toc file
--- @param opts table|nil see M.parseToc
--- @return table luaFiles Ordered list of absolute .lua file paths
function M.resolveLoadOrder(tocPath, opts)
	local base = dirname(tocPath) or "."
	return resolveEntries(M.parseToc(tocPath, opts), base)
end

-- ==========================================================================
-- Auto-mock WoW API
-- ==========================================================================

--- Known load-time WoW client API surface, safe to silently no-op when
--- unmocked. Deliberately curated rather than "any unknown global" — see
--- the module doc comment at the top of this file for why. Expect to
--- extend this as boot_sim runs against addons that reference more of the
--- WoW API at file-load top level (code inside functions never executes
--- during boot simulation, so it doesn't need to be listed here).
local KNOWN_WOW_APIS = {
	-- Blizzard namespaces/constants. Safe to include Enum here even though
	-- --mocks exists specifically to seed real Enum values: a real mocks
	-- file's globals bypass this fallback entirely (they're real _G entries,
	-- found before __index ever fires), so this only changes the no-mocks
	-- behavior — nil-vs-stub, not stub-vs-real-value.
	Enum = true,
	Constants = true,
	hooksecurefunc = true,
	securecall = true,
	RunNextFrame = true,
	SlashCmdList = true,
	StaticPopupDialogs = true,
	C_AddOns = true,
	C_Timer = true,
	C_ChatInfo = true,
	C_Item = true,
	C_CVar = true,
	C_TransmogCollection = true,
	EventRegistry = true,
	EventUtil = true,
	TooltipDataProcessor = true,
	GameTooltip = true,
	GameTooltip_AddColoredLine = true,
	ItemRefTooltip = true,
	UIParent = true,
	WorldFrame = true,
	Minimap = true,
	LinkUtil = true,
	LIGHTBLUE_FONT_COLOR = true,
	HIGHLIGHT_FONT_COLOR = true,
	NORMAL_FONT_COLOR = true,
	RED_FONT_COLOR = true,
	GREEN_FONT_COLOR = true,
	-- bit.band/bor/etc return numbers used in further bitwise math — a
	-- chain-stub return breaks that the same way an untyped stub breaks any
	-- typed use. Left as a stub anyway (not in realImplementations below):
	-- a faithful bit library is out of scope here, same reasoning as not
	-- attempting general WoW API fidelity. Addons doing real bitwise math
	-- at load time (ChatThrottleLib does) need --mocks for a real one.
	bit = true,
}

--- Build a silent, infinitely-chainable stub: any field access or call on it
--- returns another instance of itself. Used as the __index fallback for _G so
--- unmocked WoW API globals (functions, namespaces, widget methods) resolve
--- to a no-op instead of raising "attempt to call a nil value".
local function newChainStub()
	local stub
	stub = setmetatable({}, {
		__index = function()
			return stub
		end,
		__call = function()
			return stub
		end,
	})
	return stub
end

--- Build a frame mock that actually records RegisterEvent/SetScript calls, so
--- M.simulateBoot can fire ADDON_LOADED/PLAYER_ENTERING_WORLD against it afterward.
--- Any method beyond these three is a silent no-op, matching the rest of the mock.
local function newFrameMock(registry)
	local frame = { events = {} }
	local mt = {
		__index = function(_, key)
			if key == "RegisterEvent" then
				return function(_, event)
					frame.events[event] = true
				end
			elseif key == "UnregisterEvent" then
				return function(_, event)
					frame.events[event] = nil
				end
			elseif key == "SetScript" then
				return function(_, handlerName, fn)
					if handlerName == "OnEvent" then
						frame.onEvent = fn
					end
				end
			else
				return newChainStub()
			end
		end,
	}
	setmetatable(frame, mt)
	table.insert(registry, frame)
	return frame
end

--- Real recursive table copy, matching the shape of WoW's own CopyTable —
--- addon code uses its result as a real table (indexing, iterating), so a
--- chain-stub return would silently diverge the moment anything checked a
--- copied value.
local function deepCopyTable(t)
	if type(t) ~= "table" then
		return t
	end
	local copy = {}
	for k, v in pairs(t) do
		copy[k] = deepCopyTable(v)
	end
	return copy
end

--- Globals whose real return value gets used arithmetically, concatenated,
--- or as a boolean/string downstream — a generic chain-stub (a table) breaks
--- that the same way an unmocked LibStub broke `LibStub.minor < N`. These
--- get a real, typed implementation instead of a KNOWN_WOW_APIS stub.
--- frames is threaded through so CreateFrame's mock can register into the
--- same registry M.simulateBoot fires login events against.
---
--- installAutoMock (below) only applies these where nothing already set the
--- global first — i.e. an addon's own --mocks file wins for all of these
--- except CreateFrame, which always stays this one (see installAutoMock).
local function realImplementations(frames)
	return {
		CopyTable = deepCopyTable,
		CreateFrame = function()
			return newFrameMock(frames)
		end,
		GetRealmName = function()
			return "TestRealm"
		end,
		GetTime = function()
			return 0
		end,
		UnitName = function()
			return "TestUnit", nil
		end,
		UnitClass = function()
			return "Warrior", "WARRIOR", 1
		end,
		UnitRace = function()
			return "Human", "Human", 1
		end,
		GetCurrentRegion = function()
			return 1
		end,
		-- Real client behavior when not connected to a Battle.net session --
		-- also the most accurate state for a boot simulation to model.
		BNGetInfo = function() end,
		GetLocale = function()
			return "enUS"
		end,
		GetExpansionLevel = function()
			return 10
		end,
		CreateAtlasMarkup = function()
			return ""
		end,
		IsLoggedIn = function()
			return true
		end,
		IsAddOnLoaded = function()
			return true
		end,
		issecure = function()
			return true
		end,
		UnitFactionGroup = function()
			return "Alliance", "Alliance"
		end,
		GetItemInfoFromHyperlink = function()
			return 0
		end,
		wipe = function(t)
			if type(t) == "table" then
				for k in pairs(t) do
					t[k] = nil
				end
			end
			return t
		end,
	}
end

--- Install the auto-mock over _G for the duration of a boot simulation.
--- Only affects globals that aren't already real (Lua stdlib, busted helper
--- stubs, whatever the addon itself defines as it loads) — those pass through
--- untouched. Returns a restore() function that must be called afterward.
local function installAutoMock()
	local frames = {}
	local previousMt = getmetatable(_G)
	local impls = realImplementations(frames)
	local previousValues = {}

	for key, impl in pairs(impls) do
		-- CreateFrame is core simulator machinery, not just a convenience
		-- mock: M.simulateBoot's login-event firing only sees frames created
		-- through this tracked mock, so it always wins regardless of
		-- --mocks. Every other entry here defers to a real value an
		-- addon's own --mocks file already set (RPGLootFeed's busted
		-- helper, for example, defines its own CreateFrame/UnitClass/
		-- GetExpansionLevel/RunNextFrame) rather than silently overwriting
		-- it.
		if key == "CreateFrame" or rawget(_G, key) == nil then
			previousValues[key] = rawget(_G, key)
			rawset(_G, key, impl)
		end
	end

	setmetatable(_G, {
		__index = function(_, key)
			if KNOWN_WOW_APIS[key] then
				return newChainStub()
			end
			return nil
		end,
	})

	local function restore()
		setmetatable(_G, previousMt)
		for key, value in pairs(previousValues) do
			rawset(_G, key, value)
		end
	end

	return frames, restore
end

-- ==========================================================================
-- Boot simulation
-- ==========================================================================

--- Load every file in the addon's real load order and fire the two events
--- that drive addon initialization during a real client login.
--- @param tocPath string Path to the addon's .toc file
--- @param addonName string Addon name, as passed in the `local addonName, ns = ...` vararg
--- @param opts table|nil opts.exclude passed through to M.parseToc/resolveLoadOrder
--- @return boolean ok True if every file loaded and every event handler ran without error
--- @return table results Ordered list of {phase, file|event, ok, err}
function M.simulateBoot(tocPath, addonName, opts)
	local luaFiles = M.resolveLoadOrder(tocPath, opts)
	local frames, restore = installAutoMock()

	local results = {}
	local overallOk = true
	local ns = {}

	for _, luaPath in ipairs(luaFiles) do
		local fn, loadErr = loadfile(luaPath)
		if not fn then
			overallOk = false
			table.insert(results, { phase = "load", file = luaPath, ok = false, err = loadErr })
		else
			local ok, err = pcall(fn, addonName, ns)
			if not ok then
				overallOk = false
			end
			table.insert(results, { phase = "load", file = luaPath, ok = ok, err = err })
		end
	end

	for _, event in ipairs({ "ADDON_LOADED", "PLAYER_ENTERING_WORLD" }) do
		for _, frame in ipairs(frames) do
			if frame.onEvent and frame.events[event] then
				local ok, err
				if event == "ADDON_LOADED" then
					ok, err = pcall(frame.onEvent, frame, event, addonName)
				else
					-- isLogin, isReload — simulate a fresh login
					ok, err = pcall(frame.onEvent, frame, event, true, false)
				end
				if not ok then
					overallOk = false
				end
				table.insert(results, { phase = "event", event = event, ok = ok, err = err })
			end
		end
	end

	restore()

	return overallOk, results
end

M.basename = basename
M.dirname = dirname

return M
