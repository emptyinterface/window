package pricing

import (
	"fmt"
	"testing"
	"time"
)

func TestLoadTable(t *testing.T) {

	start := time.Now()
	table, err := LoadTable()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("row ct", len(table.Rows), time.Since(start))

	for _, row := range table.Rows {
		if row.OfferCode == AmazonVPCOfferCode &&
			row.Region == USEast1Region {
			fmt.Printf("%++v\n", row)
		}
	}

}
