// 支持超长行，会对其进行拆行处理
//
// @author      Liu Yongshuai<liuyongshuai@hotmail.com>
// @date        2018-11-27 19:02

package goUtils

import (
	"bytes"
	"regexp"
	"sort"
	"strings"
)

type TerminalTable struct {
	rawHeaderData       []string            //原始的表头数据
	headerFontColorFunc ColorFunc           //表头的字体颜色，默认Yellow
	rawRowData          [][]string          //原始的行的数据
	rowFontColorFunc    ColorFunc           //表格内容的字体颜色
	borderColorFunc     ColorFunc           //边框的颜色
	isUseSeparator      bool                //是否需要每行间的分隔线
	maxColumnNum        int                 //列的数量，以最多的一行的列为准
	maxColumnWidth      []int               //每列的最大宽度，对齐用的
	rowData             []*terminalTableRow //所有行
}

func NewTerminalTable() *TerminalTable {
	t := &TerminalTable{
		isUseSeparator:      true,
		headerFontColorFunc: Yellow,
		rowFontColorFunc:    nil,
		borderColorFunc:     nil,
	}
	return t
}

//是否使用行的分隔符
func (t *TerminalTable) IsUseRowSeparator(b bool) *TerminalTable {
	t.isUseSeparator = b
	return t
}

//添加表头数据
func (t *TerminalTable) SetHeader(header []string) *TerminalTable {
	if len(header) > t.maxColumnNum {
		t.maxColumnNum = len(header)
	}
	for _, h := range header {
		t.rawHeaderData = append(t.rawHeaderData, h)
	}
	return t
}

//添加表头字体颜色
func (t *TerminalTable) SetHeaderFontColor(color ColorType) *TerminalTable {
	colorFunc, ok := GetColorFunc(color)
	if ok {
		t.headerFontColorFunc = colorFunc
	}
	return t
}

//添加一行数据
func (t *TerminalTable) AddRow(row []string) *TerminalTable {
	t.rawRowData = append(t.rawRowData, row)
	if len(row) > t.maxColumnNum {
		t.maxColumnNum = len(row)
	}
	return t
}

//添加许多行数据
func (t *TerminalTable) AddRows(rows [][]string) *TerminalTable {
	for _, row := range rows {
		if len(row) > t.maxColumnNum {
			t.maxColumnNum = len(row)
		}
		t.rawRowData = append(t.rawRowData, row)
	}
	return t
}

//添加行内容字体颜色
func (t *TerminalTable) SetRowFontColor(color ColorType) *TerminalTable {
	colorFunc, ok := GetColorFunc(color)
	if ok {
		t.rowFontColorFunc = colorFunc
	}
	return t
}

//添加边框颜色
func (t *TerminalTable) SetBorderFontColor(color ColorType) *TerminalTable {
	colorFunc, ok := GetColorFunc(color)
	if ok {
		t.borderColorFunc = colorFunc
	}
	return t
}

//开始返回表格数据
func (t *TerminalTable) Render() string {
	headerLen := len(t.rawHeaderData)
	rowLen := len(t.rawRowData)
	if headerLen <= 0 && rowLen <= 0 {
		return ""
	}
	t.prepareSomething()

	//行分隔符，根据每列的最大列宽来决定
	sepBuf := bytes.Buffer{}
	sepBuf.WriteString(t.borderStr("+"))
	for _, w := range t.maxColumnWidth {
		sepBuf.WriteString(t.borderStr(strings.Repeat("-", w)))
		sepBuf.WriteString(t.borderStr("+"))
	}
	separatorLine := sepBuf.String()

	dataBuf := bytes.Buffer{}

	//第一个行分隔符，先写进去
	if t.isUseSeparator {
		dataBuf.WriteString(separatorLine)
		dataBuf.WriteString("\n")
	}

	for idx := range t.rowData {
		row := t.rowData[idx]
		rowStr := t.renderSingleRow(row)
		if rowStr == "" {
			continue
		}
		dataBuf.WriteString(rowStr)
		if t.isUseSeparator {
			dataBuf.WriteString(separatorLine)
			dataBuf.WriteString("\n")
		}
	}

	return dataBuf.String()
}

type rowType byte

const (
	rowTypeHeader rowType = iota
	rowTypeData
)

//一行数据，包含多个列
type terminalTableRow struct {
	lineNum     int     //本行各个小格子的数据行数，本行所有列的行数都是一样的
	columnWidth []int   //本行数据各列的宽度
	rowType     rowType //本行数据的类型，是表头还是数据内容
	cellList    []*terminalTableCell
}

//一个小格子里的数据
type terminalTableCell struct {
	columnNo      int      //第几列
	maxAllowWidth int      //允许的最大宽度
	cellStrList   []string //本小格子的数据，可能要分多行
}

//计算属性
func (t *TerminalTable) prepareSomething() {
	if len(t.rawHeaderData) <= 0 && len(t.rawRowData) <= 0 {
		return
	}
	t.maxColumnWidth = make([]int, t.maxColumnNum)
	headerLen := len(t.rawHeaderData)

	//把所有的行的列数补齐到一致，方便输出
	if headerLen > 0 {
		if headerLen < t.maxColumnNum {
			for i := headerLen; i < t.maxColumnNum; i++ {
				t.rawHeaderData = append(t.rawHeaderData, " ")
			}
		}
	}
	for idx, row := range t.rawRowData {
		if len(row) < t.maxColumnNum {
			for i := len(row); i < t.maxColumnNum; i++ {
				t.rawRowData[idx] = append(t.rawRowData[idx], " ")
			}
		}
	}

	//统一给各行折一下行
	if headerLen > 0 {
		tmp := wrapTableRows(t.rawHeaderData)
		if tmp != nil {
			tmp.rowType = rowTypeHeader
			t.rowData = append(t.rowData, tmp)
		}
	}
	for idx := range t.rawRowData {
		tmp := wrapTableRows(t.rawRowData[idx])
		if tmp != nil {
			tmp.rowType = rowTypeData
			t.rowData = append(t.rowData, tmp)
		}
	}
	//再统计各列的最大宽度，并给各列补齐到相同的宽度
	for idx := range t.rowData {
		row := t.rowData[idx]
		for i, w := range row.columnWidth {
			if t.maxColumnWidth[i] < w {
				t.maxColumnWidth[i] = w
			}
		}
	}

	//将每列的数据补整齐
	for rowIdx := range t.rowData {
		row := t.rowData[rowIdx]
		for cellIdx, cellUnit := range row.cellList {
			maxWidth := t.maxColumnWidth[cellIdx]
			for subCellIdx, subCellStr := range cellUnit.cellStrList {
				subCellStr = RuneFillRight(subCellStr, maxWidth)
				cellUnit.cellStrList[subCellIdx] = subCellStr
			}
		}
	}
}

//表头字体获取
func (t *TerminalTable) headerStr(str string) string {
	if t.headerFontColorFunc != nil {
		return t.headerFontColorFunc(str)
	}
	return Yellow(str)
}

//边框字体获取
func (t *TerminalTable) borderStr(str string) string {
	if t.borderColorFunc != nil {
		return t.borderColorFunc(str)
	}
	return str
}

//行内容字体获取
func (t *TerminalTable) rowStr(str string) string {
	if t.rowFontColorFunc != nil {
		return t.rowFontColorFunc(str)
	}
	return str
}

//生成一行数据的格式
func (t *TerminalTable) renderSingleRow(row *terminalTableRow) string {
	if row == nil || len(row.cellList) <= 0 {
		return ""
	}
	buf := bytes.Buffer{}
	//竖线分隔符
	verSepLine := t.borderStr("|")
	if !t.isUseSeparator {
		verSepLine = " "
	}
	//列的数量
	colNum := len(row.columnWidth)
	//小格子里的内容被拆分成了多少行，所有小格子的行都是一样的
	srowNum := len(row.cellList[0].cellStrList)
	for i := 0; i < srowNum; i++ {
		for j := 0; j < colNum; j++ {
			buf.WriteString(verSepLine)
			str := row.cellList[j].cellStrList[i]
			switch row.rowType {
			case rowTypeHeader:
				str = t.headerStr(str)
			case rowTypeData:
				str = t.rowStr(str)
			}
			buf.WriteString(str)
		}
		buf.WriteString(verSepLine)
		buf.WriteString("\n")
	}
	return buf.String()
}

//将一行数据折行，并返回最大行数，主要策略是每次都将最长的行折半，一直折到所有行的长度小于屏幕长度
func wrapTableRows(rawRow []string) (retRow *terminalTableRow) {
	if len(rawRow) <= 0 {
		return nil
	}
	retRow = &terminalTableRow{columnWidth: make([]int, len(rawRow))}
	//每一行中都有一些多余的字符，要将屏幕宽度减去这部分
	totalWidth := ScreenWidth - len(rawRow)*5
	reg, _ := regexp.Compile(`\n`)
	allWidth := 0

	cellNoMap := make(map[int]*terminalTableCell)
	var tmpCellList []*terminalTableCell

	//统计各小格子宽度
	for idx, row := range rawRow {
		//本小格子的最大宽度，要考虑小格子的数据中有换行符的情况
		maxw := 0
		tmp := reg.Split(row, -1)
		for _, t := range tmp {
			w := RuneStringWidth(t)
			if w > maxw {
				maxw = w
			}
		}
		allWidth += maxw
		cell := &terminalTableCell{
			columnNo:      idx,
			maxAllowWidth: maxw,
		}
		tmpCellList = append(tmpCellList, cell)
		cellNoMap[idx] = cell
	}

	//如果各小格子的宽度和大于屏幕宽度
	if allWidth > totalWidth {
		for {
			diff := allWidth - totalWidth
			sort.Slice(tmpCellList, func(i, j int) bool {
				return tmpCellList[i].maxAllowWidth > tmpCellList[j].maxAllowWidth
			})
			//本次总宽度消除量
			reduce := tmpCellList[0].maxAllowWidth / 2
			if reduce > diff {
				reduce = diff
			}
			tmpCellList[0].maxAllowWidth -= reduce
			allWidth -= reduce
			if allWidth <= totalWidth {
				break
			}
		}
	}

	//各小格子的最大行数
	maxLineNum := 0

	//开始对各小格子进行拆行处理
	for idx, cellStr := range rawRow {
		cellUnit := cellNoMap[idx]
		tmpCellStr := cellStr
		lineNum := 1
		if RuneStringWidth(cellStr) > cellUnit.maxAllowWidth {
			tmpCellStr, lineNum = RuneWrap(cellStr, cellUnit.maxAllowWidth)
		}
		if lineNum > maxLineNum {
			maxLineNum = lineNum
		}
		tmp := reg.Split(tmpCellStr, -1)
		for _, t := range tmp {
			cellUnit.cellStrList = append(cellUnit.cellStrList, " "+t+" ")
		}
	}

	//如果每行数据不够最大的行数，用空行补齐
	for idx := range rawRow {
		cellUnit := cellNoMap[idx]
		tmpLen := len(cellUnit.cellStrList)
		for i := 0; i < maxLineNum-tmpLen; i++ {
			cellUnit.cellStrList = append(cellUnit.cellStrList, " ")
		}
		maxCellWidth := 0
		for _, s := range cellUnit.cellStrList {
			l := RuneStringWidth(s)
			if l > maxCellWidth {
				maxCellWidth = l
			}
		}
		retRow.cellList = append(retRow.cellList, cellUnit)
		retRow.columnWidth[idx] = maxCellWidth
	}

	return retRow
}
