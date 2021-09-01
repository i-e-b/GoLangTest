package main


var bmp []byte
var renderWidth, renderHeight uint64
var bmpSize uint64
var scanBytes uint64
var frameCount uint64

func GetBufferPointer() *byte{return &bmp[0] }

func RenderFrame(){
	frameCount++

	var y,x uint64
	var i uint64
	var v byte

	for y = 0; y < renderHeight; y++ {
		i = y * scanBytes
		for x = 0; x < renderWidth; x++ {
			v = byte( y + x + frameCount )
			bmp[i+0] = v    // B
			bmp[i+1] = v	// G
			bmp[i+2] = v	// R
			bmp[i+3] = 0	// A - mostly ignored

			i += 4
		}
	}
}

func SetupRender(width, height int) {

	renderWidth = uint64(width)
	renderHeight = uint64(height)

	scanBytes = uint64(width) * 4
	bmpSize = scanBytes * uint64(height) // 32 bit argb
	bmp = make([]byte, bmpSize)

	var i uint64
	for i = 0; i < bmpSize; i+=4 {
		bmp[i+0] = 0    // B
		bmp[i+1] = 0	// G
		bmp[i+2] = 0	// R
		bmp[i+3] = 0	// A - mostly ignored
	}
}
