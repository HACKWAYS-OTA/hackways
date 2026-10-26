package programs

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var errWorkbookLimit = errors.New("workbook exceeds the bounded export size")

type workbookBuffer struct{ bytes.Buffer }

func (b *workbookBuffer) Write(p []byte) (int, error) {
	if b.Len()+len(p) > 8<<20 {
		return 0, errWorkbookLimit
	}
	return b.Buffer.Write(p)
}

type applicantWorkbook struct {
	buffer  *workbookBuffer
	archive *zip.Writer
	sheet   io.Writer
	row     int
	err     error
}

func newApplicantWorkbook(title, description string) (*applicantWorkbook, error) {
	buffer := &workbookBuffer{}
	archive := zip.NewWriter(buffer)
	archive.RegisterCompressor(zip.Deflate, func(w io.Writer) (io.WriteCloser, error) { return flate.NewWriter(w, flate.BestSpeed) })
	parts := map[string]string{
		"[Content_Types].xml":        `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types"><Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/xl/workbook.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.sheet.main+xml"/><Override PartName="/xl/worksheets/sheet1.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.worksheet+xml"/><Override PartName="/xl/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.spreadsheetml.styles+xml"/></Types>`,
		"_rels/.rels":                `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="xl/workbook.xml"/></Relationships>`,
		"xl/workbook.xml":            `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships"><bookViews><workbookView/></bookViews><sheets><sheet name="Applicants" sheetId="1" r:id="rId1"/></sheets></workbook>`,
		"xl/_rels/workbook.xml.rels": `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/worksheet" Target="worksheets/sheet1.xml"/><Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/></Relationships>`,
		"xl/styles.xml":              `<styleSheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><numFmts count="1"><numFmt numFmtId="164" formatCode="yyyy-mm-dd hh:mm"/></numFmts><fonts count="4"><font><sz val="11"/><name val="Arial"/><color rgb="FF12284A"/></font><font><b/><sz val="16"/><name val="Arial"/><color rgb="FFFFFFFF"/></font><font><b/><sz val="11"/><name val="Arial"/><color rgb="FFFFFFFF"/></font><font><sz val="10"/><name val="Arial"/><color rgb="FF526582"/></font></fonts><fills count="4"><fill><patternFill patternType="none"/></fill><fill><patternFill patternType="gray125"/></fill><fill><patternFill patternType="solid"><fgColor rgb="FF163675"/><bgColor indexed="64"/></patternFill></fill><fill><patternFill patternType="solid"><fgColor rgb="FFF2F6FC"/><bgColor indexed="64"/></patternFill></fill></fills><borders count="1"><border><left/><right/><top/><bottom/><diagonal/></border></borders><cellStyleXfs count="1"><xf numFmtId="0" fontId="0" fillId="0" borderId="0"/></cellStyleXfs><cellXfs count="8"><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0"/><xf numFmtId="0" fontId="1" fillId="2" borderId="0" xfId="0" applyAlignment="1"><alignment vertical="center"/></xf><xf numFmtId="0" fontId="2" fillId="2" borderId="0" xfId="0" applyAlignment="1"><alignment wrapText="1" vertical="center"/></xf><xf numFmtId="0" fontId="0" fillId="0" borderId="0" xfId="0" applyAlignment="1"><alignment wrapText="1" vertical="top"/></xf><xf numFmtId="0" fontId="0" fillId="3" borderId="0" xfId="0" applyAlignment="1"><alignment wrapText="1" vertical="top"/></xf><xf numFmtId="0" fontId="3" fillId="0" borderId="0" xfId="0"/><xf numFmtId="164" fontId="0" fillId="0" borderId="0" xfId="0" applyNumberFormat="1" applyAlignment="1"><alignment vertical="top"/></xf><xf numFmtId="164" fontId="0" fillId="3" borderId="0" xfId="0" applyNumberFormat="1" applyAlignment="1"><alignment vertical="top"/></xf></cellXfs><cellStyles count="1"><cellStyle name="Normal" xfId="0" builtinId="0"/></cellStyles></styleSheet>`,
	}
	for _, name := range []string{"[Content_Types].xml", "_rels/.rels", "xl/workbook.xml", "xl/_rels/workbook.xml.rels", "xl/styles.xml"} {
		writer, err := archive.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err = io.WriteString(writer, parts[name]); err != nil {
			return nil, err
		}
	}
	sheet, err := archive.Create("xl/worksheets/sheet1.xml")
	if err != nil {
		return nil, err
	}
	w := &applicantWorkbook{buffer: buffer, archive: archive, sheet: sheet}
	w.write(`<worksheet xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"><sheetViews><sheetView showGridLines="0" workbookViewId="0"><pane xSplit="1" ySplit="4" topLeftCell="B5" activePane="bottomRight" state="frozen"/><selection pane="bottomRight" activeCell="B5" sqref="B5"/></sheetView></sheetViews><sheetFormatPr defaultRowHeight="20"/><cols>`)
	for i, width := range []int{28, 38, 22, 25, 45, 24, 20, 22, 55, 70} {
		w.write(fmt.Sprintf(`<col min="%d" max="%d" width="%d" customWidth="1"/>`, i+1, i+1, width))
	}
	w.write(`</cols><sheetData>`)
	w.addRow([]string{title}, 1, 34)
	w.addRow([]string{description}, 5, 24)
	w.addRow(nil, 0, 10)
	w.addRow([]string{"Name", "Email", "WhatsApp", "City / community", "LinkedIn / profile", "Submitted (IST)", "Review decision", "Released result", "Introduction", "Event proposal"}, 2, 32)
	return w, w.err
}

func (w *applicantWorkbook) write(text string) {
	if w.err == nil {
		_, w.err = io.WriteString(w.sheet, text)
	}
}
func (w *applicantWorkbook) addRow(values []string, style, height int) {
	w.addDatedRow(values, style, height, time.Time{})
}
func (w *applicantWorkbook) addDatedRow(values []string, style, height int, submitted time.Time) {
	w.row++
	w.write(fmt.Sprintf(`<row r="%d" ht="%d" customHeight="1">`, w.row, height))
	for column, value := range values {
		if column == 5 && !submitted.IsZero() {
			local := submitted.In(time.FixedZone("IST", 19800))
			wall := time.Date(local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second(), 0, time.UTC)
			serial := wall.Sub(time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)).Hours() / 24
			w.write(fmt.Sprintf(`<c r="F%d" s="%d"><v>%.10f</v></c>`, w.row, style+3, serial))
			continue
		}
		// Explicit inline strings keep phone numbers intact and never execute formulas.
		w.write(fmt.Sprintf(`<c r="%c%d" t="inlineStr" s="%d"><is><t xml:space="preserve">`, 'A'+column, w.row, style))
		if w.err == nil {
			w.err = xml.EscapeText(w.sheet, []byte(strings.Map(func(r rune) rune {
				if r < 32 && r != '\n' && r != '\r' && r != '\t' {
					return '\uFFFD'
				}
				return r
			}, value)))
		}
		w.write(`</t></is></c>`)
	}
	w.write(`</row>`)
}
func (w *applicantWorkbook) finish() ([]byte, error) {
	w.write(fmt.Sprintf(`</sheetData><autoFilter ref="A4:J%d"/><mergeCells count="2"><mergeCell ref="A1:J1"/><mergeCell ref="A2:J2"/></mergeCells><pageMargins left="0.3" right="0.3" top="0.5" bottom="0.5" header="0.2" footer="0.2"/><pageSetup orientation="landscape" fitToWidth="1" fitToHeight="0"/></worksheet>`, w.row))
	if w.err != nil {
		return nil, w.err
	}
	if err := w.archive.Close(); err != nil {
		return nil, err
	}
	return w.buffer.Bytes(), nil
}
