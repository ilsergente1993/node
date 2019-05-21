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
	"github.com/mum4k/termdash/widgets/textinput"
	tequilapi_client "github.com/mysteriumnetwork/node/tequilapi/client"
	"github.com/olekukonko/tablewriter"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

// widgets holds the widgets used by this demo.
type widgets struct {
	gauge     *gauge.Gauge
	speedLine *linechart.LineChart
	input     *textinput.TextInput
}

// rootID is the ID assigned to the root container.
const rootID = "root"

// redrawInterval is how often termdash redraws the screen.
const redrawInterval = 250 * time.Millisecond

const speedIGraphRange = 30 //seconds

var tableProposals *tablewriter.Table

type Stats struct {
	BSent             uint64
	BReceived         uint64
	PeakDownloadSpeed uint64
	PeakUploadSpeed   uint64
}

var globalStats = Stats{0, 0, 0, 0}

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

	/*
		tickerAPI := time.NewTicker(APIUPDATEFREQ * time.Millisecond)
		tickerConsole := time.NewTicker(WINUPDATEFREQ * time.Millisecond)
		//var proposals []tequilapi_client.ProposalDTO
		var statistics tequilapi_client.StatisticsDTO
		var status tequilapi_client.StatusDTO
		var err error
		defer tickerAPI.Stop()
		defer tickerConsole.Stop()
		cont := 0
		Init()

		done := make(chan bool)
		go func() {
			time.Sleep(10 * time.Second)
			done <- true
		}()

		for {
			select {
			case d := <-done:
				fmt.Println("Done?", d)
				return
			case <-tickerAPI.C:
				//ottengo le statistiche di connessione
				statistics, err = api.ConnectionStatistics()
				if err != nil {
					fmt.Println(err)
				}
				//random data perchè sono in noop connection
				statistics.BytesReceived = uint64(rand.Intn(1000000))
				statistics.BytesSent = uint64(rand.Intn(1000000))

				status, err = api.Status()
				if err != nil {
					fmt.Println(err)
				}

				//ottengo le proposals
				proposals, err = api.Proposals()
				if err != nil {
					fmt.Println(err)
				}
			case <-tickerConsole.C: //pulisco la console
				CallClear()
				//PRINT EVERYTHING
				fmt.Println(speedGraph(statistics))
				fmt.Println(Line())
				fmt.Println(ConnectionDetails(status))
				//fmt.Println(proposalsTable(proposals))
				cont++
				fmt.Print(cont)
			}
		}*/
}

func speedGraph(statistics tequilapi_client.StatisticsDTO) string {
	//fmt.Println(statistics)
	/*vel := (math.Abs(float64(statistics.BytesSent) - speedData[len(speedData)-1])) / APIUPDATEFREQ
	globalStats.BSent = globalStats.BSent + statistics.BytesSent
	globalStats.BReceived = globalStats.BReceived + statistics.BytesReceived
	if uint64(vel) > globalStats.PeakSpeed {
		globalStats.PeakSpeed = uint64(vel)
	}
	//TODO: mantenere un numero limitato di elementi nello slice altrimenti vado in overflow
	speedData = append(speedData, float64(statistics.BytesSent))
	header := "Speed graph [current: " + ToUnit(uint64(vel)) + "/sec]\n"
	footer := "\n" +
		"TOTAL SENT: " + ToUnit(globalStats.BSent) + " | " +
		"TOTAL RECEIVED " + ToUnit(globalStats.BReceived) + " |  " +
		"SPEED PEAK " + ToUnit(globalStats.PeakSpeed) + "/sec"
	return header + asciigraph.Plot(speedData, asciigraph.Height(10), asciigraph.Width(-5), asciigraph.Offset(5)) + footer */
	return ""
}

func ConnectionDetails(status tequilapi_client.StatusDTO) string {
	return "ID: \t\t" + strconv.Itoa(status.Proposal.ID) + "\n" +
		"Provider: \t" + string(status.Proposal.ProviderID) + "\n" +
		"Country: \t" + string(status.Proposal.ServiceDefinition.LocationOriginate.Country) + "\n" +
		"Service Type: \t" + string(status.Proposal.ServiceType)

}

func proposalsTable(proposals []tequilapi_client.ProposalDTO) string {
	tableString := &strings.Builder{}
	//appendo ogni proposal alla tabella
	tableProposals.ClearRows()
	for _, p := range proposals {
		tableProposals.Append([]string{string(p.ID), p.ProviderID, p.ServiceType, p.ServiceDefinition.LocationOriginate.Country})
	}
	tableProposals.Render()
	return tableString.String()
}

func Init() {
	tableProposals = tablewriter.NewWriter(os.Stdout)

	tableProposals.SetHeader([]string{"ID", "ProviderID", "ServiceType", "ServiceDefinition"})
	tableProposals.SetBorder(false)
	//tableProposals.SetColMinWidth(1, ) //imposto larghezza minima di una colonna

	tableProposals.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor})

	tableProposals.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor})

	/*tableProposals.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{},
	tablewriter.Colors{tablewriter.Bold},
	tablewriter.Colors{tablewriter.FgHiRedColor})*/

}

// creates the used widgets
func createWidgets(ctx context.Context, c *container.Container) (*widgets, error) {
	updateText := make(chan string)
	input, err := newTextInput(updateText)
	if err != nil {
		return nil, err
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
	}, nil
}

// build the layout
func gridLayout(w *widgets) ([]container.Option, error) {
	builder := grid.New()
	builder.Add(
		grid.RowHeightPerc(95,
			grid.ColWidthPerc(70,
				grid.RowHeightPerc(30,
					grid.Widget(w.speedLine,
						container.Border(linestyle.Round),
						container.BorderTitle("Press 'q' to quit"),
					),
				),
				grid.RowHeightPerc(70),
			),
			grid.ColWidthPerc(30,
				grid.RowHeightPerc(10,
					grid.Widget(w.gauge,
						container.Border(linestyle.Round),
						container.BorderTitle("Your wallet"),
					), ),
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

/* funtions to create the widgets */

// newHeartbeat returns a line chart that displays a heartbeat-like progression.
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
	go periodic(ctx, updateInterval, func() error {
		statistics, err = api.ConnectionStatistics()
		if err != nil {
			fmt.Println(err)
		}
		//TODO: remove, random data because I test with a noop connection
		statistics.BytesReceived = uint64(rand.Intn(1000000))
		statistics.BytesSent = uint64(rand.Intn(1000000))

		// saving speed calculation and global statistics
		speed := (math.Abs(float64(statistics.BytesSent) - downloadSpeed[len(downloadSpeed)-1])) / float64(updateInterval)
		globalStats = Stats{
			globalStats.BSent + statistics.BytesSent,
			globalStats.BReceived + statistics.BytesReceived,
			globalStats.PeakDownloadSpeed,
			globalStats.PeakUploadSpeed,
		}
		if uint64(speed) > globalStats.PeakDownloadSpeed {
			globalStats.PeakDownloadSpeed = uint64(speed)
		}
		if uint64(speed) > globalStats.PeakUploadSpeed {
			globalStats.PeakUploadSpeed = uint64(speed)
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
func newGauge(ctx context.Context) (*gauge.Gauge, error) {
	g, err := gauge.New()
	if err != nil {
		return nil, err
	}

	const start = 22
	progress := start

	go periodic(ctx, 2*time.Second, func() error {
		//TODO: getting here the data
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

// rotateFloats returns a new slice with inputs rotated by step.
// I.e. for a step of one:
//   inputs[0] -> inputs[len(inputs)-1]
//   inputs[1] -> inputs[0]
// And so on.
func rotateFloats(inputs []float64, step int) []float64 {
	return append(inputs[step:], inputs[:step]...)
}
