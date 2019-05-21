package dashboard

import (
	"context"
	"fmt"
	"github.com/mum4k/termdash"
	"github.com/mum4k/termdash/cell"
	"github.com/mum4k/termdash/container"
	"github.com/mum4k/termdash/container/grid"
	"github.com/mum4k/termdash/keyboard"
	"github.com/mum4k/termdash/linestyle"
	"github.com/mum4k/termdash/terminal/termbox"
	"github.com/mum4k/termdash/terminal/terminalapi"
	"github.com/mum4k/termdash/widgets/gauge"
	"github.com/mum4k/termdash/widgets/linechart"
	"github.com/mum4k/termdash/widgets/text"
	"github.com/mum4k/termdash/widgets/textinput"
	"github.com/mysteriumnetwork/node/datasize"
	tequilapi_client "github.com/mysteriumnetwork/node/tequilapi/client"
	"math"
	"math/rand"
	"strconv"
	"time"
)

// widgets holds the widgets used by this demo.
type widgets struct {
	gauge     *gauge.Gauge
	speedLine *linechart.LineChart
	input     *textinput.TextInput
	speedText *text.Text
}

// rootID is the ID assigned to the root container.
const rootID = "root"

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

const speedIGraphRange = 30 //seconds

type Stats struct {
	BSent             uint64
	BReceived         uint64
	PeakDownloadSpeed uint64
	PeakUploadSpeed   uint64
}

var globalStats = Stats{0, 0, 0, 0} //in bytes

var api *tequilapi_client.Client

func GetDashboard(_api *tequilapi_client.Client) {
	api = _api
	t, err := termbox.New(termbox.ColorMode(terminalapi.ColorMode256))
	if err != nil {
		panic(err)
	}
	defer t.Close()

	// creating the container
	c, err := container.New(t, container.ID(rootID))
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	// creating the widgets
	widgets, err := createWidgets(ctx, c)
	if err != nil {
		panic(err)
	}
	// building the layout
	gridOpts, err := gridLayout(widgets)
	if err != nil {
		panic(err)
	}

	// updating the layout with the
	if err := c.Update(rootID, gridOpts...); err != nil {
		panic(err)
	}

	// defying the function to exit the dashboard
	quitter := func(k *terminalapi.Keyboard) {
		if k.Key == keyboard.Key('q') || k.Key == keyboard.KeyEsc {
			cancel()
		}
	}
	// running the dashboard
	err = termdash.Run(ctx, t, c, termdash.KeyboardSubscriber(quitter), termdash.RedrawInterval(redrawInterval))
	if err != nil {
		panic(err)
	}

}

// creates the used widgets
func createWidgets(ctx context.Context, c *container.Container) (*widgets, error) {
	updateText := make(chan string)
	input, err := newTextInput(updateText)
	if err != nil {
		return nil, err
	}
	speedText, err := newSpeedText(ctx)
	if err != nil {
		panic(err)
	}
	g, err := newGauge(ctx)
	if err != nil {
		return nil, err
	}
	speedLine, err := speedLine(ctx)
	if err != nil {
		return nil, err
	}

	return &widgets{
		gauge:     g,
		speedLine: speedLine,
		input:     input,
		speedText: speedText,
	}, nil
}

// build the layout
func gridLayout(w *widgets) ([]container.Option, error) {
	padding := 5
	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(95,
			grid.ColWidthPerc(70,
				grid.RowHeightPerc(30,
					grid.Widget(w.speedLine,
						container.Border(linestyle.Light),
						container.BorderTitle("Press 'q' to quit"),
						container.PaddingLeftPercent(padding), container.PaddingTopPercent(padding), container.PaddingRightPercent(padding), container.PaddingBottomPercent(padding),
					),
				),
				grid.RowHeightPerc(15,
					grid.Widget(w.speedText,
						container.Border(linestyle.Light),
						container.PaddingLeftPercent(padding), container.PaddingTopPercent(padding), container.PaddingRightPercent(padding), container.PaddingBottomPercent(padding)),
				),
				grid.RowHeightPerc(50),
			),
			grid.ColWidthPerc(30,
				grid.RowHeightPerc(10,
					grid.Widget(w.gauge,
						container.Border(linestyle.Light),
						container.BorderTitle("Your wallet"),
						container.PaddingLeftPercent(padding), container.PaddingTopPercent(padding), container.PaddingRightPercent(padding), container.PaddingBottomPercent(padding),
					)),
				grid.RowHeightPerc(90),
			),
		),
		grid.RowHeightPerc(5,
			grid.Widget(w.input),
		),
	)
	gridOpts, err := builder.Build()
	if err != nil {
		return nil, err
	}
	return gridOpts, nil
}

/* functions to create the widgets */

// speedLine returns a line chart that displays a heartbeat-like progression.
func speedLine(ctx context.Context) (*linechart.LineChart, error) {
	updateInterval := redrawInterval * 2

	// getting data
	var statistics tequilapi_client.StatisticsDTO
	var err error
	var downloadSpeed = []float64{0}
	var uploadSpeed = []float64{0}
	var nOfElements = int(speedIGraphRange / updateInterval.Seconds()) // keeping last <speedIGraphRange> seconds of data (#ofElement = updateInterval/speedIGraphRange)

	lc, err := linechart.New(
		linechart.AxesCellOpts(cell.FgColor(cell.ColorRed)),
		linechart.YLabelCellOpts(cell.FgColor(cell.ColorGreen)),
		linechart.XLabelCellOpts(cell.FgColor(cell.ColorGreen)),
	)
	if err != nil {
		return nil, err
	}

	var Xlabels = make(map[int]string)
	for i := nOfElements; i > 0; i-- {
		Xlabels[i] = " "
	}
	Xlabels[nOfElements] = "now"
	Xlabels[nOfElements-1] = "now"
	Xlabels[nOfElements+1] = "now"
	go periodic(ctx, updateInterval, func() error {
		statistics, err = api.ConnectionStatistics()
		if err != nil {
			fmt.Println(err)
		}
		//TODO: remove, random data because I test with a noop connection
		statistics.BytesReceived = uint64(rand.Intn(500000))
		statistics.BytesSent = uint64(rand.Intn(500000))

		// calculating and saving global statistics
		globalStats = Stats{
			globalStats.BSent + statistics.BytesSent,
			globalStats.BReceived + statistics.BytesReceived,
			globalStats.PeakDownloadSpeed,
			globalStats.PeakUploadSpeed,
		}

		if speedDW := (math.Abs(float64(statistics.BytesReceived) - downloadSpeed[len(downloadSpeed)-1])) / float64(updateInterval.Seconds()); uint64(speedDW) > globalStats.PeakDownloadSpeed {
			globalStats.PeakDownloadSpeed = uint64(speedDW)
		}
		if speedUP := (math.Abs(float64(statistics.BytesSent) - downloadSpeed[len(downloadSpeed)-1])) / float64(updateInterval.Seconds()); uint64(speedUP) > globalStats.PeakUploadSpeed {
			globalStats.PeakUploadSpeed = uint64(speedUP)
		}
		downloadSpeed = append(downloadSpeed, float64(statistics.BytesReceived))
		uploadSpeed = append(uploadSpeed, float64(statistics.BytesSent))
		if len(downloadSpeed) > nOfElements {
			downloadSpeed = downloadSpeed[1:]
			uploadSpeed = uploadSpeed[1:]
		}
		Xlabels[0] = strconv.Itoa(int(float64(len(downloadSpeed))*updateInterval.Seconds())) + " sec"

		if err := lc.Series("upload", uploadSpeed,
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorRed)),
		); err != nil {
			return err
		}
		return lc.Series("download", downloadSpeed,
			linechart.SeriesCellOpts(cell.FgColor(cell.ColorGreen)),
			linechart.SeriesXLabels(Xlabels),
		)
	})
	return lc, nil
}

// newGauge creates a demo Gauge widget.
func newSpeedText(ctx context.Context) (*text.Text, error) {
	data, err := text.New(text.WrapAtRunes())
	if err != nil {
		panic(err)
	}

	go periodic(ctx, 2*time.Second, func() error {
		if err := data.Write("TOTAL SENT: ", text.WriteCellOpts(cell.FgColor(cell.ColorMagenta)), text.WriteReplace()); err != nil {
			panic(err)
		}
		if err := data.Write(datasize.BitSize(globalStats.BSent * 8).String()); err != nil {
			panic(err)
		}
		if err := data.Write("   TOTAL RECEIVED: ", text.WriteCellOpts(cell.FgColor(cell.ColorMagenta))); err != nil {
			panic(err)
		}
		if err := data.Write(datasize.BitSize(globalStats.BReceived * 8).String()); err != nil {
			panic(err)
		}
		if err := data.Write("   DOWNLOAD SPEED PEAK: ", text.WriteCellOpts(cell.FgColor(cell.ColorMagenta))); err != nil {
			panic(err)
		}
		if err := data.Write(datasize.BitSize(globalStats.PeakDownloadSpeed*8).String() + "/sec"); err != nil {
			panic(err)
		}
		if err := data.Write("   UPLOAD SPEED PEAK: ", text.WriteCellOpts(cell.FgColor(cell.ColorMagenta))); err != nil {
			panic(err)
		}
		if err := data.Write(datasize.BitSize(globalStats.PeakUploadSpeed*8).String() + "/sec"); err != nil {
			panic(err)
		}

		return nil
	})
	return data, nil
}

// newGauge creates a demo Gauge widget.
func newGauge(ctx context.Context) (*gauge.Gauge, error) {
	g, err := gauge.New()
	if err != nil {
		return nil, err
	}

	const start = 22
	progress := start

	go periodic(ctx, 2*time.Second, func() error {
		//TODO: getting here the wallet consumption data
		var c gauge.Option
		if progress < 20 {
			c = gauge.Color(cell.ColorRed)
		} else {
			c = gauge.Color(cell.ColorGreen)
		}
		if err := g.Percent(progress, c); err != nil {
			return err
		}
		progress--
		if progress < 0 {
			progress = start
		}
		return nil
	})
	return g, nil
}

// newTextInput creates a new TextInput field
func newTextInput(updateText chan<- string) (*textinput.TextInput, error) {
	input, err := textinput.New(
		textinput.Label("Change text to: ", cell.FgColor(cell.ColorBlue)),
		textinput.MaxWidthCells(20),
		textinput.PlaceHolder("enter any text"),
		textinput.OnSubmit(func(text string) error {
			updateText <- text
			return nil
		}),
		textinput.ClearOnSubmit(),
	)
	if err != nil {
		return nil, err
	}
	return input, err
}

/* helpers */

// periodic executes the provided closure periodically every interval.
// Exits when the context expires.
func periodic(ctx context.Context, interval time.Duration, fn func() error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := fn(); err != nil {
				panic(err)
			}
		case <-ctx.Done():
			return
		}
	}
}
