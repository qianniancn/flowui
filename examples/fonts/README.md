# FlowUI Font Example

[简体中文](README.zh-CN.md)

This example shows how to bundle a font with a FlowUI application and how to
build smaller font files for a known set of characters. The sample uses Source
Han Sans SC and currently embeds only the Regular face. The Medium and Bold
embedding code remains commented in `main.go` for weight comparisons. The
subsetting tool is independent of that font family and can process fonts
supplied by the application author.

## Run the example

From the repository root:

```bash
go run ./examples/fonts
```

The user interface is in English. It also includes Chinese sample text so that
CJK coverage can be checked when a bundled font is selected.

When the system font source is selected on Windows, the example prefers
Microsoft YaHei, followed by Segoe UI and the generic sans-serif fallback.
Other platforms keep the generic system font families.

## Bundled fonts

The original font files are stored in `font/`. The current build embeds the
Regular subset from `font/subset/`:

- `font/subset/SourceHanSansSC-Regular-Subset.otf`

The Medium and Bold subset files remain in the directory, but their embed and
parsing code is commented out while the Regular-only memory path is evaluated.

For a production application, Regular is often sufficient. `font.Weight` can
still be requested for headings and controls, but if the exact Medium or Bold
design is important, bundle the corresponding face as well. A missing face is
matched to the closest available weight; it is not a substitute for a real
bold font.

To use a different family, replace the `go:embed` paths and the font collection
setup in `main.go`.

The sample files come from Adobe's official [Source Han Sans repository](https://github.com/adobe-fonts/source-han-sans),
release `2.005R`:

- [Regular OTF](https://github.com/adobe-fonts/source-han-sans/raw/2.005R/OTF/SimplifiedChinese/SourceHanSansSC-Regular.otf)
- [Medium OTF](https://github.com/adobe-fonts/source-han-sans/raw/2.005R/OTF/SimplifiedChinese/SourceHanSansSC-Medium.otf)
- [Bold OTF](https://github.com/adobe-fonts/source-han-sans/raw/2.005R/OTF/SimplifiedChinese/SourceHanSansSC-Bold.otf)

Download the three files into `font/` with PowerShell:

```powershell
$base = "https://github.com/adobe-fonts/source-han-sans/raw/2.005R/OTF/SimplifiedChinese"
Invoke-WebRequest "$base/SourceHanSansSC-Regular.otf" -OutFile font/SourceHanSansSC-Regular.otf
Invoke-WebRequest "$base/SourceHanSansSC-Medium.otf" -OutFile font/SourceHanSansSC-Medium.otf
Invoke-WebRequest "$base/SourceHanSansSC-Bold.otf" -OutFile font/SourceHanSansSC-Bold.otf
```

The fonts are distributed under the [SIL Open Font License 1.1](https://openfontlicense.org/).
Read the upstream [license](https://github.com/adobe-fonts/source-han-sans/blob/release/LICENSE.txt)
before redistributing them.

## Create a subset

`subset_fonts.py` is a generic wrapper around FontTools. It does not contain a
built-in character set and does not assume a language or font family. Every run
requires one or more UTF-8 text files and one or more input font paths or glob
patterns. Repeat `--text-file` to merge language-specific character lists.

Run the following commands from `examples/fonts`:

Install the dependency with the same Python interpreter that will run the
script:

```powershell
python -m pip install fonttools
```

The first-level list from the Common Standard Chinese Characters Table contains
3,500 Chinese characters:

- Project: https://github.com/shengdoushi/common-standard-chinese-characters-table
- Plain text: https://raw.githubusercontent.com/shengdoushi/common-standard-chinese-characters-table/master/level-1.txt

```powershell
Invoke-WebRequest `
  -Uri "https://raw.githubusercontent.com/shengdoushi/common-standard-chinese-characters-table/master/level-1.txt" `
  -OutFile level-1.txt
```

That list contains Chinese characters only. `charset-latin.txt` contains the
Latin alphabet, digits, ASCII punctuation, common CJK punctuation, and selected
symbols used by the sample. Keep additional languages in separate UTF-8 files
and pass each file with another `--text-file` option.

Generate subsets for the sample fonts:

```powershell
python subset_fonts.py `
  --text-file level-1.txt `
  --text-file charset-latin.txt `
  --output-dir font/subset `
  font/SourceHanSansSC-Regular.otf `
  font/SourceHanSansSC-Medium.otf `
  font/SourceHanSansSC-Bold.otf
```

Process another directory of TrueType fonts in the same way:

```powershell
python subset_fonts.py `
  --text-file level-1.txt `
  --text-file charset-latin.txt `
  --output-dir output fonts/*.ttf
```

The original files are never modified. Generated files use the original name
with `-Subset` appended before the extension. Hinting tables are removed by
default; pass `--keep-hinting` when they are required by the target platform.
Existing output files are replaced by FontTools, so choose the output directory
carefully.

With the local `level-1.txt` and `charset-latin.txt`, the generated sample
files are written to `font/subset/` and contain the 3,500 first-level Chinese
characters plus the Latin and symbol set. Add the application's other
characters before generating a production subset.

## Configure the theme

Parse the generated font and add the returned faces to the theme before starting
the application. The sample keeps system fallback enabled for characters outside
the checked-in Chinese, Latin, and symbol set:

```go
faces, err := ui.ParseFontCollection(fontData)
if err != nil {
    return err
}

theme := ui.DefaultTheme()
theme.Fonts.Collection = faces
theme.Fonts.SystemFonts = true
```

Keeping system fallback enabled lets the platform provide characters that are
not in the subset, such as uncommon ideographs, Emoji, or another language. If
reproducible rendering is more important than fallback coverage, disable it
and provide every required font face yourself. Always check the font license
before redistributing a bundled or subsetted file.
