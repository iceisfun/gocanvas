local c = canvas.new(800, 600)

-- Light gray background.
c:set_fill_color(240, 240, 240)
c:fill_rect(0, 0, 800, 600)

-- ============================================================
-- 1. Rounded rectangle (classic arcTo use case) - top left
-- ============================================================
local function rounded_rect(cx, x, y, w, h, r)
    cx:begin_path()
    cx:move_to(x + r, y)
    cx:arc_to(x + w, y,     x + w, y + h, r)  -- top-right corner
    cx:arc_to(x + w, y + h, x,     y + h, r)  -- bottom-right corner
    cx:arc_to(x,     y + h, x,     y,     r)  -- bottom-left corner
    cx:arc_to(x,     y,     x + w, y,     r)  -- top-left corner
    cx:close_path()
end

-- Filled rounded rect.
c:set_fill_color(70, 130, 220)
rounded_rect(c, 30, 30, 200, 120, 20)
c:fill()

-- Stroked rounded rect.
c:set_stroke_color(30, 80, 180)
c:set_line_width(3)
rounded_rect(c, 30, 30, 200, 120, 20)
c:stroke()

-- ============================================================
-- 2. Rounded triangle (rounded polygon) - top right
-- ============================================================
local function rounded_polygon(cx, pts, r)
    local n = #pts
    cx:begin_path()
    -- Start at the midpoint of the last-to-first edge direction.
    local lx, ly = pts[n][1], pts[n][2]
    local fx, fy = pts[1][1], pts[1][2]
    cx:move_to((lx + fx) / 2, (ly + fy) / 2)
    for i = 1, n do
        local cur = pts[i]
        local nxt = pts[(i % n) + 1]
        cx:arc_to(cur[1], cur[2], nxt[1], nxt[2], r)
    end
    cx:close_path()
end

c:set_fill_color(220, 100, 50)
rounded_polygon(c, {
    {450, 30},
    {560, 160},
    {340, 160},
}, 15)
c:fill()

c:set_stroke_color(180, 60, 20)
c:set_line_width(3)
rounded_polygon(c, {
    {450, 30},
    {560, 160},
    {340, 160},
}, 15)
c:stroke()

-- ============================================================
-- 3. Varying radii at each corner - middle row
-- ============================================================
c:set_fill_color(80, 180, 80)
c:begin_path()
local x, y, w, h = 30, 220, 250, 140
c:move_to(x + 5, y)
c:arc_to(x + w, y,     x + w, y + h, 5)   -- top-right: small
c:arc_to(x + w, y + h, x,     y + h, 40)  -- bottom-right: large
c:arc_to(x,     y + h, x,     y,     20)  -- bottom-left: medium
c:arc_to(x,     y,     x + w, y,     5)   -- top-left: small
c:close_path()
c:fill()

c:set_stroke_color(40, 120, 40)
c:set_line_width(2)
c:begin_path()
c:move_to(x + 5, y)
c:arc_to(x + w, y,     x + w, y + h, 5)
c:arc_to(x + w, y + h, x,     y + h, 40)
c:arc_to(x,     y + h, x,     y,     20)
c:arc_to(x,     y,     x + w, y,     5)
c:close_path()
c:stroke()

-- ============================================================
-- 4. Sharp vs rounded comparison - bottom half
-- ============================================================

-- Sharp star (line_to only).
c:set_fill_color(200, 200, 220)
c:set_stroke_color(100, 100, 160)
c:set_line_width(2)

local function star_points(cx_pos, cy_pos, outer, inner, n)
    local pts = {}
    for i = 0, 2 * n - 1 do
        local angle = math.pi * i / n - math.pi / 2
        local r_val = (i % 2 == 0) and outer or inner
        pts[#pts + 1] = {
            cx_pos + math.cos(angle) * r_val,
            cy_pos + math.sin(angle) * r_val,
        }
    end
    return pts
end

-- Sharp version (left).
local sharp_pts = star_points(150, 480, 80, 35, 5)
c:begin_path()
c:move_to(sharp_pts[1][1], sharp_pts[1][2])
for i = 2, #sharp_pts do
    c:line_to(sharp_pts[i][1], sharp_pts[i][2])
end
c:close_path()
c:fill()
c:begin_path()
c:move_to(sharp_pts[1][1], sharp_pts[1][2])
for i = 2, #sharp_pts do
    c:line_to(sharp_pts[i][1], sharp_pts[i][2])
end
c:close_path()
c:stroke()

-- Rounded version (right).
c:set_fill_color(220, 200, 230)
c:set_stroke_color(140, 80, 180)
local round_pts = star_points(400, 480, 80, 35, 5)
rounded_polygon(c, round_pts, 10)
c:fill()
rounded_polygon(c, round_pts, 10)
c:stroke()

-- Rounded hexagon.
c:set_fill_color(255, 210, 100)
c:set_stroke_color(200, 150, 30)
c:set_line_width(3)
local hex_pts = {}
for i = 0, 5 do
    local angle = math.pi * i / 3 - math.pi / 6
    hex_pts[#hex_pts + 1] = {
        650 + math.cos(angle) * 70,
        480 + math.sin(angle) * 70,
    }
end
rounded_polygon(c, hex_pts, 12)
c:fill()
rounded_polygon(c, hex_pts, 12)
c:stroke()

-- ============================================================
-- Labels (without font, use simple rectangles as labels).
-- ============================================================

c:save_png("out.png")
print("saved out.png")
