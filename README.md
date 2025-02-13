![Canvas](https://raw.githubusercontent.com/tdewolff/canvas/master/resources/title/title.png)

[![API reference](https://img.shields.io/badge/godoc-reference-5272B4)](https://pkg.go.dev/github.com/tdewolff/canvas?tab=doc) [![User guide](https://img.shields.io/badge/user-guide-5272B4)](https://github.com/tdewolff/canvas/wiki) [![Go Report Card](https://goreportcard.com/badge/github.com/tdewolff/canvas)](https://goreportcard.com/report/github.com/tdewolff/canvas) [![Coverage Status](https://coveralls.io/repos/github/tdewolff/canvas/badge.svg?branch=master)](https://coveralls.io/github/tdewolff/canvas?branch=master) [![Donate](https://img.shields.io/badge/patreon-donate-DFB317)](https://www.patreon.com/tdewolff)

**[API documentation](https://pkg.go.dev/github.com/tdewolff/canvas?tab=doc)**

**[User guide](https://github.com/tdewolff/canvas/wiki)**

**[Live HTMLCanvas demo](https://tdewolff.github.io/canvas/examples/html-canvas/index.html)**

Canvas is a common vector drawing target that can output SVG, PDF, EPS, raster images (PNG, JPG, GIF, ...), HTML Canvas through WASM, OpenGL, and Gio. It has a wide range of path manipulation functionality such as flattening, stroking and dashing implemented. Additionally, it has a text formatter and embeds and subsets fonts (TTF, OTF, WOFF, WOFF2, or EOT) or converts them to outlines. It can be considered a Cairo or node-canvas alternative in Go. See the example below in Figure 1 for an overview of the functionality.

![Preview](https://raw.githubusercontent.com/tdewolff/canvas/master/resources/preview/preview.png)

**Figure 1**: top-left you can see text being fitted into a box, justified using Donald Knuth's linea breaking algorithm to stretch the spaces between words to fill the whole width. You can observe a variety of styles and text decorations applied, as well as support for LTR/RTL mixing and complex scripts. In the bottom-right the word "stroke" is being stroked and drawn as a path. Top-right we see a LaTeX formula that has been converted to a path. Left of that we see an ellipse showcasing precise dashing, notably the length of e.g. the short dash is equal wherever it is on the curve. Note that the dashes themselves are elliptical arcs as well (thus exactly precise even if magnified greatly). To the right we see a closed polygon of four points being smoothed by cubic Béziers that are smooth along the whole path, and the blue line on the left shows a smoothed open path. On the bottom you can see a rotated rasterized image. The bottom-left shows path boolean operations. The result is equivalent for all renderers (PNG, PDF, SVG, etc.).

### Sponsors

Please see https://www.patreon.com/tdewolff for ways to contribute, otherwise please contact me directly!

## Recent changes
- `Context` view and coordinate view have been altered. `View` now doesn't affect the coordinate view/system. To achieve the same as before, replace `ctx.SetView(m)` by `ctx.SetView(m); ctx.SetCoordView(m)`. The change makes coordinate systems more intuitive when using in combination with views, the given coordinate reflects the coordinate where it is drawn irrespective of the view.
- `Flatten()`, `Stroke()`, and `Offset()` now require an additional `tolerance` variable, which used to be set by the `Tolerance` parameter with a default value of `0.01`. To get the original behaviour, use `Flatten(0.01)`, `Stroke(width, capper, joiner, 0.01)`, and `Offset(width, fillRule, 0.01)`.
- `Interior()` is renamed to `Fills()`
- `ParseSVG` and `MustParseSVG` are now `ParseSVGPath` and `MustParseSVGPath` to avoid confusion that it parses entire SVGs

## Features
- Path segment types: MoveTo, LineTo, QuadTo, CubeTo, ArcTo, Close
- Precise path flattening, stroking, and dashing for all segment type uing papers (see below)
- Smooth spline generation through points for open and closed paths
- Path boolean operations: AND, OR, XOR, NOT, Divide
- LaTeX to path conversion (native Go and CGO implementations available)
- Font formats support 
- - SFNT (such as TTF, OTF, WOFF, WOFF2, EOT) supporting TrueType, CFF, and CFF2 tables
- HarfBuzz for text shaping (native Go and CGO implementations available)
- FriBidi for text bidirectionality (native Go and CGO implementations available)
- Donald Knuth's line breaking algorithm for text layout
- sRGB compliance (use `SRGBColorSpace`, only available for rasterizer)
- Font rendering with gamma correction of 1.43
- Rendering targets
- - Raster images (PNG, GIF, JPEG, TIFF, BMP, WEBP)
- - PDF
- - SVG and SVGZ
- - PS and EPS
- - HTMLCanvas
- - OpenGL
- - [Gio](https://gioui.org/)
- - [Fyne](https://fyne.io/)
- Rendering sources
- - Canvas itself
- - [go-chart](https://github.com/wcharczuk/go-chart)
- - [gonum/plot](https://github.com/gonum/plot)

## Examples
**[Street Map](https://github.com/tdewolff/canvas/tree/master/examples/map)**: the centre of Amsterdam is drawn from data loaded from the Open Street Map API.

**[Mauna-Loa CO2 concentration](https://github.com/tdewolff/canvas/tree/master/examples/graph)**: using data from the Mauna-Loa observatory, carbon dioxide concentrations over time are drawn

**[PDF document](https://github.com/tdewolff/canvas/tree/master/examples/document)**: an example of a text document using the PDF backend.

**[OpenGL](https://github.com/tdewolff/canvas/tree/master/examples/opengl)**: an example using the OpenGL backend.

**[Gio](https://github.com/tdewolff/canvas/tree/master/examples/gio)**: an example using the Gio backend.

**[Fyne](https://github.com/tdewolff/canvas/tree/master/examples/fyne)**: an example using the Fyne backend.

**[TeX/PGF](https://github.com/tdewolff/canvas/tree/master/examples/tex)**: an example showing the usage of the PGF (TikZ) LaTeX package as renderer in order to generated a PDF using LaTeX.

**[go-chart](https://github.com/tdewolff/canvas/tree/master/examples/go-chart)**: an example using the [go-chart](https://github.com/wcharczuk/go-chart) library, plotting a financial graph.

**[gonum/plot](https://github.com/tdewolff/canvas/tree/master/examples/gonum-plot)**: an example using the [gonum/plot](https://github.com/gonum/plot) library.

**[HTMLCanvas](https://github.com/tdewolff/canvas/tree/master/examples/html-canvas)**: an example using the HTMLCanvas backend, see the [live demo](https://tdewolff.github.io/canvas/examples/html-canvas/index.html).

## Users
This is a non-exhaustive list of library users I've come across. PRs are welcome to extend the list!

* https://github.com/aldernero/sketchy
* https://github.com/davidhampgonsalves/quickdraw
* https://github.com/engelsjk/go-annular
* https://github.com/fullstack-lang/gongdoc
* https://github.com/jansorg/marketplace-stats
* https://github.com/kpym/marianne
* https://github.com/namsor/go-qrcode
* https://github.com/peteraba/roadmapper
* https://github.com/stv0g/vand
* https://github.com/uncopied/chirograph
* https://github.com/winkula/dragons

## Articles
* [Numerically stable quadratic formula](https://math.stackexchange.com/questions/866331/numerically-stable-algorithm-for-solving-the-quadratic-equation-when-a-is-very/2007723#2007723)
* [Quadratic Bézier length](https://malczak.linuxpl.com/blog/quadratic-bezier-curve-length/)
* [Bézier spline through open path](https://www.particleincell.com/2012/bezier-splines/)
* [Bézier spline through closed path](http://www.jacos.nl/jacos_html/spline/circular/index.html)
* [Point inclusion in polygon test](https://wrf.ecse.rpi.edu/Research/Short_Notes/pnpoly.html)

#### My own

* [Arc length parametrization](https://tacodewolff.nl/posts/20190525-arc-length/)

#### Papers

* [M. Walter, A. Fournier, Approximate Arc Length Parametrization, Anais do IX SIBGRAPHI (1996), p. 143--150](https://www.visgraf.impa.br/sibgrapi96/trabs/pdf/a14.pdf)
* [T.F. Hain, et al., Fast, precise flattening of cubic Bézier path and offset curves, Computers & Graphics 29 (2005). p. 656--666](https://doi.org/10.1016/j.cag.2005.08.002)
* [M. Goldapp, Approximation of circular arcs by cubic polynomials, Computer Aided Geometric Design 8 (1991), p. 227--238](https://doi.org/10.1016/0167-8396%2891%2990007-X)
* [L. Maisonobe, Drawing and elliptical arc using polylines, quadratic or cubic Bézier curves (2003)](https://spaceroots.org/documents/ellipse/elliptical-arc.pdf)
* [S.H. Kim and Y.J. Ahn, An approximation of circular arcs by quartic Bezier curves, Computer-Aided Design 39 (2007, p. 490--493)](https://doi.org/10.1016/j.cad.2007.01.004)
* [D.E. Knuth and M.F. Plass, Breaking Paragraphs into Lines, Software: Practive and Experience 11 (1981), p. 1119--1184]()

## License
Released under the [MIT license](LICENSE.md).

Be aware that Fribidi uses the LGPL license.
