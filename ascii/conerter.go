package ascii

import (
	"github.com/qeesung/image2ascii/convert"
)

func ConvertImage(path string) string {
	converter := convert.NewImageConverter()
	options := convert.DefaultOptions
	options.Colored = false // enale color codes
	options.FixedWidth = 30 // Set width
	options.FixedHeight = 8 // Set height
	options.Reversed = false
	return "[#87CEFA]" + converter.ImageFile2ASCIIString(path, &options) + "[white]"
}
