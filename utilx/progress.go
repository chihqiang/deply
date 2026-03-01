package utilx

import (
	"fmt"
	"github.com/schollz/progressbar/v3"
	"os"
)

func NewProgress(totalSize int64, description string) *progressbar.ProgressBar {
	return progressbar.NewOptions64(
		totalSize,
		progressbar.OptionSetDescription(description),
		progressbar.OptionSetWidth(40),
		progressbar.OptionShowBytes(true),
		progressbar.OptionOnCompletion(func() {
			_, _ = fmt.Fprint(os.Stderr, "\n")
		}),
	)
}
