package dashboard

import (
	"fmt"
	"github.com/guptarohit/asciigraph"
	tequilapi_client "github.com/mysteriumnetwork/node/tequilapi/client"
	"github.com/olekukonko/tablewriter"
	"math/rand"
	"os"
	"strings"
	"time"
)

func GetDashboard(api *tequilapi_client.Client) {
	var data []float64
	cont := 1
	updated := 150
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
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
		case <-ticker.C:
			if updated == 150 {
				updated = 0
				//pulisco la console
				CallClear()

				//creo il grafico
				cont++
				for i := 0; i < cont; i++ {
					data = append(data, float64(rand.Intn(100)))
				}
				graph := asciigraph.Plot(data, asciigraph.Height(10), asciigraph.Width(20), asciigraph.Offset(5))
				fmt.Println(graph)

				//creo la tabella
				table := tablewriter.NewWriter(os.Stdout)

				//ottengo le proposals
				proposals, err := api.Proposals()
				if err != nil {
					fmt.Println(err)
				}

				//costruisco la multi-slice che poi converto in [][]
				//[4]string{p.ID), p.ProviderID, p.ServiceType, p.ServiceDefinition.LocationOriginate.Country
				for _, p := range proposals {
					table.Append([]string{string(p.ID), p.ProviderID, p.ServiceType, p.ServiceDefinition.LocationOriginate.Country})
				}

				//creo la tabella
				tableString := &strings.Builder{}

				table.SetHeader([]string{"ID", "ProviderID", "ServiceType", "ServiceDefinition"})
				table.SetFooter([]string{"", "", "Total", "$146.93"}) // Add Footer
				table.SetBorder(false)                                // Set Border to false

				table.SetHeaderColor(tablewriter.Colors{tablewriter.Bold, tablewriter.BgGreenColor},
					tablewriter.Colors{tablewriter.FgHiRedColor, tablewriter.Bold, tablewriter.BgBlackColor},
					tablewriter.Colors{tablewriter.BgRedColor, tablewriter.FgWhiteColor},
					tablewriter.Colors{tablewriter.BgCyanColor, tablewriter.FgWhiteColor})

				table.SetColumnColor(tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiRedColor},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgHiBlackColor},
					tablewriter.Colors{tablewriter.Bold, tablewriter.FgBlackColor})

				table.SetFooterColor(tablewriter.Colors{}, tablewriter.Colors{},
					tablewriter.Colors{tablewriter.Bold},
					tablewriter.Colors{tablewriter.FgHiRedColor})

				//table.AppendBulk(data)
				table.SetColMinWidth(1, 100)
				table.Render()
				fmt.Println(tableString.String())
			}
			fmt.Print("=")
			updated++
		}
	}

}
