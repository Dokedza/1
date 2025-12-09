package pdf

import (
	"1/storage"
	"bytes"
	"fmt"
	"time"

	"github.com/jung-kurt/gofpdf"
)

func GenerateReport(sets []storage.LinkSet) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Linc status report")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 10, fmt.Sprintf("Generated: %s", time.Now().Format("2006-01-02 15:04:05")))
	pdf.Ln(20)

	for _, set := range sets {
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(40, 10, fmt.Sprintf("Set #%d (Created: %s)", set.ID, set.CreatedAt.Format("2006-01-02 15:04:05")))
		pdf.Ln(8)

		pdf.SetFont("Arial", "B", 10)
		pdf.CellFormat(120, 8, "URL", "1", 0, "", false, 0, "")
		pdf.CellFormat(50, 8, "Status", "1", 0, "", false, 0, "")
		pdf.Ln(8)

		pdf.SetFont("Arial", "", 10)
		for _, link := range set.Links {
			pdf.CellFormat(120, 8, link.URL, "1", 0, "", false, 0, "")

			if link.Status == storage.StatusAvailable {
				pdf.SetTextColor(0, 128, 0)
			} else if link.Status == storage.StatusUnavailable {
				pdf.SetTextColor(255, 0, 0)
			} else {
				pdf.SetTextColor(128, 128, 128)
			}
			pdf.CellFormat(50, 8, string(link.Status), "1", 0, "", false, 0, "")
			pdf.SetTextColor(0, 0, 0)
			pdf.Ln(8)
		}
		pdf.Ln(10)
	}

	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
