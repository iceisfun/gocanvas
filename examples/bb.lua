local c = canvas.new(800, 600)
local img = canvas.load_image("data/src.png")
local font = canvas.load_font("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf")

-- Draw the source image as background, scaled to fill.
c:draw_image(img, 0, 0, img.width, img.height, 0, 0, 800, 600)

-- Detection style presets.
local green = {
    font = font, font_size = 14, line_width = 2, padding = 4,
    stroke_color = {0, 255, 0},
    fill_color = {0, 0, 0, 180},
    text_color = {255, 255, 255},
}
local cyan = {
    font = font, font_size = 13, line_width = 2, padding = 3,
    stroke_color = {0, 220, 255},
    fill_color = {0, 60, 80, 200},
    text_color = {200, 255, 255},
}
local yellow = {
    font = font, font_size = 12, line_width = 2, padding = 3,
    stroke_color = {255, 220, 0},
    fill_color = {60, 50, 0, 200},
    text_color = {255, 255, 200},
}
local red = {
    font = font, font_size = 13, line_width = 2, padding = 3,
    stroke_color = {255, 60, 60},
    fill_color = {80, 0, 0, 200},
    text_color = {255, 200, 200},
}

-- Simulated detections.
c:draw_labeled_box("barn owl  0.97",     150, 5,  480, 600, green)
c:draw_labeled_box("cowboy hat  0.89",   130, 10,  560, 180, cyan)
c:draw_labeled_box("eye  0.82",          245, 220, 70,  60,  yellow)
c:draw_labeled_box("eye  0.79",          400, 220, 65,  65,  yellow)
c:draw_labeled_box("beak  0.74",         300, 300, 90,  80,  red)
c:draw_labeled_box("wing  0.71",         480, 330, 270, 260, cyan)

c:save_png("out.png")
print("saved out.png")
