package dashboard

import (
	"fmt"
	"github.com/guptarohit/asciigraph"
	tequilapi_client "github.com/mysteriumnetwork/node/tequilapi/client"
	"github.com/olekukonko/tablewriter"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

const WINWIDTH = 100       //number of chars
const APIUPDATEFREQ = 500  //milliseconds
const WINUPDATEFREQ = 1000 //milliseconds (must be WINUPDATEFREQ > APIUPDATEFREQ)

var speedData []float64
var table *tablewriter.Table

type Stats struct {
	BSent     uint64
	BReceived uint64
	PeakSpeed uint64
}

var globalStats = Stats{0, 0, 0}

func GetDashboard(api *tequilapi_client.Client) {
	tickerAPI := time.NewTicker(APIUPDATEFREQ * time.Millisecond)
	tickerConsole := time.NewTicker(WINUPDATEFREQ * time.Millisecond)
	//var proposals []tequilapi_client.ProposalDTO
	var statistics tequilapi_client.StatisticsDTO
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

			//ottengo le proposals
			/*proposals, err = api.Proposals()
			if err != nil {
				fmt.Println(err)
			}*/
		case <-tickerConsole.C: //pulisco la console
			CallClear()
			//PRINT EVERYTHING
			fmt.Println(speedGraph(statistics))
			fmt.Println()
			//fmt.Println(proposalsTable(proposals))
			cont++
			fmt.Print(cont)
		}
	}
}

func speedGraph(statistics tequilapi_client.StatisticsDTO) string {
	//fmt.Println(statistics)
	vel := (math.Abs(float64(statistics.BytesSent) - speedData[len(speedData)-1])) / APIUPDATEFREQ
	globalStats.BSent = globalStats.BSent + statistics.BytesSent
	globalStats.BReceived = globalStats.BReceived + statistics.BytesReceived
	if uint64(vel) > globalStats.PeakSpeed {
		globalStats.PeakSpeed = uint64(vel)
	}
	//TODO: mantenere un numero limitato di elementi nello slice altrimenti vado in overflow
	speedData = append(speedData, float64(statistics.BytesSent))
	header := "Speed graph [current: " + strconv.FormatFloat(vel, 'f', 2, 64) + "KB/sec]\n"
	footer := "\n" +
		"TOTAL SENT: " + strconv.FormatFloat(float64(globalStats.BSent/100), 'f', 2, 640) + " KB  |  " +
		"TOTAL RECEIVED " + strconv.FormatFloat(float64(globalStats.BReceived/1000), 'f', 2, 64) + " KB  |  " +
		"SPEED PEAK " + strconv.FormatFloat(float64(globalStats.PeakSpeed), 'f', 2, 64) + " KB/sec"
	return header + asciigraph.Plot(speedData, asciigraph.Height(10), asciigraph.Width(WINWIDTH-5), asciigraph.Offset(5)) + footer
}

func proposalsTable(proposals []tequilapi_client.ProposalDTO) string {
	tableString := &strings.Builder{}
	//appendo ogni proposal alla tabella
	table.ClearRows()
	for _, p := range proposals {
		table.Append([]string{string(p.ID), p.ProviderID, p.ServiceType, p.ServiceDefinition.LocationOriginate.Country})
	}
	table.Render()
	return tableString.String()
}

func Init() {
	speedData = append(speedData, 1)

	table = tablewriter.NewWriter(os.Stdout)

	table.SetHeader([]string{"ID", "ProviderID", "ServiceType", "ServiceDefinition"})
	table.SetBorder(false)
	//table.SetColMinWidth(1, WINWIDTH) //imposto larghezza minima di una colonna

	table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
		tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold, tablewriter.BgBlackColor},
		tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
		tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor})

	table.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
		tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor})

	/*table.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{},
	tablewriter.Colors{tablewriter.Bold},
	tablewriter.Colors{tablewriter.FgHiRedColor})*/

}
