package main
import (
    "bytes"
    "image"
    "image/png"
    "io/ioutil"
)

func main() {
    const height = 256
    const width = 256
    img := image.NewNRGBA(image.Rect(0, 0, width, height))
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            var red uint8 = uint8(255.0 * (x % 64) / 64.0)
            var green uint8 = uint8(255.0 * (y % 64) / 64.0)
            var blue uint8 = uint8(128)
            i := 4 * (y * width + x)
            img.Pix[i] = red
            img.Pix[i + 1] = green
            img.Pix[i + 2] = blue
            img.Pix[i + 3] = 255
        }
    }
    var buffer bytes.Buffer
    error := png.Encode(&buffer, img)
    if error != nil {
        panic(error)
    }
    
    ioutil.WriteFile("test.png", buffer.Bytes(), 0777)
}