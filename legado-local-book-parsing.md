# Legado 本地书籍格式解析规则梳理

本文档基于当前 `legado` 源码整理，关注本地书籍格式解析：`txt`、`epub`、`pdf`、`mobi`、`azw3`、`azw`、`umd`，以及压缩包导入。目的是给后续自研阅读器复用解析思路、规则和数据结构。

## Web 阅读器落地归属

在当前 Web 阅读器架构中，本文档整理的 Legado 本地书籍解析规则主要应落在后端，而不是前端。

后端负责：

- 文件鉴权、文件流读取和真实路径隔离。
- 书籍元数据抽取，例如书名、作者、简介、封面。
- 章节列表生成和章节定位信息保存。
- 章节正文读取和正文归一化。
- 图片、封面、内嵌资源读取。
- TXT 编码识别、目录正则选择、字节偏移分章。
- EPUB 目录解析、fragment 截取、跨 XHTML 拼接和图片资源解析。
- MOBI / AZW3 / AZW / UMD 等浏览器端支持较弱格式的解析。

前端负责：

- 阅读器 UI、目录面板、翻页或滚动交互。
- 字号、主题、行距等阅读设置。
- 请求后端章节列表、章节正文和资源图片。
- 节流上报阅读进度。

不建议把这些解析规则主要放在前端，原因是：

- 大文件解析容易造成浏览器内存压力。
- TXT 解析依赖字节偏移、编码识别和分块扫描，后端更可靠。
- MOBI / AZW3 / UMD 的浏览器生态弱，不适合放前端实现。
- 公共图书只存一份文件，解析结果应可被多个用户共享。
- 图书文件必须经过后端鉴权，不应暴露真实文件路径。

推荐渐进式落地：

- 第一期：TXT 解析放后端；EPUB 可以先由前端 `epub.js` 阅读；PDF 可以先由前端 `pdf.js` 渲染。
- 第二期：EPUB 元数据、目录、章节正文和图片资源迁移到后端解析。
- 第三期：MOBI / AZW3 / UMD 放后端解析；PDF 如采用 Legado 图片式阅读，再迁移为后端分页和页面图片渲染。

建议后端模块位置：

```text
backend/internal/parser/
  parser.go
  txt/
  epub/
  pdf/
  mobi/
  umd/
  htmlfmt/
```

建议统一接口使用 Go 风格抽象：

```go
type LocalBookParser interface {
    ReadMetadata(file BookFile) (*BookMetadata, error)
    ReadChapters(file BookFile) ([]Chapter, error)
    ReadContent(file BookFile, chapter Chapter) (*Content, error)
    ReadResource(file BookFile, href string) (io.ReadCloser, string, error)
}
```

前端不直接解析或读取真实文件路径，而是通过阅读 API 消费后端解析结果：

```text
GET /api/v1/reader/books/:bookId/chapters
GET /api/v1/reader/books/:bookId/chapters/:chapterId/content
GET /api/v1/reader/books/:bookId/resources
```

## 核心入口

本地书籍解析入口在 `legado/app/src/main/java/io/legado/app/model/localBook/LocalBook.kt`。

统一接口在 `BaseLocalBookParse.kt`：

- `upBookInfo(book)`：读取或更新书籍元数据，如书名、作者、简介、封面。
- `getChapterList(book)`：生成章节列表。
- `getContent(book, chapter)`：读取章节正文。
- `getImage(book, href)`：按正文里的图片引用读取图片流。

`LocalBook.getChapterList()` 和 `LocalBook.getContent()` 按扩展名分派：

- `.epub` 使用 `EpubFile`。
- `.umd` 使用 `UmdFile`。
- `.pdf` 使用 `PdfFile`。
- `.mobi`、`.azw3`、`.azw` 使用 `MobiFile`。
- 其他本地文本默认使用 `TextFile`，实际主要是 `.txt`。

支持的本地书籍扩展名正则在 `AppPattern.bookFileRegex`：

```regex
.*\.(txt|epub|umd|pdf|mobi|azw3|azw)
```

压缩包支持类型在 `AppPattern.archiveFileRegex`：

```regex
.*\.(zip|rar|7z)$
```

导入压缩包时，`LocalBook.importArchiveFile()` 会解压后筛选符合 `bookFileRegex` 的条目，再逐个导入为本地书籍。若解压出的临时文件丢失，后续会根据原压缩包重新导入匹配 `originName` 的文件。

## 通用数据结构

解析器最终都把不同格式归一到 `Book` 和 `BookChapter`。

`Book` 关键字段：

- `bookUrl`：本地文件 URI 或路径，是书籍唯一地址。
- `originName`：导入时的原始文件名，用于判断格式。
- `name`、`author`、`intro`：书籍元数据。
- `coverUrl`：封面缓存路径。
- `charset`：TXT 编码。
- `tocUrl`：TXT 自动选出的目录规则，格式为 `正则 + spaceChars + replacement`。
- `wordCount`、`totalChapterNum`、`latestChapterTitle`：字数和目录状态。

`BookChapter` 关键字段：

- `index`：章节序号。
- `title`：章节标题。
- `url`：章节定位标识。EPUB 是 href，TXT 是生成的 MD5，PDF 是 `pdf_0`，MOBI 是 `index:href`。
- `start`、`end`：TXT 使用的字节偏移范围。
- `startFragmentId`、`endFragmentId`：EPUB 使用的锚点范围。
- `isVolume`：是否是卷标题或空正文目录节点。
- `wordCount`：章节字数。
- 变量 `nextUrl`：EPUB/MOBI 用来知道当前章节到哪个资源或 section 截止。

章节列表生成后，`LocalBook.getChapterList()` 会统一去重、重排 `index`、填充空标题为 `无标题章节`，并更新书籍当前章节、最新章节、章节总数和更新时间。

## TXT 解析

核心代码：`model/localBook/TextFile.kt`。

### 编码识别

TXT 首次解析目录时会读取文件前 `512000` 字节。

- 如果 `book.charset` 为空，或文件被判定已修改，调用 `EncodingDetect.getEncode()` 自动识别编码。
- `Book.fileCharset()` 默认回退到 `UTF-8`。
- 解析时检测 UTF-8 BOM。若前三个字节是 BOM，正文偏移从 `3` 开始。
- 读取内容时按最终 `charset` 把字节转成字符串。

### 目录规则来源

TXT 目录规则来自数据库 `txtTocRules`。如果数据库为空，会导入 `assets/defaultData/txtTocRule.json`。

`TxtTocRule` 字段：

- `name`：规则名。
- `rule`：章节匹配正则。
- `replacement`：可选 JS，用于把匹配到的原始标题改写成显示标题。
- `enable`：是否启用。
- `serialNumber`：排序。

默认启用的规则覆盖这些常见格式：

- `目录(去空白)`：匹配前面有空白的 `序章`、`楔子`、`正文`、`终章`、`后记`、`尾声`、`番外`、`第...章/节/卷/集`。
- `目录`：匹配行首 0 到 4 个空格或制表符后的标准章节标题。
- `数字 分隔符 标题名称`：匹配 `1、标题`、`12: 标题`。
- `大写数字 分隔符 标题名称`：匹配 `一、标题`、`二十四章 标题`。
- `正文 标题/序号`：匹配 `正文 标题`。
- `Chapter/Section/Part/Episode 序号 标题`：匹配英文章节格式。
- `特殊符号 序号 标题`：匹配 `【第一章 标题`。
- `特殊符号 标题(单个)`：匹配 `☆标题`、`★标题`。
- `章/卷 序号 标题`：匹配 `卷五 标题`、`章12 标题`。
- `书名 括号 序号`：匹配 `书名(12)`。
- `书名 序号`：匹配 `书名 12`。
- `字数分割 分节阅读`：匹配 `分节阅读`、`第一页`。

默认未启用但可备用：匹配简介、古典/轻小说备用、纯数字标题、顶格标题、双标题前向/后向、激进通用规则、空正则兜底规则。

### 可直接移植的 TXT 默认规则

下面是 `assets/defaultData/txtTocRule.json` 的默认目录规则。另一个实现可以直接把它们作为初始规则集导入。`enable=true` 的规则参与自动识别；`enable=false` 的规则默认不参与，但应允许用户手动启用。

| 顺序 | 默认启用 | 名称 | 正则 |
| --- | --- | --- | --- |
| 0 | 是 | 目录(去空白) | `(?<=[　\s])(?:序章|楔子|正文(?!完|结)|终章|后记|尾声|番外|第\s{0,4}[\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]+?\s{0,4}(?:章|节(?!课)|卷|集(?![合和]))).{0,30}$` |
| 1 | 是 | 目录 | `^[ 　\t]{0,4}(?:序章|楔子|正文(?!完|结)|终章|后记|尾声|番外|第\s{0,4}[\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]+?\s{0,4}(?:章|节(?!课)|卷|集(?![合和])|部(?![分赛游])|篇(?!张))).{0,30}$` |
| 2 | 否 | 目录(匹配简介) | `(?<=[　\s])(?:(?:内容|文章)?简介|文案|前言|序章|楔子|正文(?!完|结)|终章|后记|尾声|番外|第\s{0,4}[\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]+?\s{0,4}(?:章|节(?!课)|卷|集(?![合和])|部(?![分赛游])|回(?![合来事去])|场(?![和合比电是])|篇(?!张))).{0,30}$` |
| 3 | 否 | 目录(古典、轻小说备用) | `^[ 　\t]{0,4}(?:序章|楔子|正文(?!完|结)|终章|后记|尾声|番外|第\s{0,4}[\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]+?\s{0,4}(?:章|节(?!课)|卷|集(?![合和])|部(?![分赛游])|回(?![合来事去])|场(?![和合比电是])|话|篇(?!张))).{0,30}$` |
| 4 | 否 | 数字(纯数字标题) | `(?<=[　\s])\d+\.?[ 　\t]{0,4}$` |
| 5 | 否 | 大写数字(纯数字标题) | `(?<=[　\s])[零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,12}[ 　\t]{0,4}$` |
| 6 | 否 | 数字混合(纯数字标题) | `(?<=[　\s])[零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟\d]{1,12}[ 　\t]{0,4}$` |
| 7 | 是 | 数字 分隔符 标题名称 | `^[ 　\t]{0,4}\d{1,5}[:：,.， 、_—\-].{1,30}$` |
| 8 | 是 | 大写数字 分隔符 标题名称 | `^[ 　\t]{0,4}(?:序章|楔子|正文(?!完|结)|终章|后记|尾声|番外|[零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,8}章?)[ 、_—\-].{1,30}$` |
| 9 | 否 | 数字混合 分隔符 标题名称 | `^[ 　\t]{0,4}(?:序章|楔子|正文(?!完|结)|终章|后记|尾声|番外|[零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,8}章?[ 、_—\-]|\d{1,5}章?[:：,.， 、_—\-]).{0,30}$` |
| 10 | 是 | 正文 标题/序号 | `^[ 　\t]{0,4}正文[ 　]{1,4}.{0,20}$` |
| 11 | 是 | Chapter/Section/Part/Episode 序号 标题 | `^[ 　\t]{0,4}(?:[Cc]hapter|[Ss]ection|[Pp]art|ＰＡＲＴ|[Nn][oO][.、]|[Ee]pisode|(?:内容|文章)?简介|文案|前言|序章|楔子|正文(?!完|结)|终章|后记|尾声|番外)\s{0,4}\d{1,4}.{0,30}$` |
| 12 | 否 | Chapter(去简介) | `^[ 　\t]{0,4}(?:[Cc]hapter|[Ss]ection|[Pp]art|ＰＡＲＴ|[Nn][Oo]\.|[Ee]pisode)\s{0,4}\d{1,4}.{0,30}$` |
| 13 | 是 | 特殊符号 序号 标题 | `(?<=[\s　])[【〔〖「『〈［\[](?:第|[Cc]hapter)[\d零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,10}[章节].{0,20}$` |
| 14 | 否 | 特殊符号 标题(成对) | `(?<=[\s　]{0,4})(?:[\[〈「『〖〔《（【\(].{1,30}[\)】）》〕〗』」〉\]]?|(?:内容|文章)?简介|文案|前言|序章|楔子|正文(?!完|结)|终章|后记|尾声|番外)[ 　]{0,4}$` |
| 15 | 是 | 特殊符号 标题(单个) | `(?<=[\s　]{0,4})(?:[☆★✦✧].{1,30}|(?:内容|文章)?简介|文案|前言|序章|楔子|正文(?!完|结)|终章|后记|尾声|番外)[ 　]{0,4}$` |
| 16 | 是 | 章/卷 序号 标题 | `^[ \t　]{0,4}(?:(?:内容|文章)?简介|文案|前言|序章|楔子|正文(?!完|结)|终章|后记|尾声|番外|[卷章][\d零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,8})[ 　]{0,4}.{0,30}$` |
| 17 | 否 | 顶格标题 | `^\S.{1,20}$` |
| 18 | 否 | 双标题(前向) | `(?m)(?<=[ \t　]{0,4})第[\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,8}章.{0,30}$(?=[\s　]{0,8}第[\d零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,8}章)` |
| 19 | 否 | 双标题(后向) | `(?m)(?<=[ \t　]{0,4}第[\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,8}章.{0,30}$[\s　]{0,8})第[\d零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,8}章.{0,30}$` |
| 20 | 是 | 书名 括号 序号 | `^[一-龥]{1,20}[ 　\t]{0,4}[(（][\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,8}[)）][ 　\t]{0,4}$` |
| 21 | 是 | 书名 序号 | `^[一-龥]{1,20}[ 　\t]{0,4}[\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]{1,8}[ 　\t]{0,4}$` |
| 22 | 否 | 特定字符 标题 特定符号 | `(?<=\={3,6}).{1,40}?(?=\=)` |
| 23 | 是 | 字数分割 分节阅读 | `(?<=[ 　\t]{0,4})(?:.{0,15}分[页节章段]阅读[-_ ]|第\s{0,4}[\d零一二两三四五六七八九十百千万]{1,6}\s{0,4}[页节]).{0,30}$` |
| 24 | 否 | 通用规则 | `(?im)^.{0,6}(?:[引楔]子|正文(?!完|结)|[引序前]言|[序终]章|扉页|[上中下][部篇卷]|卷首语|后记|尾声|番外|={2,4}|第\s{0,4}[\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟]+?\s{0,4}(?:章|节(?!课)|卷|页[、 　]|集(?![合和])|部(?![分是门落])|篇(?!张))).{0,40}$|^.{0,6}[\d〇零一二两三四五六七八九十百千万壹贰叁肆伍陆柒捌玖拾佰仟a-z]{1,8}[、. 　].{0,20}$` |
| 99 | 否 | 默认分章规则 | 空字符串 |

这些默认规则的 `replacement` 均为空；标题默认使用正则命中的原文。若后续允许用户添加自定义规则，需要保留 `replacement` JS 扩展点。

### 自动选择目录规则

流程：

1. 读取前 `512000` 字节作为 `blockContent`。
2. 遍历所有启用规则，使用 `Pattern.MULTILINE` 编译正则。
3. 对每条规则统计可用匹配数 `csNum` 和疑似误匹配数 `numE`。
4. 只有 `csNum >= numE * 3`，且比当前最佳规则多超过 `overRuleCount = 2` 个匹配时，才替换最佳规则。
5. 如果最佳规则匹配数超过 `70`，提前停止。
6. 最终保存到 `book.tocUrl`，格式为 `rule + spaceChars + replacement`。

可直接实现的伪代码：

```text
bestRule = null
maxNum = -1
for rule in enabledRulesSortedBySerialNumber:
    pattern = compile(rule.rule, MULTILINE)
    matcher = pattern.findAll(firstBlockText)
    start = 0
    csNum = 0
    numE = 0
    lastTitle = null
    for match in matcher:
        contentLength = match.start - start
        if start == 0 or contentLength > 1000:
            title = applyReplacement(match.text, rule.replacement, csNum + 1, lastTitle, contentLength)
            if title is not empty:
                lastTitle = title
                csNum += 1
            start = match.end
        else if contentLength < 100:
            numE += 1
    if csNum >= numE * 3 and csNum > maxNum + 2:
        maxNum = csNum
        bestRule = rule
        if maxNum > 70:
            break
return bestRule
```

注意：`contentLength < 100` 的误匹配计数只在“两个疑似标题间正文太短”时发生，用于压低把卷名、短句、广告误识别为章节标题的规则。

### JS replacement

如果 `replacement` 为空，章节标题就是正则匹配到的原文。

如果不为空，会交给 Rhino 执行，JS 上下文提供：

- `result`：当前匹配到的标题文本。
- `book`：替换用书籍对象。
- `index`：当前章节序号，从 1 开始。
- `prevTitle`：前一个标题。
- `prevLength`：前一段内容长度。
- `lastVolumeTitle`：最近卷标题。
- `java.putVolume(title)`：JS 可主动插入卷标题。

JS 返回空字符串时，该目录命中会被跳过。

### 按目录规则分章

TXT 分章以字节偏移为准，不用字符串下标。

流程：

1. 用 `bufferSize = 512000` 分块读取文件。
2. 当前块刚好读满时，从块末尾往前找换行字节 `0x0a`，避免把一行标题截断到两个块里。
3. 把当前块按 `charset` 转成字符串，用 `Pattern.MULTILINE` 匹配标题。
4. 每次匹配到新标题，把上一个章节的 `end` 设为当前标题前的字节偏移。
5. 新章节的 `start` 设为标题结束后的字节偏移。
6. 首个标题前有内容时创建 `前言` 章节；如果前言过长，`book.intro` 只取前 600 字。
7. 两个标题之间正文为空时，把上一个章节标为 `isVolume = true`。
8. 章节 `wordCount` 按字符串长度格式化，整书 `wordCount` 也累计。

更接近源码的分章伪代码：

```text
toc = []
curOffset = 0
bufferStart = 3
read first 3 bytes
if hasUtf8Bom(first3):
    bufferStart = 0
    curOffset = 3

while read bytes into buffer[bufferStart..bufferSize):
    end = bufferStart + readLength
    if end == bufferSize:
        end = lastIndexOfLineFeed(buffer, from=end-1) or end
    blockText = decode(buffer[0..end], charset)
    moveRemainingBytesToHead(buffer, from=end, to=bufferStart + readLength)
    bufferStart = bufferStart + readLength - end
    length = end
    seekPos = 0
    for match in pattern.findAll(blockText):
        chapterStartInText = match.start
        chapterContent = blockText.substring(seekPos, chapterStartInText)
        chapterContentBytes = byteLength(chapterContent, charset)
        titleBytes = byteLength(match.text, charset)
        title = applyReplacement(match.text, replacementJs, toc.size + 1, previousTitle, chapterContent.length)
        if title is empty:
            continue
        if toc is empty and chapterStartInText != 0:
            add preface chapter from curOffset to curOffset + chapterContentBytes
            add new chapter title with start = curOffset + chapterContentBytes + titleBytes
        else:
            last = toc.lastOrNull()
            if last exists:
                last.end = calculated byte offset before current title
                if chapterContent is blank:
                    last.isVolume = true
            add new chapter title with start = last.end + titleBytes
        seekPos += chapterContent.length + match.text.length
    curOffset += length
    if toc has last:
        toc.last.end = curOffset
```

实现时要特别注意两套位置单位：正则匹配位置是字符串下标，章节 `start/end` 必须是文件字节偏移。源码通过 `chapterContent.toByteArray(charset).size` 和标题字节长度把字符串匹配结果换算成字节偏移。

长章节拆分：

- 使用目录规则时，单章字节长度超过 `maxLengthWithToc = 102400` 且 `book.getSplitLongChapter()` 为 true，会把原章节改成卷标题，再调用无规则拆分。
- 子章标题格式为 `原章节标题(1)`、`原章节标题(2)`。

### 无目录规则分章

如果没有选中目录规则，或 `book.tocUrl` 为空，会走无规则拆分。

- 每章最大 `maxLengthWithNoToc = 10 * 1024` 字节左右。
- 超过长度后向后查找换行符作为章末。
- 标题格式为 `第{blockPos}章({chapterPos})`。
- 最后一段小于等于 100 字节时合并到上一章，否则单独成章。

### 正文读取

`TextFile.getContent()` 按章节的 `start/end` 字节偏移读取：

- 内部维护一个 `txtBuffer`，大小 `8 * 1024 * 1024`。
- 如果章节范围不在缓存内，会重新定位到 `txtBufferSize * (start / txtBufferSize)` 读取一块。
- 如果章节跨缓存边界，直接从文件流跳到 `start` 后读取 `end-start` 字节。
- 字节转字符串后，把开头连续空白和换行替换成全角缩进 `　　`。

复用建议：TXT 应保存字节偏移，不要只保存字符串下标。否则多字节编码、BOM、分块读取时容易错位。

## EPUB 解析

核心代码：`model/localBook/EpubFile.kt`。

依赖 `me.ag2s.epublib`，并使用自定义 `AndroidZipFile` 支持 `ParcelFileDescriptor` 懒加载。

### 打开文件

EPUB 本质是 zip 包。Legado 通过 `BookHelp.getBookPFD(book)` 获取文件描述符，再调用 `EpubReader().readEpubLazy(zipFile, "utf-8")`。

### 元数据和封面

`upBookInfo()` 从 EPUB metadata 读取：

- `metadata.firstTitle` 作为书名。
- 书名为空时，使用文件名去掉 `.epub`。
- `metadata.authors[0]` 作为作者，并去掉首尾多余逗号。
- `metadata.descriptions[0]` 作为简介。如果描述是 XML/HTML，则用 Jsoup 提取纯文本。
- 封面优先使用 `epubBook.coverImage.inputStream`，解码后压缩为 JPEG，质量 `90`。

### 目录解析

优先使用 EPUB 的 `tableOfContents.tocReferences`。

- 如果 TOC 为空或解析失败，退回 `spine.spineReferences`。
- spine 回退模式下，章节 URL 是 resource href，标题优先用 resource title，其次解析 XHTML 的 `<title>`。
- 如果第一章前存在封面、引言、扉页等 HTML 内容，`parseFirstPage()` 会在正式 TOC 前补充卷首章节。
- 正常 TOC 模式下，递归解析 `TOCReference` 及 children。
- 每个章节记录 `ref.completeHref`、`ref.fragmentId`。
- 上一章会记录当前章为 `nextUrl`，上一章的 `endFragmentId` 也会设为当前章的 `startFragmentId`。
- 有 children 的章节会标为 `isVolume = true`。

### 正文截取

`EpubFile.getContent()` 根据当前章 href、下一章 href、起止 fragment 截取内容。

规则：

- `currentChapterFirstResourceHref = chapter.url.substringBeforeLast("#")`。
- `nextChapterFirstResourceHref = chapter.nextUrl.substringBeforeLast("#")`。
- 从 `epubBook.contents` 中找到当前资源开始收集。
- 如果章节跨多个 XHTML，会继续拼接后续资源，直到下一章资源。
- 如果同一 XHTML 内有多个章节，使用 `startFragmentId` 和 `endFragmentId` 截取片段。
- 特殊处理 `titlepage.xhtml` 或 href 包含 `cover` 的资源，直接转为 `<img src="cover.jpeg" />`。

HTML 清洗规则：

- 删除 `<script>`、`<style>`。
- 删除正文最终结果中的 `<title>`。
- 删除 `[style*=display:none]`。
- 多个 `cover.jpeg` 只保留第一个。
- EPUB 的 `<image xlink:href="...">` 改成 HTML `<img src="...">`。
- 图片相对路径用当前资源 href 解析成规范 href，并 URL decode。
- 如果用户设置删除 ruby 标签，则删除 `rp`、`rt`。
- 如果用户设置删除 H 标签，则删除 `h1` 到 `h6`。
- 最后调用 `HtmlFormatter.formatKeepImg()`，保留图片，其他 HTML 转成阅读器文本换行和缩进。

图片读取：

- `href == "cover.jpeg"` 返回 `epubBook.coverImage.inputStream`。
- 其他图片从 `epubBook.resources.getByHref(decodedHref)` 读取。

复用建议：EPUB 不要只按 TOC 节点读取单个文件。很多 EPUB 一个 XHTML 内含多章，必须支持 fragment 截取；也有一章跨多个 XHTML，必须支持拼接到下一章前。

## PDF 解析

核心代码：`model/localBook/PdfFile.kt`。

Legado 的 PDF 不是做文本抽取，而是把页面渲染为图片阅读。

### 打开文件

- `content://` URI 使用 `contentResolver.openFileDescriptor(uri, "r")`。
- 文件路径使用 `ParcelFileDescriptor.open(file, MODE_READ_ONLY)`。
- 通过 Android `PdfRenderer` 打开。

### 分章规则

- 常量 `PAGE_SIZE = 10`。
- 每 10 页作为一个章节。
- 章节数为 `ceil(pageCount / 10.0)`。
- 标题是 `分段_{index}`。
- URL 是 `pdf_{index}`。

### 正文规则

PDF 章节正文不是文本，而是图片占位 HTML：

```html
<img src="0" >
<img src="1" >
```

对第 `n` 个章节，页码范围是：

- `start = chapter.index * PAGE_SIZE`。
- `end = min((chapter.index + 1) * PAGE_SIZE, pageCount)`。

### 页面渲染

`getImage(href)` 把 `href` 当成页码 index：

- 打开 `PdfRenderer.Page`。
- 创建宽度为屏幕宽度的 bitmap。
- 高度按 `screenWidth * page.height / page.width` 等比例计算。
- 背景刷白。
- 使用 `RENDER_MODE_FOR_DISPLAY` 渲染。
- 转为 `InputStream` 返回。

封面使用第 0 页渲染后保存为 JPEG，质量 `90`。

复用建议：如果自研阅读器需要 PDF 文字选择、搜索、重排，需要额外接入 PDF 文本抽取库。Legado 当前规则只满足图片式阅读。

## MOBI / AZW3 / AZW 解析

核心代码：`model/localBook/MobiFile.kt` 和 `lib/mobi`。

`.mobi`、`.azw3`、`.azw` 都通过 `Book.isMobi` 归到 `MobiFile`。`MobiReader` 根据 MOBI header 判断是 KF6 还是 KF8：

- `mobi.version >= 8` 视为 KF8。
- 如果版本不是 8，会继续检查 EXTH 中的 KF8 boundary 信息，存在则按 KF8。
- KF8 对应 AZW3 的主要内容模型。

### 元数据和封面

`upBookInfo()`：

- 书名来自 `mobiBook.metadata.title`。
- 书名为空时，文件名去掉 `.mobi`、`.azw3`、`.azw`。
- 作者取 `metadata.author.first()`。
- 简介取 `metadata.description`，并用 `HtmlFormatter.format()` 清洗。
- 封面使用 `mobiBook.getCover()` 返回的字节，解码后保存 JPEG，质量 `90`。

### 目录解析

KF6 和 KF8 都优先使用 `mobiBook.toc`。

如果 section id map 中没有第 0 个 section，会补一个卷首章节：

- KF6 取第一个 section。
- KF8 取第一个 href 非空的 section。
- 标题从 section HTML 的 `<title>` 读取，失败则为 `卷首`。

递归 TOC：

- 每个 TOC 节点生成一个 `BookChapter`。
- `title = ref.label`。
- `url = "${chapterList.size}:${ref.href}"`。
- 有子节点则 `isVolume = true`。
- 上一章记录当前章为 `nextUrl`。
- 如果上一章是卷标题，且和当前章 href 相同，上一章 url 加 `skip:`，读取正文时返回空字符串，避免卷标题重复显示正文。

### KF6 正文

规则：

- 如果章节是 `isVolume` 且 url 以 `skip:` 开头，返回空字符串。
- 按章节 href 找到 section。
- 先追加当前 section 文本。
- 继续追加后续 section，直到遇到 `nextUrl`、遇到下一个 TOC section，或没有后续 section。
- 用 Jsoup 清洗：删除 `<title>`、`[style*=display:none]`。
- `img[recindex]` 改写成 `<img src="recindex:{recindex}">`。
- 调用 `HtmlFormatter.formatKeepImg()`，再删除 XML declaration 和 doctype。

### KF8 / AZW3 正文

规则：

- 如果章节是 `isVolume` 且 url 以 `skip:` 开头，返回空字符串。
- 按章节 href 找到 section。
- 解析下一章 href 的 position URI，用于判断 fragment 级截止点。
- 使用 `getTextByHref(chapter.url, nextSectionHref)` 截取当前 section 内文本。
- 继续追加后续 linear section，直到遇到下一章 fragment、下一章 section、下一个 TOC section，或没有后续 section。
- 删除 `<title>` 和隐藏元素。
- 调用 `HtmlFormatter.formatKeepImg()`，再删除 XML declaration 和 doctype。

图片读取：

- KF6/KF8 都调用 `getResourceByHref(href)?.inputStream()`。
- KF6 里 `recindex:` 图片由内部资源解析处理。

复用建议：AZW3 基本按 KF8 路径处理。MOBI 系列不能只按文件扩展名判断内部结构，需要读取 header，区分 KF6/KF8；AZW3 内容更接近 EPUB 的 HTML spine/fragment 模型。

## UMD 解析

核心代码：`model/localBook/UmdFile.kt`。

依赖 `me.ag2s.umdlib`。

- `UmdReader().read(input)` 读取整本书。
- 元数据来自 `umdBook.header`：`title`、`author`、`bookType`。
- 封面来自 `umdBook.cover.coverData`。
- 目录来自 `umdBook.chapters.titles`。
- 章节标题用 `chapters.getTitle(index)`。
- 章节 URL 是 index 字符串。
- 正文用 `chapters.getContentString(chapter.index)`。
- 不支持正文图片，`getImage()` 返回 null。

## HTML 正文归一化

核心代码：`utils/HtmlFormatter.kt`。

`HtmlFormatter.format()`：

- `&nbsp;`、`&ensp;`、`&emsp;` 转空格。
- 删除 `&thinsp;`、零宽字符等不可打印字符。
- `div`、`p`、`br`、`hr`、`h1-h6`、`article`、`dd`、`dl` 等块级标签转换行。
- 删除 HTML 注释。
- 删除普通 HTML 标签。
- 多个换行和空白折叠为换行加全角缩进 `　　`。
- 开头空白替换为 `　　`，末尾空白删除。

`HtmlFormatter.formatKeepImg()`：

- 先执行 `format()`，但保留 `<img>` 标签。
- 从 `src`、`data-src`、`data-original`、`srcset` 等属性提取图片地址。
- 统一输出 `<img src="...">`。
- 如果有基础 URL，会转成绝对 URL。

因此 EPUB、MOBI、AZW3 的最终正文形态基本是：普通文本 + 换行缩进 + `<img src="...">`。

## 自研阅读器建议抽象

可以把解析器抽象成同 Legado 类似的接口：

```kotlin
interface LocalBookParser {
    fun readMetadata(file: BookFile): BookMetadata
    fun readChapters(file: BookFile): List<Chapter>
    fun readContent(file: BookFile, chapter: Chapter): String?
    fun readResource(file: BookFile, href: String): InputStream?
}
```

建议统一中间模型：

- `BookMetadata(title, author, intro, coverHrefOrBytes, charset)`。
- `Chapter(index, title, locator, start, end, startFragmentId, endFragmentId, nextLocator, isVolume, wordCount)`。
- `Content` 使用 HTML 子集：纯文本、换行、`<img src="...">`。

实现优先级建议：

1. 先实现 TXT：编码识别、目录正则、字节偏移、章节缓存。
2. 再实现 EPUB：zip/opf/ncx/nav/spine、fragment 截取、图片资源读取。
3. PDF 可先按 Legado 图片式阅读实现，每 10 页一章。
4. MOBI/AZW3 建议直接复用成熟库或移植 `lib/mobi` 思路，重点区分 KF6/KF8。
5. UMD 需求低，可最后实现或依赖现成库。

## 关键源码位置

- `legado/app/src/main/java/io/legado/app/model/localBook/LocalBook.kt`：本地书籍导入、格式分派、统一章节后处理。
- `legado/app/src/main/java/io/legado/app/model/localBook/BaseLocalBookParse.kt`：本地解析器接口。
- `legado/app/src/main/java/io/legado/app/model/localBook/TextFile.kt`：TXT 编码、目录正则、字节偏移分章、正文读取。
- `legado/app/src/main/assets/defaultData/txtTocRule.json`：TXT 默认目录规则。
- `legado/app/src/main/java/io/legado/app/model/localBook/EpubFile.kt`：EPUB 元数据、目录、fragment 截取、图片资源。
- `legado/app/src/main/java/io/legado/app/model/localBook/PdfFile.kt`：PDF 每 10 页分段、页面渲染为图片。
- `legado/app/src/main/java/io/legado/app/model/localBook/MobiFile.kt`：MOBI/AZW/AZW3 目录、正文、图片入口。
- `legado/app/src/main/java/io/legado/app/lib/mobi/`：MOBI/KF6/KF8 底层解析。
- `legado/app/src/main/java/io/legado/app/model/localBook/UmdFile.kt`：UMD 解析。
- `legado/app/src/main/java/io/legado/app/utils/HtmlFormatter.kt`：HTML 到阅读器正文的归一化。
- `legado/app/src/main/java/io/legado/app/constant/AppPattern.kt`：支持格式和压缩包正则。
- `legado/app/src/main/java/io/legado/app/help/book/BookExtensions.kt`：格式识别扩展属性。
