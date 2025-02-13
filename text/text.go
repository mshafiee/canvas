package text

import (
	"fmt"

	"github.com/tdewolff/canvas/font"
)

type ScriptItem struct {
	Script
	Text string
}

// ScriptItemizer divides the string in parts for each different script.
func ScriptItemizer(runes []rune, embeddingLevels []int) []ScriptItem {
	if len(runes) == 0 {
		return []ScriptItem{}
	}

	i := 0
	items := []ScriptItem{}
	scripts := []Script{ScriptInherited} // script stack for embedding levels
	for j, r := range runes {
		curLevel := len(scripts) - 1
		curScript := scripts[curLevel]
		script, level := LookupScript(r), embeddingLevels[j]
		if len(scripts) <= level {
			for len(scripts) < level {
				scripts = append(scripts, ScriptInherited)
			}
			scripts = append(scripts, script)
		} else {
			if script != ScriptInherited && script != ScriptCommon {
				scripts[level] = script
			} else {
				script = scripts[level]
			}
			scripts = scripts[:level+1]
		}

		if j != 0 && (curLevel != level || curScript != script && curScript != ScriptInherited && curScript != ScriptCommon && script != ScriptInherited && script != ScriptCommon) {
			items = append(items, ScriptItem{
				Script: curScript,
				Text:   string(runes[i:j]),
			})
			i = j
		}
	}
	items = append(items, ScriptItem{
		Script: scripts[len(scripts)-1],
		Text:   string(runes[i:]),
	})
	return items
}

// Glyph is a shaped glyph for the given font and font size. It specified the glyph ID, the cluster ID, its X and Y advance and offset in font units, and its representation as text.
type Glyph struct {
	SFNT *font.SFNT
	Size float64
	Script
	Vertical bool // is false for Latin/Mongolian/etc in a vertical layout

	ID       uint16
	Cluster  uint32
	XAdvance int32
	YAdvance int32
	XOffset  int32
	YOffset  int32
	Text     rune
}

func (g Glyph) Advance() float64 {
	if !g.Vertical {
		return float64(g.XAdvance) * g.Size / float64(g.SFNT.Head.UnitsPerEm)
	} else {
		return float64(-g.YAdvance) * g.Size / float64(g.SFNT.Head.UnitsPerEm)
	}
}

func (g Glyph) String() string {
	return fmt.Sprintf("%s GID=%v Cluster=%v Adv=(%v,%v) Off=(%v,%v)", string(g.Text), g.ID, g.Cluster, g.XAdvance, g.YAdvance, g.XOffset, g.YOffset)
}

func (g Glyph) Rotation() Rotation {
	rot := NoRotation
	if !g.Vertical {
		rot = ScriptRotation(g.Script)
		if rot == NoRotation {
			rot = CW
		}
	}
	return rot
}

// TODO: implement Liang's (soft) hyphenation algorithm? Add \u00AD at opportunities, unless \u2060 or \uFEFF is present

// IsParagraphSeparator returns true for paragraph separator runes.
func IsParagraphSeparator(r rune) bool {
	// line feed, vertical tab, form feed, carriage return, next line, line separator, paragraph separator
	return 0x0A <= r && r <= 0x0D || r == 0x85 || r == '\u2028' || r == '\u2029'
}

func IsSpacelessScript(script Script) bool {
	// missing: S'gaw Karen
	return script == Han || script == Hangul || script == Katakana || script == Khmer || script == Lao || script == PhagsPa || script == Brahmi || script == TaiTham || script == NewTaiLue || script == TaiLe || script == TaiViet || script == Thai || script == Tibetan || script == Myanmar
}

func IsVerticalScript(script Script) bool {
	return script == Bopomofo || script == EgyptianHieroglyphs || script == Hiragana || script == Katakana || script == Han || script == Hangul || script == MeroiticCursive || script == MeroiticHieroglyphs || script == Mongolian || script == Ogham || script == OldTurkic || script == PhagsPa || script == Yi
}

type Rotation float64

const (
	NoRotation Rotation = 0.0
	CW         Rotation = -90.0
	CCW        Rotation = 90.0
)

func ScriptRotation(script Script) Rotation {
	if script == Mongolian || script == PhagsPa {
		return CW
	} else if script == Ogham || script == OldTurkic {
		return CCW
	}
	return NoRotation
}
