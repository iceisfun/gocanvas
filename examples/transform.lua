local c = canvas.new(800, 600)
local img = canvas.load_image("data/src.png")
local size = 120

-- Dark background
c:set_fill_color(40, 40, 50)
c:fill_rect(0, 0, 800, 600)

-- Labels (drawn as colored rects since we don't have fonts loaded)
-- 1) Original, no transform
c:set_stroke_color(255, 255, 255)
c:set_line_width(1)
c:stroke_rect(50, 50, size, size)
c:draw_image(img, 0, 0, img.width, img.height, 50, 50, size, size)

-- 2) Scaled 2x
c:save()
c:translate(300, 50)
c:scale(2, 2)
c:draw_image(img, 0, 0, img.width, img.height, 0, 0, size, size)
c:restore()

-- 3) Rotated 30 degrees
c:save()
c:translate(150, 380)
c:rotate(math.pi / 6)
c:draw_image(img, 0, 0, img.width, img.height, -size / 2, -size / 2, size, size)
c:restore()

-- 4) Rotated 90 degrees and scaled non-uniformly
c:save()
c:translate(400, 420)
c:rotate(math.pi / 2)
c:scale(1.5, 0.75)
c:draw_image(img, 0, 0, img.width, img.height, -size / 2, -size / 2, size, size)
c:restore()

-- 5) Fan of rotated copies
for i = 0, 11 do
    c:save()
    c:translate(650, 350)
    c:rotate(i * math.pi / 6)
    c:set_global_alpha(0.7)
    c:draw_image(img, 0, 0, img.width, img.height, 50, -20, 40, 40)
    c:restore()
end
c:set_global_alpha(1.0)

c:save_png("out.png")
print("saved out.png")
