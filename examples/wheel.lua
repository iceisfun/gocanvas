local c = canvas.new(600, 600)
local w, h = 600, 600

-- Black background
c:set_fill_color(0, 0, 0)
c:fill_rect(0, 0, w, h)

-- Colored pie slices
local step = math.pi * 0.1
for i = 0, 19 do
    local r = i * step
    c:set_fill_color(
        math.floor(r * 10) % 256,
        math.floor(r * 20) % 256,
        math.floor(r * 40) % 256
    )
    c:begin_path()
    c:move_to(w * 0.5, h * 0.5)
    c:arc(w * 0.5, h * 0.5, math.min(w, h) * 0.4, r, r + step)
    c:close_path()
    c:fill()
end

-- White ring outline
c:set_stroke_color(255, 255, 255)
c:set_line_width(10)
c:begin_path()
c:arc(w * 0.5, h * 0.5, math.min(w, h) * 0.4, 0, math.pi * 2)
c:stroke()

c:save_png("out.png")
print("saved out.png")
