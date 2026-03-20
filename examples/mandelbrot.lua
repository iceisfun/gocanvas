-- Mandelbrot set rendered with set_pixel.

local W, H = 800, 600
local c = canvas.new(W, H)

-- Viewport in the complex plane.
local cx, cy = -0.5, 0.0
local zoom = 1.2
local aspect = W / H
local x_min = cx - zoom * aspect
local x_max = cx + zoom * aspect
local y_min = cy - zoom
local y_max = cy + zoom

local max_iter = 256

-- Color palette: smooth HSV-style cycling.
local function hsv(h, s, v)
    local i = math.floor(h * 6) % 6
    local f = h * 6 - math.floor(h * 6)
    local p = v * (1 - s)
    local q = v * (1 - f * s)
    local t = v * (1 - (1 - f) * s)
    if i == 0 then return v, t, p
    elseif i == 1 then return q, v, p
    elseif i == 2 then return p, v, t
    elseif i == 3 then return p, q, v
    elseif i == 4 then return t, p, v
    else return v, p, q
    end
end

for py = 0, H - 1 do
    local ci = y_min + (y_max - y_min) * py / H
    for px = 0, W - 1 do
        local cr = x_min + (x_max - x_min) * px / W
        local zr, zi = 0.0, 0.0
        local iter = 0

        while zr * zr + zi * zi <= 4.0 and iter < max_iter do
            local tr = zr * zr - zi * zi + cr
            zi = 2.0 * zr * zi + ci
            zr = tr
            iter = iter + 1
        end

        if iter == max_iter then
            c:set_pixel(px, py, 0, 0, 0)
        else
            -- Smooth coloring.
            local log2 = math.log(2)
            local mu = iter + 1 - math.log(math.log(zr * zr + zi * zi) / log2) / log2
            local hue = (mu / max_iter * 4) % 1.0
            local r, g, b = hsv(hue, 0.9, 1.0)
            c:set_pixel(px, py,
                math.floor(r * 255),
                math.floor(g * 255),
                math.floor(b * 255))
        end
    end
end

c:save_png("out.png")
