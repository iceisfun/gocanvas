local c = canvas.new(600, 500)

-- Dark background
c:set_fill_color(30, 30, 30)
c:fill_rect(0, 0, 600, 500)

-- Load the source image
local img = canvas.load_image("data/src.png")
print(string.format("loaded %dx%d image", img.width, img.height))

-- Full image scaled to fit
c:draw_image(img, 0, 0, img.width, img.height, 10, 10, 180, 180)

-- Zoom into the owl's face (center crop)
c:draw_image(img, 250, 200, 500, 500, 200, 10, 180, 180)

-- Top-left quarter only
c:draw_image(img, 0, 0, img.width / 2, img.height / 2, 390, 10, 180, 180)

-- Bottom strip stretched wide
c:draw_image(img, 0, img.height * 0.75, img.width, img.height * 0.25, 10, 210, 570, 100)

-- Row of tiny thumbnails
for i = 0, 6 do
    c:draw_image(img, 0, 0, img.width, img.height,
        10 + i * 82, 330, 75, 75)
end

-- Draw with reduced alpha
c:set_global_alpha(0.4)
c:draw_image(img, 0, 0, img.width, img.height, 200, 400, 200, 80)
c:set_global_alpha(1.0)

c:save_png("out.png")
print("saved out.png")
