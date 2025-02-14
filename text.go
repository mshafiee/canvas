package canvas

import (
	"image"
	"image/color"
	"math"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/tdewolff/canvas/font"
	"github.com/tdewolff/canvas/text"
	canvasText "github.com/tdewolff/canvas/text"
)

// TextAlign specifies how the text should align or whether it should be justified.
type TextAlign int

// see TextAlign
const (
	Left TextAlign = iota
	Right
	Center
	Middle
	Top
	Bottom
	Justify
)

func (ta TextAlign) String() string {
	switch ta {
	case Left:
		return "Left"
	case Right:
		return "Right"
	case Center:
		return "Center"
	case Middle:
		return "Middle"
	case Top:
		return "Top"
	case Bottom:
		return "Bottom"
	case Justify:
		return "Justify"
	}
	return "Invalid(" + strconv.Itoa(int(ta)) + ")"
}

// VerticalAlign specifies how the object should align vertically when embedded in text.
type VerticalAlign int

// see VerticalAlign
const (
	Baseline VerticalAlign = iota
	FontTop
	FontMiddle
	FontBottom
)

func (valign VerticalAlign) String() string {
	switch valign {
	case Baseline:
		return "Baseline"
	case FontTop:
		return "FontTop"
	case FontMiddle:
		return "FontMiddle"
	case FontBottom:
		return "FontBottom"
	}
	return "Invalid(" + strconv.Itoa(int(valign)) + ")"
}

// WritingMode specifies how the text lines should be laid out.
type WritingMode int

// see WritingMode
const (
	HorizontalTB WritingMode = iota
	VerticalRL
	VerticalLR
)

func (wm WritingMode) String() string {
	switch wm {
	case HorizontalTB:
		return "HorizontalTB"
	case VerticalRL:
		return "VerticalRL"
	case VerticalLR:
		return "VerticalLR"
	}
	return "Invalid(" + strconv.Itoa(int(wm)) + ")"
}

// TextOrientation specifies how horizontal text should be oriented within vertical text, or how vertical-only text should be laid out in horizontal text.
type TextOrientation int

// see TextOrientation
const (
	Natural TextOrientation = iota // turn horizontal text 90deg clockwise for VerticalRL, and counter clockwise for VerticalLR
	Upright                        // split characters and lay them out upright
)

func (orient TextOrientation) String() string {
	switch orient {
	case Natural:
		return "Natural"
	case Upright:
		return "Upright"
	}
	return "Invalid(" + strconv.Itoa(int(orient)) + ")"
}

// Text holds the representation of a text object.
type Text struct {
	lines []line
	fonts map[*Font]bool
	WritingMode
	TextOrientation
	width, height float64
	text          string
	Overflows     bool // true if lines stick out of the box
}

type line struct {
	y     float64
	spans []TextSpan
}

// Heights returns the maximum top, ascent, descent, and bottom heights of the line, where top and bottom are equal to ascent and descent respectively with added line spacing.
func (l line) Heights(mode WritingMode) (float64, float64, float64, float64) {
	top, ascent, descent, bottom := 0.0, 0.0, 0.0, 0.0
	if mode == HorizontalTB {
		for _, span := range l.spans {
			if span.IsText() {
				spanTop, spanAscent, spanDescent, spanBottom := span.Face.heights(mode)
				top = math.Max(top, spanTop)
				ascent = math.Max(ascent, spanAscent)
				descent = math.Max(descent, spanDescent)
				bottom = math.Max(bottom, spanBottom)
			} else {
				for _, obj := range span.Objects {
					spanAscent, spanDescent := obj.Heights(span.Face)
					lineSpacing := span.Face.Metrics().LineGap
					top = math.Max(top, spanAscent+lineSpacing)
					ascent = math.Max(ascent, spanAscent)
					descent = math.Max(descent, spanDescent)
					bottom = math.Max(bottom, spanDescent+lineSpacing)
				}
			}
		}
	} else {
		width := 0.0
		for _, span := range l.spans {
			if span.IsText() {
				for _, glyph := range span.Glyphs {
					if glyph.Vertical {
						width = math.Max(width, 1.2*span.Face.mmPerEm*float64(glyph.SFNT.GlyphAdvance(glyph.ID))) // TODO: what left/right padding should upright characters in a vertical layout have?
					} else {
						spanTop, spanAscent, spanDescent, spanBottom := span.Face.heights(mode)
						top = math.Max(top, spanTop)
						ascent = math.Max(ascent, spanAscent)
						descent = math.Max(descent, spanDescent)
						bottom = math.Max(bottom, spanBottom)
					}
				}
			} else {
				for _, obj := range span.Objects {
					width = math.Max(width, obj.Width)
				}
			}
		}
		top = math.Max(top, width/2.0)
		ascent = math.Max(ascent, width/2.0)
		descent = math.Max(descent, width/2.0)
		bottom = math.Max(bottom, width/2.0)
	}
	return top, ascent, descent, bottom
}

// TextSpan is a span of text.
type TextSpan struct {
	X         float64
	Width     float64
	Face      *FontFace
	Text      string
	Glyphs    []canvasText.Glyph
	Direction canvasText.Direction
	Rotation  canvasText.Rotation

	Objects []TextSpanObject
}

// IsText returns true if the text span is text and not objects (such as images or paths).
func (span *TextSpan) IsText() bool {
	return len(span.Objects) == 0
}

// TextSpanObject is an object that can be used within a text span. It is a wrapper around Canvas and can thus draw anything to be mixed with text, such as images (emoticons) or paths (symbols).
type TextSpanObject struct {
	*Canvas
	X, Y          float64
	Width, Height float64
	VAlign        VerticalAlign
}

// Heights returns the ascender and descender values of the span object.
func (obj TextSpanObject) Heights(face *FontFace) (float64, float64) {
	switch obj.VAlign {
	case FontTop:
		ascent := face.Metrics().Ascent
		return ascent, -(ascent - obj.Height)
	case FontMiddle:
		ascent, descent := face.Metrics().Ascent, face.Metrics().Descent
		return (ascent - descent + obj.Height) / 2.0, -(ascent - descent - obj.Height) / 2.0
	case FontBottom:
		descent := face.Metrics().Descent
		return -descent + obj.Height, descent
	}
	return obj.Height, 0.0 // Baseline
}

// View returns the object's view to be placed within the text line.:
func (obj TextSpanObject) View(x, y float64, face *FontFace) Matrix {
	_, bottom := obj.Heights(face)
	return Identity.Translate(x+obj.X, y+obj.Y-bottom)
}

////////////////////////////////////////////////////////////////

func itemizeString(log string) []canvasText.ScriptItem {
	logRunes := []rune(log)
	embeddingLevels := canvasText.EmbeddingLevels(logRunes)
	return canvasText.ScriptItemizer(logRunes, embeddingLevels)
}

// NewTextLine is a simple text line using a single font face, a string (supporting new lines) and horizontal alignment (Left, Center, Right). The text's baseline will be drawn on the current coordinate.
func NewTextLine(face *FontFace, s string, halign TextAlign) *Text {
	t := &Text{
		fonts: map[*Font]bool{face.Font: true},
		text:  s,
	}

	ascent, descent, spacing := face.Metrics().Ascent, face.Metrics().Descent, face.Metrics().LineGap

	i := 0
	y := 0.0
	skipNext := false
	for j, r := range s + "\n" {
		if canvasText.IsParagraphSeparator(r) {
			if skipNext {
				skipNext = false
				i++
				continue
			}
			if i < j {
				ppem := face.PPEM(DefaultResolution)
				lineWidth := 0.0
				line := line{y: y, spans: []TextSpan{}}
				for _, item := range itemizeString(s[i:j]) {
					glyphs, direction := face.Font.shaper.Shape(item.Text, ppem, face.Direction, face.Script, face.Language, face.Font.features, face.Font.variations)
					width := face.textWidth(glyphs)
					line.spans = append(line.spans, TextSpan{
						X:         lineWidth,
						Width:     width,
						Face:      face,
						Text:      item.Text,
						Glyphs:    glyphs,
						Direction: direction,
					})
					lineWidth += width
				}
				if halign == Center || halign == Middle {
					for k := range line.spans {
						line.spans[k].X = -lineWidth / 2.0
					}
				} else if halign == Right {
					for k := range line.spans {
						line.spans[k].X = -lineWidth
					}
				}
				t.lines = append(t.lines, line)
			}
			y += ascent + descent + spacing
			i = j + utf8.RuneLen(r)
			skipNext = r == '\r' && j+1 < len(s) && s[j+1] == '\n'
		}
	}
	return t
}

// NewTextBox is an advanced text formatter that will format text placement based on the settings. It takes a single font face, a string, the width or height of the box (can be zero to disable), horizontal and vertical alignment (Left, Center, Right, Top, Bottom or Justify), text indentation for the first line and line stretch (percentage to stretch the line based on the line height).
func NewTextBox(face *FontFace, s string, width, height float64, halign, valign TextAlign, indent, lineStretch float64) *Text {
	rt := NewRichText(face)
	rt.WriteString(s)
	return rt.ToText(width, height, halign, valign, indent, lineStretch)
}

type indexer []int

func (indexer indexer) index(loc int) int {
	for index, start := range indexer {
		if loc < start {
			return index - 1
		}
	}
	return len(indexer) - 1
}

// RichText allows to build up a rich text with text spans of different font faces and fitting that into a box using Donald Knuth's line breaking algorithm.
type RichText struct {
	*strings.Builder
	locs   indexer // faces locations in string by number of runes
	faces  []*FontFace
	mode   WritingMode
	orient TextOrientation

	defaultFace *FontFace
	objects     []TextSpanObject
}

// NewRichText returns a new rich text with the given default font face.
func NewRichText(face *FontFace) *RichText {
	if face == nil {
		panic("FontFace cannot be nil")
	}
	return &RichText{
		Builder:     &strings.Builder{},
		locs:        indexer{0},
		faces:       []*FontFace{face},
		mode:        HorizontalTB,
		orient:      Natural,
		defaultFace: face,
	}
}

// Reset resets the rich text to its initial state.
func (rt *RichText) Reset() {
	rt.Builder.Reset()
	rt.locs = rt.locs[:1]
	rt.faces = rt.faces[:1]
}

// SetWritingMode sets the writing mode.
func (rt *RichText) SetWritingMode(mode WritingMode) {
	rt.mode = mode
}

// SetTextOrientation sets the text orientation of non-CJK between CJK.
func (rt *RichText) SetTextOrientation(orient TextOrientation) {
	rt.orient = orient
}

// SetFace sets the font face.
func (rt *RichText) SetFace(face *FontFace) {
	if face == nil {
		panic("FontFace cannot be nil")
	}
	rt.setFace(face)
}

func (rt *RichText) setFace(face *FontFace) {
	if face == rt.faces[len(rt.faces)-1] {
		return
	}
	prevLoc := rt.locs[len(rt.locs)-1]
	if rt.Len()-prevLoc == 0 {
		rt.locs = rt.locs[:len(rt.locs)-1]
		rt.faces = rt.faces[:len(rt.faces)-1]
	}
	rt.locs = append(rt.locs, len([]rune(rt.String())))
	rt.faces = append(rt.faces, face)
}

// SetFaceSpan sets the font face between start and end measured in bytes.
func (rt *RichText) SetFaceSpan(face *FontFace, start, end int) {
	// TODO: optimize when face already is on (part of) the span
	if end <= start || rt.Len() <= start {
		return
	} else if rt.Len() < end {
		end = rt.Len()
	}

	k := 0
	i, j := 0, len(rt.locs)-1
	for k < len(rt.locs) {
		if rt.locs[k] < start {
			i = k
		}
		if end <= rt.locs[k] {
			j = k - 1
			break
		}
		k++
	}
	rt.locs[j] = len([]rune(rt.String()[:end]))
	rt.locs = append(rt.locs[:i], append(indexer{len([]rune(rt.String()[:start]))}, rt.locs[j:]...)...)
	rt.faces = append(rt.faces[:i], append([]*FontFace{face}, rt.faces[j:]...)...)
}

// Add adds a string with a given font face.
func (rt *RichText) Add(face *FontFace, text string) *RichText {
	rt.SetFace(face)
	rt.WriteString(text)
	return rt
}

// AddCanvas adds a canvas object that can have paths/images/texts.
func (rt *RichText) AddCanvas(c *Canvas, valign VerticalAlign) *RichText {

	width, height := c.Size()
	face := rt.faces[len(rt.faces)-1]
	rt.setFace(nil)
	rt.WriteRune(rune(len(rt.objects)))
	rt.objects = append(rt.objects, TextSpanObject{
		Canvas: c,
		Width:  width,
		Height: height,
		VAlign: valign,
	})
	rt.setFace(face)
	return rt
}

// AddPath adds a path.
func (rt *RichText) AddPath(path *Path, col color.RGBA, valign VerticalAlign) *RichText {
	style := DefaultStyle
	style.Fill.Color = col
	bounds := path.Bounds()
	c := New(bounds.X+bounds.W, bounds.Y+bounds.H)
	c.RenderPath(path, style, Identity)
	rt.AddCanvas(c, valign)
	return rt
}

// AddImage adds an image.
func (rt *RichText) AddImage(img image.Image, res Resolution, valign VerticalAlign) *RichText {
	bounds := img.Bounds().Size()
	c := New(float64(bounds.X)/res.DPMM(), float64(bounds.Y)/res.DPMM())
	c.RenderImage(img, Identity.Scale(1.0/res.DPMM(), 1.0/res.DPMM()))
	rt.AddCanvas(c, valign)
	return rt
}

// AddLaTeX adds a LaTeX formula.
func (rt *RichText) AddLaTeX(s string) error {
	p, err := ParseLaTeX(s)
	if err != nil {
		return err
	}
	rt.AddPath(p, Black, Baseline)
	return nil
}

func scriptDirection(mode WritingMode, orient TextOrientation, script canvasText.Script, direction canvasText.Direction) (canvasText.Direction, canvasText.Rotation) {
	if direction == canvasText.TopToBottom || direction == canvasText.BottomToTop {
		if mode == HorizontalTB {
			direction = canvasText.LeftToRight
		} else {
			direction = canvasText.TopToBottom
		}
	} else if mode != HorizontalTB {
		// unknown, left to right, right to left
		direction = canvasText.TopToBottom
	}
	rotation := canvasText.NoRotation
	if mode != HorizontalTB {
		if !canvasText.IsVerticalScript(script) && orient == Natural {
			direction = canvasText.LeftToRight
			rotation = canvasText.CW
		} else if rotation = canvasText.ScriptRotation(script); rotation != canvasText.NoRotation {
			direction = canvasText.LeftToRight
		}
	}
	return direction, rotation
}

// ToText takes the added text spans and fits them within a given box of certain width and height using Donald Knuth's line breaking algorithm.
func (rt *RichText) ToText(width, height float64, halign, valign TextAlign, indent, lineStretch float64) *Text {
	log := rt.String()
	logRunes := []rune(log)
	embeddingLevels := canvasText.EmbeddingLevels(logRunes)

	// itemize string by font face and script
	texts := []string{}
	scripts := []canvasText.Script{}
	faces := []*FontFace{}
	i := 0       // index into logRunes
	curFace := 0 // index into rt.faces
	for j := range logRunes {
		nextFace := rt.locs.index(j)
		if nextFace != curFace {
			if rt.faces[curFace] == nil {
				// path/image objects
				texts = append(texts, string(logRunes[i:j]))
				scripts = append(scripts, canvasText.ScriptInvalid)
				faces = append(faces, nil)
			} else {
				// text
				items := canvasText.ScriptItemizer(logRunes[i:j], embeddingLevels[i:j])
				for _, item := range items {
					texts = append(texts, item.Text)
					scripts = append(scripts, item.Script)
					faces = append(faces, rt.faces[curFace])
				}
			}
			curFace = nextFace
			i = j
		}
	}
	if i < len(logRunes) {
		if rt.faces[curFace] == nil {
			// path/image objects
			texts = append(texts, string(logRunes[i:]))
			scripts = append(scripts, canvasText.ScriptInvalid)
			faces = append(faces, nil)
		} else {
			// text
			items := canvasText.ScriptItemizer(logRunes[i:], embeddingLevels[i:])
			for _, item := range items {
				texts = append(texts, item.Text)
				scripts = append(scripts, item.Script)
				faces = append(faces, rt.faces[curFace])
			}
		}
	}

	// shape text into glyphs and keep index into texts and faces
	clusterOffset := uint32(0)
	glyphIndices := indexer{} // indexes glyphs into texts and faces
	glyphs := []canvasText.Glyph{}
	directions := make([]canvasText.Direction, len(texts))
	rotations := make([]canvasText.Rotation, len(texts))
	for k, text := range texts {
		face := faces[k]
		script := scripts[k]
		direction := canvasText.DirectionInvalid
		rotation := canvasText.NoRotation
		var glyphsString []canvasText.Glyph
		if face == nil {
			// path/image objects
			for i, r := range text {
				obj := rt.objects[r]
				ppem := float64(rt.defaultFace.Font.SFNT.Head.UnitsPerEm)
				xadv, yadv := obj.Width, obj.Height
				if rt.mode != HorizontalTB {
					yadv = -yadv
				}
				glyphsString = append(glyphsString, canvasText.Glyph{
					SFNT:     rt.defaultFace.Font.SFNT,
					Size:     rt.defaultFace.Size,
					Script:   script,
					Vertical: rt.mode != HorizontalTB,
					ID:       uint16(r),
					Cluster:  clusterOffset + uint32(i),
					XAdvance: int32(xadv * ppem / rt.defaultFace.Size),
					YAdvance: int32(yadv * ppem / rt.defaultFace.Size),
				})
			}
		} else {
			// text
			ppem := face.PPEM(DefaultResolution)
			direction, rotation = scriptDirection(rt.mode, rt.orient, script, face.Direction)
			glyphsString, direction = face.Font.shaper.Shape(text, ppem, direction, script, face.Language, face.Font.features, face.Font.variations)
			for i := range glyphsString {
				glyphsString[i].SFNT = face.Font.SFNT
				glyphsString[i].Size = face.Size
				glyphsString[i].Script = script
				glyphsString[i].Vertical = direction == canvasText.TopToBottom || direction == canvasText.BottomToTop
				glyphsString[i].Cluster += clusterOffset
				if rt.mode != HorizontalTB {
					if script == canvasText.Mongolian {
						glyphsString[i].YOffset += int32(face.Font.SFNT.Hhea.Descender)
					} else if rotation != canvasText.NoRotation {
						// center horizontal text by x-height when rotated in vertical layout
						glyphsString[i].YOffset -= int32(face.Font.SFNT.OS2.SxHeight) / 2
					} else if rt.orient == Upright && rotation == canvasText.NoRotation && !canvasText.IsVerticalScript(script) {
						// center horizontal text vertically when upright in vertical layout
						glyphsString[i].YOffset = -(int32(face.Font.SFNT.Head.UnitsPerEm) + int32(face.Font.SFNT.OS2.SxHeight)) / 2
					}
				}
			}
		}

		if direction == canvasText.RightToLeft || direction == canvasText.BottomToTop {
			// reverse right-to-left and bottom-to-top glyph order for line breaking purposes
			// this is required when mixing e.g. LTR and RTL scripts where line breaking should
			// treat the RTL words in the logical order. We undo this later on.
			for i := 0; i < len(glyphsString)/2; i++ {
				glyphsString[i], glyphsString[len(glyphsString)-1-i] = glyphsString[len(glyphsString)-1-i], glyphsString[i]
			}
		}

		glyphIndices = append(glyphIndices, len(glyphs))
		glyphs = append(glyphs, glyphsString...)
		clusterOffset += uint32(len(text))
		directions[k] = direction
		rotations[k] = rotation
	}

	if rt.mode != HorizontalTB {
		width, height = height, width
		halign, valign = valign, halign
		if halign == Top {
			halign = Left
		} else if halign == Bottom {
			halign = Right
		}
		if valign == Left {
			valign = Top
		} else if valign == Right {
			valign = Bottom
		}
	}

	align := canvasText.Left
	if halign == Justify {
		align = canvasText.Justified
	}

	// break glyphs into lines following Donald Knuth's line breaking algorithm
	looseness := 0
	items := canvasText.GlyphsToItems(glyphs, indent, align)

	var breaks []*canvasText.Breakpoint
	var overflows bool
	if width != 0.0 {
		var ok bool
		breaks, ok = canvasText.Linebreak(items, width, looseness)
		overflows = !ok
	} else if len(items) == 0 {
		breaks = append(breaks, &canvasText.Breakpoint{Position: 0, Width: 0.0})
	} else {
		lineWidth := 0.0
		for i, item := range items {
			if item.Type != canvasText.PenaltyType {
				lineWidth += item.Width
			} else if item.Penalty <= -canvasText.Infinity {
				breaks = append(breaks, &canvasText.Breakpoint{Position: i, Width: lineWidth})
				lineWidth = 0.0
			}
		}
	}

	// clean up items, remove penalties/glues that were not chosen as breaks, this concatenates adjacent boxes and thus spans
	var j int
	i, j = 0, 0 // index into: glyphs, breaks/lines
	shift := 0  // break index shift
	if 0 < len(items) && items[0].Width == 0.0 {
		// remove empty indent box
		items = items[1:]
		shift++
	}
	for k := 0; k < len(items); k++ {
		size := items[k].Size
		if k == breaks[j].Position-shift {
			// keep breaking item
			breaks[j].Position -= shift
			j++
		} else if 0 < k && items[k].Type == canvasText.GlueType && 0 < j && k-1 == breaks[j-1].Position {
			// put spaces at the beginning of the line into the break
			items[k-1].Size += items[k].Size
			items = append(items[:k], items[k+1:]...)
			shift++
			k--
		} else if k+1 < len(items) && items[k].Type == canvasText.GlueType && k+1 == breaks[j].Position-shift {
			// put spaces at the end of the line into the break
			items[k+1].Size += items[k].Size
			items = append(items[:k], items[k+1:]...)
			shift++
			k--
		} else if items[k].Type == canvasText.PenaltyType && items[k].Size == 0 {
			// remove non-breaking penalties
			items = append(items[:k], items[k+1:]...)
			shift++
			k--
		} else if items[k].Type == canvasText.GlueType && items[k].Size == 0 && breaks[j].Ratio == 0.0 {
			// remove empty glues
			items = append(items[:k], items[k+1:]...)
			shift++
			k--
		} else if 0 < k && items[k].Type == canvasText.GlueType && items[k-1].Type == canvasText.GlueType {
			// merge glues
			items[k-1].Width += items[k].Width
			items[k-1].Stretch += items[k].Stretch
			items[k-1].Shrink += items[k].Shrink
			items[k-1].Size += items[k].Size
			items = append(items[:k], items[k+1:]...)
			shift++
			k -= 2 // parse it again in case we have a box-glue pair
		} else if 0 < k && items[k].Type == canvasText.BoxType && items[k-1].Type == canvasText.BoxType {
			// merge boxes
			items[k-1].Width += items[k].Width
			items[k-1].Size += items[k].Size
			items = append(items[:k], items[k+1:]...)
			shift++
			k--
		} else if 0 < k && items[k].Type == canvasText.GlueType && (breaks[j].Ratio == 0.0 || items[k].Stretch == 0.0 && items[k].Shrink == 0.0) && items[k-1].Type == canvasText.BoxType {
			// merge glue with box when glue is the width of a space
			items[k-1].Type = canvasText.BoxType
			items[k-1].Width += items[k].Width
			items[k-1].Size += items[k].Size
			items = append(items[:k], items[k+1:]...)
			shift++
			k--
		}
		i += size
	}

	// build up lines
	t := &Text{
		lines:           []line{{}},
		fonts:           map[*Font]bool{},
		WritingMode:     rt.mode,
		TextOrientation: rt.orient,
		width:           width,
		height:          height,
		text:            log,
		Overflows:       overflows,
	}
	glyphs = append(glyphs, canvasText.Glyph{Cluster: uint32(len(log))}) // makes indexing easier

	i, j = 0, 0      // index into: glyphs, breaks/lines
	x, y := 0.0, 0.0 // both positive toward the bottom right
	lineSpacing := 1.0 + lineStretch
	if halign == Right {
		x += width - breaks[j].Width
	} else if halign == Center || halign == Middle {
		x += (width - breaks[j].Width) / 2.0
	}
	for position, item := range items {
		if position == breaks[j].Position {
			if 0 < len(t.lines[j].spans) { // not if there is an empty first line
				// add spaces to previous span
				for _, glyph := range glyphs[i : i+item.Size] {
					t.lines[j].spans[len(t.lines[j].spans)-1].Text += string(glyph.Text)
				}

				// hyphenate at breakpoint
				if item.Type == canvasText.PenaltyType && item.Size == 1 && glyphs[i].Text == '\u00AD' {
					span := &t.lines[j].spans[len(t.lines[j].spans)-1]
					id := span.Face.Font.GlyphIndex('-')
					glyph := canvasText.Glyph{
						SFNT:     span.Face.Font.SFNT,
						Size:     span.Face.Size,
						ID:       id,
						XAdvance: int32(span.Face.Font.GlyphAdvance(id)),
						Text:     '-',
					}
					span.Glyphs = append(span.Glyphs, glyph)
					span.Width += span.Face.textWidth([]canvasText.Glyph{glyph})
					span.Text += "-"
				}
			}

			var ascent, descent, bottom float64
			if len(t.lines[j].spans) == 0 {
				_, ascent, descent, bottom = faces[glyphIndices.index(i)].heights(rt.mode)
			} else {
				_, ascent, descent, bottom = t.lines[j].Heights(rt.mode)
			}
			if 0 < j {
				ascent *= lineSpacing
				// don't stretch descent for possible last line
			}
			bottom *= lineSpacing

			if height != 0.0 && height < y+ascent+descent {
				// doesn't fit or at the end of items
				t.lines = t.lines[:len(t.lines)-1]
				if 0 < j {
					t.text = log[:glyphs[i].Cluster]
				} else {
					t.text = ""
					y = 0.0
				}
				break
			}
			t.lines[j].y = y + ascent
			y += ascent + bottom
			if position == len(items)-1 {
				break
			}

			t.lines = append(t.lines, line{})
			if j+1 < len(breaks) {
				j++
			}
			x = 0.0
			if halign == Right {
				x += width - breaks[j].Width
			} else if halign == Center || halign == Middle {
				x += (width - breaks[j].Width) / 2.0
			}
		} else if item.Type == canvasText.BoxType {
			// find index k into faces/texts
			// find a,b index range into glyphs
			a := i
			dx := 0.0
			k := glyphIndices.index(i)
			for b := i + 1; b <= i+item.Size; b++ {
				nextK := glyphIndices.index(b)
				if nextK != k || b == i+item.Size {
					face := faces[k]
					ac, bc := glyphs[a].Cluster, glyphs[b].Cluster

					var w float64
					var objects []TextSpanObject
					if face != nil {
						// text
						w = face.textWidth(glyphs[a:b])
						t.fonts[face.Font] = true
					} else {
						// path/image object, only one glyph is ever selected; b-a == 1
						if 0 < len(t.lines[j].spans) {
							face = t.lines[j].spans[len(t.lines[j].spans)-1].Face
						} else {
							face = rt.defaultFace
						}
						for _, glyph := range glyphs[a:b] {
							obj := rt.objects[glyph.ID]
							if rt.mode == HorizontalTB {
								obj.X = w
								w += obj.Width
							} else {
								obj.X = -obj.Width / 2.0
								obj.Y = -w - obj.Height
								w += obj.Height
							}
							objects = append(objects, obj)
						}
					}

					if directions[k] == canvasText.RightToLeft || directions[k] == canvasText.BottomToTop {
						// reverse right-to-left and bottom-to-top glyph order
						// this undoes the previous reversal for line breaking purposed
						for i := 0; i < (b-a)/2; i++ {
							glyphs[a+i], glyphs[b-1-i] = glyphs[b-1-i], glyphs[a+i]
						}
					}

					s := log[ac:bc]
					t.lines[j].spans = append(t.lines[j].spans, TextSpan{
						X:         x + dx,
						Width:     w,
						Face:      face,
						Text:      s,
						Objects:   objects,
						Glyphs:    glyphs[a:b],
						Direction: directions[k],
						Rotation:  rotations[k],
					})

					if directions[k] == canvasText.RightToLeft || directions[k] == canvasText.BottomToTop {
						// reverse right-to-left and bottom-to-top span order in line
						// this undoes the previous reversal for line breaking purposed
						last := len(t.lines[j].spans) - 1
						first := last
						for ; 0 < first; first-- {
							if t.lines[j].spans[first-1].Direction != canvasText.RightToLeft && t.lines[j].spans[first-1].Direction != canvasText.BottomToTop {
								break
							}
						}
						if first < last {
							space := x + dx - t.lines[j].spans[first].X - t.lines[j].spans[first].Width
							t.lines[j].spans[last].X = t.lines[j].spans[last-1].X
							for i := first; i < last; i++ {
								t.lines[j].spans[i].X += w + space
							}
						}
					}

					k = nextK
					a = b
					dx += w
				}
			}
			x += item.Width
		} else if item.Type == canvasText.GlueType {
			width := item.Width
			if 0.0 <= breaks[j].Ratio {
				if !math.IsInf(item.Stretch, 0.0) {
					width += breaks[j].Ratio * item.Stretch
				}
			} else if !math.IsInf(item.Shrink, 0.0) {
				width += breaks[j].Ratio * item.Shrink
			}
			x += width

			// add spaces to previous span
			if 0 < len(t.lines[j].spans) { // don't add if there is an empty first line
				for _, glyph := range glyphs[i : i+item.Size] {
					t.lines[j].spans[len(t.lines[j].spans)-1].Text += string(glyph.Text)
				}
			}
		}
		i += item.Size
	}

	if 0 < j {
		// remove line gap of last line
		_, _, descent, bottom := t.lines[j-1].Heights(rt.mode)
		y += -bottom*lineSpacing + descent
	}

	// vertical align
	if rt.mode == VerticalRL {
		if valign == Top {
			valign = Bottom
		} else if valign == Bottom {
			valign = Top
		}
	}
	if valign == Center || valign == Middle || valign == Bottom {
		dy := height - y
		if valign == Center || valign == Middle {
			dy /= 2.0
		}
		for j := range t.lines {
			t.lines[j].y += dy
		}
	} else if valign == Justify {
		ddy := (height - y) / float64(len(t.lines)-1)
		dy := 0.0
		for j := range t.lines {
			t.lines[j].y += dy
			dy += ddy
		}
	}
	if rt.mode == VerticalRL {
		for j := range t.lines {
			t.lines[j].y = height - t.lines[j].y
		}
	}
	return t
}

// Empty returns true if there are no text lines or text spans.
func (t *Text) Empty() bool {
	for _, line := range t.lines {
		if len(line.spans) != 0 {
			return false
		}
	}
	return true
}

// Size returns the width and height of a text box. Either can be zero when unspecified.
func (t *Text) Size() (float64, float64) {
	return t.width, t.height
}

// Heights returns the top and bottom position of the first and last line respectively.
func (t *Text) Heights() (float64, float64) {
	if len(t.lines) == 0 {
		return 0.0, 0.0
	}
	firstLine := t.lines[0]
	lastLine := t.lines[len(t.lines)-1]
	_, ascent, _, _ := firstLine.Heights(t.WritingMode)
	_, _, descent, _ := lastLine.Heights(t.WritingMode)
	return -firstLine.y + ascent, lastLine.y + descent
}

// Bounds returns the bounding rectangle that defines the text box.
func (t *Text) Bounds() Rect {
	if len(t.lines) == 0 || len(t.lines[0].spans) == 0 {
		return Rect{}
	}
	rect := Rect{}
	for _, line := range t.lines {
		for _, span := range line.spans {
			// TODO: vertical text
			rect = rect.Add(Rect{span.X, -line.y - span.Face.Metrics().Descent, span.Width, span.Face.Metrics().Ascent + span.Face.Metrics().Descent})
		}
	}
	return rect
}

// OutlineBounds returns the rectangle that contains the entire text box, i.e. the glyph outlines (slow).
func (t *Text) OutlineBounds() Rect {
	if len(t.lines) == 0 || len(t.lines[0].spans) == 0 {
		return Rect{}
	}
	r := Rect{}
	for _, line := range t.lines {
		for _, span := range line.spans {
			// TODO: vertical text
			p, _, err := span.Face.toPath(span.Glyphs, span.Face.PPEM(DefaultResolution))
			if err != nil {
				panic(err)
			}
			spanBounds := p.Bounds()
			spanBounds = spanBounds.Move(Point{span.X, -line.y})
			r = r.Add(spanBounds)
		}
	}
	t.WalkDecorations(func(_ Paint, p *Path) {
		r = r.Add(p.Bounds())
	})
	return r
}

// Fonts returns the list of fonts used.
func (t *Text) Fonts() []*Font {
	fonts := []*Font{}
	fontNames := []string{}
	fontMap := map[string]*Font{}
	for font := range t.fonts {
		name := font.Name()
		fontNames = append(fontNames, name)
		fontMap[name] = font
	}
	sort.Strings(fontNames)
	for _, name := range fontNames {
		fonts = append(fonts, fontMap[name])
	}
	return fonts
}

// MostCommonFontFace returns the most common FontFace of the text.
func (t *Text) MostCommonFontFace() *FontFace {
	fonts := map[*Font]int{}
	sizes := map[float64]int{}
	styles := map[FontStyle]int{}
	variants := map[FontVariant]int{}
	colors := map[color.RGBA]int{}
	for _, line := range t.lines {
		for _, span := range line.spans {
			fonts[span.Face.Font]++
			sizes[span.Face.Size]++
			styles[span.Face.Style]++
			variants[span.Face.Variant]++
			if span.Face.Fill.IsColor() {
				colors[span.Face.Fill.Color]++ // TODO: also for patterns or other fill paints
			}
		}
	}
	if len(fonts) == 0 {
		return nil
	}

	font, size, style, variant, col := (*Font)(nil), 0.0, FontRegular, FontNormal, Black
	for key, val := range fonts {
		if fonts[font] < val {
			font = key
		}
	}
	for key, val := range sizes {
		if sizes[size] < val {
			size = key
		}
	}
	for key, val := range styles {
		if styles[style] < val {
			style = key
		}
	}
	for key, val := range variants {
		if variants[variant] < val {
			variant = key
		}
	}
	for key, val := range colors {
		if colors[col] < val {
			col = key
		}
	}

	face := font.Face(size*ptPerMm, col)
	face.Style = style
	face.Variant = variant
	return face
}

type decorationSpan struct {
	deco  FontDecorator
	fill  Paint
	x     float64
	width float64
	face  *FontFace // biggest face
}

// WalkDecorations calls the callback for each color of decoration used per line.
func (t *Text) WalkDecorations(callback func(fill Paint, deco *Path)) {
	// TODO: vertical text
	// accumulate paths with colors for all lines
	fs := []Paint{}
	ps := []*Path{}
	for _, line := range t.lines {
		// track active decorations, when finished draw and append to accumulated paths
		active := []decorationSpan{}
		for k, span := range line.spans {
			foundActive := make([]bool, len(active))
			for _, spanDeco := range span.Face.Deco {
				found := false
				for i, deco := range active {
					if reflect.DeepEqual(span.Face.Fill, deco.fill) && reflect.DeepEqual(deco.deco, spanDeco) {
						// extend decoration
						active[i].width = span.X + span.Width - active[i].x
						if active[i].face.Size < span.Face.Size {
							active[i].face = span.Face
						}
						foundActive[i] = true
						found = true
						break
					}
				}
				if !found {
					// add new decoration
					active = append(active, decorationSpan{
						deco:  spanDeco,
						fill:  span.Face.Fill,
						x:     span.X,
						width: span.Width,
						face:  span.Face,
					})
				}
			}

			if k == len(line.spans)-1 {
				foundActive = make([]bool, len(active))
			}

			di := 0
			for i, found := range foundActive {
				if !found {
					// remove active decoration and draw it
					decoSpan := active[i-di]
					xOffset := span.Face.mmPerEm * float64(span.Face.XOffset)
					yOffset := span.Face.mmPerEm * float64(span.Face.YOffset)
					p := decoSpan.deco.Decorate(decoSpan.face, decoSpan.width)
					p = p.Translate(decoSpan.x+xOffset, -line.y+yOffset)

					foundFill := false
					for j, fill := range fs {
						if reflect.DeepEqual(fill, decoSpan.fill) {
							ps[j] = ps[j].Append(p)
							foundFill = true
						}
					}
					if !foundFill {
						fs = append(fs, decoSpan.fill)
						ps = append(ps, p)
					}

					active = append(active[:i-di], active[i-di+1:]...)
					di++
				}
			}
		}
	}

	for i := 0; i < len(ps); i++ {
		callback(fs[i], ps[i])
	}
}

// WalkLines calls the callback for each text line.
func (t *Text) WalkLines(callback func(float64, []TextSpan)) {
	for _, line := range t.lines {
		callback(-line.y, line.spans)
	}
}

// WalkSpans calls the callback for each text span per line.
func (t *Text) WalkSpans(callback func(float64, float64, TextSpan)) {
	for _, line := range t.lines {
		for _, span := range line.spans {
			xOffset := span.Face.mmPerEm * float64(span.Face.XOffset)
			yOffset := span.Face.mmPerEm * float64(span.Face.YOffset)
			if t.WritingMode == HorizontalTB {
				callback(span.X+xOffset, -line.y+yOffset, span)
			} else {
				callback(line.y+xOffset, -span.X+yOffset, span)
			}
		}
	}
}

// RenderAsPath renders the text and its decorations converted to paths, calling r.RenderPath.
func (t *Text) RenderAsPath(r Renderer, m Matrix, resolution Resolution) {
	t.WalkDecorations(func(paint Paint, p *Path) {
		style := DefaultStyle
		style.Fill = paint
		r.RenderPath(p, style, m)
	})

	for _, line := range t.lines {
		for _, span := range line.spans {
			x, y := span.X, -line.y
			if t.WritingMode != HorizontalTB {
				x, y = line.y, -span.X
			}

			if span.IsText() {
				style := DefaultStyle
				style.Fill = span.Face.Fill
				p, _, err := span.Face.toPath(span.Glyphs, span.Face.PPEM(resolution))
				if err != nil {
					panic(err)
				}
				p = p.Transform(Identity.Rotate(float64(span.Rotation)))
				if resolution != 0.0 && span.Face.Hinting != font.NoHinting && span.Rotation == text.NoRotation {
					// grid-align vertically on pixel raster, this improves font sharpness
					_, dy := m.Pos()
					dy += y
					y += float64(int(dy*resolution.DPMM()+0.5))/resolution.DPMM() - dy
				}
				p = p.Translate(x, y)
				r.RenderPath(p, style, m)
			} else {
				for _, obj := range span.Objects {
					obj.RenderViewTo(r, m.Mul(obj.View(x, y, span.Face)))
				}
			}
		}
	}
}

// String returns the content of the text box.
func (t *Text) String() string {
	return t.text
}
