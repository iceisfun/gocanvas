local font = canvas.load_font("/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf")
local c = canvas.new(720, 560)

-- White background
c:set_fill_color(255, 255, 255)
c:fill_rect(0, 0, 720, 560)

-- Title
c:set_font(font, 22)
c:set_fill_color(0, 0, 0)
c:set_text_align("center")
c:set_text_baseline("top")
c:fill_text("Text Alignment Demo", 360, 10)

-- Grid of all 3 aligns x 4 baselines
local aligns = {"left", "center", "right"}
local baselines = {"alphabetic", "top", "middle", "bottom"}

local col_width = 220
local row_height = 120
local x_start = 130
local y_start = 80

c:set_font(font, 18)

for col, align in ipairs(aligns) do
    for row, baseline in ipairs(baselines) do
        local x = x_start + (col - 1) * col_width
        local y = y_start + (row - 1) * row_height

        -- Draw crosshairs at anchor point
        c:set_stroke_color(200, 80, 80)
        c:set_line_width(1)
        c:begin_path()
        c:move_to(x - 20, y)
        c:line_to(x + 20, y)
        c:stroke()
        c:begin_path()
        c:move_to(x, y - 20)
        c:line_to(x, y + 20)
        c:stroke()

        -- Small red dot at anchor
        c:set_fill_color(200, 50, 50)
        c:begin_path()
        c:arc(x, y, 3, 0, math.pi * 2)
        c:fill()

        -- Draw the text with alignment
        c:set_text_align(align)
        c:set_text_baseline(baseline)
        c:set_fill_color(0, 0, 0)
        c:set_font(font, 18)
        c:fill_text("Align", x, y)

        -- Label the combination below
        c:set_text_align("center")
        c:set_text_baseline("top")
        c:set_fill_color(100, 100, 100)
        c:set_font(font, 10)
        c:fill_text(align .. " / " .. baseline, x, y + 30)
    end
end

-- Reset alignment for footer
c:set_text_align("center")
c:set_text_baseline("top")
c:set_fill_color(80, 80, 80)
c:set_font(font, 12)
c:fill_text("Red crosshairs show the anchor point (x, y) passed to fill_text", 360, 530)

c:save_png("out.png")
print("saved out.png")
