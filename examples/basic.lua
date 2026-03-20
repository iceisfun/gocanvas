local c = canvas.new(400, 300)

-- Light gray background
c:set_fill_color(240, 240, 240)
c:fill_rect(0, 0, 400, 300)

-- Red filled rectangle
c:set_fill_color(220, 50, 50)
c:fill_rect(20, 20, 120, 80)

-- Blue stroked rectangle
c:set_stroke_color(50, 50, 220)
c:set_line_width(3)
c:stroke_rect(170, 20, 120, 80)

-- Green circle
c:set_fill_color(50, 180, 50)
c:begin_path()
c:arc(80, 200, 50, 0, math.pi * 2)
c:fill()

-- Orange triangle
c:set_fill_color(255, 165, 0)
c:begin_path()
c:move_to(250, 280)
c:line_to(350, 280)
c:line_to(300, 180)
c:close_path()
c:fill()

-- Semi-transparent purple overlay
c:set_fill_color(128, 0, 255, 100)
c:fill_rect(60, 60, 200, 150)

c:save_png("out.png")
print("saved out.png")
