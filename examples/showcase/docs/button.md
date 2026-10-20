# Button · 按钮

按钮用于触发保存、提交、取消等操作。普通按钮直接传文字；图标或组合内容通过 `Content` 提供。

## 基本用法

```go
import "github.com/qianniancn/flowui/widget/button"

button.New("保存").OnClick(func() { send(Save{}) })
```

## 构造函数

```go
button.New(label) // 返回 button.Widget
```

| 参数 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `label` | `string` | 必填 | 按钮文字，同时作为无障碍名称。普通按钮直接传字符串。 |

需要显式身份时（列表、动态重排）再链式调用 `.Key("...")`；默认身份由构造位置自动推导。

链式方法返回配置后的新值；把返回值放入布局。点击回调、状态和业务数据由应用提供。

## 常见用法

### 加载与禁用

把模型状态传给按钮，避免重复提交。

```go
button.New("提交").
    Loading(m.Saving).
    Disabled(!m.Valid).
    OnClick(func() { send(Save{}) })
```

### 纯图标与组合内容

`IconOnly` 负责尺寸；构造时的文字仍用于无障碍名称。图标包与本项目的 `icon` 组件配合使用。

```go
// import "github.com/qianniancn/flowui-icons-lucide"
// import "github.com/qianniancn/flowui/widget/icon"
button.New("添加项目").
    Content(icon.New(lucide.Plus)).IconOnly().
    OnClick(addProject)

button.New("保存").Content(
    layout.Row(icon.New(lucide.Save).Size(16), text.New("保存")).Gap(8),
).OnClick(save)
```

### 宽度与局部样式

`FullWidth` 占满可用宽度。需要修改单个按钮外观时，先配置行为，再用 `.Style(...)` 追加外观。

```go
button.New("继续").
    FullWidth().OnClick(next)

button.New("保存").OnClick(save).Style(style.New().Radius(8).PaddingX(20))
```

## 参数与方法

| 参数 / 方法 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `Variant(value)` | `button.Variant` | `Primary` | 设置操作的视觉层级，支持下方列出的 8 种变体。 |
| `Size(value)` | `button.Size` | `Medium` | Small / Medium / Large；尺寸由当前主题定义。 |
| `Disabled(value)` | `bool` | `false` | 禁用操作；不会触发 OnClick，也不能通过键盘激活。 |
| `Loading(value)` | `bool` | `false` | 显示加载指示器并阻止点击；应用负责开启和结束加载状态。 |
| `FullWidth()` | 无参数 | 未开启 | 占满父布局允许的宽度；默认按内容和主题内边距确定宽度。 |
| `IconOnly()` | 无参数 | 未开启 | 使用正方形图标按钮尺寸；图标内容仍需通过 Content 提供。 |
| `Content(value)` | `ui.Widget` | label 文本 | 替换可见内容，可传图标或组合布局；保留原有无障碍名称。 |
| `Label(value)` | `string` | 构造时的 label | 只修改无障碍名称，不改变当前可见文字或内容。 |

## 变体与尺寸

| 变体 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `Primary` | `button.Variant` | 默认变体 | 主色实心，适合保存、提交等主要操作。 |
| `Secondary` | `button.Variant` | — | 中性底色，适合次要操作。 |
| `Tertiary` | `button.Variant` | — | 更弱的底色，适合低强调操作。 |
| `Outline` | `button.Variant` | — | 描边按钮，适合取消或并列的次要操作。 |
| `Ghost` | `button.Variant` | — | 轻量按钮，适合工具栏和辅助操作。 |
| `Danger` | `button.Variant` | — | 危险色实心，适合删除等破坏性操作。 |
| `DangerSoft` | `button.Variant` | — | 危险色的柔和底色版本。 |
| `DangerOutline` | `button.Variant` | — | 危险色的描边版本。 |

默认主题高度：**Small 24dp、Medium 32dp、Large 40dp**。`IconOnly` 的宽高相等；自定义主题和实例样式可调整这些尺寸。

## 点击事件与键盘

| 方法 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `OnClick(fn)` | `func()` | `nil` | 鼠标、触摸或键盘激活时调用；禁用或加载期间不调用。未设置回调时不执行业务操作。 |

```go
button.New("保存").
    OnClick(func() { send(Save{}) })
```

Tab / Shift+Tab 移动焦点，获得焦点后可用 Enter 或空格激活。纯图标按钮也应提供明确的操作名称。

`OnClick` 在界面更新过程中同步调用。耗时操作交给应用的异步任务，通过消息更新模型，再把新状态传给 `Loading` / `Disabled`。

## 按钮组 API

```go
button.Group(
    button.New("上一步").OnClick(previous),
    button.New("下一步").OnClick(next),
).Variant(button.Outline).
    Size(button.Small).
    Separators(true)
```

| 参数 / 方法 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `Group(buttons...)` | `...button.Widget` | 空组 | 组合多个按钮；每个按钮仍需唯一 key 和自己的回调。 |
| `Orientation(value)` | `button.GroupOrientation` | `GroupHorizontal` | GroupHorizontal 横向排列；GroupVertical 纵向排列。 |
| `Variant(value)` | `button.Variant` | `Primary` | 设置组内按钮的默认变体。 |
| `Size(value)` | `button.Size` | `Medium` | 设置组内按钮的默认尺寸。 |
| `Disabled(value)` | `bool` | `false` | 设置组内按钮的默认禁用状态。 |
| `FullWidth()` | 无参数 | 未开启 | 占满可用宽度，横向排列时按钮等分宽度。 |
| `Separators(value)` | `bool` | `false` | 显示相邻按钮之间的分隔线。 |

单个按钮显式设置的 `Variant`、`Size`、`Disabled` 优先于组设置。例如 `Disabled(false)` 可以覆盖组的 `Disabled(true)`。按钮组表示一组操作；互斥选择请使用 Segmented 或 RadioGroup。

## 使用注意

- `key` 标识按钮本身，不要用点击次数或经常变化的文字作为 key。列表中优先使用记录的稳定 ID。
- `Loading` 不会自动启动、结束或执行任务；即使没有 `OnClick`，也可以展示加载状态。
- `Content` 只替换显示内容；如果操作含义也变了，再调用 `Label` 更新无障碍名称。
- 普通按钮文字不需要 `text.New`；只有图标与文字等组合内容才需要 `Content`。
- 本页代码中的 `m`、`send`、`Save`、回调和模型字段需由应用定义；按钮不负责业务状态管理。

