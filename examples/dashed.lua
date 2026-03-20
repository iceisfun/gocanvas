local c = canvas.new(800, 500)

-- Dark background
c:set_fill_color(25, 25, 35)
c:fill_rect(0, 0, 800, 500)

--------------------------------------------------------------
-- 1) Filled polygon with dashed perimeter
--------------------------------------------------------------

-- Hexagon
local cx, cy, r = 200, 200, 120
local sides = 6

c:begin_path()
for i = 0, sides - 1 do
    local angle = (i * 2 * math.pi / sides) - math.pi / 2
    local x = cx + r * math.cos(angle)
    local y = cy + r * math.sin(angle)
    if i == 0 then
        c:move_to(x, y)
    else
        c:line_to(x, y)
    end
end
c:close_path()

-- Fill with translucent teal
c:set_fill_color(0, 200, 180, 60)
c:fill()

-- Dashed stroke
c:set_line_dash({15, 8})
c:set_stroke_color(0, 255, 220)
c:set_line_width(3)
c:stroke()

-- Pentagon with different dash pattern
local cx2, cy2, r2 = 500, 200, 100

c:begin_path()
for i = 0, 4 do
    local angle = (i * 2 * math.pi / 5) - math.pi / 2
    local x = cx2 + r2 * math.cos(angle)
    local y = cy2 + r2 * math.sin(angle)
    if i == 0 then
        c:move_to(x, y)
    else
        c:line_to(x, y)
    end
end
c:close_path()

c:set_fill_color(200, 100, 255, 50)
c:fill()

-- Dot-dash pattern
c:set_line_dash({20, 6, 4, 6})
c:set_stroke_color(200, 150, 255)
c:set_line_width(2)
c:stroke()

--------------------------------------------------------------
-- 2) Glowing dashed lines using layered shadows
--------------------------------------------------------------
c:set_line_dash({20, 12})

-- Outer glow layer (wide, blurred)
c:set_shadow({
    color = {0, 180, 255, 200},
    blur = 12,
    offset_x = 0,
    offset_y = 0,
})
c:set_stroke_color(0, 140, 255)
c:set_line_width(3)

c:begin_path()
c:move_to(80, 400)
c:line_to(250, 370)
c:line_to(400, 420)
c:line_to(550, 360)
c:line_to(720, 400)
c:stroke()

c:clear_shadow()

-- Bright core on top (no shadow, thinner)
c:set_stroke_color(150, 220, 255)
c:set_line_width(1)
c:begin_path()
c:move_to(80, 400)
c:line_to(250, 370)
c:line_to(400, 420)
c:line_to(550, 360)
c:line_to(720, 400)
c:stroke()

-- Glowing dashed circle
c:set_shadow({
    color = {255, 60, 120, 180},
    blur = 10,
    offset_x = 0,
    offset_y = 0,
})
c:set_line_dash({12, 8})
c:set_stroke_color(255, 80, 140)
c:set_line_width(3)

c:begin_path()
c:arc(680, 180, 60, 0, math.pi * 2)
c:stroke()

c:clear_shadow()

-- Bright core
c:set_stroke_color(255, 180, 200)
c:set_line_width(1)
c:begin_path()
c:arc(680, 180, 60, 0, math.pi * 2)
c:stroke()

-- Reset dash for clean state
c:set_line_dash({})

c:save_png("out.png")
print("saved out.png")
