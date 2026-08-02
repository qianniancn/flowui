# FlowUI 字体示例

[English](README.md)

这个示例演示如何在 FlowUI 应用中嵌入字体，以及如何根据实际使用的字符
生成更小的字体文件。示例使用思源黑体简体中文，目前只嵌入 Regular 字重。
Medium 和 Bold 的嵌入代码仍保留在 `main.go` 中，但暂时注释掉，方便后续比较
字重和内存占用。子集化工具不绑定具体字体或语言，也可以处理应用作者提供的其他字体。

## 运行示例

在仓库根目录执行：

```bash
go run ./examples/fonts
```

示例界面使用英文，同时保留了中文文本作为 CJK 字形覆盖测试。切换字体来源
后，可以比较内置字体与系统字体回退的显示效果。

在 Windows 上切换到系统字体时，示例会优先使用 Microsoft YaHei（微软雅黑），
然后回退到 Segoe UI 和通用的无衬线字体。其他平台继续使用系统通用字体族名。

## 内置字体

原始字体文件位于 `font/` 目录。当前构建只通过 `go:embed` 嵌入
`font/subset/` 中的 Regular 子集字体：

- `font/subset/SourceHanSansSC-Regular-Subset.otf`

Medium 和 Bold 子集文件仍保留在目录中，但对应的嵌入和解析代码在评估
Regular-only 内存路径期间暂时注释掉。

在生产应用中，通常只需要 Regular。标题和控件仍然可以请求
`font.Weight`，但如果需要准确的 Medium 或 Bold 字形，应同时嵌入对应字重。
缺少目标字重时，Gio 会匹配最接近的可用字重，但这不能替代真正的粗体字体。

使用其他字体时，替换 `main.go` 中的 `go:embed` 路径和字体集合配置即可。

示例中的字体来自 Adobe 官方的
[Source Han Sans 仓库](https://github.com/adobe-fonts/source-han-sans)，固定使用
`2.005R` 版本：

- [Regular OTF](https://github.com/adobe-fonts/source-han-sans/raw/2.005R/OTF/SimplifiedChinese/SourceHanSansSC-Regular.otf)
- [Medium OTF](https://github.com/adobe-fonts/source-han-sans/raw/2.005R/OTF/SimplifiedChinese/SourceHanSansSC-Medium.otf)
- [Bold OTF](https://github.com/adobe-fonts/source-han-sans/raw/2.005R/OTF/SimplifiedChinese/SourceHanSansSC-Bold.otf)

使用 PowerShell 将三个文件下载到 `font/` 目录：

```powershell
$base = "https://github.com/adobe-fonts/source-han-sans/raw/2.005R/OTF/SimplifiedChinese"
Invoke-WebRequest "$base/SourceHanSansSC-Regular.otf" -OutFile font/SourceHanSansSC-Regular.otf
Invoke-WebRequest "$base/SourceHanSansSC-Medium.otf" -OutFile font/SourceHanSansSC-Medium.otf
Invoke-WebRequest "$base/SourceHanSansSC-Bold.otf" -OutFile font/SourceHanSansSC-Bold.otf
```

字体采用 [SIL Open Font License 1.1](https://openfontlicense.org/) 发布。
在重新分发字体前，请阅读上游的
[许可证文件](https://github.com/adobe-fonts/source-han-sans/blob/release/LICENSE.txt)。

## 生成字体子集

`subset_fonts.py` 是对 FontTools 的通用封装。它不内置字符集，也不假设字体
家族或语言；每次运行都需要提供一个或多个 UTF-8 字符文件，以及一个或多个
字体路径或通配符。重复使用 `--text-file` 可以合并不同语言的字符表。

下面的命令都需要在 `examples/fonts` 目录中执行：

请使用运行脚本的同一个 Python 解释器安装依赖：

```powershell
python -m pip install fonttools
```

《通用规范汉字表》一级字表包含 3500 个常用汉字：

- 项目：https://github.com/shengdoushi/common-standard-chinese-characters-table
- 纯文本：https://raw.githubusercontent.com/shengdoushi/common-standard-chinese-characters-table/master/level-1.txt

```powershell
Invoke-WebRequest `
  -Uri "https://raw.githubusercontent.com/shengdoushi/common-standard-chinese-characters-table/master/level-1.txt" `
  -OutFile level-1.txt
```

该字表只有中文字符。`charset-latin.txt` 包含示例所需的英文字母、数字、ASCII
标点、常用中文标点和部分符号。其他语言可以单独保存为 UTF-8 文件，再通过另一个
`--text-file` 传入，不需要修改 Python 工具。

为示例中的字体生成子集：

```powershell
python subset_fonts.py `
  --text-file level-1.txt `
  --text-file charset-latin.txt `
  --output-dir font/subset `
  font/SourceHanSansSC-Regular.otf `
  font/SourceHanSansSC-Medium.otf `
  font/SourceHanSansSC-Bold.otf
```

也可以处理其他目录中的 TrueType 字体：

```powershell
python subset_fonts.py `
  --text-file level-1.txt `
  --text-file charset-latin.txt `
  --output-dir output fonts/*.ttf
```

原始字体文件不会被修改。生成文件会在扩展名前追加 `-Subset`。默认会移除
hinting 表以减小文件体积；如果目标平台需要保留 hinting，请添加
`--keep-hinting`。

如果输出文件已经存在，FontTools 会直接替换它，因此请谨慎选择输出目录。

使用当前目录中的 `level-1.txt` 和 `charset-latin.txt` 生成后，示例字体会写入
`font/subset/`，每个文件包含 3500 个一级常用汉字以及示例所需的英文和符号。
正式发布前，请通过额外的字符文件补充应用实际使用的其他字符。

## 配置主题

在启动应用前解析生成的字体，并将返回的字体面加入主题。由于当前字符表
包含中文、英文和部分符号，示例仍保留系统字体回退，以显示字符集之外的生僻字、
Emoji 和其他语言字符：

```go
faces, err := ui.ParseFontCollection(fontData)
if err != nil {
    return err
}

theme := ui.DefaultTheme()
theme.Fonts.Collection = faces
theme.Fonts.SystemFonts = true
```

建议保留系统字体回退。这样，子集未覆盖的生僻字、Emoji 或其他语言字符，
可以由操作系统字体补齐。如果更重视跨平台渲染一致性，则可以关闭系统回退，
并自行提供应用所需的全部字体。发布内置字体或子集字体前，请确认字体许可证
允许这样使用和再分发。
