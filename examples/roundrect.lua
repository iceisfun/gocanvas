-- Rounded rectangle showcase: UI-style panels, buttons, cards, and gauges.
local c = canvas.new(800, 600)

-- Dark background
c:set_fill_color(30, 30, 40)
c:fill_rect(0, 0, 800, 600)

-- ── Header panel ──────────────────────────────────────────────
c:set_fill_color(45, 50, 65)
c:fill_round_rect(20, 15, 760, 55, 12)
c:set_fill_color(200, 210, 230)
c:fill_round_rect(30, 25, 10, 35, 5)   -- accent bar
c:set_fill_color(100, 120, 160)
c:fill_round_rect(50, 25, 120, 35, 8)  -- title badge

-- ── Sidebar ───────────────────────────────────────────────────
c:set_fill_color(40, 44, 58)
c:fill_round_rect(20, 85, 180, 495, 14)

-- Sidebar nav items
local nav_labels = {"Dashboard", "Sensors", "Alerts", "Settings"}
for i, _ in ipairs(nav_labels) do
    local y = 100 + (i - 1) * 55
    if i == 1 then
        -- Active item highlight
        c:set_fill_color(60, 130, 220, 60)
        c:fill_round_rect(30, y, 160, 42, 10)
        c:set_stroke_color(60, 130, 220)
        c:set_line_width(2)
        c:stroke_round_rect(30, y, 160, 42, 10)
    else
        c:set_fill_color(55, 60, 78)
        c:fill_round_rect(30, y, 160, 42, 10)
    end
end

-- ── Main content area ─────────────────────────────────────────

-- Status cards row
local card_colors = {
    {46, 204, 113},   -- green
    {52, 152, 219},   -- blue
    {231, 76, 60},    -- red
    {241, 196, 15},   -- yellow
}
local card_w = 155
local card_h = 100
for i = 1, 4 do
    local x = 215 + (i - 1) * (card_w + 15)
    local y = 85
    local col = card_colors[i]

    -- Card background
    c:set_fill_color(50, 55, 72)
    c:fill_round_rect(x, y, card_w, card_h, 12)

    -- Colored accent strip at top
    c:set_fill_color(col[1], col[2], col[3])
    c:fill_round_rect(x, y, card_w, 6, 3)

    -- Inner value display
    c:set_fill_color(col[1], col[2], col[3], 40)
    c:fill_round_rect(x + 12, y + 40, card_w - 24, 45, 8)
end

-- ── Large panel with gauge-style indicators ───────────────────
c:set_fill_color(42, 46, 62)
c:fill_round_rect(215, 200, 350, 280, 16)

-- Panel border
c:set_stroke_color(70, 80, 100)
c:set_line_width(1)
c:stroke_round_rect(215, 200, 350, 280, 16)

-- Gauge bars inside the panel
for i = 0, 5 do
    local y = 225 + i * 40
    local bar_w = 300
    local fill_pct = 0.3 + 0.1 * i

    -- Track
    c:set_fill_color(35, 38, 50)
    c:fill_round_rect(240, y, bar_w, 22, 11)

    -- Fill (pill-shaped: radius = height / 2)
    local fill_w = bar_w * fill_pct
    if fill_w > 22 then
        local r = math.min(fill_w / 2, 150 + i * 20)
        local g = math.max(220 - i * 30, 80)
        local b = 80
        c:set_fill_color(r, g, b)
        c:fill_round_rect(240, y, fill_w, 22, 11)
    end
end

-- ── Right-side panel ──────────────────────────────────────────
c:set_fill_color(42, 46, 62)
c:fill_round_rect(580, 200, 200, 280, 16)
c:set_stroke_color(70, 80, 100)
c:set_line_width(1)
c:stroke_round_rect(580, 200, 200, 280, 16)

-- Stacked info tiles
for i = 0, 3 do
    local y = 215 + i * 62
    c:set_fill_color(55, 60, 80)
    c:fill_round_rect(595, y, 170, 50, 10)

    -- Small colored dot
    local dot_col = card_colors[(i % 4) + 1]
    c:set_fill_color(dot_col[1], dot_col[2], dot_col[3])
    c:fill_round_rect(605, y + 18, 14, 14, 7) -- circle via pill
end

-- ── Bottom button bar ─────────────────────────────────────────
local btn_colors = {
    {52, 152, 219},   -- blue
    {46, 204, 113},   -- green
    {155, 89, 182},   -- purple
    {231, 76, 60},    -- red
}
for i = 1, 4 do
    local x = 215 + (i - 1) * 145
    local col = btn_colors[i]

    -- Button
    c:set_fill_color(col[1], col[2], col[3])
    c:fill_round_rect(x, 500, 130, 40, 8)

    -- Button hover highlight (lighter inner rect)
    c:set_fill_color(255, 255, 255, 30)
    c:fill_round_rect(x + 3, 502, 124, 20, 6)
end

-- ── Radius comparison strip at very bottom ────────────────────
c:set_fill_color(200, 210, 230, 120)
local radii = {0, 4, 8, 16, 30, 50}
for i, r in ipairs(radii) do
    local x = 20 + (i - 1) * 130
    c:set_fill_color(60 + i * 25, 90 + i * 20, 180, 180)
    c:fill_round_rect(x, 555, 120, 35, r)
    c:set_stroke_color(200, 210, 230)
    c:set_line_width(1)
    c:stroke_round_rect(x, 555, 120, 35, r)
end

c:save_png("out.png")
print("saved out.png")
