local c = canvas.new(800, 600)
local img = canvas.load_image("data/src.png")

-- Draw the owl as background
c:draw_image(img, 0, 0, img.width, img.height, 0, 0, 800, 600)

-- Concave star polygon over the owl
c:set_fill_color(255, 50, 50, 80)
c:set_stroke_color(255, 255, 0)
c:set_line_width(3)

local cx, cy, outer, inner = 250, 300, 150, 60
local points = 5

c:begin_path()
for i = 0, points * 2 - 1 do
    local angle = (i * math.pi / points) - math.pi / 2
    local r = outer
    if i % 2 == 1 then r = inner end
    local x = cx + r * math.cos(angle)
    local y = cy + r * math.sin(angle)
    if i == 0 then
        c:move_to(x, y)
    else
        c:line_to(x, y)
    end
end
c:close_path()
c:fill()
c:stroke()

-- Arrow-shaped concave polygon
c:set_fill_color(50, 120, 255, 100)
c:set_stroke_color(200, 220, 255)
c:set_line_width(2)

c:begin_path()
c:move_to(500, 100)
c:line_to(700, 250)
c:line_to(620, 250)
c:line_to(620, 450)
c:line_to(580, 450)
c:line_to(580, 250)
c:line_to(500, 250)
c:close_path()
c:fill()
c:stroke()

-- L-shaped concave polygon
c:set_fill_color(50, 220, 100, 90)
c:set_stroke_color(180, 255, 200)
c:set_line_width(2)

c:begin_path()
c:move_to(450, 350)
c:line_to(550, 350)
c:line_to(550, 400)
c:line_to(500, 400)
c:line_to(500, 500)
c:line_to(450, 500)
c:close_path()
c:fill()
c:stroke()

c:save_png("out.png")
print("saved out.png")
