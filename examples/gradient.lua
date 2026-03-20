local c = canvas.new(600, 400)

-- 1. Linear gradient background (sunset colors)
local bg = canvas.linear_gradient(0, 0, 0, 400)
bg:add_color_stop(0.0, 255, 100, 50)    -- warm orange
bg:add_color_stop(0.4, 255, 60, 80)     -- coral
bg:add_color_stop(0.7, 120, 40, 140)    -- purple
bg:add_color_stop(1.0, 20, 20, 60)      -- dark navy
c:set_fill_gradient(bg)
c:fill_rect(0, 0, 600, 400)

-- 2. Rectangle with horizontal linear gradient
local horiz = canvas.linear_gradient(50, 0, 250, 0)
horiz:add_color_stop(0.0, 0, 200, 255)
horiz:add_color_stop(1.0, 0, 50, 150)
c:set_fill_gradient(horiz)
c:fill_rect(50, 30, 200, 80)

-- 3. Rectangle with diagonal linear gradient
local diag = canvas.linear_gradient(300, 30, 550, 110)
diag:add_color_stop(0.0, 255, 255, 100)
diag:add_color_stop(0.5, 100, 255, 100)
diag:add_color_stop(1.0, 100, 100, 255)
c:set_fill_gradient(diag)
c:fill_rect(300, 30, 250, 80)

-- 4. Radial gradient (spotlight/orb)
local orb = canvas.radial_gradient(200, 240, 10, 200, 260, 100)
orb:add_color_stop(0.0, 255, 255, 200)  -- bright center
orb:add_color_stop(0.5, 255, 180, 50)   -- golden
orb:add_color_stop(1.0, 200, 50, 0, 0)  -- transparent edge
c:set_fill_gradient(orb)
c:begin_path()
c:arc(200, 260, 100, 0, math.pi * 2)
c:fill()

-- 5. Another radial gradient (cool orb)
local cool = canvas.radial_gradient(450, 240, 5, 450, 260, 80)
cool:add_color_stop(0.0, 200, 255, 255)
cool:add_color_stop(0.6, 50, 100, 255)
cool:add_color_stop(1.0, 10, 10, 80, 0)
c:set_fill_gradient(cool)
c:begin_path()
c:arc(450, 260, 80, 0, math.pi * 2)
c:fill()

-- 6. Gradient with transform (rotated rectangle)
local rot = canvas.linear_gradient(0, 0, 80, 0)
rot:add_color_stop(0.0, 255, 50, 50)
rot:add_color_stop(1.0, 50, 50, 255)
c:save()
c:translate(300, 280)
c:rotate(math.pi / 6)
c:set_fill_gradient(rot)
c:fill_rect(-40, -30, 80, 60)
c:restore()

-- 7. Stroke with gradient
local sg = canvas.linear_gradient(30, 370, 570, 370)
sg:add_color_stop(0.0, 255, 100, 100)
sg:add_color_stop(0.5, 100, 255, 100)
sg:add_color_stop(1.0, 100, 100, 255)
c:set_stroke_gradient(sg)
c:set_line_width(4)
c:stroke_rect(30, 350, 540, 40)

c:save_png("out.png")
print("saved out.png")
