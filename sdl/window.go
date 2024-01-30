package sdl

import (
	"fmt"
	"unsafe"
	
	"github.com/veandco/go-sdl2/sdl"
	"uk.ac.bris.cs/gameoflife/util"
)

type Window struct {
	Width, Height int32
	window        *sdl.Window
	renderer      *sdl.Renderer
	texture       *sdl.Texture
	pixels        []byte
}

func filterEvent(e sdl.Event, userdata interface{}) bool {
	return e.GetType() == sdl.KEYDOWN || e.GetType() == sdl.QUIT
}

func NewWindow(width, height int32) *Window {
	err := sdl.Init(sdl.INIT_EVERYTHING)
	util.Check(err)
	window, err := sdl.CreateWindow("GOL GUI", sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, width, height, sdl.WINDOW_SHOWN)
	util.Check(err)
	renderer, err := sdl.CreateRenderer(window, -1, sdl.WINDOW_SHOWN)
	util.Check(err)
	sdl.SetHint(sdl.HINT_RENDER_SCALE_QUALITY, "linear")
	err = renderer.SetLogicalSize(width, height)
	util.Check(err)
	texture, err := renderer.CreateTexture(sdl.PIXELFORMAT_ARGB8888, sdl.TEXTUREACCESS_STATIC, width, height)
	util.Check(err)

	sdl.SetEventFilterFunc(filterEvent, nil)
	return &Window{
		width,
		height,
		window,
		renderer,
		texture,
		make([]byte, width*height*4),
	}
}

func (w *Window) Destroy() {
	err := w.texture.Destroy()
	util.Check(err)
	err = w.renderer.Destroy()
	util.Check(err)
	err = w.window.Destroy()
	util.Check(err)
	sdl.Quit()
}

func (w *Window) RenderFrame() {
	err := w.texture.Update(nil, unsafe.Pointer(&w.pixels[0]), int(w.Width*4))
	util.Check(err)
	err = w.renderer.Clear()
	util.Check(err)
	err = w.renderer.Copy(w.texture, nil, nil)
	util.Check(err)
	w.renderer.Present()
}

func (w *Window) PollEvent() sdl.Event {
	return sdl.PollEvent()
}

func (w *Window) SetPixel(x, y int) {
	width := int(w.Width)
	w.pixels[4*(y*width+x)+0] = 0xFF
	w.pixels[4*(y*width+x)+1] = 0xFF
	w.pixels[4*(y*width+x)+2] = 0xFF
	w.pixels[4*(y*width+x)+3] = 0xFF
}

func (w *Window) FlipPixel(x, y int) {
	if x < 0 || y < 0 || x >= int(w.Width) || y >= int(w.Height) {
		panic(fmt.Sprintf("CellFlipped event at (%d, %d) is outside the bounds of the window.", x, y))
	}

	width := int(w.Width)
	w.pixels[4*(y*width+x)+0] = ^w.pixels[4*(y*width+x)+0]
	w.pixels[4*(y*width+x)+1] = ^w.pixels[4*(y*width+x)+1]
	w.pixels[4*(y*width+x)+2] = ^w.pixels[4*(y*width+x)+2]
	w.pixels[4*(y*width+x)+3] = ^w.pixels[4*(y*width+x)+3]
}

func (w *Window) CountPixels() int {
	count := 0
	for i := 0; i < int(w.Width) * int(w.Height) * 4; i += 4 {
		if w.pixels[i] == 0xFF {
			count++
		}
	}
	return count
}

func (w *Window) ClearPixels() {
	for i := range w.pixels {
		w.pixels[i] = 0
	}
}
